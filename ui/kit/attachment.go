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
	imageSource        *attachmentImageSource
	mediaProgress      *ProgressCircleView
	mediaOverlay       el.View
	titleShimmer       *ShimmerTextView
	titleStatus        *AttachmentStatus
	descriptionStatus  *AttachmentStatus
	description        *string
	content            el.View
	actions            []el.View
	styles             [6]func(*el.DivEl)
	density            AttachmentSize
	vertical           bool
	mediaRatio         *float32
	hideMedia          bool
	hideContent        bool
	hideActions        bool
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
func (v *AttachmentView) Media(view el.View) *AttachmentView {
	v.stopMediaLoad()
	v.imageSource = nil
	v.media = view
	return v
}

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
	tile := v.vertical && v.hideContent && !v.hideMedia
	emptyMain := v.hideMedia && v.hideContent
	state, _, _ := v.statusDetail(status)
	_, detail, color := v.statusDetail(v.partStatus(AttachmentPartDescription))
	if v.description != nil {
		detail = *v.description
	}

	info := el.Div().Grow().W(el.Dp(0)).Gap(theme.SpaceXs).Items(el.Stretch)
	if v.content != nil {
		info.Child(v.content.Render(cx))
	} else {
		if v.titleShimmer == nil {
			v.titleShimmer = ShimmerText(v.name).MaxLines(1)
		}
		v.titleShimmer.SetText(v.name)
		title := v.part(AttachmentPartTitle, id+"/title", el.Div().TextSize(metrics.title)).Child(v.titleShimmer.Enabled(!v.hideContent && v.partStatus(AttachmentPartTitle).IsInProgress()).Render(cx))
		description := v.part(AttachmentPartDescription, id+"/description", el.Div().TextSize(metrics.description).TextColor(color)).Child(el.Text(detail).MaxLines(1))
		info.Child(title, description)
		if status.IsUploading() {
			info.Child(el.Div().H(el.Dp(4)).Rounded(theme.RadiusFull).Bg(theme.Subtle).Items(el.Start).Child(
				el.Div().H(el.Dp(4)).Rounded(theme.RadiusFull).Bg(theme.Primary).W(el.Frac(v.progress))))
		}
	}

	media := el.Div().ID(id + "/media").NoShrink().Rounded(theme.RadiusMd).Bg(theme.Highlight).Center()
	if v.imageSource != nil {
		media.Role("group").Child(v.sourceMedia(cx))
		if !v.vertical || v.previewRatio() == 0 {
			media.Size(el.Dp(metrics.media))
		}
		if overlay := v.mediaStatus(cx, id, status); overlay != nil {
			media.Child(overlay)
		}
	} else if v.media == nil {
		media.Size(el.Dp(metrics.media))
		if status.IsInProgress() && !v.hideMedia {
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
		if v.vertical && v.previewRatio() > 0 {
			media.Child(el.Div().ID(id + "/preview").Absolute().Top(0).Left(0).WFull().HFull().Center().Child(v.verticalMedia(cx)))
		} else {
			media.Child(v.media.Render(cx))
		}
		media.Role("group").MaxW(el.Full)
		if overlay := v.mediaStatus(cx, id, status); overlay != nil {
			media.Child(overlay)
		}
	}
	if v.mediaOverlay != nil {
		media.Role("group").Child(el.Div().ID(id + "/media-overlay").Absolute().Top(0).Left(0).WFull().HFull().Center().Child(v.mediaOverlay.Render(cx)))
	}
	if v.vertical && v.previewRatio() > 0 {
		media.WFull().H(el.Auto).AspectRatio(v.previewRatio()).Role("group")
	}
	if tile {
		media.Rounded(theme.RadiusLg - 1)
	}
	media = v.part(AttachmentPartMedia, id+"/media", media)
	if v.hideMedia {
		media.Hidden(true)
	}
	main := el.Div().ID(id + "/main").Grow().W(el.Dp(0)).Row().Items(el.Center).Gap(metrics.gap)
	if v.vertical {
		main.Col().W(el.Full).Flex(0).Items(el.Stretch)
		info.W(el.Full).Flex(0)
	}

	info = v.part(AttachmentPartContent, id+"/info", info)
	if v.hideContent {
		info.Hidden(true)
	}
	if emptyMain {
		main.Hidden(true)
	}
	if v.onOpen != nil && status.IsComplete() {
		main.Child(el.Div().ID(id+"/open").Absolute().Top(0).Left(0).WFull().HFull().
			Role("button").Name(v.name).Border(1, theme.Surface).Rounded(theme.RadiusSm).CursorPointer().Focusable(true).OnClick(v.onOpen).
			FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }))
	}
	main.Child(media, info)
	if tile && cx.Focused(id+"/open") {
		main.Child(el.Div().Absolute().Top(0).Left(0).WFull().HFull().Border(2, theme.Primary).Rounded(theme.RadiusLg - 1))
	}
	card := surface().ID(id).Disabled(v.disabled).Role("attachment").Name(v.name).Value(state).
		Row().Items(el.Center).Gap(metrics.gap).P(metrics.padding).W(el.Dp(metrics.width)).MinH(el.Dp(metrics.height)).MaxW(el.Full).Child(main)
	if tile {
		card.P(0).MinH(el.Dp(0))
	}
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
	if v.vertical && !emptyMain {
		actions.Absolute().Top(metrics.padding).Right(metrics.padding).MaxW(el.Full).Justify(el.End).Bg(theme.Surface).Rounded(theme.RadiusSm)
	}
	actions = v.part(AttachmentPartActions, id+"/actions", actions)
	if v.hideActions {
		actions.Hidden(true)
	}
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
