package motion

import (
	"image"
	"image/color"
	"time"

	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// PointerGlowStyle controls a cursor-origin spotlight and ripple. Zero values
// select defaults suitable for a card or button.
type PointerGlowStyle struct {
	Color          color.NRGBA
	Radius         float32
	FadeDuration   time.Duration
	RippleDuration time.Duration
	RippleWidth    float32
}

// PointerGlow paints a soft spotlight that follows the pointer and a ripple
// on hover-enter or press. Keep it in view state. It uses a pass-through
// pointer handler, so the decorated element's own input continues to work.
type PointerGlow struct {
	tag              int
	x, y             float32
	rippleX, rippleY float32
	rippleStart      time.Time
	ripple           bool
	hovered          bool
	initialized      bool
	opacity          Value
	style            PointerGlowStyle
	drawOpacity      float32
	drawRipple       float32
	enterPending     bool
	reduced          bool
	decorate         func(core.C, func())
}

// Origin returns the most recent pointer position in element-local dp.
func (p *PointerGlow) Origin() (x, y float32) { return p.x, p.y }

// Active reports whether a fade or ripple still needs animation frames.
func (p *PointerGlow) Active() bool { return p.opacity.Animating() || p.ripple }

// Apply gives e the stable id and decorates it with the pointer effect. Hover
// state comes from Context.Hovered; the pass-through handler only records local
// pointer coordinates. Apply owns e's Decorate callback.
func (p *PointerGlow) Apply(cx *el.Context, id string, e *el.DivEl, style PointerGlowStyle) *el.DivEl {
	if e == nil {
		return nil
	}
	e.ID(id)
	hovered := cx.Hovered(id)
	p.style = normalizePointerGlow(style)
	if !p.initialized {
		p.opacity.Set(0)
		p.initialized = true
	}
	if p.decorate == nil {
		p.decorate = p.paint
	}
	p.enterPending = hovered && !p.hovered
	p.hovered = hovered
	p.reduced = el.ReducedMotion()
	if p.reduced {
		p.opacity.Set(0)
		p.ripple = false
		p.drawOpacity, p.drawRipple = 0, 0
	} else {
		target := float32(0)
		if hovered {
			target = 1
		}
		var fadeActive bool
		p.drawOpacity, fadeActive = p.opacity.toAt(cx.Now(), target, Tween(p.style.FadeDuration, EaseOut), false)
		p.drawRipple = p.rippleProgress(cx.Now())
		if fadeActive || p.ripple {
			cx.Animating()
		}
	}
	e.Decorate(p.decorate)
	return e
}

func normalizePointerGlow(style PointerGlowStyle) PointerGlowStyle {
	if style.Color == (color.NRGBA{}) {
		style.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 120}
	}
	if style.Radius <= 0 {
		style.Radius = 84
	}
	if style.FadeDuration <= 0 {
		style.FadeDuration = 180 * time.Millisecond
	}
	if style.RippleDuration <= 0 {
		style.RippleDuration = 520 * time.Millisecond
	}
	if style.RippleWidth <= 0 {
		style.RippleWidth = 2
	}
	return style
}

func (p *PointerGlow) paint(gtx core.C, draw func()) {
	scale := gtx.Metric.PxPerDp
	if scale <= 0 {
		scale = 1
	}
	started := false
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &p.tag, Kinds: pointer.Enter | pointer.Leave | pointer.Move | pointer.Press})
		if !ok {
			break
		}
		pe, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		switch pe.Kind {
		case pointer.Enter:
			p.x, p.y = pe.Position.X/scale, pe.Position.Y/scale
			if p.enterPending && !p.reduced {
				p.startRipple(gtx.Now, p.x, p.y)
				started = true
			}
		case pointer.Move:
			p.x, p.y = pe.Position.X/scale, pe.Position.Y/scale
		case pointer.Press:
			p.x, p.y = pe.Position.X/scale, pe.Position.Y/scale
			if !p.reduced {
				p.startRipple(gtx.Now, p.x, p.y)
				started = true
			}
		}
	}
	if p.enterPending && !started && !p.reduced {
		p.startRipple(gtx.Now, p.x, p.y)
		started = true
	}
	p.enterPending = false

	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &p.tag)
	pass.Pop()
	area.Pop()

	draw()
	clipArea := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	if p.drawOpacity > 0 {
		p.drawSpotlight(gtx, p.style, p.drawOpacity, scale)
	}
	progress := p.drawRipple
	if started {
		progress = 0
	}
	p.paintRipple(gtx, p.style, scale, progress)
	clipArea.Pop()
	if started {
		// The press/enter is handled after Render. Request exactly the first
		// follow-up frame here; Apply drives all remaining live frames.
		gtx.Execute(op.InvalidateCmd{})
	}
}

func (p *PointerGlow) startRipple(now time.Time, x, y float32) {
	p.rippleX, p.rippleY = x, y
	p.rippleStart, p.ripple = now, true
}

func (p *PointerGlow) drawSpotlight(gtx core.C, style PointerGlowStyle, opacity, scale float32) {
	x, y, radius := p.x*scale, p.y*scale, style.Radius*scale
	// Fixed concentric discs approximate a radial falloff without textures or
	// per-frame allocation. Draw outside-in so the centre is brightest.
	for i := 5; i >= 1; i-- {
		fraction := float32(i) / 5
		r := radius * fraction
		alpha := opacity * float32(style.Color.A) / 255 * (1 - fraction*.72) / 4
		c := style.Color
		c.A = alphaByte(alpha)
		rect := image.Rect(roundInt(x-r), roundInt(y-r), roundInt(x+r), roundInt(y+r))
		paint.FillShape(gtx.Ops, c, clip.Ellipse(rect).Op(gtx.Ops))
	}
}

func (p *PointerGlow) rippleProgress(now time.Time) float32 {
	if !p.ripple {
		return 1
	}
	progress := float32(now.Sub(p.rippleStart)) / float32(p.style.RippleDuration)
	if progress >= 1 || progress < 0 {
		p.ripple = false
		return 1
	}
	return progress
}

func (p *PointerGlow) paintRipple(gtx core.C, style PointerGlowStyle, scale, progress float32) {
	if !p.ripple || progress < 0 || progress >= 1 {
		return
	}
	eased := EaseOut(progress)
	radius := style.Radius * scale * (0.16 + .9*eased)
	x, y := p.rippleX*scale, p.rippleY*scale
	rect := image.Rect(roundInt(x-radius), roundInt(y-radius), roundInt(x+radius), roundInt(y+radius))
	c := style.Color
	c.A = alphaByte((1 - progress) * float32(style.Color.A) / 255)
	path := clip.Ellipse(rect).Path(gtx.Ops)
	paint.FillShape(gtx.Ops, c, clip.Stroke{Path: path, Width: style.RippleWidth * scale}.Op())
}

func alphaByte(a float32) uint8 {
	return uint8(clamp01(a)*255 + .5)
}

func roundInt(x float32) int {
	if x < 0 {
		return int(x - .5)
	}
	return int(x + .5)
}
