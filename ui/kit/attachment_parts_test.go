package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestAttachmentContentActionsAndStyles(t *testing.T) {
	for _, scale := range []int{1, 2} {
		opens, custom, removes := 0, 0, 0
		a := Attachment("file", 1024).Vertical(true).OnOpen(func() { opens++ }).OnRemove(func() { removes++ })
		action := Button("Inspect", func() { custom++ })
		input := Input("Note")
		actions := []el.View{action, nil, input}
		a.Actions(actions...)
		actions[0] = Button("Wrong", nil)
		a.Content(el.ViewFunc(func(*el.Context) el.Element { return el.Div().Child(el.Text("Custom metadata"), el.Text("Version 2")) }))
		h := renderView(a, 280, scale)
		if shown(h, "Wrong") || shown(h, "1.0 KB") {
			t.Fatal("slice ownership or default metadata")
		}
		click(t, h, "Inspect")
		if custom != 1 || opens != 0 {
			t.Fatal("action bubbled to open")
		}
		click(t, h, "Custom metadata")
		if opens != 1 {
			t.Fatal("content did not open")
		}
		clickClass(t, h, "Editor", "Note")
		h.Type("a")
		a.PartStyle(AttachmentPartRoot, func(e *el.DivEl) { e.P(6).W(el.Dp(260)).Role("wrong").Name("wrong") }).
			PartStyle(AttachmentPartContent, func(e *el.DivEl) { e.Gap(12) }).
			PartStyle(AttachmentPartActions, func(e *el.DivEl) { e.Gap(10) })
		h.Frame()
		h.Key(key.NameRightArrow, 0)
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "" {
			t.Fatal("style lost action focus")
		}
		n, ok := semanticNode(h, "attachment")
		if !ok || n.Desc.Label != "file" || n.Desc.Bounds.Dx() != 260*scale {
			t.Fatal("root style or identity", n.Desc)
		}
		a.Content(nil).PartStyle(AttachmentPartRoot, nil).PartStyle(AttachmentPartContent, nil).PartStyle(AttachmentPartActions, nil)
		h.Frame()
		h.Type("b")
		if input.Value() != "b" || !shown(h, "1.0 KB") {
			t.Fatal("restoration lost state or defaults")
		}
		a.Actions()
		h.Frame()
		if shown(h, "Inspect") || shown(h, "Note") {
			t.Fatal("custom actions not cleared")
		}
		click(t, h, "移除 file")
		if removes != 1 || opens != 1 {
			t.Fatal("built-in action lost")
		}
	}
}

func TestAttachmentCustomActionsDisabled(t *testing.T) {
	calls := 0
	a := Attachment("file", 0).Actions(Button("Action", func() { calls++ }))
	a.SetDisabled(true)
	h := renderView(a, 300, 1)
	click(t, h, "Action")
	if calls != 0 {
		t.Fatal("disabled custom action")
	}
	a.SetDisabled(false)
	h.Frame()
	click(t, h, "Action")
	if calls != 1 {
		t.Fatal("restored custom action")
	}
	a.PartStyle(AttachmentPart(255), func(*el.DivEl) { t.Fatal("invalid style part called") })
	h.Frame()
}
