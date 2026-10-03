package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestAttachmentGroupScrollRequests(t *testing.T) {
	for _, scale := range []int{1, 2} {
		a, b, c := Attachment("a", 0), Attachment("b", 0), Attachment("c", 0)
		group := AttachmentGroup(a, b, c).Gap(12)
		group.ScrollTo(150)
		var cx *el.Context
		h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return group.Render(ctx) }), 300, scale)
		h.Frame()
		h.Frame()
		offset, view, content := group.ScrollState(cx)
		if offset != 150 || view != 300 || content != 720 {
			t.Fatal("initial request", scale, offset, view, content)
		}
		group.ScrollTo(math.MaxFloat32)
		h.Frame()
		h.Frame()
		offset, _, _ = group.ScrollState(cx)
		if offset != 420 {
			t.Fatal("end clamp", offset)
		}
		group.ScrollTo(-20)
		h.Frame()
		h.Frame()
		offset, _, _ = group.ScrollState(cx)
		if offset != 0 {
			t.Fatal("negative clamp")
		}
		group.ScrollTo(120)
		group.ScrollTo(float32(math.NaN()))
		group.ScrollTo(float32(math.Inf(1)))
		h.Frame()
		h.Frame()
		offset, _, _ = group.ScrollState(cx)
		if offset != 120 {
			t.Fatal("invalid offset replaced request")
		}
		group.ScrollTo(400)
		group.SetItems(a)
		h.Frame()
		h.Frame()
		offset, _, content = group.ScrollState(cx)
		if offset != 0 || content != 232 {
			t.Fatal("request did not use new content", offset, content)
		}
	}
}

func TestAttachmentGroupDeferredScrollDoesNotActivate(t *testing.T) {
	opens := 0
	group := AttachmentGroup(Attachment("a", 0), Attachment("b", 0).OnOpen(func() { opens++ }))
	hidden := true
	group.ScrollTo(1000)
	var cx *el.Context
	h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return el.Div().Hidden(hidden).Child(group.Render(ctx)) }), 300, 1)
	if !group.scrollPending {
		t.Fatal("hidden render consumed request")
	}
	hidden = false
	h.Frame()
	h.Frame()
	h.Frame()
	offset, _, _ := group.ScrollState(cx)
	if offset != 170 || opens != 0 {
		t.Fatal("deferred scroll", offset, opens)
	}
	group.SetDisabled(true)
	group.ScrollTo(0)
	h.Frame()
	h.Frame()
	offset, _, _ = group.ScrollState(cx)
	if offset != 0 {
		t.Fatal("programmatic scroll blocked by disabled actions")
	}
}
