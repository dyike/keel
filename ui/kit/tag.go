package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

// TagView displays a label. Removal and selection are not part of this version.
type TagView struct {
	text    string
	tone    Tone
	primary bool
}

type TagColor uint8

const (
	TagDefault TagColor = iota
	TagPrimary
	TagSuccess
	TagWarning
	TagDanger
)

func (v *TagView) Color(c TagColor) *TagView {
	v.primary = c == TagPrimary
	switch c {
	case TagSuccess:
		v.tone = Success
	case TagWarning:
		v.tone = Warning
	case TagDanger:
		v.tone = Danger
	default:
		v.tone = Neutral
	}
	return v
}
func tint(c color.NRGBA, alpha uint8) color.NRGBA {
	b := theme.Surface
	mix := func(x, y uint8) uint8 { return uint8((uint32(x)*uint32(alpha) + uint32(y)*uint32(255-alpha)) / 255) }
	return color.NRGBA{R: mix(c.R, b.R), G: mix(c.G, b.G), B: mix(c.B, b.B), A: 255}
}
func Tag(text string) *TagView          { return &TagView{text: text} }
func (v *TagView) SetText(s string)     { v.text = s }
func (v *TagView) Tone(t Tone) *TagView { v.tone = t; v.primary = false; return v }
func (v *TagView) Render(*el.Context) el.Element {
	c, name := v.tone.color(), v.tone.name()
	if v.tone == Neutral {
		name = "default"
	}
	if v.primary {
		c = theme.PrimaryText
		name = "primary"
	}
	return el.Div().Role("tag").Name(v.text).Value(name).Px(8).Py(4).Rounded(12).Bg(tint(c, 24)).Child(el.Text(v.text).TextSize(float32(theme.SmallSize)).TextColor(c))
}
