package kit

import (
	"image"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// A title icon sits before the title and survives message reuse; without
// one the title starts at the panel's edge.
func TestDialogTitleIcon(t *testing.T) {
	plain := Dialog("")
	iconic := Dialog("").Icon(IconWarning)
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(plain.Render(cx), iconic.Render(cx))
	})))
	plain.Alert("没有图标", "正文", nil)
	h.Frame()
	h.Frame()
	without := innermost(h, "没有图标")
	plain.SetValue(false)
	iconic.ConfirmDanger("删除订单", "确定删除？", "删除", nil)
	h.Frame()
	h.Frame()
	with := innermost(h, "删除订单")
	if with.Min.X-without.Min.X < 22 {
		t.Fatalf("title not moved past the icon: %v vs %v", with, without)
	}
	if iconic.icon != IconWarning || iconic.iconSet {
		t.Fatal("message reuse dropped the icon or set a tone")
	}
	iconic.IconTone(ToneWarning).Icon(IconNone)
	h.Frame()
	if b := innermost(h, "删除订单"); b.Min.X != without.Min.X {
		t.Fatalf("IconNone kept the space: %v", b)
	}
}

// innermost is the smallest element with the name: the title text rather
// than the dialog named after it.
func innermost(h *uitest.Harness, name string) image.Rectangle {
	var best image.Rectangle
	for _, n := range h.Router.AppendSemantics(nil) {
		if n.Desc.Label == name && (best.Empty() || n.Desc.Bounds.Dx()*n.Desc.Bounds.Dy() < best.Dx()*best.Dy()) {
			best = n.Desc.Bounds
		}
	}
	return best
}
