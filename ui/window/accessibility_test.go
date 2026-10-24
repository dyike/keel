package window

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/loop"
)

// While an accessibility client reads the window, rows scrolled out of view
// are elements of their scroll area, at their real place; agents still see
// only what shows.
func TestAccessibilityElementsOutOfView(t *testing.T) {
	w := openTest(t, Options{Width: 300, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		list := el.Div().HFull().ScrollY()
		for i := range 40 {
			list.Child(el.Div().P(4).OnClick(func() {}).Child(el.Text(fmt.Sprintf("Row %d", i))))
		}
		return el.Div().HFull().Child(el.Text("Title"), list)
	}))})
	find := func(els []Element, name string) *Element {
		for i := range els {
			if els[i].Name == name {
				return &els[i]
			}
		}
		return nil
	}
	agent := w.snapshot()
	if find(agent, "Row 39") != nil {
		t.Fatal("agents see a row scrolled out of view")
	}
	for _, e := range agent {
		if e.Role == "scrollarea" {
			t.Fatal("agents see a scroll area")
		}
	}
	loop.SetAccessibility(true)
	defer loop.SetAccessibility(false)
	w.render()
	els := elementsOf(w.virt.router.AppendSemantics(nil), w.window(), true)
	row := find(els, "Row 39")
	if row == nil || row.Role != "button" {
		t.Fatalf("row out of view missing: %+v", els)
	}
	area := els[row.parent]
	if area.Role != "scrollarea" || row.viewport != area.box().Intersect(area.viewport) {
		t.Fatalf("row's container %+v, viewport %v", area, row.viewport)
	}
	if row.Y < area.Y+area.Height || row.Height <= 0 {
		t.Fatalf("row at %d, height %d; scroll area ends at %d", row.Y, row.Height, area.Y+area.Height)
	}
	if title := find(els, "Title"); title == nil || title.parent != -1 {
		t.Fatalf("title outside the scroll area: %+v", title)
	}
	if again := w.snapshot(); len(again) != len(agent) {
		t.Fatalf("agent snapshot changed with accessibility on: %d elements, was %d", len(again), len(agent))
	}
}

// The element holding keyboard focus is marked: a button reached with Tab,
// and a text field through the editor node inside it.
func TestAccessibilityFocus(t *testing.T) {
	w := openTest(t, Options{Width: 300, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Col().Child(
			el.Div().P(4).OnClick(func() {}).Child(el.Text("One")),
			el.Div().P(4).OnClick(func() {}).Child(el.Text("Two")),
			el.Input().Placeholder("Name"),
		)
	}))})
	focused := func() []string {
		t.Helper()
		w.render()
		var names []string
		for _, e := range elementsOf(w.virt.router.AppendSemantics(nil), w.window(), true) {
			if e.focused {
				names = append(names, e.Role+" "+e.Name)
			}
		}
		return names
	}
	if f := focused(); len(f) != 0 {
		t.Fatalf("focused before any input: %v", f)
	}
	for _, want := range []string{"button One", "button Two"} {
		if err := w.press("tab"); err != nil {
			t.Fatal(err)
		}
		if f := focused(); len(f) != 1 || f[0] != want {
			t.Fatalf("after Tab, focused %v, want %s", f, want)
		}
	}
	w.click(element(t, w, "Name").center())
	if f := focused(); len(f) != 1 || f[0] != "textbox Name" {
		t.Fatalf("after clicking the field, focused %v", f)
	}
}
