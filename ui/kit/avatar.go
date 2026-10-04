package kit

import (
	"context"
	"hash/fnv"
	"image"
	"image/color"
	"strings"
	"time"
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
	source   string
	cancel   context.CancelFunc
	revision uint64
	loading  bool
	imageErr error

	// Appearance; zero values keep the defaults.
	radius      *float32
	bg, fg      color.NRGBA
	border      float32
	borderColor color.NRGBA
	placeholder IconName
	style       func(*el.DivEl)
}

type AvatarStatus string

const (
	AvatarOnline  AvatarStatus = "online"
	AvatarBusy    AvatarStatus = "busy"
	AvatarOffline AvatarStatus = "offline"
)

func (v *AvatarView) Status(s AvatarStatus) *AvatarView { v.status = s; return v }
func Avatar(name string) *AvatarView {
	return &AvatarView{name: name, size: 40, placeholder: IconUser}
}

// Rounded sets the corner radius in dp; avatars are circles by default. An
// app or team avatar is often a rounded square: Rounded(theme.RadiusLg).
func (v *AvatarView) Rounded(dp float32) *AvatarView {
	if dp >= 0 && finiteNumber(float64(dp)) {
		v.radius = &dp
	}
	return v
}

// Colors sets the background behind initials or the placeholder, and their
// color. Zero colors keep the defaults: a color picked from the name, and
// the theme's text color.
func (v *AvatarView) Colors(bg, fg color.NRGBA) *AvatarView { v.bg, v.fg = bg, fg; return v }

// Border draws a ring of width dp, such as a Surface-colored ring that
// separates overlapping avatars. Zero removes it.
func (v *AvatarView) Border(width float32, c color.NRGBA) *AvatarView {
	if width >= 0 && finiteNumber(float64(width)) {
		v.border, v.borderColor = width, c
	}
	return v
}

// Placeholder is the icon shown with neither an image nor a name, IconUser by
// default; IconNone shows nothing.
func (v *AvatarView) Placeholder(name IconName) *AvatarView { v.placeholder = name; return v }

// Style adjusts the avatar's box after its default styling on every Render,
// for anything the other options do not cover. Nil removes it.
func (v *AvatarView) Style(fn func(*el.DivEl)) *AvatarView { v.style = fn; return v }
func (v *AvatarView) SetName(name string)                  { v.name = name }

// Size accepts a diameter in dp, clamped to 16..256. Invalid values are ignored.
func (v *AvatarView) Size(dp float32) *AvatarView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.size = max(16, min(256, dp))
	}
	return v
}

// Image sets already decoded pixels. Nil or empty images restore the initials.
// Load images outside the UI lock and assign them through core.Update.
func (v *AvatarView) Image(img image.Image) *AvatarView {
	v.stopLoad()
	v.source, v.imageErr = "", nil
	v.setPixels(img)
	return v
}

func (v *AvatarView) setPixels(img image.Image) {
	v.hasImage = img != nil && !img.Bounds().Empty()
	if v.hasImage {
		v.pixels = paint.NewImageOp(img)
	} else {
		v.pixels = paint.ImageOp{}
	}
}

// Source asynchronously loads an image URL or local path. Empty restores initials.
// Repeating the same source does not reload; Retry restarts a failed request.
func (v *AvatarView) Source(source string) *AvatarView {
	if source != "" && source == v.source {
		return v
	}
	v.stopLoad()
	v.source, v.imageErr = source, nil
	v.setPixels(nil)
	if source != "" {
		v.startLoad()
	}
	return v
}
func (v *AvatarView) Loading() bool     { return v.loading }
func (v *AvatarView) ImageError() error { return v.imageErr }

// Retry reloads the current source, including cancelling any in-flight request.
func (v *AvatarView) Retry() {
	if v.source == "" {
		return
	}
	v.stopLoad()
	v.imageErr = nil
	v.setPixels(nil)
	v.startLoad()
}
func (v *AvatarView) stopLoad() {
	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}
	v.revision++
	v.loading = false
}
func (v *AvatarView) startLoad() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	v.cancel, v.loading = cancel, true
	revision, source := v.revision, v.source
	go func() {
		defer cancel()
		media, err := loadImageMedia(ctx, source)
		core.Update(func() {
			if v.revision != revision {
				return
			}
			v.loading, v.cancel, v.imageErr = false, nil, err
			if err == nil {
				v.setPixels(media.still)
			}
		})
	}()
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
func (v *AvatarView) Render(cx *el.Context) el.Element {
	label := strings.TrimSpace(v.name)
	if label == "" {
		label = "Avatar"
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(v.name))
	colors := []color.NRGBA{theme.Subtle, theme.Highlight, theme.Surface}
	bg := colors[hash.Sum32()%uint32(len(colors))]
	if v.bg.A > 0 {
		bg = v.bg
	}
	fg := theme.Text
	if v.fg.A > 0 {
		fg = v.fg
	}
	radius := v.size / 2
	if v.radius != nil {
		radius = *v.radius
	}
	box := el.Div().Size(el.Dp(v.size)).Rounded(radius).Bg(bg).Role("avatar").Name(label).Value(string(v.status)).Center()
	if v.border > 0 {
		box.Border(v.border, v.borderColor)
	}
	switch {
	case v.hasImage:
		pixels := v.pixels
		box.Child(el.Widget(core.Func(func(gtx core.C) core.D {
			return (widget.Image{Src: pixels, Fit: widget.Cover, Position: layout.Center, Scale: 1}).Layout(gtx)
		})).Size(el.Dp(v.size)))
	case strings.TrimSpace(v.name) == "":
		if v.placeholder != IconNone {
			box.Child(Icon(v.placeholder).Size(v.size * .55).Color(fg).Render(cx))
		}
	default:
		box.Child(el.Text(avatarInitials(v.name)).TextSize(v.size * .36).TextColor(fg).MaxLines(1))
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
	if v.style != nil {
		v.style(box)
	}
	return box
}
