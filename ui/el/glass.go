package el

import (
	"image"
	"image/color"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// GlassStyle selects the optical treatment of a Liquid Glass element surface.
type GlassStyle uint8

const (
	GlassNone GlassStyle = iota
	GlassRegular
	GlassClear
	GlassProminent
)

// GlassSurface configures a control-level Liquid Glass material drawn over
// the window backdrop or underlying UI content.
type GlassSurface struct {
	Style   GlassStyle
	Tint    color.NRGBA
	Boost   float32 // 0 normal, >0 hover/active specular lift
	Spacing float32 // merge proximity in dp for native NSGlassEffectContainerView (0 = default)
	HasSpec bool    // true when SpecX/SpecY carry a pointer/touch highlight focus
	SpecX   float32 // normalized 0..1 within the element
	SpecY   float32
}

func (e *engine) nativeGlassWindow() core.NativeGlassControls {
	if e.opaqueBackdrop {
		return nil
	}
	w, ok := core.CurrentWindow().(core.NativeGlassControls)
	if !ok || !w.NativeGlassEnabled() {
		return nil
	}
	return w
}

func (e *engine) paintGlassShadow(st Style, rect image.Rectangle, radius corners) {
	if rect.Empty() || st.glass.Style == GlassNone {
		return
	}
	shadowAlpha := uint8(24)
	if isDarkSurface() {
		shadowAlpha = 56
	}
	e.paintShadow(theme.Elevation{Blur: 10, Offset: 3}, color.NRGBA{R: 15, G: 23, B: 42, A: shadowAlpha}, 4, rect, radius)
}

func isDarkSurface() bool {
	bg := theme.Bg
	// Perceived luminance (Rec. 601).
	return int(bg.R)*299+int(bg.G)*587+int(bg.B)*114 < 128*1000
}

func clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func (e *engine) paintGlass(st Style, rect image.Rectangle, radius corners, interactive bool) {
	if rect.Empty() || st.glass.Style == GlassNone {
		return
	}
	ops := e.gtx.Ops
	g := st.glass
	dark := isDarkSurface()
	boost := g.Boost
	if boost < 0 {
		boost = 0
	} else if boost > 1 {
		boost = 1
	}

	// Native macOS 26+ NSGlassEffectView + NSGlassEffectContainerView path:
	// register the control's visible window-space rect so AppKit renders the
	// genuine Liquid Glass lens behind the transparent Gio Metal layer.
	if nw := e.nativeGlassWindow(); nw != nil {
		abs := image.Rectangle{Min: e.origin, Max: e.origin.Add(rect.Size())}.Intersect(e.visible)
		if !abs.Empty() {
			scale := e.gtx.Metric.PxPerDp
			if scale <= 0 {
				scale = 1
			}
			regStyle := core.GlassRegionRegular
			if g.Style == GlassClear {
				regStyle = core.GlassRegionClear
			}
			tint := g.Tint
			if g.Style == GlassProminent && tint.A == 0 {
				tint = theme.Primary
			}
			nw.AddGlassRegion(core.GlassRegion{
				X:            float32(abs.Min.X) / scale,
				Y:            float32(abs.Min.Y) / scale,
				Width:        float32(abs.Dx()) / scale,
				Height:       float32(abs.Dy()) / scale,
				CornerRadius: float32(radius.max()) / scale,
				Style:        regStyle,
				Tint:         tint,
				Interactive:  interactive,
				Spacing:      g.Spacing,
			})
		}
		// Subtle hover/active specular lift over the native glass so interactive
		// state remains crisp on both macOS 26 and macOS 27.
		if boost > 0 {
			bodyClip := radius.rrect(rect).Push(ops)
			liftA := clampByte(int(boost * 26))
			paint.ColorOp{Color: color.NRGBA{R: 255, G: 255, B: 255, A: liftA}}.Add(ops)
			paint.PaintOp{}.Add(ops)
			bodyClip.Pop()
		}
		return
	}

	w, h := float32(rect.Dx()), float32(rect.Dy())
	topPt := f32.Pt(float32(rect.Min.X), float32(rect.Min.Y))
	botPt := f32.Pt(float32(rect.Min.X), float32(rect.Max.Y))

	// 2. Translucent glass body gradient.
	bodyClip := radius.rrect(rect).Push(ops)
	var cTop, cBot color.NRGBA
	switch g.Style {
	case GlassProminent:
		tint := g.Tint
		if tint.A == 0 {
			tint = theme.Primary
		}
		lift := int(boost * 24)
		cTop = color.NRGBA{
			R: clampByte(int(tint.R) + 22 + lift),
			G: clampByte(int(tint.G) + 22 + lift),
			B: clampByte(int(tint.B) + 22 + lift),
			A: clampByte(218 + int(boost*22)),
		}
		cBot = color.NRGBA{
			R: clampByte(int(tint.R) - 10 + lift/2),
			G: clampByte(int(tint.G) - 10 + lift/2),
			B: clampByte(int(tint.B) - 10 + lift/2),
			A: clampByte(192 + int(boost*24)),
		}
	case GlassClear:
		if dark {
			cTop = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(30 + int(boost*28))}
			cBot = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(12 + int(boost*16))}
		} else {
			cTop = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(108 + int(boost*45))}
			cBot = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(48 + int(boost*32))}
		}
	default: // GlassRegular
		if dark {
			cTop = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(46 + int(boost*32))}
			cBot = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(22 + int(boost*18))}
		} else {
			cTop = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(176 + int(boost*40))}
			cBot = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(108 + int(boost*34))}
		}
	}
	paint.LinearGradientOp{Stop1: topPt, Stop2: botPt, Color1: cTop, Color2: cBot}.Add(ops)
	paint.PaintOp{}.Add(ops)

	// Optional translucent tint wash on Regular/Clear glass.
	if g.Style != GlassProminent && g.Tint.A > 0 {
		wash := g.Tint
		if wash.A > 72 {
			wash.A = 72
		}
		paint.ColorOp{Color: wash}.Add(ops)
		paint.PaintOp{}.Add(ops)
	}

	// 3. Upper-lens curved specular sheen.
	sheenAlpha := 82 + int(boost*44)
	if dark {
		sheenAlpha = 38 + int(boost*34)
	}
	if g.Style == GlassProminent {
		sheenAlpha = 76 + int(boost*40)
	}
	sheenEnd := f32.Pt(float32(rect.Min.X), float32(rect.Min.Y)+h*0.48)
	paint.LinearGradientOp{
		Stop1:  topPt,
		Stop2:  sheenEnd,
		Color1: color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(sheenAlpha)},
		Color2: color.NRGBA{R: 255, G: 255, B: 255, A: 0},
	}.Add(ops)
	paint.PaintOp{}.Add(ops)

	// Interactive pointer glow spot when cursor/touch moves over the glass.
	if g.HasSpec && boost > 0 {
		sx := float32(rect.Min.X) + g.SpecX*w
		sy := float32(rect.Min.Y) + g.SpecY*h
		ex := float32(rect.Min.X) + (1-g.SpecX)*w
		ey := float32(rect.Min.Y) + (1-g.SpecY)*h
		glowA := clampByte(int(boost * 48))
		paint.LinearGradientOp{
			Stop1:  f32.Pt(sx, sy),
			Stop2:  f32.Pt(ex, ey),
			Color1: color.NRGBA{R: 255, G: 255, B: 255, A: glowA},
			Color2: color.NRGBA{R: 255, G: 255, B: 255, A: 0},
		}.Add(ops)
		paint.PaintOp{}.Add(ops)
	}
	bodyClip.Pop()

	// 4. Directional optical specular rim stroke.
	rimW := max(e.dp(1), 1)
	half := rimW / 2
	insetRect := image.Rectangle{
		Min: rect.Min.Add(image.Pt(half, half)),
		Max: rect.Max.Sub(image.Pt(rimW-half, rimW-half)),
	}
	if insetRect.Empty() {
		return
	}
	rimPath := radius.inset(half).rrect(insetRect).Path(ops)
	rimClip := clip.Stroke{Path: rimPath, Width: float32(rimW)}.Op().Push(ops)

	var rimStart, rimEnd f32.Point
	if g.HasSpec {
		rimStart = f32.Pt(float32(rect.Min.X)+g.SpecX*w, float32(rect.Min.Y)+g.SpecY*h)
		rimEnd = f32.Pt(float32(rect.Min.X)+(1-g.SpecX)*w, float32(rect.Min.Y)+(1-g.SpecY)*h)
	} else {
		rimStart = f32.Pt(float32(rect.Min.X)+w*0.15, float32(rect.Min.Y))
		rimEnd = f32.Pt(float32(rect.Min.X)+w*0.85, float32(rect.Max.Y))
	}
	var rimC1, rimC2 color.NRGBA
	switch {
	case g.Style == GlassProminent:
		rimC1 = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(196 + int(boost*45))}
		rimC2 = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(44 + int(boost*24))}
	case dark:
		rimC1 = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(128 + int(boost*64))}
		rimC2 = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(28 + int(boost*20))}
	default:
		rimC1 = color.NRGBA{R: 255, G: 255, B: 255, A: clampByte(238 + int(boost*17))}
		rimC2 = color.NRGBA{R: 15, G: 23, B: 42, A: clampByte(34 + int(boost*14))}
	}
	paint.LinearGradientOp{Stop1: rimStart, Stop2: rimEnd, Color1: rimC1, Color2: rimC2}.Add(ops)
	paint.PaintOp{}.Add(ops)
	rimClip.Pop()
}
