package kit

import (
	"image"
	"strings"
	"unicode"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// AvatarView displays decoded image pixels or initials in a fixed-size circle.
type AvatarView struct {
	name     string
	size     float32
	pixels   paint.ImageOp
	hasImage bool
}

func Avatar(name string) *AvatarView      { return &AvatarView{name: name, size: 40} }
func (v *AvatarView) SetName(name string) { v.name = name }

// Size accepts a diameter in dp, clamped to 16..256. Invalid values are ignored.
func (v *AvatarView) Size(dp float32) *AvatarView {
	if dp > 0 {
		v.size = max(16, min(256, dp))
	}
	return v
}

// Image sets already decoded pixels. Nil or empty images restore the initials.
// Load images outside the UI lock and assign them through core.Update.
func (v *AvatarView) Image(img image.Image) *AvatarView {
	v.hasImage = img != nil && !img.Bounds().Empty()
	if v.hasImage {
		v.pixels = paint.NewImageOp(img)
	} else {
		v.pixels = paint.ImageOp{}
	}
	return v
}
func avatarInitials(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return "?"
	}
	first := []rune(words[0])
	if len(words) == 1 {
		return string(unicode.ToUpper(first[0]))
	}
	last := []rune(words[len(words)-1])
	return string([]rune{unicode.ToUpper(first[0]), unicode.ToUpper(last[0])})
}
func (v *AvatarView) Render(*el.Context) el.Element {
	label := strings.TrimSpace(v.name)
	if label == "" {
		label = "Avatar"
	}
	box := el.Div().Size(el.Dp(v.size)).Rounded(v.size / 2).Bg(theme.Subtle).Role("image").Name(label).Center()
	if v.hasImage {
		pixels := v.pixels
		box.Value("loaded").Child(el.Widget(core.Func(func(gtx core.C) core.D {
			return (widget.Image{Src: pixels, Fit: widget.Cover, Position: layout.Center, Scale: 1}).Layout(gtx)
		})).Size(el.Dp(v.size)))
	} else {
		box.Value("initials").Child(el.Text(avatarInitials(v.name)).TextSize(v.size * .36).TextColor(theme.Text).MaxLines(1))
	}
	return box
}
