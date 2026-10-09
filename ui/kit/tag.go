package kit

import (
	"image/color"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// TagAppearance defines normal, selected and border colors, resolved each frame.
type TagAppearance struct {
	Background, Foreground, Border, SelectedBackground color.NRGBA
}

// TagView displays an optional selectable/removable label.
type TagView struct {
	text                           string
	tone                           Tone
	selectable, selected, disabled bool
	onRemove                       func()
	onChange                       func(bool)
	outline                        bool
	height                         float32
	radius                         *float32
	content                        el.View
	appearance                     func(TagAppearance) TagAppearance
}

func tint(c color.NRGBA, alpha uint8) color.NRGBA {
	b := theme.Surface
	mix := func(x, y uint8) uint8 { return uint8((uint32(x)*uint32(alpha) + uint32(y)*uint32(255-alpha)) / 255) }
	return color.NRGBA{R: mix(c.R, b.R), G: mix(c.G, b.G), B: mix(c.B, b.B), A: 255}
}
func Tag(text string) *TagView                     { return &TagView{text: text} }
func (v *TagView) SetText(s string)                { v.text = s }
func (v *TagView) Tone(t Tone) *TagView            { v.tone = t; return v }
func (v *TagView) OnRemove(fn func()) *TagView     { v.onRemove = fn; return v }
func (v *TagView) Selectable() *TagView            { v.selectable = true; return v }
func (v *TagView) OnChange(fn func(bool)) *TagView { v.onChange = fn; return v }
func (v *TagView) Value() bool                     { return v.selected }
func (v *TagView) SetValue(b bool)                 { v.selected = b }
func (v *TagView) SetDisabled(b bool)              { v.disabled = b }

// Outline uses a transparent background until selected.
func (v *TagView) Outline(on bool) *TagView { v.outline = on; return v }

// Size sets minimum height in dp; long labels can wrap to additional lines.
func (v *TagView) Size(dp float32) *TagView {
	if dp >= 16 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// Rounded sets a corner radius in dp, including zero for square corners.
func (v *TagView) Rounded(dp float32) *TagView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.radius = &dp
	}
	return v
}

// Content replaces the visible text with display-only content; nil restores it.
// The constructor text remains the accessible name and action label.
func (v *TagView) Content(content el.View) *TagView { v.content = content; return v }

// Appearance transforms theme-derived colors each frame. Nil restores them.
func (v *TagView) Appearance(fn func(TagAppearance) TagAppearance) *TagView {
	v.appearance = fn
	return v
}
func (v *TagView) Render(cx *el.Context) el.Element {
	c, name := v.tone.color(), v.tone.name()

	a := TagAppearance{Background: tint(c, 24), Foreground: c, Border: c, SelectedBackground: tint(c, 48)}
	if v.outline {
		a.Background = color.NRGBA{}
	}
	if v.appearance != nil {
		a = v.appearance(a)
	}
	radius := float32(theme.RadiusFull)
	if v.radius != nil {
		radius = *v.radius
	}
	font, px, py := float32(theme.SmallSize), float32(theme.SpaceMd), float32(theme.SpaceXs)
	if v.height > 0 && v.height <= 22 {
		font, px, py = theme.TextXs, theme.SpaceSm, theme.SpaceXxs
	} else if v.height >= 32 {
		font = theme.TextBody
	}
	box := el.Div().Role("tag").Name(v.text).Value(name).Disabled(v.disabled).Row().Items(el.Center).Rounded(radius).Bg(a.Background).MaxW(el.Full)
	if v.outline || v.appearance != nil && a.Border.A > 0 {
		box.Border(1, a.Border)
	}
	label := el.Div().Px(px).Py(py).MinH(el.Dp(v.height)).MinW(el.Dp(0)).Justify(el.Center).TextSize(font).TextColor(a.Foreground).DisabledStyle(func(s *el.Style) { s.TextColor(theme.Muted) })
	if v.content != nil {
		label.Child(v.content.Render(cx))
	} else {
		label.Child(el.Text(v.text))
	}
	if v.selectable {
		box.Selected(v.selected)
		if v.selected {
			box.Bg(a.SelectedBackground)
		}
		label.Name(locale.Current().Name(locale.Current().Toggle, v.text)).OnClick(func() {
			v.selected = !v.selected
			if v.onChange != nil {
				v.onChange(v.selected)
			}
		})
	}
	box.Child(label)
	if v.onRemove != nil {
		remove := func() {
			if !v.disabled && v.onRemove != nil {
				v.onRemove()
			}
		}
		box.Child(el.Div().ID("remove").Name(locale.Current().Name(locale.Current().Remove, v.text)).P(theme.SpaceXs).NoShrink().OnClick(remove).OnKey(func(e el.KeyEvent) bool {
			if e.Name == string(key.NameDeleteBackward) || e.Name == string(key.NameDeleteForward) {
				if e.State == el.KeyRelease {
					remove()
				}
				return true
			}
			return false
		}).Child(Icon(IconClose).Size(14).Color(a.Foreground).Render(cx)))
	}
	return box
}
