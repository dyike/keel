package kit

import (
	"hash/fnv"
	"image"
	"image/color"
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
	status   AvatarStatus
}

type AvatarStatus string

const (
	AvatarOnline  AvatarStatus = "online"
	AvatarBusy    AvatarStatus = "busy"
	AvatarOffline AvatarStatus = "offline"
)

func (v *AvatarView) Status(s AvatarStatus) *AvatarView { v.status = s; return v }
func Avatar(name string) *AvatarView                    { return &AvatarView{name: name, size: 40} }
func (v *AvatarView) SetName(name string)               { v.name = name }

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

// avatarInitials takes the first letters of the first two words. Han names
// and mixed-script names ("AI 助手") use only the first word, never pairing a
// Latin letter with a Han character; a short acronym stays whole ("AI").
func avatarInitials(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return "?"
	}
	first := []rune(words[0])
	if len(words) == 1 || unicode.Is(unicode.Han, first[0]) || unicode.Is(unicode.Han, []rune(words[1])[0]) {
		if len(first) == 2 && strings.ToUpper(words[0]) == words[0] && !unicode.Is(unicode.Han, first[0]) {
			return words[0]
		}
		return string(unicode.ToUpper(first[0]))
	}
	last := []rune(words[1])
	return string([]rune{unicode.ToUpper(first[0]), unicode.ToUpper(last[0])})
}
func (v *AvatarView) Render(*el.Context) el.Element {
	label := strings.TrimSpace(v.name)
	if label == "" {
		label = "Avatar"
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(v.name))
	colors := []color.NRGBA{theme.Subtle, theme.Highlight, theme.Surface}
	bg := colors[hash.Sum32()%uint32(len(colors))]
	box := el.Div().Size(el.Dp(v.size)).Rounded(v.size / 2).Bg(bg).Role("avatar").Name(label).Value(string(v.status)).Center()
	if v.hasImage {
		pixels := v.pixels
		box.Child(el.Widget(core.Func(func(gtx core.C) core.D {
			return (widget.Image{Src: pixels, Fit: widget.Cover, Position: layout.Center, Scale: 1}).Layout(gtx)
		})).Size(el.Dp(v.size)))
	} else {
		box.Child(el.Text(avatarInitials(v.name)).TextSize(v.size * .36).TextColor(theme.Text).MaxLines(1))
	}
	if v.status != "" {
		c := theme.Muted
		switch v.status {
		case AvatarOnline:
			c = theme.Success
		case AvatarBusy:
			c = theme.Warning
		}
		d := v.size * .22
		box.Child(el.Div().Absolute().Right(v.size*.14).Bottom(v.size*.14).Size(el.Dp(d)).Rounded(d/2).Border(1, theme.Surface).Bg(c))
	}
	return box
}
