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
