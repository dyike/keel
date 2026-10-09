package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"math"
	"testing"
)

func TestAttachmentGroupScrollAndRemoval(t *testing.T) {
	for _, scale := range []int{1, 2} {
		opened, removed := 0, 0
		first := Attachment("first", 1024)
		first.SetProgress(.4)
		last := Attachment("last", 2048).OnOpen(func() { opened++ })
		views := []el.View{first, nil, last}
		group := AttachmentGroup(views...).Name("Files").Gap(12)
		views[0] = nil
		copy := group.Items()
		copy[0] = nil
		if len(group.Items()) != 2 || group.Items()[0] != first {
			t.Fatal("slice ownership")
		}
		last.OnRemove(func() { removed++; group.SetItems(first) })
		var cx *el.Context
		h := renderView(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return group.Render(c) }), 300, scale)
		h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(30*float32(scale), 20*float32(scale)), Scroll: f32.Pt(1000*float32(scale), 0)})
		h.Frame()
		h.Frame()
		offset, view, total := cx.ScrollStateX(autoID("attachment-group", group))
		if view != 300 || total != 488 || offset != 188 {
			t.Fatal("group geometry", scale, offset, view, total)
		}
		click(t, h, "last")
		if opened != 1 {
			t.Fatal("scrolled open hit area")
		}
		click(t, h, locale.Current().Name(locale.Current().Remove, "last"))
		h.Frame()
		h.Frame()
		if removed != 1 || opened != 1 || first.progress != .4 {
			t.Fatal("remove changed other state")
		}
		offset, _, _ = cx.ScrollStateX(autoID("attachment-group", group))
		if offset != 0 {
			t.Fatal("removed item left scroll offset", offset)
		}
		group.SetItems()
		h.Frame()
		if shown(h, "first") {
			t.Fatal("empty group retains item")
		}
	}
}

func TestAttachmentGroupDisabledAndGap(t *testing.T) {
	calls := 0
	a := Attachment("a", 0).OnOpen(func() { calls++ })
	group := AttachmentGroup(a).Gap(0)
	group.Gap(-1).Gap(float32(math.NaN())).Gap(float32(math.Inf(1)))
	if group.gap != 0 {
		t.Fatal("invalid gap")
	}
	group.SetDisabled(true)
	h := renderView(group, 300, 1)
	click(t, h, "a")
	if calls != 0 {
		t.Fatal("disabled group opened attachment")
	}
	group.SetDisabled(false)
	h.Frame()
	click(t, h, "a")
	if calls != 1 || a.disabled {
		t.Fatal("group disabled mutated item")
	}
}
