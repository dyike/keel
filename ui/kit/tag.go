package kit

import (
	"image/color"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// TagView displays an optional selectable/removable label.
type TagView struct {
	text                           string
	tone                           Tone
	selectable, selected, disabled bool
	onRemove                       func()
	onChange                       func(bool)
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
func (v *TagView) Render(cx *el.Context) el.Element {
	c, name := v.tone.color(), v.tone.name()

	box := el.Div().Role("tag").Name(v.text).Value(name).Disabled(v.disabled).Row().Items(el.Center).Rounded(theme.RadiusFull).Bg(tint(c, 24))
	label := el.Div().Px(theme.SpaceMd).Py(theme.SpaceXs).Child(el.Text(v.text).TextSize(float32(theme.SmallSize)).TextColor(c))
	if v.selectable {
		box.Selected(v.selected)
		if v.selected {
			box.Bg(tint(c, 48))
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
		box.Child(el.Div().ID("remove").Name(locale.Current().Name(locale.Current().Remove, v.text)).P(theme.SpaceXs).OnClick(remove).OnKey(func(e el.KeyEvent) bool {
			if e.Name == string(key.NameDeleteBackward) || e.Name == string(key.NameDeleteForward) {
				if e.State == el.KeyRelease {
					remove()
				}
				return true
			}
			return false
		}).Child(Icon(IconClose).Size(14).Render(cx)))
	}
	return box
}
