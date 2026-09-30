package window

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
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

func TestAutomationScroll(t *testing.T) {
	var rows []core.Widget
	for i := 1; i <= 40; i++ {
		rows = append(rows, widget.Text(fmt.Sprintf("row %d", i)))
	}
	w := openTest(t, Options{Width: 300, Height: 200, Content: layout.Column(rows...)})
	before := element(t, w, "row 1").Y
	w.scroll(element(t, w, "row 1").center(), 300)
	w.snapshot()
	if after := element(t, w, "row 1").Y; after >= before-200 {
		t.Fatalf("row 1 moved from y=%d to y=%d; expected about 300dp up", before, after)
	}
}

func TestAutomationTabMovesFocus(t *testing.T) {
	a, b := widget.Input("a"), widget.Input("b")
	w := openTest(t, Options{Content: layout.Column(a, b)})
	w.click(element(t, w, "a").center())
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.typeText("x"); err != nil {
		t.Fatal(err)
	}
	if a.Value() != "" || b.Value() != "x" {
		t.Fatalf("a=%q b=%q; Tab should have moved focus to b", a.Value(), b.Value())
	}
}

func TestAutomationCloseFromCallback(t *testing.T) {
	var w *Window
	closed := false
	w = openTest(t, Options{
		Content: widget.Button("close", func() { w.Close() }),
		OnClose: func() { closed = true },
	})
	other := openTest(t, Options{Content: widget.Text("stay")}) // keeps the process alive
	w.click(element(t, w, "close").center())
	if !closed || !w.Closed() || other.Closed() {
		t.Fatalf("closed=%t w.Closed=%t other.Closed=%t", closed, w.Closed(), other.Closed())
	}
}

func TestAutomationDisabledAndChecked(t *testing.T) {
	b := widget.Button("save", nil)
	b.SetDisabled(true)
	w := openTest(t, Options{Content: layout.Column(b, widget.Checkbox("agree", true))})
	if e := element(t, w, "save"); !e.Disabled {
		t.Errorf("disabled button reported enabled: %+v", e)
	}
	if e := element(t, w, "agree"); e.Checked == nil || !*e.Checked {
		t.Errorf("checked box reported %+v", e)
	}
}

// Agent actions must redraw the real windows even when no component callback
// runs: a checkbox without OnChange changes state silently.
func TestAutomationRedrawsRealWindows(t *testing.T) {
	redraws := 0
	key := new(int)
	loop.Register(key, func() { redraws++ })
	defer loop.Unregister(key)
	w := openTest(t, Options{Content: widget.Checkbox("silent", false)})
	redraws = 0
	w.click(element(t, w, "silent").center())
	if redraws == 0 {
		t.Fatal("clicking a checkbox without OnChange did not invalidate real windows")
	}
}
