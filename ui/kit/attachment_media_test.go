package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestAttachmentMediaLayoutAndActions(t *testing.T) {
	for _, scale := range []int{1, 2} {
		opened, removed, canceled, retried := 0, 0, 0, 0
		preview := el.ViewFunc(func(*el.Context) el.Element { return el.Div().Name("Preview").Size(el.Dp(48)).Child(el.Text("IMG")) })
		a := Attachment("photo.png", 1024).Media(preview).OnOpen(func() { opened++ }).OnRemove(func() { removed++ }).OnCancel(func() { canceled++ }).OnRetry(func() { retried++ })
		h := renderView(a, 220, scale)
		media := bounds(h, "Preview")
		// The outer attachment uses the same accessible name as its open button.
		if media.Empty() {
			t.Fatal("missing preview")
		}
		click(t, h, "Preview")
		if opened != 1 {
			t.Fatal("preview did not open", opened)
		}
		a.Vertical(true)
		h.Frame()
		vmedia := bounds(h, "Preview")
		detail := bounds(h, "1.0 KB")
		if detail.Min.Y < vmedia.Max.Y {
			t.Fatal("metadata not below media", vmedia, detail)
		}
		h.Key(key.NameSpace, 0)
		if opened != 2 {
			t.Fatal("orientation lost keyboard focus", opened)
		}
		click(t, h, locale.Current().Name(locale.Current().Remove, "photo.png"))
		if removed != 1 || opened != 2 {
			t.Fatal("remove activated preview")
		}
		a.SetProgress(.4)
		h.Frame()
		click(t, h, "Preview")
		if opened != 2 {
			t.Fatal("upload preview opened")
		}
		click(t, h, locale.Current().Name(locale.Current().Cancel, "photo.png"))
		click(t, h, locale.Current().Name(locale.Current().Retry, "photo.png"))
		if canceled != 1 || retried != 1 || a.progress != 0 {
			t.Fatal("upload controls")
		}
		a.Media(nil).Vertical(false)
		h.Frame()
		if shown(h, "Preview") {
			t.Fatal("media not removed")
		}
		a.SetProgress(-1)
		a.Media(preview)
		a.SetDisabled(true)
		h.Frame()
		click(t, h, "Preview")
		if opened != 2 {
			t.Fatal("disabled preview activated")
		}
		n, ok := semanticNode(h, "attachment")
		if !ok || n.Desc.Bounds.Dx() > 220*scale {
			t.Fatal("card overflow", n.Desc.Bounds)
		}
	}
}
