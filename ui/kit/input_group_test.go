package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestInputGroupLabelActionsAndDisabled(t *testing.T) {
	in := Input("")
	calls, actions := 0, 0
	group := InputGroup("Query", in).OnChange(func(string) { calls++ }).Suffix(Button("Go", func() { actions++ }))
	h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(260)).Child(group.Render(cx)) })
	click(t, h, "Query")
	h.Type("hello")
	if group.Value() != "hello" || calls != 1 {
		t.Fatalf("label did not focus editor: %q %d", group.Value(), calls)
	}
	click(t, h, "Go")
	if actions != 1 {
		t.Fatal("suffix action")
	}
	group.SetError("required")
	h.Frame()
	if !shown(h, "required") {
		t.Fatal("missing error")
	}
	group.SetDisabled(true)
	h.Frame()
	click(t, h, "Go")
	h.Type("ignored")
	if group.Value() != "hello" || actions != 1 {
		t.Fatal("disabled group")
	}
	group.SetDisabled(false)
	group.SetValue("program")
	h.Frame()
	if in.Value() != "program" || calls != 1 {
		t.Fatal("program value callback")
	}
}

func TestInputGroupDynamicAddonsPreserveEditorFocus(t *testing.T) {
	in := Input("")
	group := InputGroup("Query", in)
	h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(260)).Child(group.Render(cx)) })
	clickClass(t, h, "Editor", "Query")
	h.Type("a")
	group.Prefix(Icon(IconSearch))
	h.Frame()
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 1, End: 1}, Text: "b"})
	h.Frame()
	group.Prefix(nil)
	group.Suffix(Button("Go", func() {}))
	h.Frame()
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 2, End: 2}, Text: "c"})
	h.Frame()
	if in.Value() != "abc" {
		t.Fatalf("addon changed editor identity: %q", in.Value())
	}
}

func TestInputGroupFormNameAndInputDisable(t *testing.T) {
	in := Input("")
	group := InputGroup("", in)
	form := Form().Field("Customer", group, func() string { return "required" })
	h := page(form)
	clickClass(t, h, "Editor", "Customer")
	h.Type("name")
	if in.Value() != "name" {
		t.Fatal("form association")
	}
	in.SetDisabled(true)
	h.Frame()
	h.Type("x")
	if in.Value() != "name" {
		t.Fatal("input disabled")
	}
}

func TestInputGroupBlockAddonsAndFocus(t *testing.T) {
	in := TextArea("").Rows(2)
	actions := 0
	group := InputGroup("Message", in).
		Addon("heading", InputGroupBlockStart, text("Draft")).
		Addon("counter", InputGroupBlockEnd, text("Counter")).
		Addon("send", InputGroupBlockEnd, Button("Send", func() { actions++ }).Size(24))
	h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(260)).Child(group.Render(cx)) })
	editor, _ := node(h, "Message")
	// The label and group share a name; use text around the editor to establish ordering.
	if editor.Desc.Bounds.Empty() || bounds(h, "Draft").Min.Y >= bounds(h, "Counter").Min.Y || bounds(h, "Counter").Min.Y >= bounds(h, "Send").Min.Y {
		t.Fatal("block addons out of order")
	}
	click(t, h, "Draft")
	h.Type("first")
	if in.Value() != "first" {
		t.Fatal("addon text did not focus editor")
	}
	group.Addon("heading", InputGroupBlockStart, text("Updated"))
	group.Addon("counter", InputGroupBlockEnd, nil)
	h.Frame()
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 5, End: 5}, Text: "\nsecond"})
	h.Frame()
	if in.Value() != "first\nsecond" || shown(h, "Counter") || !shown(h, "Updated") {
		t.Fatal("addon update lost state or content")
	}
	click(t, h, "Send")
	h.Type("ignored")
	if actions != 1 || in.Value() != "first\nsecond" {
		t.Fatal("button action returned focus to editor")
	}
	in.SetReadOnly(true)
	h.Frame()
	click(t, h, "Updated")
	h.Type("ignored")
	click(t, h, "Send")
	if actions != 2 || in.Value() != "first\nsecond" {
		t.Fatal("read-only affected action or allowed editing")
	}
	group.SetDisabled(true)
	h.Frame()
	click(t, h, "Send")
	if actions != 2 {
		t.Fatal("disabled block action")
	}
	group.SetDisabled(false)
	in.SetDisabled(true)
	h.Frame()
	click(t, h, "Send")
	if actions != 2 {
		t.Fatal("input disable did not cover addon")
	}
}

func TestInputGroupAddonOrderNarrowLayoutAndKeyboard(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		group := InputGroup("Query", Input("")).Prefix(text("P")).Suffix(text("S")).
			Addon("start1", InputGroupInlineStart, text("A")).Addon("start2", InputGroupInlineStart, text("B")).
			Addon("end1", InputGroupInlineEnd, text("C")).Addon("end2", InputGroupInlineEnd, text("D")).
			Addon("action", InputGroupBlockEnd, Button("Go", func() { calls++ }).Size(24))
		disabled := false
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(group.Render(cx)) }), 200, scale)
		if !(bounds(h, "P").Min.X < bounds(h, "A").Min.X && bounds(h, "A").Min.X < bounds(h, "B").Min.X && bounds(h, "C").Min.X < bounds(h, "D").Min.X && bounds(h, "D").Min.X < bounds(h, "S").Min.X) {
			t.Fatal("inline addon order")
		}
		if b := bounds(h, "Go"); b.Max.X > 200*scale {
			t.Fatal("block addon overflow", b)
		}
		clickClass(t, h, "Editor", "Query")
		h.Router.MoveFocus(key.FocusForward)
		h.Frame()
		h.Key(key.NameReturn, 0)
		if calls != 1 {
			t.Fatal("Tab/Enter did not reach block action")
		}
		disabled = true
		h.Frame()
		click(t, h, "Go")
		if calls != 1 {
			t.Fatal("ancestor disable ignored")
		}
	}
}
