package kit

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ColorFormat is how a ColorPicker shows and takes its value.
type ColorFormat uint8

const (
	ColorHex ColorFormat = iota // #2563EB
	ColorRGB                    // rgb(37, 99, 235)
	ColorHSL                    // hsl(221, 83%, 53%)
)

// FormatColor writes c in a format, with alpha when alpha is set:
// #2563EBCC, rgba(37, 99, 235, 0.8), hsla(221, 83%, 53%, 0.8).
func FormatColor(c color.NRGBA, f ColorFormat, alpha bool) string {
	a := strconv.FormatFloat(math.Round(float64(c.A)/255*100)/100, 'f', -1, 64)
	switch f {
	case ColorRGB:
		if alpha {
			return fmt.Sprintf("rgba(%d, %d, %d, %s)", c.R, c.G, c.B, a)
		}
		return fmt.Sprintf("rgb(%d, %d, %d)", c.R, c.G, c.B)
	case ColorHSL:
		h, s, l := rgbToHSL(c.R, c.G, c.B)
		hs := fmt.Sprintf("%d, %d%%, %d%%", int(math.Round(h))%360, int(math.Round(s*100)), int(math.Round(l*100)))
		if alpha {
			return "hsla(" + hs + ", " + a + ")"
		}
		return "hsl(" + hs + ")"
	}
	return hexOf(c, alpha)
}

// Format chooses the format the picker shows: its fields, and the text on a
// popup trigger. Users can switch it in the panel. The default is ColorHex.
func (p *ColorPickerView) Format(f ColorFormat) *ColorPickerView {
	if f <= ColorHSL {
		p.format = f
		p.syncDrafts()
	}
	return p
}

// CurrentFormat is the format now shown, after any switch by the user.
func (p *ColorPickerView) CurrentFormat() ColorFormat { return p.format }

// Text is the value in the current format, such as "rgb(37, 99, 235)".
func (p *ColorPickerView) Text() string { return FormatColor(p.Value(), p.format, p.alpha) }

// Placement sets where the popup panel opens beside its trigger (popup
// pickers only), as for a Popover: Placement(el.Top, el.End).
func (p *ColorPickerView) Placement(side el.Side, align el.Align) *ColorPickerView {
	p.side, p.align, p.placed = side, align, true
	if p.popover != nil {
		p.popover.Placement(side, align)
	}
	return p
}

// syncDrafts refills the text fields from the value.
func (p *ColorPickerView) syncDrafts() {
	c := p.Value()
	p.hex = hexOf(c, p.alpha)
	p.parts[0], p.parts[1], p.parts[2] = strconv.Itoa(int(c.R)), strconv.Itoa(int(c.G)), strconv.Itoa(int(c.B))
	if p.format == ColorHSL {
		h, s, l := rgbToHSL(c.R, c.G, c.B)
		if p.s > 0 && p.v > 0 || h != 0 {
			h = p.h // keep the picker's hue through grays, as the square does
		}
		p.parts[0], p.parts[1], p.parts[2] = strconv.Itoa(int(math.Round(h))%360), strconv.Itoa(int(math.Round(s*100))), strconv.Itoa(int(math.Round(l*100)))
	}
	p.parts[3] = strconv.Itoa(int(math.Round(p.a * 100)))
}

// commitDraft applies what was typed in the current format's fields, or
// restores them if it is not a color.
func (p *ColorPickerView) commitDraft() {
	if p.disabled {
		return
	}
	if p.format == ColorHex {
		p.commitHex()
		return
	}
	n := [4]int{}
	for i := range n {
		v, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(p.parts[i], "%")))
		if err != nil {
			p.syncDrafts()
			return
		}
		n[i] = v
	}
	limit := [3]int{255, 255, 255}
	if p.format == ColorHSL {
		limit = [3]int{360, 100, 100}
	}
	for i := range 3 {
		n[i] = min(max(n[i], 0), limit[i])
	}
	a := uint8(math.Round(p.a * 255))
	if p.alpha {
		a = uint8(math.Round(float64(min(max(n[3], 0), 100)) / 100 * 255))
	}
	c := color.NRGBA{R: uint8(n[0]), G: uint8(n[1]), B: uint8(n[2]), A: a}
	if p.format == ColorHSL {
		r, g, b := hslToRGB(float64(n[0]%360), float64(n[1])/100, float64(n[2])/100)
		c.R, c.G, c.B = r, g, b
	}
	if c != p.Value() {
		p.SetValue(c)
		if p.format == ColorHSL && n[1] > 0 {
			p.h = float64(n[0] % 360) // the typed hue, not the one rounding gives back
		}
		p.changed()
	}
	p.syncDrafts()
}

// fields is the value entry for the current format, under the bars.
func (p *ColorPickerView) fields(cx *el.Context, id string, focused bool) el.Element {
	m := p.metrics()
	text := locale.Current()
	box := el.Div().ID(id + "/fields").Gap(theme.SpaceSm).Items(el.Stretch)
	// The format switch: HEX RGB HSL.
	sw := el.Div().Row().Gap(theme.SpaceXxs).Role("radiogroup").Name(text.ColorFormat)
	for f, name := range []string{"HEX", "RGB", "HSL"} {
		f := ColorFormat(f)
		on := p.format == f
		b := el.Div().ID(id+"/format/"+name).Role("radio").Name(name+" "+text.ColorFormat).Selected(on).Px(theme.SpaceSm).Py(theme.SpaceXxs).
			Rounded(theme.RadiusSm).TextSize(theme.TextXs).Focusable(true).CursorPointer().
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).Border(1, color.NRGBA{}).
			OnClick(func() {
				if !p.disabled && p.format != f {
					p.commitDraft()
					p.format = f
					p.syncDrafts()
				}
			}).Child(el.Text(name))
		if on {
			b.Bg(theme.Highlight).TextColor(theme.PrimaryText)
		} else {
			b.TextColor(theme.Muted).Hover(func(s *el.Style) { s.Bg(theme.Subtle) })
		}
		sw.Child(b)
	}
	box.Child(sw)
	preview := el.Div().Size(el.Dp(m.control)).NoShrink().Rounded(theme.RadiusMd).Border(1, theme.Border).Bg(p.Value())
	row := el.Div().Row().Items(el.Center).Gap(theme.SpaceSm).Child(preview)
	field := func(slot, name string, bind *string, filter string, maxLen int) el.Element {
		in := fieldText(el.Input().ID(id + "/" + slot).Name(name).Bind(bind)).Filter(filter).MaxLen(maxLen).
			OnSubmit(func(string) { p.commitDraft() })
		frame := fieldFrame(id+"/"+slot+"box", focused && cx.FocusWithin(id+"/"+slot), false, p.disabled, false).
			FocusOnPress(id + "/" + slot).Grow().W(el.Dp(0))
		(&InputView{size: InputSize(p.size)}).applySize(in, frame)
		return frame.Child(in)
	}
	switch p.format {
	case ColorHex:
		maxLen := 7
		if p.alpha {
			maxLen = 9
		}
		row.Child(field("hex", "HEX", &p.hex, "#0123456789abcdefABCDEF", maxLen))
	default:
		names := [3]string{"R", "G", "B"}
		if p.format == ColorHSL {
			names = [3]string{"H", "S", "L"}
		}
		for i, n := range names {
			row.Child(field("part"+strconv.Itoa(i), n, &p.parts[i], "0123456789", 3))
		}
		if p.alpha {
			row.Child(field("part3", "A", &p.parts[3], "0123456789", 3))
		}
	}
	return box.Child(row)
}

// rgbToHSL converts to hue 0..360 and saturation, lightness 0..1.
func rgbToHSL(r8, g8, b8 uint8) (h, s, l float64) {
	r, g, b := float64(r8)/255, float64(g8)/255, float64(b8)/255
	mx, mn := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	l = (mx + mn) / 2
	d := mx - mn
	if d == 0 {
		return 0, 0, l
	}
	s = d / (1 - math.Abs(2*l-1))
	switch mx {
	case r:
		h = 60 * math.Mod((g-b)/d, 6)
	case g:
		h = 60 * ((b-r)/d + 2)
	default:
		h = 60 * ((r-g)/d + 4)
	}
	if h < 0 {
		h += 360
	}
	return h, s, l
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
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
	to8 := func(v float64) uint8 { return uint8(math.Round(math.Min(1, math.Max(0, v+m)) * 255)) }
	return to8(r), to8(g), to8(b)
}
