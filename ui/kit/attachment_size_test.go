package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestAttachmentSizesAndFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		a := Attachment("file", 1024).OnRemove(func() { calls++ })
		h := renderView(a, 400, scale)
		click(t, h, "移除 file")
		previousHeight := 0
		for _, tc := range []struct {
			size  AttachmentSize
			width int
		}{
			{AttachmentSizeXSmall, 176}, {AttachmentSizeSmall, 200}, {AttachmentSizeMedium, 232}, {AttachmentSizeLarge, 272},
		} {
			a.Size(tc.size)
			h.Frame()
			n, ok := semanticNode(h, "attachment")
			if !ok || n.Desc.Bounds.Dx() != tc.width*scale || n.Desc.Bounds.Dy() <= previousHeight {
				t.Fatal("size geometry", scale, tc.size, n.Desc.Bounds)
			}
			previousHeight = n.Desc.Bounds.Dy()
			h.Key(key.NameSpace, 0)
		}
		if calls != 5 {
			t.Fatal("size changes lost button focus", calls)
		}
		before := bounds(h, "file")
		a.Size(255)
		h.Frame()
		if bounds(h, "file") != before {
			t.Fatal("invalid size applied")
		}
		a.PartStyle(AttachmentPartRoot, func(e *el.DivEl) { e.W(el.Dp(180)) }).PartStyle(AttachmentPartMedia, func(e *el.DivEl) { e.Size(el.Dp(50)) })
		h.Frame()
		n, _ := semanticNode(h, "attachment")
		if n.Desc.Bounds.Dx() != 180*scale {
			t.Fatal("part override lost")
		}
	}
}

func TestAttachmentSizesNarrowVerticalAndState(t *testing.T) {
	for _, size := range []AttachmentSize{AttachmentSizeXSmall, AttachmentSizeSmall, AttachmentSizeMedium, AttachmentSizeLarge} {
		a := Attachment("long filename.pdf", 8192).Size(size).Vertical(true).OnCancel(func() {})
		a.SetProgress(.6)
		h := renderView(a, 160, 1)
		n, _ := node(h, "long filename.pdf")
		if n.Desc.Bounds.Dx() > 160 || n.Desc.Bounds.Dy() == 0 {
			t.Fatal("narrow bounds", n.Desc.Bounds)
		}
		click(t, h, "取消 long filename.pdf")
		if a.Status() != AttachmentStatusCanceled {
			t.Fatal("size lost upload action")
		}
	}
}
