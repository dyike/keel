package kit

import (
	"math"
	"strconv"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// AttachmentView is a file card: name, size, upload progress or an error,
// and optional open and remove actions.
type AttachmentView struct {
	name               string
	media              el.View
	vertical           bool
	size               int64
	progress           float32 // 0..1 while uploading, < 0 when done
	err                string
	onOpen             func()
	onRemove           func()
	onCancel           func()
	onRetry            func()
	canceled, disabled bool
}

// Attachment shows a file of size bytes.
func Attachment(name string, size int64) *AttachmentView {
	return &AttachmentView{name: name, size: size, progress: -1}
}
func (v *AttachmentView) OnOpen(fn func()) *AttachmentView   { v.onOpen = fn; return v }
func (v *AttachmentView) OnRemove(fn func()) *AttachmentView { v.onRemove = fn; return v }

func (v *AttachmentView) OnCancel(fn func()) *AttachmentView { v.onCancel = fn; return v }
func (v *AttachmentView) OnRetry(fn func()) *AttachmentView  { v.onRetry = fn; return v }
func (v *AttachmentView) SetDisabled(on bool)                { v.disabled = on }

// Media replaces the file icon with a display view (for example kit.Image).
// Nil restores the default icon. Media should not contain interactive controls;
// the completed attachment's OnOpen handles activation of the whole preview.
func (v *AttachmentView) Media(view el.View) *AttachmentView { v.media = view; return v }

// Vertical stacks the preview above metadata and actions. False restores the row.
func (v *AttachmentView) Vertical(on bool) *AttachmentView { v.vertical = on; return v }

// SetProgress shows an upload at fraction p (0..1); a negative p means done.
func (v *AttachmentView) SetProgress(p float32) {
	if math.IsNaN(float64(p)) || math.IsInf(float64(p), 0) {
		return
	}
	v.canceled, v.err = false, ""
	if p >= 0 {
		p = min(p, 1)
	}
	v.progress = p
}

// SetError shows why the upload failed; "" clears it.
func (v *AttachmentView) SetError(msg string) { v.err = msg }

// FileSize formats a byte count as B, KB, MB or GB.
func FileSize(n int64) string {
	n = max(0, n)
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
	id := autoID("attachment", v)
	state, detail, color := "", FileSize(v.size), theme.Muted
	switch {
	case v.canceled:
		state, detail = "canceled", text.Canceled
	case v.err != "":
		state, detail, color = "error", v.err, theme.DangerText
	case v.progress >= 0:
		pct := strconv.Itoa(int(v.progress*100+0.5)) + "%"
		state, detail = text.Uploading+" "+pct, text.Uploading+" "+pct
	}
	info := el.Div().ID(id+"/info").Grow().W(el.Dp(0)).Gap(theme.SpaceXs).Items(el.Stretch).Child(
		el.Text(v.name).MaxLines(1),
		el.Text(detail).TextSize(theme.TextSm).TextColor(color).MaxLines(1),
	)
	if v.progress >= 0 && v.err == "" {
		info.Child(el.Div().H(el.Dp(4)).Rounded(theme.RadiusFull).Bg(theme.Subtle).Items(el.Start).Child(
			el.Div().H(el.Dp(4)).Rounded(theme.RadiusFull).Bg(theme.Primary).W(el.Frac(v.progress))))
	}
	media := el.Div().ID(id + "/media").NoShrink().Rounded(theme.RadiusMd).Bg(theme.Highlight).Center()
	if v.media == nil {
		media.Size(el.Dp(36)).Child(Icon(IconCopy).Size(18).Color(theme.PrimaryText).Render(cx))
	} else {
		media.MaxW(el.Full).Child(v.media.Render(cx))
	}
	main := el.Div().ID(id+"/open").Grow().W(el.Dp(0)).Row().Items(el.Center).Gap(10).Child(media, info)
	if v.vertical {
		main.Col().W(el.Full).Flex(0).Items(el.Stretch)
		info.W(el.Full).Flex(0)
	}

	if v.onOpen != nil && v.progress < 0 && v.err == "" && !v.canceled {
		main.Role("button").Name(v.name).Border(1, theme.Surface).Rounded(theme.RadiusSm).CursorPointer().Focusable(true).OnClick(v.onOpen).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) })
	}
	card := surface().ID(id).Disabled(v.disabled).Role("attachment").Name(v.name).Value(state).
		Row().Items(el.Center).Gap(10).P(10).W(el.Dp(280)).MaxW(el.Full).Child(main)
	if v.vertical {
		card.Col().Items(el.Stretch)
	}
	actions := el.Div().ID(id + "/actions").Row().NoShrink().Gap(theme.SpaceXs).Items(el.Center)
	if v.vertical {
		actions.Wrap().Justify(el.End)
	}
	if v.progress >= 0 && v.err == "" && !v.canceled && v.onCancel != nil {
		actions.Child(Button("", func() {
			if v.disabled || v.progress < 0 || v.canceled || v.err != "" {
				return
			}
			v.canceled, v.progress = true, -1
			v.onCancel()
		}).Name(text.Name(text.Cancel, v.name)).Icon(IconClose).Variant(ButtonGhost).Size(24).Render(cx))
	}
	if (v.err != "" || v.canceled) && v.onRetry != nil {
		actions.Child(Button(text.Retry, func() {
			if v.disabled || (v.err == "" && !v.canceled) {
				return
			}
			v.SetProgress(0)
			v.onRetry()
		}).Name(text.Name(text.Retry, v.name)).Variant(ButtonGhost).Size(24).Render(cx))
	}
	if v.onRemove != nil {
		actions.Child(Button("", v.onRemove).Name(text.Name(text.Remove, v.name)).Icon(IconClose).Variant(ButtonGhost).Size(24).Render(cx))
	}
	if v.onRemove != nil || (v.progress >= 0 && v.err == "" && !v.canceled && v.onCancel != nil) || ((v.err != "" || v.canceled) && v.onRetry != nil) {
		card.Child(actions)
	}
	return card
}
