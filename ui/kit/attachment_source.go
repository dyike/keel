package kit

import (
	"context"
	"time"

	"gioui.org/layout"
	"gioui.org/op/paint"
	giowidget "gioui.org/widget"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

type attachmentImageSource struct {
	source   string
	cancel   context.CancelFunc
	revision uint64
	loading  bool
	err      error
	pixels   paint.ImageOp
}

// MediaSource asynchronously loads a URL, data URL or local path using
// core.DecodeImage. Repeating a source does not reload it; empty restores the
// default icon. Media(view) also cancels any source request.
func (v *AttachmentView) MediaSource(source string) *AttachmentView {
	if v.imageSource != nil && v.imageSource.source == source {
		return v
	}
	v.stopMediaLoad()
	v.media = nil
	v.imageSource = nil
	if source != "" {
		v.imageSource = &attachmentImageSource{source: source}
		v.RetryMedia()
	}
	return v
}
func (v *AttachmentView) MediaLoading() bool { return v.imageSource != nil && v.imageSource.loading }
func (v *AttachmentView) MediaError() error {
	if v.imageSource != nil {
		return v.imageSource.err
	}
	return nil
}

// RetryMedia reloads the current source. It does not retry the attachment upload.
func (v *AttachmentView) RetryMedia() {
	if v.imageSource == nil {
		return
	}
	v.stopMediaLoad()
	s := v.imageSource
	s.err, s.loading, s.pixels = nil, true, paint.ImageOp{}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	s.cancel = cancel
	revision, source := s.revision, s.source
	go func() {
		defer cancel()
		media, err := loadImageMedia(ctx, source)
		core.Update(func() {
			if v.imageSource != s || s.revision != revision {
				return
			}
			s.loading, s.cancel, s.err = false, nil, err
			if err == nil {
				s.pixels = paint.NewImageOp(media.still)
			}
		})
	}()
}
func (v *AttachmentView) stopMediaLoad() {
	if s := v.imageSource; s != nil {
		if s.cancel != nil {
			s.cancel()
			s.cancel = nil
		}
		s.revision++
		s.loading = false
	}
}

func (v *AttachmentView) sourceMedia(cx *el.Context) el.Element {
	s := v.imageSource
	id := autoID("attachment", v) + "/source"
	box := el.Div().ID(id).Absolute().Top(0).Left(0).WFull().HFull().Center()
	text := locale.Current()
	switch {
	case s.loading:
		box.Role("image").Name(v.name).Value("loading")
		if !v.hideMedia && !v.Status().IsInProgress() && !v.Status().IsFailed() {
			box.Child(Spinner().Size(20).Label("").Render(cx))
		}
	case s.err != nil:
		box.Role("group").Name(v.name).Value("image-error")
		if v.Status().IsInProgress() || v.Status().IsFailed() {
			box.Child(Icon(IconError).Size(20).Color(theme.DangerText).Render(cx))
		} else {
			box.Child(Button("", func() {
				if v.disabled || v.MediaError() == nil || v.Status().IsInProgress() || v.Status().IsFailed() {
					return
				}
				v.RetryMedia()
			}).ID(id + "/retry").Name(text.Name(text.Retry, v.name)).Icon(IconRetry).Variant(ButtonGhost).Size(24).Render(cx))
		}
	default:
		src := s.pixels
		box.Role("image").Name(v.name).Value("loaded").Child(el.Widget(core.Func(func(gtx core.C) core.D {
			return giowidget.Image{Src: src, Fit: giowidget.Cover, Position: layout.Center, Scale: 1 / gtx.Metric.PxPerDp}.Layout(gtx)
		})).WFull().HFull())
	}
	return box.Bg(theme.Highlight)
}
