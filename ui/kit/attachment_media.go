package kit

import (
	"image/color"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// mediaStatus overlays the preview without contributing to its measured size.
// The media group's clip follows AttachmentPartMedia's final corner radius.
func (v *AttachmentView) mediaStatus(cx *el.Context, id string, status AttachmentStatus) el.Element {
	if !status.IsInProgress() && !status.IsFailed() {
		v.mediaProgress = nil
		return nil
	}
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	shade := color.NRGBA{A: 112}
	overlay := el.Div().ID(id + "/media-status").Absolute().Top(0).Left(0).WFull().HFull().Center()
	text := locale.Current()
	if status.IsFailed() {
		v.mediaProgress = nil
		shade.A = 160
		if v.onRetry != nil {
			overlay.Child(el.Div().ID(id+"/media-retry").Role("button").Name(text.Name(text.Retry, v.name)).
				Size(el.Dp(28)).Center().Rounded(theme.RadiusFull).Bg(color.NRGBA{A: 160}).Border(2, color.NRGBA{}).
				Focusable(true).CursorPointer().OnClick(v.retry).
				FocusStyle(func(s *el.Style) { s.BorderColor(white) }).
				Hover(func(s *el.Style) { s.Bg(color.NRGBA{A: 220}) }).
				Child(Icon(IconRetry).Size(18).Color(white).Render(cx)))
		} else {
			overlay.Child(Icon(IconBan).Size(24).Color(white).Render(cx))
		}
	} else {
		if v.mediaProgress == nil {
			v.mediaProgress = ProgressCircle("")
		}
		v.mediaProgress.SetLabel(text.Name(text.Uploading, v.name))
		v.mediaProgress.SetValue(v.progress)
		v.mediaProgress.SetIndeterminate(status.IsProcessing())
		if status.IsProcessing() {
			v.mediaProgress.SetLabel(text.Name(text.AttachmentProcessing, v.name))
		}
		overlay.Child(v.mediaProgress.Size(24).Color(white).Render(cx))
	}
	return overlay.Bg(shade)
}

func (v *AttachmentView) retry() {
	if v.disabled || v.onRetry == nil || !(v.Status().IsFailed() || v.Status() == AttachmentStatusCanceled) {
		return
	}
	v.SetProgress(0)
	v.onRetry()
}
