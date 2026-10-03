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
	mediaProgress      *ProgressCircleView
	content            el.View
	actions            []el.View
	styles             [6]func(*el.DivEl)
	density            AttachmentSize
	vertical           bool
	size               int64
	status             AttachmentStatus
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
	v.SetStatus(AttachmentStatusComplete)
	if p >= 0 {
		p = min(p, 1)
		v.status = AttachmentStatusUploading
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
	status := v.Status()
	metrics := v.metrics()
	state, detail, color := "", FileSize(v.size), theme.Muted
	switch status {
	case AttachmentStatusCanceled:
		state, detail = "canceled", text.Canceled
	case AttachmentStatusPending:
		state, detail = "pending", text.AttachmentPending
	case AttachmentStatusProcessing:
		state, detail = "processing", text.AttachmentProcessing
	case AttachmentStatusFailed:
		state, detail, color = "error", v.err, theme.DangerText
		if detail == "" {
			detail = text.AttachmentFailed
		}
	case AttachmentStatusUploading:
		pct := strconv.Itoa(int(v.progress*100+0.5)) + "%"
		state, detail = text.Uploading+" "+pct, text.Uploading+" "+pct
	}

	info := el.Div().Grow().W(el.Dp(0)).Gap(theme.SpaceXs).Items(el.Stretch)
	if v.content != nil {
		info.Child(v.content.Render(cx))
	} else {
		title := v.part(AttachmentPartTitle, id+"/title", el.Div().TextSize(metrics.title)).Child(el.Text(v.name).MaxLines(1))
		description := v.part(AttachmentPartDescription, id+"/description", el.Div().TextSize(metrics.description).TextColor(color)).Child(el.Text(detail).MaxLines(1))
		info.Child(title, description)
		if status.IsUploading() {
			info.Child(el.Div().H(el.Dp(4)).Rounded(theme.RadiusFull).Bg(theme.Subtle).Items(el.Start).Child(
				el.Div().H(el.Dp(4)).Rounded(theme.RadiusFull).Bg(theme.Primary).W(el.Frac(v.progress))))
		}
	}

	media := el.Div().ID(id + "/media").NoShrink().Rounded(theme.RadiusMd).Bg(theme.Highlight).Center()
	if v.media == nil {
		media.Size(el.Dp(metrics.media))
		if status.IsInProgress() {
			media.Child(Spinner().Size(metrics.media / 2).Label("").Render(cx))
		} else if status.IsFailed() {
			icon := IconError
			if v.onRetry == nil {
				icon = IconBan
			}
			tint := theme.DangerText
			tint.A = 24
			media.Bg(tint).Child(Icon(icon).Size(metrics.media / 2).Color(theme.DangerText).Render(cx))
		} else {
			media.Child(Icon(IconCopy).Size(metrics.media / 2).Color(theme.PrimaryText).Render(cx))
		}
	} else {
		media.Role("group").MaxW(el.Full).Child(v.media.Render(cx))
		if overlay := v.mediaStatus(cx, id, status); overlay != nil {
			media.Child(overlay)
		}
	}
	media = v.part(AttachmentPartMedia, id+"/media", media)
	main := el.Div().ID(id+"/open").Grow().W(el.Dp(0)).Row().Items(el.Center).Gap(metrics.gap).Child(media, info)
	if v.vertical {
		main.Col().W(el.Full).Flex(0).Items(el.Stretch)
		info.W(el.Full).Flex(0)
	}

	info = v.part(AttachmentPartContent, id+"/info", info)
	if v.onOpen != nil && status.IsComplete() {
		main.Role("button").Name(v.name).Border(1, theme.Surface).Rounded(theme.RadiusSm).CursorPointer().Focusable(true).OnClick(v.onOpen).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) })
	}
	card := surface().ID(id).Disabled(v.disabled).Role("attachment").Name(v.name).Value(state).
		Row().Items(el.Center).Gap(metrics.gap).P(metrics.padding).W(el.Dp(metrics.width)).MinH(el.Dp(metrics.height)).MaxW(el.Full).Child(main)
	if status.IsPending() {
		card.BorderDashed(true)
	}
	if status.IsFailed() {
		card.Border(1, theme.DangerText)
	}
	if v.vertical {
		card.Col().Items(el.Stretch)
	}
	actions := el.Div().ID(id + "/actions").Row().NoShrink().Gap(theme.SpaceXs).Items(el.Center)
	if v.vertical {
		actions.Wrap().Justify(el.End)
	}
	actions = v.part(AttachmentPartActions, id+"/actions", actions)
	hasCustom := false
	for i, view := range v.actions {
		if view != nil {
			actions.Child(el.Div().ID(id + "/action/" + strconv.Itoa(i)).Child(view.Render(cx)))
			hasCustom = true
		}
	}
	if status.IsInProgress() && v.onCancel != nil {
		actions.Child(Button("", func() {
			if v.disabled || !v.Status().IsInProgress() {
				return
			}
			v.SetStatus(AttachmentStatusCanceled)
			v.onCancel()
		}).Name(text.Name(text.Cancel, v.name)).Icon(IconClose).Variant(ButtonGhost).Size(metrics.action).Render(cx))
	}
	if (status.IsFailed() || status == AttachmentStatusCanceled) && v.onRetry != nil {
		actions.Child(Button(text.Retry, v.retry).Name(text.Name(text.Retry, v.name)).Variant(ButtonGhost).Size(metrics.action).Render(cx))
	}
	if v.onRemove != nil {
		actions.Child(Button("", v.onRemove).Name(text.Name(text.Remove, v.name)).Icon(IconClose).Variant(ButtonGhost).Size(metrics.action).Render(cx))
	}
	if hasCustom || v.onRemove != nil || (status.IsInProgress() && v.onCancel != nil) || ((status.IsFailed() || status == AttachmentStatusCanceled) && v.onRetry != nil) {
		card.Child(actions)
	}
	return v.part(AttachmentPartRoot, id, card).Role("attachment").Name(v.name).Value(state).Disabled(v.disabled).MaxW(el.Full)
}
