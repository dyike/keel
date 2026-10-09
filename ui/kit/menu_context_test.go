package kit

import (
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

func menuContextBindings(t *testing.T) string {
	t.Helper()
	const action = "test.menu.context"
	core.Bind(action, "f6")
	core.BindIn("menu-editor", action, "f7")
	core.BindIn("menu-other", action, "f8")
	t.Cleanup(func() {
		core.Bind(action)
		core.ClearBindingIn("menu-editor", action)
		core.ClearBindingIn("menu-other", action)
	})
	return action
}

func TestMenuContextHintsFirstFrameRebindingAndSubmenus(t *testing.T) {
	action := menuContextBindings(t)
	sub := Menu().ActionItem("Nested", action, nil)
	m := Menu().ActionItem("Save", action, nil).Sub("More", sub)
	m.Trigger(Button("Open", m.Toggle))
	m.SetValue(true)
	context := "menu-editor"
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().KeyContext(context).Child(m.Render(cx))
	}), 640, 1)
	if !shown(h, "f7") || shown(h, "f6") {
		t.Fatal("initial open did not use trigger context")
	}
	context = "menu-other"
	h.Frame()
	if !shown(h, "f8") || shown(h, "f7") {
		t.Fatal("current frame used stale context")
	}
	click(t, h, "More")
	h.Frame()
	if !shown(h, "Nested") || !shown(h, "f8") {
		t.Fatal("submenu did not inherit context")
	}
	core.BindIn("menu-other", action, "f9")
	h.Frame()
	if !shown(h, "f9") || shown(h, "f8") {
		t.Fatal("live rebinding did not update")
	}
	core.BindIn("menu-other", action)
	h.Frame()
	if shown(h, "f9") || shown(h, "f6") {
		t.Fatal("disabled binding fell back to global")
	}
	core.ClearBindingIn("menu-other", action)
	h.Frame()
	if !shown(h, "f6") {
		t.Fatal("cleared override did not restore global binding")
	}
}

func TestMenuExplicitActionTargetAndMissingTargets(t *testing.T) {
	action := menuContextBindings(t)
	calls := 0
	m := Menu().ActionContext("target").ActionItem("Run", action, func() { calls++ })
	m.Trigger(Button("Open", m.Toggle))
	m.SetValue(true)
	hidden, disabled := false, false
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		// The explicitly targeted element is built after the menu.
		return el.Div().KeyContext("menu-other").Child(m.Render(cx),
			el.Div().ID("target").KeyContext("menu-editor").Hidden(hidden).Disabled(disabled))
	}), 640, 1)
	if !shown(h, "f7") || shown(h, "f8") {
		t.Fatal("explicit context did not win")
	}
	hidden = true
	h.Frame()
	if shown(h, "f7") || shown(h, "f8") || shown(h, "f6") {
		t.Fatal("hidden target showed hint")
	}
	hidden, disabled = false, true
	h.Frame()
	if shown(h, "f7") {
		t.Fatal("disabled target showed hint")
	}
	m.ActionContext("missing")
	h.Frame()
	if shown(h, "f6") || shown(h, "f8") {
		t.Fatal("missing target fell back")
	}
	m.ActionContext("")
	h.Frame()
	if !shown(h, "f8") {
		t.Fatal("reset did not restore trigger context")
	}
	click(t, h, "Run")
	if calls != 1 || m.Value() {
		t.Fatal("context hint changed command dispatch")
	}
}

func TestContextActionDispatchMatchesHints(t *testing.T) {
	action := menuContextBindings(t)
	input := Input("Editor")
	other := Input("Outside")
	calls := 0
	context := "menu-editor"
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.ActionAt("target", action, func() { calls++ })
		return el.Div().Child(
			el.Div().ID("target").KeyContext(context).Child(input.Render(cx)), other.Render(cx))
	}), 640, 1)
	clickClass(t, h, "Editor", "Editor")
	h.Key(key.Name("F6"), 0)
	h.Key(key.Name("F7"), 0)
	if calls != 1 {
		t.Fatal("context action used global binding or did not run", calls)
	}
	context = "menu-other"
	h.Frame()
	h.Key(key.Name("F7"), 0)
	h.Key(key.Name("F8"), 0)
	if calls != 2 {
		t.Fatal("context switch did not change dispatch", calls)
	}
	clickClass(t, h, "Editor", "Outside")
	h.Key(key.Name("F8"), 0)
	if calls != 2 {
		t.Fatal("scoped action ran outside target")
	}
}

func TestContextNestedHandlersPreferInnerAndRespectUnbinding(t *testing.T) {
	action := menuContextBindings(t)
	input := Input("Editor")
	outer, inner := 0, 0
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.ActionAt("outer", action, func() { outer++ })
		cx.ActionAt("inner", action, func() { inner++ })
		return el.Div().ID("outer").Child(el.Div().ID("inner").KeyContext("menu-editor").Child(input.Render(cx)))
	}), 640, 1)
	clickClass(t, h, "Editor", "Editor")
	h.Key(key.Name("F6"), 0)
	h.Key(key.Name("F7"), 0)
	if outer != 0 || inner != 1 {
		t.Fatal("outer handler bypassed inner scope", outer, inner)
	}
	core.BindIn("menu-editor", action)
	h.Frame()
	h.Key(key.Name("F6"), 0)
	h.Key(key.Name("F7"), 0)
	if outer != 0 || inner != 1 {
		t.Fatal("empty binding failed to suppress outer handler", outer, inner)
	}
}
