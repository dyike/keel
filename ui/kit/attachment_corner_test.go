package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestAttachmentRemoveCorner(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, size := range []AttachmentSize{AttachmentSizeXSmall, AttachmentSizeSmall, AttachmentSizeMedium, AttachmentSizeLarge} {
			removed, opened, acted := 0, 0, 0
			a := Attachment("file", 0).Size(size).RemoveOnHover(false).OnRemove(func() { removed++ }).OnOpen(func() { opened++ }).Actions(Button("Other", func() { acted++ }).Size(20))
			group := AttachmentGroup(a)
			h := renderView(group, 400, scale)
			name := locale.Current().Name(locale.Current().Remove, "file")
			check := func() {
				t.Helper()
				card := bounds(h, "file")
				button := bounds(h, name)
				if button.Dx() != int(a.metrics().action)*scale || button.Dx() != button.Dy() || button.Min.X+button.Dx()/2 != card.Max.X || button.Min.Y+button.Dy()/2 != card.Min.Y {
					t.Fatal("corner geometry", scale, size, card, button)
				}
				if button.Min.Y < 0 {
					t.Fatal("overhang outside group", button)
				}
				// Activate the portion outside the card's right edge, still inside the disc.
				h.Click(float32(button.Max.X-3*scale), float32(button.Min.Y+button.Dy()/2))
			}
			check()
			if removed != 1 || opened != 0 {
				t.Fatal("overhang clipped or opened", removed, opened)
			}
			a.Vertical(true)
			h.Frame()
			h.Key(key.NameSpace, 0)
			if removed != 2 {
				t.Fatal("orientation lost remove focus", removed)
			}
			check()
			click(t, h, "Other")
			if acted != 1 || removed != 3 || opened != 0 {
				t.Fatal("actions overlapped", acted, removed, opened)
			}
			group.SetDisabled(true)
			h.Frame()
			click(t, h, name)
			if removed != 3 {
				t.Fatal("ancestor disabled ignored")
			}
			group.SetDisabled(false)
			a.PartStyle(AttachmentPartRoot, func(e *el.DivEl) { e.Hidden(true) })
			h.Frame()
			if shown(h, name) || shown(h, "file") {
				t.Fatal("root hidden left corner visible")
			}
			a.PartStyle(AttachmentPartRoot, nil)
			h.Frame()
			click(t, h, name)
			if removed != 4 {
				t.Fatal("root restore lost corner")
			}
		}
	}
}

func TestAttachmentCornerInStretchedParent(t *testing.T) {
	for _, width := range []int{180, 400} {
		a := Attachment("file", 0).OnRemove(func() {})
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().WFull().Child(a.Render(cx)) }), width, 1)
		card := bounds(h, "file")
		button := bounds(h, locale.Current().Name(locale.Current().Remove, "file"))
		if button.Min.X+button.Dx()/2 != card.Max.X || button.Max.X > width {
			t.Fatal("corner follows parent instead of card", width, card, button)
		}
	}
}
