package kit

import (
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestMenuActionItemRoutesToHandler(t *testing.T) {
	const save, quit = "test.route.save", "test.route.quit"
	var got []string
	inner := Menu().ActionItem("Save", save, nil).ActionItem("Quit", quit, nil)
	inner.Trigger(Button("Editor menu", inner.Toggle))
	outer := Menu().ActionItem("Save", save, nil)
	outer.Trigger(Button("App menu", outer.Toggle))
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.Action(save, func() { got = append(got, "global save") })
		cx.Action(quit, func() { got = append(got, "quit") })
		cx.ActionAt("editor", save, func() { got = append(got, "editor save") })
		return el.Div().Child(
			el.Div().ID("editor").Child(inner.Render(cx)),
			outer.Render(cx),
		)
	}), 640, 1)
	choose := func(trigger, item string) {
		t.Helper()
		click(t, h, trigger)
		h.Frame()
		click(t, h, item)
		h.Frame()
	}
	// Inside the editor the scoped handler wins, with no key bound at all.
	choose("Editor menu", "Save")
	// Outside it, the global handler runs.
	choose("App menu", "Save")
	// An action with only a global handler runs from anywhere.
	choose("Editor menu", "Quit")
	want := []string{"editor save", "global save", "quit"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	// A callback still takes precedence over routing.
	own := 0
	m := Menu().ActionItem("Own", save, func() { own++ })
	m.Trigger(Button("Own menu", m.Toggle))
	h = renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.Action(save, func() { got = append(got, "should not run") })
		return m.Render(cx)
	}), 640, 1)
	click(t, h, "Own menu")
	h.Frame()
	click(t, h, "Own")
	if own != 1 || len(got) != 3 {
		t.Fatal("explicit callback", own, got)
	}
}
