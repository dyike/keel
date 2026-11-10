package el

import (
	"image"
	"image/color"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/semantic"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/internal/animation"
	"github.com/dyike/keel/ui/theme"
)

// ContentTransition controls changes to a Text's content. The default crossfades.
// Numeric rolls changed clusters on a single line; Down reverses the direction.
// Morph moves shared shaping clusters to their new positions, fading unmatched
// clusters. Numeric and Morph are mutually exclusive. Springs/tweens must finish.
type ContentTransition struct {
	Numeric, Down, Morph bool
	Animation            Animation
}

// ContentTransition animates text and range-color changes in paint-only frames.
// The first declaration is immediate. Layout and semantics use the new text;
// old glyphs are drawing only. Shimmer takes precedence. Stable IDs preserve state.
// Reduced motion, density changes and an intervening unpainted frame snap.
func (t *TextEl) ContentTransition(options ContentTransition) *TextEl {
	if options.Numeric && options.Morph {
		panic("el: content transition cannot be both numeric and morph")
	}
	if options.Animation.Kind != animation.KindSpring && options.Animation.Kind != animation.KindTween || options.Animation.Repeat != 0 {
		panic("el: content transition requires a finite spring or tween")
	}
	options.Animation = defaultAnimation(options.Animation)
	t.node().content = &options
	return t
}

// A resting label retains only its declaration, never a glyph/path/bitmap cache.
type textContentSnapshot struct {
	value        string
	params       text.Parameters
	shaper       *text.Shaper
	color        color.NRGBA
	ranges       []TextRange
	origin, size image.Point
	metric       unit.Metric
}

func (s textContentSnapshot) equal(t textContentSnapshot) bool {
	return s.value == t.value && s.params == t.params && s.shaper == t.shaper && s.color == t.color &&
		s.origin == t.origin && s.size == t.size && s.metric == t.metric && slices.Equal(s.ranges, t.ranges)
}
func (s textContentSnapshot) owned() textContentSnapshot {
	s.value = strings.Clone(s.value)
	s.ranges = slices.Clone(s.ranges)
	return s
}

type contentGlyph struct {
	key          string
	shaper       *text.Shaper
	glyphs       []text.Glyph
	position     f32.Point
	stretch      f32.Point // additional per-axis scale; zero means uniform
	raster       f32.Point // preserve native raster phase across a shared-box remap
	rasterSet    bool
	viewport     image.Rectangle
	scale, alpha float32
	em           float32
	color        color.NRGBA
	// Departures keep their original opacity clock when content changes again.
	exit *contentExit
}
type contentExit struct {
	value      animation.Value
	transition Animation
}
type contentTrack struct {
	from, to contentGlyph
}

type textContentState struct {
	last     textContentSnapshot
	frame    uint64
	progress animation.Value
	tracks   []contentTrack
	clip     image.Rectangle
}

func (e *engine) contentSnapshot(n *Node, inner image.Rectangle) textContentSnapshot {
	g := e.gtx
	g.Constraints = layout.Constraints{Max: inner.Size()}
	lb := e.label(n, n.text)
	return textContentSnapshot{value: n.text, params: labelParameters(g, lb), shaper: lb.Shaper, color: lb.Color, ranges: n.textRanges,
		origin: inner.Min.Add(image.Pt(0, e.textShift(n, scriptOf(n.text)))), size: inner.Size(), metric: e.m}
}

func (e *engine) paintTextContent(n *Node, inner image.Rectangle) bool {
	if !e.frameSource.Enabled() {
		return false
	}
	next := e.contentSnapshot(n, inner)
	st := e.store.get(n.key)
	s := st.content
	snap := ReducedMotion() || n.content.Animation.Duration < 0
	shared := e.sharedContent(n)
	if shared != nil {
		defer func() { shared.recordContent(st.content, n.size, e.gtx.Now) }()
	}
	handoff := shared != nil && shared.contentHandoff && !snap
	if handoff && s == nil {
		s = new(textContentState)
		st.content = s
		if e.contentStates == nil {
			e.contentStates = make(map[*elemState]struct{})
		}
		e.contentStates[st] = struct{}{}
	}
	if handoff {
		from, to := shared.contentFrom, next.shape()
		s.tracks = pairContent(from, to, *n.content, e.gtx.Now, next)
		s.clip = image.Rectangle{Min: next.origin, Max: next.origin.Add(next.size)}
		for _, g := range append(from, to...) {
			s.clip = s.clip.Union(g.viewport).Union(g.inkBounds())
		}
		s.last, s.frame = next.owned(), e.paintSerial
		s.progress.Set(0)
	}
	if !handoff && (s == nil || s.frame+1 < e.paintSerial || s.last.metric != next.metric || s.last.shaper != next.shaper || snap) {
		if s == nil {
			s = new(textContentState)
			st.content = s
			if e.contentStates == nil {
				e.contentStates = make(map[*elemState]struct{})
			}
			e.contentStates[st] = struct{}{}
		}
		if !s.last.equal(next) {
			s.last = next.owned()
		}
		s.frame, s.tracks = e.paintSerial, nil
		s.progress.Set(1)
		return false
	}
	s.frame = e.paintSerial
	now := e.gtx.Now
	if !s.last.equal(next) {
		// Width/typography changes without a content change follow layout
		// immediately. A later text edit still starts from this current layout.
		changed := s.last.value != next.value || !slices.Equal(s.last.ranges, next.ranges)
		if !changed && len(s.tracks) == 0 {
			s.last = next.owned()
			return false
		}
		var from []contentGlyph
		if len(s.tracks) > 0 {
			p, _ := s.progress.Sample(now, 1, n.content.Animation, false)
			for _, track := range s.tracks {
				g := track.at(min(max(p, 0), 1), now)
				if g.alpha > 1.0/255 {
					from = append(from, g)
				}
			}
		} else {
			from = s.last.shape()
		}
		to := next.shape()
		s.tracks = pairContent(from, to, *n.content, now, next)
		s.clip = image.Rectangle{Min: s.last.origin, Max: s.last.origin.Add(s.last.size)}.Union(image.Rectangle{Min: next.origin, Max: next.origin.Add(next.size)})
		// Include glyph overhang (italic, accents and emoji), as normal labels do.
		for _, g := range append(from, to...) {
			s.clip = s.clip.Union(g.inkBounds())
		}
		s.last = next.owned()
		s.progress.Set(0)
	}
	if len(s.tracks) == 0 {
		return false
	}
	p, active := s.progress.Sample(now, 1, n.content.Animation, false)
	if !active {
		s.tracks = nil
		return false
	}
	p = min(max(p, 0), 1)
	defer clip.Rect(s.clip).Push(e.gtx.Ops).Pop()
	semantic.LabelOp(n.text).Add(e.gtx.Ops)
	for _, track := range s.tracks {
		track.at(p, now).paint(e.gtx.Ops)
	}
	e.frameSource.Execute(op.InvalidateCmd{})
	if r := paintingRoot; r != nil {
		r.motionDue = true
	}
	return true
}

// shape preserves complete clusters and their native/Go shaping results. The
// shaper emits clusters in rune order, including right-to-left runs.
func (s textContentSnapshot) shape() []contentGlyph {
	s.shaper.LayoutString(s.params, s.value)
	var out []contentGlyph
	var cluster []text.Glyph
	var padding image.Rectangle
	byteAt, runeAt := 0, 0
	for g, ok := s.shaper.NextGlyph(); ok; g, ok = s.shaper.NextGlyph() {
		if int(g.Y)-g.Ascent.Ceil() > s.size.Y {
			break
		}
		padding.Min.X = min(padding.Min.X, g.Bounds.Min.X.Floor())
		padding.Max.X = max(padding.Max.X, (g.Bounds.Max.X - g.Advance).Ceil())
		padding.Min.Y = min(padding.Min.Y, (g.Bounds.Min.Y + g.Ascent).Floor())
		padding.Max.Y = max(padding.Max.Y, (g.Bounds.Max.Y - g.Descent).Ceil())
		cluster = append(cluster, g)
		if g.Flags&text.FlagClusterBreak == 0 {
			continue
		}
		start := byteAt
		for i := 0; i < int(g.Runes) && byteAt < len(s.value); i++ {
			_, n := utf8.DecodeRuneInString(s.value[byteAt:])
			byteAt += n
		}
		nextRune := runeAt + int(g.Runes)
		c := s.color
		if g.Flags&text.FlagTruncator == 0 {
			for _, r := range s.ranges {
				if r.Start < r.End && r.Start < nextRune && r.End > runeAt {
					c = r.Color
				}
			}
		}
		key := strings.Clone(s.value[start:byteAt])
		if g.Flags&(text.FlagTruncator|text.FlagParagraphBreak) != 0 {
			key = ""
		}
		if g.Flags&text.FlagParagraphBreak == 0 && len(cluster) > 0 && g.Runes > 0 {
			first := cluster[0]
			out = append(out, contentGlyph{key: key, shaper: s.shaper, glyphs: cluster,
				position: f32.Pt(float32(s.origin.X)+float32(first.X)/64, float32(s.origin.Y)+float32(first.Y)), scale: 1, alpha: 1, em: float32(s.params.PxPerEm) / 64, color: c})
			cluster = nil
		} else {
			cluster = cluster[:0]
		}
		runeAt = nextRune
	}
	viewport := image.Rectangle{Min: s.origin.Add(padding.Min), Max: s.origin.Add(s.size).Add(padding.Max)}
	for i := range out {
		out[i].viewport = viewport
	}
	return out
}

func (g contentGlyph) inkBounds() image.Rectangle {
	var out image.Rectangle
	if len(g.glyphs) == 0 {
		return out
	}
	first := g.glyphs[0]
	sx, sy := g.scale*(1+g.stretch.X), g.scale*(1+g.stretch.Y)
	for _, glyph := range g.glyphs {
		x := g.position.X + float32(glyph.X-first.X)/64*sx
		y := g.position.Y + float32(glyph.Y-first.Y)*sy
		r := image.Rect(int(math.Floor(float64(x+float32(glyph.Bounds.Min.X)/64*sx))), int(math.Floor(float64(y+float32(glyph.Bounds.Min.Y)/64*sy))),
			int(math.Ceil(float64(x+float32(glyph.Bounds.Max.X)/64*sx))), int(math.Ceil(float64(y+float32(glyph.Bounds.Max.Y)/64*sy))))
		out = out.Union(r)
	}
	return out
}

func (g contentGlyph) paint(ops *op.Ops) {
	if g.alpha <= 0 || len(g.glyphs) == 0 {
		return
	}
	defer clip.Rect(g.viewport).Push(ops).Pop()
	raster := g.position
	if g.rasterSet {
		raster = g.raster
	}
	sx, sy := g.scale*(1+g.stretch.X), g.scale*(1+g.stretch.Y)
	if sx != 1 || sy != 1 || raster != g.position {
		transform := f32.NewAffine2D(sx, 0, g.position.X-raster.X*sx, 0, sy, g.position.Y-raster.Y*sy)
		defer op.Affine(transform).Push(ops).Pop()
	}
	if g.alpha < 1 {
		defer paint.PushOpacity(ops, g.alpha).Pop()
	}
	theme.PaintGlyphs(ops, g.shaper, g.glyphs, g.color, raster)
}

func (t contentTrack) at(p float32, now time.Time) contentGlyph {
	g := t.to
	lerp := func(a, b int) int { return int(math.Round(float64(float32(a) + (float32(b)-float32(a))*p))) }
	g.viewport = image.Rect(lerp(t.from.viewport.Min.X, t.to.viewport.Min.X), lerp(t.from.viewport.Min.Y, t.to.viewport.Min.Y), lerp(t.from.viewport.Max.X, t.to.viewport.Max.X), lerp(t.from.viewport.Max.Y, t.to.viewport.Max.Y))
	g.position = t.from.position.Add(t.to.position.Sub(t.from.position).Mul(p))
	g.scale = t.from.scale + (t.to.scale-t.from.scale)*p
	g.stretch = t.from.stretch.Add(t.to.stretch.Sub(t.from.stretch).Mul(p))
	if t.from.rasterSet || t.to.rasterSet {
		a, b := t.from.position, t.to.position
		if t.from.rasterSet {
			a = t.from.raster
		}
		if t.to.rasterSet {
			b = t.to.raster
		}
		g.raster, g.rasterSet = a.Add(b.Sub(a).Mul(p)), true
	}
	g.alpha = t.from.alpha + (t.to.alpha-t.from.alpha)*p
	g.color = mixContentColor(t.from.color, t.to.color, p)
	if g.exit != nil {
		alpha, _ := g.exit.value.Sample(now, 0, g.exit.transition, false)
		g.alpha = min(g.alpha, max(alpha, 0))
	}

	return g
}

func sameContentShape(a, b contentGlyph) bool {
	if a.shaper != b.shaper || len(a.glyphs) != len(b.glyphs) {
		return false
	}
	for i, g := range a.glyphs {
		if g.ID != b.glyphs[i].ID || g.Offset != b.glyphs[i].Offset || g.X-a.glyphs[0].X != b.glyphs[i].X-b.glyphs[0].X {
			return false
		}
	}
	return true
}

// Premultiplied mixing avoids introducing black fringes from transparent colors.
func mixContentColor(a, b color.NRGBA, p float32) color.NRGBA {
	aa, ba := float32(a.A), float32(b.A)
	alpha := aa + (ba-aa)*p
	if alpha <= 0 {
		return color.NRGBA{}
	}
	ch := func(x, y uint8) uint8 {
		return uint8(min(max((float32(x)*aa*(1-p)+float32(y)*ba*p)/alpha, 0), 255) + .5)
	}
	return color.NRGBA{R: ch(a.R, b.R), G: ch(a.G, b.G), B: ch(a.B, b.B), A: uint8(alpha + .5)}
}

func pairContent(from, to []contentGlyph, options ContentTransition, now time.Time, next textContentSnapshot) []contentTrack {
	if options.Numeric && (!singleLineContent(from) || !singleLineContent(to)) {
		options.Numeric = false
	}
	pairs := contentPairs(from, to, options)
	exitClocks := make(map[float32]*contentExit)
	taken := make([]bool, len(from))
	tracks := make([]contentTrack, 0, len(from)+len(to))
	orders := make([]int, 0, len(from)+len(to))
	roll := float32(next.params.PxPerEm) / 64
	if next.params.LineHeightScale > 0 {
		roll *= next.params.LineHeightScale
	}
	if options.Down {
		roll = -roll
	}
	for j, dest := range to {
		start := dest
		start.alpha = 0
		if i := pairs[j]; i >= 0 {
			source := from[i]
			taken[i] = true
			start.position, start.color, start.alpha = source.position, source.color, source.alpha
			start.viewport = source.viewport
			start.stretch, start.raster, start.rasterSet = source.stretch, source.raster, source.rasterSet
			if dest.em > 0 && source.em > 0 {
				start.scale = source.scale * source.em / dest.em
			}
			if !sameContentShape(source, dest) {
				// A changed font/ligature outline crossfades while moving along
				// the same route instead of substituting the new shape at t=0.
				oldEnd := source
				oldEnd.position, oldEnd.alpha, oldEnd.viewport = dest.position, 0, dest.viewport
				oldEnd.stretch = dest.stretch
				if source.em > 0 && dest.em > 0 {
					oldEnd.scale = dest.em / source.em
				}
				tracks = append(tracks, departingContent(source, oldEnd, now, options.Animation, exitClocks))
				orders = append(orders, i*2)
				start.alpha = 0
			}
		} else if options.Numeric {
			start.position.Y += roll
		}
		tracks = append(tracks, contentTrack{from: start, to: dest})
		order := len(from)*2 + j
		if pairs[j] >= 0 {
			order = pairs[j]*2 + 1
		}
		orders = append(orders, order)
	}
	for i, source := range from {
		if taken[i] {
			continue
		}
		dest := source
		dest.alpha = 0
		if options.Numeric {
			dest.position.Y -= roll
		}
		tracks = append(tracks, departingContent(source, dest, now, options.Animation, exitClocks))
		orders = append(orders, i*2)
	}
	// Keep the previously painted order at interruption; new zero-alpha
	// particles can follow it without changing that first frame.
	indices := make([]int, len(tracks))
	for i := range indices {
		indices[i] = i
	}
	slices.SortFunc(indices, func(a, b int) int { return orders[a] - orders[b] })
	ordered := make([]contentTrack, len(tracks))
	for i, j := range indices {
		ordered[i] = tracks[j]
	}
	return ordered
}

// Match whole shaping clusters. Numeric labels preserve the common prefix and
// suffix. Morph first preserves order (bounded LCS), then pairs moved clusters.
func contentPairs(from, to []contentGlyph, options ContentTransition) []int {
	pairs := make([]int, len(to))
	for i := range pairs {
		pairs[i] = -1
	}
	if !options.Numeric && !options.Morph {
		return pairs
	}
	equal := func(i, j int) bool { return from[i].key != "" && from[i].key == to[j].key }
	used := make([]bool, len(from))
	if options.Numeric {
		// Multiple lines and bidirectional runs use a crossfade.
		for _, gs := range [][]contentGlyph{from, to} {
			for _, g := range gs {
				if len(gs) > 0 && g.position.Y != gs[0].position.Y || len(g.glyphs) > 0 && g.glyphs[0].Flags&text.FlagTowardOrigin != 0 {
					return pairs
				}
			}
		}
		i := 0
		for i < len(from) && i < len(to) && equal(i, i) {
			pairs[i] = i
			i++
		}
		for a, b := len(from)-1, len(to)-1; a >= i && b >= i && equal(a, b); a, b = a-1, b-1 {
			pairs[b] = a
		}
		return pairs
	}
	if len(from) > 0 && len(to) > 0 && len(from) <= 65536/len(to) {
		w := len(to) + 1
		table := make([]uint16, (len(from)+1)*w)
		for i := len(from) - 1; i >= 0; i-- {
			for j := len(to) - 1; j >= 0; j-- {
				if equal(i, j) {
					table[i*w+j] = table[(i+1)*w+j+1] + 1
				} else {
					table[i*w+j] = max(table[(i+1)*w+j], table[i*w+j+1])
				}
			}
		}
		for i, j := 0, 0; i < len(from) && j < len(to); {
			if equal(i, j) {
				pairs[j], used[i] = i, true
				i++
				j++
			} else if table[(i+1)*w+j] >= table[i*w+j+1] {
				i++
			} else {
				j++
			}
		}
	}
	// Linear-space queues bound matching work for longer labels.
	remaining := make(map[string][]int)
	for i, g := range from {
		if !used[i] && g.key != "" {
			remaining[g.key] = append(remaining[g.key], i)
		}
	}
	for j, g := range to {
		if pairs[j] < 0 {
			if list := remaining[g.key]; len(list) > 0 {
				pairs[j] = list[0]
				remaining[g.key] = list[1:]
			}
		}
	}
	return pairs
}

func singleLineContent(gs []contentGlyph) bool {
	for _, g := range gs {
		if g.position.Y != gs[0].position.Y || len(g.glyphs) > 0 && g.glyphs[0].Flags&text.FlagTowardOrigin != 0 {
			return false
		}
	}
	return true
}

func (e *engine) sweepContentStates() {
	for st := range e.contentStates {
		if st.content == nil || st.content.frame != e.paintSerial {
			st.content = nil
			delete(e.contentStates, st)
		}
	}
}

// Past ink may outlive a label's intrinsic width (including a change to "").
// Propagate just that overflow for culling; it never affects layout or input.
func (e *engine) placeContentOverflow(n *Node) {
	var bounds image.Rectangle
	if n.content != nil {
		if st := e.store.of(n); st != nil && st.content != nil {
			s := st.content
			bounds = image.Rectangle{Min: s.last.origin, Max: s.last.origin.Add(s.last.size)}
			if len(s.tracks) > 0 {
				bounds = bounds.Union(s.clip)
			}
		}
	}
	for _, element := range n.children {
		c := element.node()
		if !c.style.hidden && c.contentOverflow != nil {
			at := c.pos.Add(image.Pt(e.dp(c.style.translate[0]), e.dp(c.style.translate[1])))
			bounds = bounds.Union(c.contentOverflow.Add(at))
		}
	}
	box := image.Rectangle{Max: n.size}
	if n.style.scrollX || n.style.scrollY || n.style.revealSet {
		bounds = bounds.Intersect(box)
	}
	if bounds.Empty() || bounds.In(box) {
		n.contentOverflow = nil
	} else {
		n.contentOverflow = &bounds
	}
}

func departingContent(source, dest contentGlyph, now time.Time, tr Animation, cache map[float32]*contentExit) contentTrack {
	dest.exit = source.exit
	if dest.exit == nil {
		dest.exit = cache[source.alpha]
		if dest.exit == nil {
			dest.exit = &contentExit{transition: tr}
			dest.exit.value.Set(source.alpha)
			dest.exit.value.Sample(now, 0, tr, false)
			cache[source.alpha] = dest.exit
		}
	}
	return contentTrack{from: source, to: dest}
}
