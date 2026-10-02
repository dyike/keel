package kit

import (
	"fmt"
	"image"
	"image/color"
	"math"
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
	p.swatches = colors
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
	if p.disabled {
		return
	}
	h, s, v, a = math.Mod(h+360, 360), clamp01(s), clamp01(v), clamp01(a)
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
				apply(x, y)
			}
			return true
		}
	}
	thumb := func(left, top float32) el.Element {
		return el.Div().Absolute().Left(left-7).Top(top-7).Size(el.Dp(14)).Rounded(7).Border(2, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	}
	const svH = 150
	sv := el.Div().Role("slider").Name(text.ColorShade).Value(strconv.Itoa(int(p.s*100)) + "," + strconv.Itoa(int(p.v*100))).
		W(el.Dp(pickerWidth)).H(el.Dp(svH)).Rounded(6).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnDrag(func(e el.DragEvent) { p.set(p.h, float64(e.X/e.W), 1-float64(e.Y/e.H), p.a) }).
		OnKey(keys(0.02, 0.02, func(dx, dy float64) { p.set(p.h, p.s+dx, p.v+dy, p.a) })).
		Child(el.Widget(core.Func(func(gtx core.C) core.D {
			box := image.Rectangle{Max: gtx.Constraints.Max}
			paint.FillShape(gtx.Ops, pure(), clip.Rect(box).Op())
			gradient(gtx, box, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 255, G: 255, B: 255}, false)
			gradient(gtx, box, color.NRGBA{}, color.NRGBA{A: 255}, true)
			return core.D{Size: box.Max}
		})).W(el.Dp(pickerWidth)).H(el.Dp(svH)))
	// Thumbs sit beside the sliders in a plain box: a slider clips to its own
	// bounds, which would cut a thumb at the edge in half.
	svBox := el.Div().W(el.Dp(pickerWidth)).H(el.Dp(svH)).Child(sv, thumb(float32(p.s)*pickerWidth, float32(1-p.v)*svH))
	bar := func(name, value string, x float64, draw func(gtx core.C, box image.Rectangle), drag func(f float64), step func(d float64)) el.Element {
		slider := el.Div().Role("slider").Name(name).Value(value).W(el.Dp(pickerWidth)).H(el.Dp(14)).Rounded(7).Focusable(true).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
			OnDrag(func(e el.DragEvent) { drag(float64(e.X / e.W)) }).
			OnKey(keys(1, 0, func(dx, _ float64) { step(dx) })).
			Child(el.Widget(core.Func(func(gtx core.C) core.D {
				box := image.Rectangle{Max: gtx.Constraints.Max}
				defer clip.UniformRRect(box, box.Dy()/2).Push(gtx.Ops).Pop()
				draw(gtx, box)
				return core.D{Size: box.Max}
			})).W(el.Dp(pickerWidth)).H(el.Dp(14)))
		return el.Div().W(el.Dp(pickerWidth)).H(el.Dp(14)).Child(slider, thumb(float32(x)*pickerWidth, 7))
	}
	hue := bar(text.Hue, strconv.Itoa(int(p.h)), p.h/360, func(gtx core.C, box image.Rectangle) {
		w := float32(box.Dx()) / 6
		for i := 0; i < 6; i++ {
			r0, g0, b0 := hsvToRGB(float64(i)*60, 1, 1)
			r1, g1, b1 := hsvToRGB(math.Mod(float64(i+1)*60, 360), 1, 1) // 360° is red again
			seg := image.Rect(int(float32(i)*w), 0, int(float32(i+1)*w+1), box.Dy())
			gradient(gtx, seg, color.NRGBA{R: r0, G: g0, B: b0, A: 255}, color.NRGBA{R: r1, G: g1, B: b1, A: 255}, false)
		}
	}, func(f float64) { p.set(f*360, p.s, p.v, p.a) }, func(d float64) { p.set(p.h+d*2, p.s, p.v, p.a) })
	col := el.Div().ID(id).Disabled(p.disabled).Role("group").Name(hexOf(p.Value(), p.alpha)).Gap(10).W(el.Dp(pickerWidth)).Items(el.Stretch).Child(svBox, hue)
	if p.alpha {
		col.Child(bar(text.Opacity, strconv.Itoa(int(math.Round(p.a*100)))+"%", p.a, func(gtx core.C, box image.Rectangle) {
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
	preview := el.Div().Size(el.Dp(32)).NoShrink().Rounded(6).Border(1, theme.Border).Bg(p.Value())
	hex := el.Input().ID(id + "/hex").Name("HEX").Bind(&p.hex).Filter("#0123456789abcdefABCDEF").MaxLen(maxLen).Grow().
		OnSubmit(func(string) { p.commitHex() })
	col.Child(el.Div().Row().Items(el.Center).Gap(8).Child(preview, hex))
	if len(p.swatches) > 0 {
		row := el.Div().Row().Gap(6)
		for i, c := range p.swatches {
			c := c
			if i > 0 && i%8 == 0 {
				col.Child(row)
				row = el.Div().Row().Gap(6)
			}
			row.Child(el.Div().Role("button").Name(hexOf(c, p.alpha)).Size(el.Dp(24)).Rounded(4).Bg(c).Border(1, theme.Border).
				CursorPointer().Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
				OnClick(func() {
					if !p.disabled {
						p.SetValue(c)
						p.changed()
					}
				}))
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
