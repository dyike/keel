package kit

import (
	"slices"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// Context predicates choose bindings by where focus is, and Kbd.At shows the
// binding as it applies to an element.
func TestKeyContextPredicatesAndKbdAt(t *testing.T) {
	const act = "test.pred.format"
	t.Cleanup(func() {
		core.Bind(act)
		for _, c := range []string{"Editor", "Editor && !ReadOnly", "Pane > Editor", "Terminal || Shell"} {
			core.ClearBindingIn(c, act)
		}
	})
	core.Bind(act, "f1")
	core.BindIn("Editor", act, "f2")
	core.BindIn("Editor && !ReadOnly", act, "f3")
	core.BindIn("Pane > Editor", act, "f4")
	core.BindIn("Terminal || Shell", act, "f5")
	for _, c := range []struct {
		path []string
		want string
	}{
		{[]string{"Editor"}, "f3"},                    // writable editor: the latest match wins
		{[]string{"Editor ReadOnly"}, "f2"},           // read-only: only the plain Editor binding holds
		{[]string{"Editor", "Pane"}, "f4"},            // inside a pane: bound last among matches
		{[]string{"Shell"}, "f5"},                     // either side of ||
		{[]string{"Sidebar"}, "f1"},                   // no context binding: the global one
		{[]string{"Toolbar", "Editor", "Pane"}, "f4"}, // an outer level matches when the inner does not
	} {
		if got := core.BindingsIn(act, c.path...); !slices.Equal(got, []string{c.want}) {
			t.Errorf("%v = %v, want %s", c.path, got, c.want)
		}
	}
	for _, bad := range []string{"Editor &&", "(Editor", "> Editor", "Editor & Pane", "!"} {
		if core.BindIn(bad, act, "f6") == nil {
			t.Errorf("accepted predicate %q", bad)
		}
	}

	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Gap(8).Child(
			el.Div().KeyContext("Pane").Child(el.Div().ID("code").KeyContext("Editor").Child(el.Text("代码"))),
			KbdFor(act).At("code").Render(cx),
			KbdFor(act).At("missing").Render(cx),
		)
	})))
	h.Frame()
	h.Frame()
	if !shown(h, "f4") {
		t.Fatal("Kbd.At does not show the binding at the element")
	}
	if shown(h, "f1") {
		t.Fatal("Kbd.At for a missing element showed the global binding")
	}
}
