package kit

import (
	"math"
	"slices"

	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// VariableListView virtualizes naturally sized rows. Keys identify data across
// insertion and reordering; the caller keeps row data and calls SetKeys when
// its order changes. Visible rows are measured after layout, with estimates
// used for rows that have not been visited. Heights and prefix sums are cached.
type VariableListView struct {
	keys             []string
	indices          map[string]int
	sizes            []float32
	sums             heightSums
	measured         map[string]float32
	estimate, height float32
	fill, disabled   bool
	horizontal       bool
	viewportWidth    float32
	row              func(*el.Context, int) el.Element
	width            int
	metric           unit.Metric
	anchor           string
	lastOffset       float32
	anchorDelta      float32
	restore          bool
	reveal           string
	revealAlign      ScrollAlign
	followEnd        bool
	end, endApplied  int
	role             string
}

// VariableList makes a variable-height list. Keys must be unique and nonempty;
// invalid keys panic before changing the list. estimate is a positive row
// height in dp, used only until a row is measured.
func VariableList(keys []string, estimate float32, row func(*el.Context, int) el.Element) *VariableListView {
	if estimate <= 0 || math.IsNaN(float64(estimate)) || math.IsInf(float64(estimate), 0) {
		estimate = 40
	}
	v := &VariableListView{estimate: estimate, height: 320, row: row, measured: make(map[string]float32)}
	v.SetKeys(keys)
	return v
}

func (v *VariableListView) ID() string { return autoID("variable-list", v) }
func (v *VariableListView) Count() int { return len(v.keys) }
func (v *VariableListView) Height(dp float32) *VariableListView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
		if !v.horizontal {
			v.fill = false
		}
	}
	return v
}
func (v *VariableListView) Fill() *VariableListView { v.fill = true; return v }
func (v *VariableListView) SetDisabled(on bool)     { v.disabled = on }

// SetKeys copies the new order, retains measurements for surviving keys, and
// keeps the current first visible row at the same screen position. If that
// row is removed, the row at its old index becomes the new anchor.
func (v *VariableListView) SetKeys(keys []string) {
	indices := make(map[string]int, len(keys))
	for i, k := range keys {
		if k == "" {
			panic("kit.VariableList: empty key")
		}
		if _, ok := indices[k]; ok {
			panic("kit.VariableList: duplicate key " + k)
		}
		indices[k] = i
	}
	oldIndex := v.indices[v.anchor]
	if _, ok := indices[v.anchor]; !ok {
		v.anchor, v.anchorDelta = "", 0
		if len(keys) > 0 {
			v.anchor = keys[min(oldIndex, len(keys)-1)]
		}
	}
	v.keys = slices.Clone(keys)
	v.indices = indices
	for k := range v.measured {
		if _, ok := indices[k]; !ok {
			delete(v.measured, k)
		}
	}
	v.rebuild()
	v.restore = true
	if _, ok := indices[v.reveal]; !ok {
		v.reveal = ""
	}
}

func (v *VariableListView) rebuild() {
	v.sizes = make([]float32, len(v.keys))
	for i, k := range v.keys {
		h := v.estimate
		if measured, ok := v.measured[k]; ok {
			h = measured
		}
		v.sizes[i] = h
	}
	v.sums.reset(v.sizes)
}

// Invalidate drops cached heights after offscreen data or font changes. With
// no keys it invalidates all rows. Visible size changes are detected each frame.
func (v *VariableListView) Invalidate(keys ...string) {
	if len(keys) == 0 {
		clear(v.measured)
		v.rebuild()
	} else {
		for _, k := range keys {
			if i, ok := v.indices[k]; ok {
				delete(v.measured, k)
				v.sums.add(i, v.estimate-v.sizes[i])
				v.sizes[i] = v.estimate
			}
		}
	}
	v.restore = true
}

// ScrollTo minimally reveals row i, retrying after measurement or first mount.
func (v *VariableListView) ScrollTo(cx *el.Context, i int) {
	if i < 0 || i >= len(v.keys) {
		return
	}
	v.ScrollToKey(cx, v.keys[i])
}

// ScrollToKey minimally reveals the row with key, independent of its current
// index. Missing keys are ignored, leaving any pending reveal unchanged.
func (v *VariableListView) ScrollToKey(cx *el.Context, key string) {
	v.ScrollToKeyAlign(cx, key, ScrollNearest)
}

// ScrollToAlign scrolls row i to the top, center or bottom of the viewport,
// as far as the content allows; rows not yet measured use their estimate
// and settle once measured.
func (v *VariableListView) ScrollToAlign(cx *el.Context, i int, align ScrollAlign) {
	if i < 0 || i >= len(v.keys) {
		return
	}
	v.ScrollToKeyAlign(cx, v.keys[i], align)
}

// ScrollToKeyAlign is ScrollToAlign by stable key.
func (v *VariableListView) ScrollToKeyAlign(cx *el.Context, key string, align ScrollAlign) {
	if _, ok := v.indices[key]; !ok {
		return
	}
	v.reveal, v.revealAlign = key, align
	cx.After(revealKey{v.ID()}, 0, func() {})
}

func (v *VariableListView) Render(cx *el.Context) el.Element {
	id := v.ID()
	off, view, content := virtualState(cx, id, v.horizontal)
	painted := view > 0
	if !painted {
		view = virtualExtent(v.horizontal, v.viewportWidth, v.height)
		if v.fill {
			w, h := cx.ViewportSize()
			viewport := h
			if v.horizontal {
				viewport = w
			}
			view = max(view, viewport)
		}
	}
	atEnd := v.reveal == "" && (v.followEnd && (!painted || off+view >= content-4) || v.end != v.endApplied)
	preserveAnchor := v.anchor != "" && v.reveal == "" && (v.restore || off == v.lastOffset)
	total := v.sums.prefix(len(v.keys))
	if v.restore {
		if i, ok := v.indices[v.anchor]; ok {
			off = v.sums.prefix(i) + min(v.anchorDelta, max(v.sizes[i]-1, 0))
		}
	}
	if i, ok := v.indices[v.reveal]; ok {
		off = alignedOffset(v.revealAlign, off, view, total, v.sums.prefix(i), v.sums.prefix(i+1))
	}
	if atEnd {
		off = max(total-view, 0)
	}
	off = min(max(off, 0), max(total-view, 0))
	if v.restore || v.reveal != "" || atEnd {
		virtualScroll(cx, id, v.horizontal, off)
	}
	if !painted {
		cx.After(revealKey{id}, 0, func() {})
	}
	first := v.sums.at(max(off-view, 0))
	last := min(len(v.keys), v.sums.at(off+2*view)+1)
	box := el.Div().ID(id).Focusable(true).Disabled(v.disabled).Items(el.Stretch)
	if v.horizontal {
		box.Row().ScrollX().H(el.Dp(v.height))
		if v.fill {
			box.Grow().MinW(el.Dp(1))
		} else {
			box.W(el.Dp(virtualExtent(true, v.viewportWidth, v.height)))
		}
	} else {
		box.ScrollY()
		if v.viewportWidth > 0 {
			box.W(el.Dp(v.viewportWidth))
		}
		if v.fill {
			box.Grow().MinH(el.Dp(1))
		} else {
			box.H(el.Dp(v.height))
		}
		if v.followEnd && v.reveal == "" {
			box.StickToBottom()
		}
	}
	if v.role != "" {
		box.Role(v.role)
	}
	leading := v.sums.prefix(first)
	box.Child(virtualSpacer(v.horizontal, leading))
	built := make([]el.Element, 0, last-first)
	for i := first; i < last; i++ {
		row := el.Div().ID(v.keys[i]).NoShrink().Items(el.Stretch)
		if v.horizontal {
			row.MinW(el.Dp(1)).H(el.Full)
		} else {
			row.MinH(el.Dp(1))
		}
		if v.row != nil {
			row.Child(v.row(cx, i))
		}
		built = append(built, row)
		box.Child(row)
	}
	box.Child(virtualSpacer(v.horizontal, total-v.sums.prefix(last)))
	box.Decorate(func(gtx core.C, draw func()) {
		draw()
		if !gtx.Enabled() {
			return
		}
		actual, _, _ := virtualState(cx, id, v.horizontal)
		heights := make([]float32, len(built))
		start := leading
		// Use actual laid-out starts to preserve the visible row even when a row
		// above it changed size in this frame, rather than trusting stale estimates.
		found := preserveAnchor && actual == off
		for j, row := range built {
			width, height := cx.LayoutSize(row)
			if v.horizontal {
				height = width
			}
			heights[j] = max(height, 1)
			if !found && actual >= start && actual < start+heights[j] {
				v.anchor, v.anchorDelta = v.keys[first+j], actual-start
				found = true
			}
			start += heights[j]
		}
		if !found && len(v.keys) > 0 {
			i := min(v.sums.at(actual), len(v.keys)-1)
			v.anchor, v.anchorDelta = v.keys[i], max(actual-v.sums.prefix(i), 0)
		}
		v.lastOffset = actual
		v.restore = false
		cross := gtx.Constraints.Max.X
		if v.horizontal {
			cross = gtx.Constraints.Max.Y
		}
		changed := v.width != cross || v.metric != gtx.Metric
		if changed {
			clear(v.measured)
			v.rebuild()
			v.width, v.metric = cross, gtx.Metric
		}
		for j, h := range heights {
			i := first + j
			if v.sizes[i] != h {
				v.sums.add(i, h-v.sizes[i])
				v.sizes[i] = h
				changed = true
			}
			v.measured[v.keys[i]] = h
		}
		if changed {
			v.restore = true
			gtx.Execute(op.InvalidateCmd{})
		}
		if atEnd && painted && !changed {
			v.endApplied = v.end
		}
		if i, ok := v.indices[v.reveal]; ok && painted && !changed && i >= first && i < last {
			v.reveal = ""
		}
	})
	return box
}

// Fenwick prefix sums keep changes and offset lookup logarithmic; rebuilding
// after SetKeys or a width change is linear and does not build row elements.
type heightSums struct{ tree []float64 }

func (s *heightSums) reset(sizes []float32) {
	s.tree = make([]float64, len(sizes)+1)
	for i, h := range sizes {
		j := i + 1
		s.tree[j] += float64(h)
		p := j + (j & -j)
		if p < len(s.tree) {
			s.tree[p] += s.tree[j]
		}
	}
}
func (s *heightSums) add(i int, delta float32) {
	for i++; i < len(s.tree); i += i & -i {
		s.tree[i] += float64(delta)
	}
}
func (s *heightSums) prefix(end int) float32 {
	var sum float64
	for ; end > 0; end -= end & -end {
		sum += s.tree[end]
	}
	return float32(sum)
}
func (s *heightSums) at(offset float32) int {
	index := 0
	var sum float64
	step := 1
	for step < len(s.tree) {
		step <<= 1
	}
	for ; step > 0; step >>= 1 {
		next := index + step
		if next < len(s.tree) && sum+s.tree[next] <= float64(offset) {
			index = next
			sum += s.tree[next]
		}
	}
	return index
}
