package kit

import (
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// A keymap action runs from its key, its key shows in Kbd and menus, and a
// rebind moves all three.
func TestKeymapActionKbdAndMenu(t *testing.T) {
	t.Cleanup(func() { core.Bind("test.save") })
	if err := core.Bind("test.save", "ctrl+s"); err != nil {
		t.Fatal(err)
	}
	saves := 0
	m := Menu().ActionItem("保存", "test.save", func() { saves++ })
	m.Trigger(Button("文件", m.Toggle))
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element {
		cx.Action("test.save", func() { saves++ })
		return el.Div().P(20).Gap(8).Child(KbdFor("test.save").Render(cx), KbdFor("test.unbound").Render(cx), m.Render(cx))
	})))
	h.Frame()
	h.Key("S", key.ModCtrl)
	h.Frame()
	if saves != 1 {
		t.Fatalf("ctrl+s ran %d saves", saves)
	}
	if !shown(h, "ctrl+s") {
		t.Fatal("KbdFor does not show the binding")
	}
	core.Bind("test.save", "ctrl+k")
	h.Frame()
	h.Key("S", key.ModCtrl)
	h.Key("K", key.ModCtrl)
	h.Frame()
	if saves != 2 {
		t.Fatalf("after rebinding, saves = %d", saves)
	}
	if shown(h, "ctrl+s") || !shown(h, "ctrl+k") {
		t.Fatal("KbdFor did not follow the rebinding")
	}
	click(t, h, "文件")
	h.Frame()
	if n, ok := semanticNode(h, "menuitem"); !ok || n.Desc.Label == "" {
		t.Fatal("menu did not open")
	}
	count := 0
	for _, n := range h.Router.AppendSemantics(nil) {
		if n.Desc.Label == "ctrl+k" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected the binding in Kbd and the menu, found %d", count)
	}
}
