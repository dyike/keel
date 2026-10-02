package kit

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
	"strings"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

const pickerWidth = 240

// ColorPickerView picks a color: drag in the square for saturation and
// brightness, along the bars for hue and (with Alpha) opacity, type a hex
// code, or click a swatch. Focused, the square and bars move with the arrow
// keys. The color is kept as HSV, so hue survives passing through gray.
type ColorPickerView struct {
	h, s, v, a float64 // hue 0..360, the rest 0..1
	alpha      bool
	disabled   bool
	swatches   []color.NRGBA
	hex        string
	focused    bool
	onChange   func(color.NRGBA)
}

func ColorPicker() *ColorPickerView {
	p := &ColorPickerView{a: 1}
	p.SetValue(color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff})
	return p
}

// Alpha adds an opacity bar; the hex code then has eight digits.
func (p *ColorPickerView) Alpha() *ColorPickerView {
	p.alpha = true
	p.hex = hexOf(p.Value(), true)
	return p
}

// Swatches offers preset colors under the picker.
func (p *ColorPickerView) Swatches(colors ...color.NRGBA) *ColorPickerView {
	p.swatches = slices.Clone(colors)
	return p
}
func (p *ColorPickerView) OnChange(fn func(color.NRGBA)) *ColorPickerView { p.onChange = fn; return p }

// SetDisabled blocks user input. Disabling cancels an uncommitted HEX draft;
// SetValue remains available and never calls OnChange.
func (p *ColorPickerView) SetDisabled(on bool) {
	p.disabled = on
	if on {
		p.focused = false
		p.hex = hexOf(p.Value(), p.alpha)
	}
}

// Value is the current color.
func (p *ColorPickerView) Value() color.NRGBA {
	r, g, b := hsvToRGB(p.h, p.s, p.v)
	return color.NRGBA{R: r, G: g, B: b, A: uint8(math.Round(p.a * 255))}
}

// SetValue shows c without calling OnChange.
func (p *ColorPickerView) SetValue(c color.NRGBA) {
	h, s, v := rgbToHSV(c.R, c.G, c.B)
	if s > 0 && v > 0 {
		p.h = h // keep the hue of grays and black
	}
	p.s, p.v, p.a = s, v, float64(c.A)/255
	p.hex = hexOf(c, p.alpha)
}

func (p *ColorPickerView) changed() {
	c := p.Value()
	p.hex = hexOf(c, p.alpha)
	if p.onChange != nil {
		p.onChange(c)
	}
}

func (p *ColorPickerView) set(h, s, v, a float64) {
	if p.disabled || !finiteNumber(h) || !finiteNumber(s) || !finiteNumber(v) || !finiteNumber(a) {
		return
	}
	h, s, v, a = math.Mod(math.Mod(h, 360)+360, 360), clamp01(s), clamp01(v), clamp01(a)
	if h == p.h && s == p.s && v == p.v && a == p.a {
		return
	}
	p.h, p.s, p.v, p.a = h, s, v, a
	p.changed()
}

func clamp01(x float64) float64 { return math.Min(1, math.Max(0, x)) }

func hexOf(c color.NRGBA, alpha bool) string {
	if alpha {
		return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
	}
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// parseHex reads #RGB, #RRGGBB or #RRGGBBAA, with or without #.
func parseHex(s string) (color.NRGBA, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) == 6 {
		s += "FF"
	}
	if len(s) != 8 {
		return color.NRGBA{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.NRGBA{}, false
	}
	return color.NRGBA{R: uint8(n >> 24), G: uint8(n >> 16), B: uint8(n >> 8), A: uint8(n)}, true
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	u := func(f float64) uint8 { return uint8(math.Round((f + m) * 255)) }
	return u(r), u(g), u(b)
}

func rgbToHSV(r8, g8, b8 uint8) (h, s, v float64) {
	r, g, b := float64(r8)/255, float64(g8)/255, float64(b8)/255
	mx, mn := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	d := mx - mn
	switch {
	case d == 0:
		h = 0
	case mx == r:
		h = 60 * math.Mod((g-b)/d, 6)
	case mx == g:
		h = 60 * ((b-r)/d + 2)
	default:
		h = 60 * ((r-g)/d + 4)
	}
	if h < 0 {
		h += 360
	}
	if mx > 0 {
		s = d / mx
	}
	return h, s, mx
}

// gradient fills the box from c0 to c1, left to right or top to bottom.
func gradient(gtx core.C, box image.Rectangle, c0, c1 color.NRGBA, vertical bool) {
	end := f32.Pt(float32(box.Max.X), float32(box.Min.Y))
	if vertical {
		end = f32.Pt(float32(box.Min.X), float32(box.Max.Y))
	}
	paint.LinearGradientOp{Stop1: f32.Pt(float32(box.Min.X), float32(box.Min.Y)), Color1: c0, Stop2: end, Color2: c1}.Add(gtx.Ops)
	defer clip.Rect(box).Push(gtx.Ops).Pop()
	paint.PaintOp{}.Add(gtx.Ops)
}

func (p *ColorPickerView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	id := autoID("color", p)
	if p.focused && !cx.Enabled(id) {
		p.focused = false
		p.hex = hexOf(p.Value(), p.alpha)
	}
	focused := !p.disabled && cx.FocusWithin(id+"/hex")
	if p.focused && !focused {
		// Decide after this frame's disabled ancestry and modal state are known.
		cx.AfterEnabled(id, pickerBlurKey{id}, 0, func() { p.commitHex(); p.focused = false })
	} else {
		p.focused = focused
	}
	pure := func() color.NRGBA { r, g, b := hsvToRGB(p.h, 1, 1); return color.NRGBA{R: r, G: g, B: b, A: 255} }
	keys := func(dx, dy float64, apply func(dx, dy float64)) func(el.KeyEvent) bool {
		return func(e el.KeyEvent) bool {
			var x, y float64
			switch key.Name(e.Name) {
			case key.NameLeftArrow:
				x = -dx
			case key.NameRightArrow:
				x = dx
			case key.NameUpArrow:
				y = dy
			case key.NameDownArrow:
				y = -dy
			default:
				return false
			}
			if e.State == el.KeyPress {
				if e.Modifiers.Contain(key.ModShift) {
					x *= 10
					y *= 10
				}
				apply(x, y)
			}
			return true
		}
	}
	thumb := func(x, y float64) func(core.C, func()) {
		return func(gtx core.C, draw func()) {
			draw()
			r := float32(gtx.Dp(6))
			w, h := float32(gtx.Constraints.Max.X), float32(gtx.Constraints.Max.Y)
			pad := float32(gtx.Dp(8))
			xx, yy := min(w-pad, max(pad, float32(x)*w)), min(h-pad, max(pad, float32(y)*h))
			rect := image.Rect(int(xx-r), int(yy-r), int(xx+r), int(yy+r))
			path := clip.Ellipse(rect).Path(gtx.Ops)
			paint.FillShape(gtx.Ops, color.NRGBA{A: 255}, clip.Stroke{Path: path, Width: float32(gtx.Dp(3))}.Op())
			paint.FillShape(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, clip.Stroke{Path: path, Width: float32(gtx.Dp(1))}.Op())
		}
	}
	const svH = 150
	sv := el.Div().ID(id+"/shade").Border(1, theme.Border).Role("slider").Name(text.ColorShade).Value(strconv.Itoa(int(p.s*100)) + "," + strconv.Itoa(int(p.v*100))).
		WFull().H(el.Dp(svH)).Rounded(theme.RadiusMd).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnDrag(func(e el.DragEvent) {
			if e.Canceled || e.W <= 0 || e.H <= 0 {
				return
			}
			p.set(p.h, float64(e.X/e.W), 1-float64(e.Y/e.H), p.a)
		}).
		OnKey(keys(0.02, 0.02, func(dx, dy float64) { p.set(p.h, p.s+dx, p.v+dy, p.a) })).
		Child(el.Widget(core.Func(func(gtx core.C) core.D {
			box := image.Rectangle{Max: gtx.Constraints.Max}
			paint.FillShape(gtx.Ops, pure(), clip.Rect(box).Op())
			gradient(gtx, box, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 255, G: 255, B: 255}, false)
			gradient(gtx, box, color.NRGBA{}, color.NRGBA{A: 255}, true)
			return core.D{Size: box.Max}
		})).WFull().H(el.Dp(svH)))
	// Paint thumbs using the laid-out width, with a black/white ring that stays
	// inside the bounds even at an endpoint.
	svBox := el.Div().WFull().H(el.Dp(svH)).Child(sv).Decorate(thumb(p.s, 1-p.v))
	bar := func(slot, name, value string, x float64, draw func(gtx core.C, box image.Rectangle), drag func(f float64), step func(d float64)) el.Element {
		slider := el.Div().ID(id+"/"+slot).Border(1, theme.Border).Center().Role("slider").Name(name).Value(value).WFull().H(el.Dp(24)).Rounded(theme.RadiusFull).Focusable(true).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
			OnDrag(func(e el.DragEvent) {
				if e.Canceled || e.W <= 0 {
					return
				}
				drag(float64(e.X / e.W))
			}).
			OnKey(func(e el.KeyEvent) bool {
				switch key.Name(e.Name) {
				case key.NameHome, key.NameEnd:
					if e.State == el.KeyPress {
						if key.Name(e.Name) == key.NameHome {
							drag(0)
						} else {
							drag(1)
						}
					}
					return true
				case key.NamePageUp, key.NamePageDown:
					if e.State == el.KeyPress {
						if key.Name(e.Name) == key.NamePageUp {
							step(10)
						} else {
							step(-10)
						}
					}
					return true
				}
				return keys(1, 1, func(dx, dy float64) { step(dx + dy) })(e)
			}).
			Child(el.Widget(core.Func(func(gtx core.C) core.D {
				box := image.Rectangle{Max: gtx.Constraints.Max}
				defer clip.UniformRRect(box, box.Dy()/2).Push(gtx.Ops).Pop()
				draw(gtx, box)
				return core.D{Size: box.Max}
			})).WFull().H(el.Dp(14)))
		return el.Div().WFull().H(el.Dp(24)).Child(slider).Decorate(thumb(x, .5))
	}
	hue := bar("hue", text.Hue, strconv.Itoa(int(p.h)), p.h/360, func(gtx core.C, box image.Rectangle) {
		w := float32(box.Dx()) / 6
		for i := 0; i < 6; i++ {
			r0, g0, b0 := hsvToRGB(float64(i)*60, 1, 1)
			r1, g1, b1 := hsvToRGB(math.Mod(float64(i+1)*60, 360), 1, 1) // 360° is red again
			seg := image.Rect(int(float32(i)*w), 0, int(float32(i+1)*w+1), box.Dy())
			gradient(gtx, seg, color.NRGBA{R: r0, G: g0, B: b0, A: 255}, color.NRGBA{R: r1, G: g1, B: b1, A: 255}, false)
		}
	}, func(f float64) { p.set(min(359.999, max(0, f*360)), p.s, p.v, p.a) }, func(d float64) { p.set(p.h+d*2, p.s, p.v, p.a) })
	col := el.Div().ID(id).Disabled(p.disabled).Role("group").Name(hexOf(p.Value(), p.alpha)).Gap(10).W(el.Dp(pickerWidth)).MaxW(el.Full).Items(el.Stretch).Child(svBox, hue)
	if p.alpha {
		col.Child(bar("alpha", text.Opacity, strconv.Itoa(int(math.Round(p.a*100)))+"%", p.a, func(gtx core.C, box image.Rectangle) {
			paint.FillShape(gtx.Ops, theme.Subtle, clip.Rect(box).Op())
			c := p.Value()
			c0 := c
			c0.A, c.A = 0, 255
			gradient(gtx, box, c0, c, false)
		}, func(f float64) { p.set(p.h, p.s, p.v, f) }, func(d float64) { p.set(p.h, p.s, p.v, p.a+d*0.02) }))
	}
	maxLen := 7
	if p.alpha {
		maxLen = 9
	}
	preview := el.Div().Size(el.Dp(32)).NoShrink().Rounded(theme.RadiusMd).Border(1, theme.Border).Bg(p.Value())
	hex := fieldText(el.Input().ID(id + "/hex").Name("HEX").Bind(&p.hex)).Filter("#0123456789abcdefABCDEF").MaxLen(maxLen).
		OnSubmit(func(string) { p.commitHex() })
	col.Child(el.Div().Row().Items(el.Center).Gap(theme.SpaceMd).Child(preview, fieldFrame(id+"/hexbox", focused, false, p.disabled, false).FocusOnPress(id+"/hex").Grow().W(el.Dp(0)).Child(hex)))
	if len(p.swatches) > 0 {
		row := el.Div().Row().Wrap().Gap(theme.SpaceSm)
		for i, c := range p.swatches {
			swatch := el.Div().ID(id+"/swatch/"+strconv.Itoa(i)).Role("button").Name(hexOf(c, p.alpha)).Selected(c == p.Value()).
				Size(el.Dp(24)).NoShrink().Rounded(theme.RadiusSm).Bg(c).Border(1, theme.Border).Center().CursorPointer().Focusable(true).
				FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).OnClick(func() {
				if !p.disabled && c != p.Value() {
					p.SetValue(c)
					p.changed()
				}
			})
			if c == p.Value() {
				swatch.Child(Icon(IconCheck).Size(14).Color(pickerContrast(c)).Render(cx))
			}
			row.Child(swatch)
		}
		col.Child(row)
	}
	return col
}

type pickerBlurKey struct{ id string }

func (p *ColorPickerView) commitHex() {
	if p.disabled {
		return
	}
	c, ok := parseHex(p.hex)
	if !ok {
		p.hex = hexOf(p.Value(), p.alpha)
		return
	}
	if !p.alpha {
		c.A = uint8(math.Round(p.a * 255))
	}
	if c != p.Value() {
		p.SetValue(c)
		p.changed()
	}
	p.hex = hexOf(p.Value(), p.alpha)
}

// Choose black or white by contrast against the swatch composited on the surface.
func pickerContrast(c color.NRGBA) color.NRGBA {
	alpha := float64(c.A) / 255
	channel := func(x, b uint8) float64 {
		v := (float64(x)*alpha + float64(b)*(1-alpha)) / 255
		if v <= .04045 {
			return v / 12.92
		}
		return math.Pow((v+.055)/1.055, 2.4)
	}
	lum := .2126*channel(c.R, theme.Surface.R) + .7152*channel(c.G, theme.Surface.G) + .0722*channel(c.B, theme.Surface.B)
	if lum > .179 {
		return color.NRGBA{A: 255}
	}
	return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
}
