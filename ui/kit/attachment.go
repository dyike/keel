package kit

import (
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// AttachmentView is a file card: name, size, upload progress or an error,
// and optional open and remove actions.
type AttachmentView struct {
	name     string
	size     int64
	progress float32 // 0..1 while uploading, < 0 when done
	err      string
	onOpen   func()
	onRemove func()
}

// Attachment shows a file of size bytes.
func Attachment(name string, size int64) *AttachmentView {
	return &AttachmentView{name: name, size: size, progress: -1}
}
func (v *AttachmentView) OnOpen(fn func()) *AttachmentView   { v.onOpen = fn; return v }
func (v *AttachmentView) OnRemove(fn func()) *AttachmentView { v.onRemove = fn; return v }

// SetProgress shows an upload at fraction p (0..1); a negative p means done.
func (v *AttachmentView) SetProgress(p float32) {
	if p >= 0 {
		p = min(p, 1)
	}
	v.progress = p
}

// SetError shows why the upload failed; "" clears it.
func (v *AttachmentView) SetError(msg string) { v.err = msg }

// FileSize formats a byte count as B, KB, MB or GB.
func FileSize(n int64) string {
	const k = 1024
	switch {
	case n < k:
		return strconv.FormatInt(n, 10) + " B"
	case n < k*k:
		return strconv.FormatFloat(float64(n)/k, 'f', 1, 64) + " KB"
	case n < k*k*k:
		return strconv.FormatFloat(float64(n)/(k*k), 'f', 1, 64) + " MB"
	}
	return strconv.FormatFloat(float64(n)/(k*k*k), 'f', 1, 64) + " GB"
}

func (v *AttachmentView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	state, detail, color := "", FileSize(v.size), theme.Muted
	switch {
	case v.err != "":
		state, detail, color = "error", v.err, theme.DangerText
	case v.progress >= 0:
		pct := strconv.Itoa(int(v.progress*100+0.5)) + "%"
		state, detail = text.Uploading+" "+pct, text.Uploading+" "+pct
	}
	info := el.Div().Grow().W(el.Dp(0)).Gap(4).Items(el.Stretch).Child(
		el.Text(v.name).MaxLines(1),
		el.Text(detail).TextSize(12).TextColor(color).MaxLines(1),
	)
	if v.progress >= 0 && v.err == "" {
		info.Child(el.Div().H(el.Dp(4)).Rounded(2).Bg(theme.Subtle).Items(el.Start).Child(
			el.Div().H(el.Dp(4)).Rounded(2).Bg(theme.Primary).W(el.Frac(v.progress))))
	}
	card := surface().Role("attachment").Name(v.name).Value(state).Row().Items(el.Center).Gap(10).P(10).W(el.Dp(280)).MaxW(el.Full).
		Child(el.Div().Size(el.Dp(36)).NoShrink().Rounded(6).Bg(theme.Highlight).Center().
			Child(Icon(IconCopy).Size(18).Color(theme.PrimaryText).Render(cx)), info)
	if v.onOpen != nil {
		card.CursorPointer().Focusable(true).OnClick(v.onOpen)
	}
	if v.onRemove != nil {
		card.Child(Button("", v.onRemove).Name(text.Name(text.Remove, v.name)).Icon(IconClose).Variant(ButtonGhost).Size(24).Render(cx))
	}
	return card
}
