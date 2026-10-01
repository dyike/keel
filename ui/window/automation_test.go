package window

import (
	"fmt"
	"image"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/kit"
)

// openTest opens a virtual window as automation mode would, without a socket.
func openTest(t *testing.T, o Options) *Window {
	t.Helper()
	w := newWindow(o)
	openVirtual(w, true)
	t.Cleanup(func() { closeVirtual(w) })
	return w
}

func element(t *testing.T, w *Window, name string) Element {
	t.Helper()
	e, err := w.find("", name)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// views stacks views in an el.Embed, which the window pads and scrolls.
func views(vs ...el.View) core.Widget {
	return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		box := el.Div().Gap(8).Items(el.Start)
		for _, v := range vs {
			box.Child(v.Render(cx))
		}
		return box
	}))
}

func text(s string) el.View {
	return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) })
}

func TestAutomationImageRoleAndClick(t *testing.T) {
	clicked := false
	img := kit.Image(image.NewNRGBA(image.Rect(0, 0, 120, 40)), "产品图").OnClick(func() { clicked = true })
	w := openTest(t, Options{Content: views(img)})
	e := element(t, w, "产品图")
	if e.Role != "image" || e.Value != "loaded" {
		t.Fatalf("image lost accessible role/state: %+v", e)
	}
	w.click(e.center())
	if !clicked {
		t.Fatal("image click not delivered")
	}
}

func TestAutomationScroll(t *testing.T) {
	var rows []el.View
	for i := 1; i <= 40; i++ {
		rows = append(rows, text(fmt.Sprintf("row %d", i)))
	}
	w := openTest(t, Options{Width: 300, Height: 600, Content: views(rows...)})
	before := element(t, w, "row 12").Y
	w.scroll(element(t, w, "row 1").center(), 300)
	w.snapshot()
	if after := element(t, w, "row 12").Y; after > before-250 {
		t.Fatalf("row 12 moved from y=%d to y=%d; expected about 300dp up", before, after)
	}
	// Scrolled out of view, so no longer listed: only visible elements are.
	for _, e := range w.snapshot() {
		if e.Name == "row 1" {
			t.Fatalf("row 1 is scrolled out of view but still listed at %+v", e)
		}
	}
}

func TestAutomationTabMovesFocus(t *testing.T) {
	a, b := kit.Input("a"), kit.Input("b")
	w := openTest(t, Options{Content: views(a, b)})
	w.click(element(t, w, "a").center())
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.typeText("x"); err != nil {
		t.Fatal(err)
	}
	w.snapshot()
	if a.Value() != "" || b.Value() != "x" {
		t.Fatalf("a=%q b=%q; Tab should have moved focus to b", a.Value(), b.Value())
	}
}

func TestAutomationCloseFromCallback(t *testing.T) {
	var w *Window
	closed := false
	w = openTest(t, Options{
		Content: views(kit.Button("close", func() { w.Close() })),
		OnClose: func() { closed = true },
	})
	other := openTest(t, Options{Content: views(text("stay"))}) // keeps the process alive
	w.click(element(t, w, "close").center())
	if !closed || !w.Closed() || other.Closed() {
		t.Fatalf("closed=%t w.Closed=%t other.Closed=%t", closed, w.Closed(), other.Closed())
	}
}

func TestAutomationDisabledCheckedAndMixed(t *testing.T) {
	b := kit.Button("save", nil)
	b.SetDisabled(true)
	all := kit.Checkbox("全部项目", false)
	all.SetMixed(true)
	all.SetDisabled(true)
	w := openTest(t, Options{Content: views(b, kit.Checkbox("agree", true), all)})
	if e := element(t, w, "save"); !e.Disabled {
		t.Errorf("disabled button reported enabled: %+v", e)
	}
	if e := element(t, w, "agree"); e.Checked == nil || !*e.Checked {
		t.Errorf("checked box reported %+v", e)
	}
	if e := element(t, w, "全部项目"); e.Role != "checkbox" || e.Value != "mixed" || !e.Disabled {
		t.Errorf("missing mixed/disabled state: %+v", e)
	}
}

// Agent actions must redraw the real windows even when no component callback
// runs: a checkbox without OnChange changes state silently.
func TestAutomationRedrawsRealWindows(t *testing.T) {
	redraws := 0
	key := new(int)
	loop.Register(key, func() { redraws++ })
	defer loop.Unregister(key)
	w := openTest(t, Options{Content: views(kit.Checkbox("silent", false))})
	redraws = 0
	w.click(element(t, w, "silent").center())
	if redraws == 0 {
		t.Fatal("clicking a checkbox without OnChange did not invalidate real windows")
	}
}

// The first request after launch may be a key press: the shadow has not
// rendered yet, so no handler would receive it.
func TestAutomationShortcutAsFirstRequest(t *testing.T) {
	n := 0
	w := openTest(t, Options{Content: views(text("x")), Shortcuts: map[string]func(){"mod+n": func() { n++ }}})
	if err := w.press("mod+n"); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("shortcut fired %d times", n)
	}
}
