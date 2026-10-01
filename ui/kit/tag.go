package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// TagView displays a label. Removal and selection are not part of this version.
type TagView struct {
	text string
	tone Tone
}

func Tag(text string) *TagView          { return &TagView{text: text} }
func (v *TagView) SetText(s string)     { v.text = s }
func (v *TagView) Tone(t Tone) *TagView { v.tone = t; return v }
func (v *TagView) Render(*el.Context) el.Element {
	return el.Div().Role("tag").Name(v.text).Value(v.tone.name()).Px(8).Py(4).Rounded(4).Border(1, v.tone.color()).Bg(theme.Surface).Child(el.Text(v.text).TextSize(float32(theme.SmallSize)).TextColor(v.tone.color()))
}
