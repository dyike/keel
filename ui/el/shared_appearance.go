package el

import (
	"image/color"
	"math"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/ui/theme"
)

// Only paint values are retained. Layout and interaction variants belong to
// the destination declaration.
type sharedGradient struct {
	from, to   color.NRGBA
	start, end f32.Point // unit-box coordinates, independent of aspect ratio
}
type sharedAppearance struct {
	gradient                 sharedGradient
	gradientSet              bool
	shadowOffset, shadowBlur float32 // fractions of the shorter side
	shadowColor              color.NRGBA
	shadowLayers             float32
	background, border, text color.NRGBA
	corners                  [4]float32 // fractions of the shorter side
	borderWidth              float32
	opacity                  float32
}

func (e *engine) sharedAppearance(n *Node) sharedAppearance {
	if n.palette != nil {
		defer theme.Scope(*n.palette)()
	}
	st := n.style
	a := sharedAppearance{
		background: st.bg.value, border: st.borderColor,
		text:    st.text.color.or(e.paintTextColor.or(n.textStyle.color.or(theme.Text))),
		opacity: 1,
	}
	if st.gradient != nil {
		a.gradientSet = true
		g := *st.gradient
		angle := float64(g.Angle) * math.Pi / 180
		dx, dy := math.Cos(angle), math.Sin(angle)
		w, h := float64(n.size.X), float64(n.size.Y)
		half := (math.Abs(w*dx) + math.Abs(h*dy)) / 2
		a.gradient = sharedGradient{g.From, g.To, f32.Pt(float32(.5-dx*half/w), float32(.5-dy*half/h)), f32.Pt(float32(.5+dx*half/w), float32(.5+dy*half/h))}
	}
	if st.opacitySet {
		a.opacity = st.opacity
	}
	side := float32(min(n.size.X, n.size.Y))
	if side > 0 {
		for i, r := range e.cornersOf(st, n.size) {
			a.corners[i] = float32(r) / side
		}
		a.borderWidth = float32(e.dp(st.borderWidth)) / side
		if st.shadow != nil {
			a.shadowOffset = float32(e.dp(st.shadow.Offset)) / side
			a.shadowBlur = float32(e.dp(st.shadow.Blur)) / side
			a.shadowColor = theme.Shadow
			a.shadowLayers = float32(min(max(e.dp(st.shadow.Blur)/2, 3), 16))
		}
	}
	return a
}

func mixSharedAppearance(a, b sharedAppearance, p float32) sharedAppearance {
	p = min(max(p, 0), 1)
	out := sharedAppearance{
		background:  mixContentColor(a.background, b.background, p),
		border:      mixContentColor(a.border, b.border, p),
		text:        mixContentColor(a.text, b.text, p),
		borderWidth: a.borderWidth + (b.borderWidth-a.borderWidth)*p,
		opacity:     a.opacity + (b.opacity-a.opacity)*p,
	}
	out.shadowOffset = a.shadowOffset + (b.shadowOffset-a.shadowOffset)*p
	out.shadowBlur = a.shadowBlur + (b.shadowBlur-a.shadowBlur)*p
	out.shadowColor = mixContentColor(a.shadowColor, b.shadowColor, p)
	out.shadowLayers = a.shadowLayers + (b.shadowLayers-a.shadowLayers)*p
	if a.gradientSet || b.gradientSet {
		ga, gb := a.gradient, b.gradient
		if !a.gradientSet {
			ga = gb
			ga.from, ga.to = a.background, a.background
		}
		if !b.gradientSet {
			gb = ga
			gb.from, gb.to = b.background, b.background
		}
		out.gradientSet = true
		out.gradient = sharedGradient{mixContentColor(ga.from, gb.from, p), mixContentColor(ga.to, gb.to, p), ga.start, ga.end}
		if ga.start != gb.start || ga.end != gb.end {
			if p == 1 {
				out.gradient.start, out.gradient.end = gb.start, gb.end
			} else if p > 0 {
				va, vb := ga.end.Sub(ga.start).Mul(.5), gb.end.Sub(gb.start).Mul(.5)
				angle := math.Atan2(float64(va.Y), float64(va.X))
				angle += math.Remainder(math.Atan2(float64(vb.Y), float64(vb.X))-angle, 2*math.Pi) * float64(p)
				length := math.Hypot(float64(va.X), float64(va.Y))*(1-float64(p)) + math.Hypot(float64(vb.X), float64(vb.Y))*float64(p)
				v := f32.Pt(float32(math.Cos(angle)*length), float32(math.Sin(angle)*length))
				center := f32.Pt(.5, .5)
				out.gradient.start, out.gradient.end = center.Sub(v), center.Add(v)
			}
		}
	}
	for i := range out.corners {
		out.corners[i] = a.corners[i] + (b.corners[i]-a.corners[i])*p
	}
	return out
}

// Apply before hover/focus/disabled styles: interactive feedback stays immediate.
func (e *engine) sharedPaintStyle(n *Node, st Style) Style {
	if n.sharedLayout == nil {
		return st
	}
	s := e.sharedLayouts[n.sharedLayout.identity]
	if s == nil || s.owner != n.key || !s.lookActive {
		return st
	}
	a := s.look
	st.bg, st.borderColor = some(a.background), a.border
	if a.gradientSet {
		st.sharedGradient = &s.look.gradient
	}
	// ContentTransition interpolates glyph colors with its own content tracks.
	if n.content == nil {
		st.text.color = some(a.text)
	}
	pxPerDp := e.m.PxPerDp
	if pxPerDp == 0 {
		pxPerDp = 1
	}
	side := float32(min(n.size.X, n.size.Y)) / pxPerDp
	var radii [4]float32
	for i, r := range a.corners {
		radii[i] = r * side
	}
	st.corners = some(radii)
	st.borderWidth = a.borderWidth * side
	return st
}

func (e *engine) sharedOpacity(n *Node) float32 {
	if n.sharedLayout != nil {
		if s := e.sharedLayouts[n.sharedLayout.identity]; s != nil && s.owner == n.key && s.lookActive {
			return s.look.opacity
		}
	}
	if n.style.opacitySet {
		return n.style.opacity
	}
	return 1
}

func (e *engine) sharedShadow(n *Node) (theme.Elevation, color.NRGBA, int, bool) {
	if n.sharedLayout != nil {
		if s := e.sharedLayouts[n.sharedLayout.identity]; s != nil && s.owner == n.key && s.lookActive {
			scale := e.m.PxPerDp
			if scale == 0 {
				scale = 1
			}
			side := float32(min(n.size.X, n.size.Y)) / scale
			a := s.look
			return theme.Elevation{Offset: a.shadowOffset * side, Blur: a.shadowBlur * side}, a.shadowColor, max(1, int(math.Round(float64(a.shadowLayers)))), a.shadowColor.A > 0
		}
	}
	if n.style.shadow != nil {
		return *n.style.shadow, theme.Shadow, 0, true
	}
	return theme.Elevation{}, color.NRGBA{}, 0, false
}
