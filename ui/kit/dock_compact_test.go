package kit

import (
	"reflect"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestDockCompactPreservesDesktopLayoutAndInput(t *testing.T) {
	input := Input("Query")
	d := Dock(nil).Panel(DockPanel{ID: "editor", Title: "Editor", View: input}, DockCenter).
		Panel(DockPanel{ID: "files", Title: "Files", View: viewFunc(func(*el.Context) el.Element { return el.Text("file list") })}, DockLeft)
	d.Split("files", "editor", DockPlacementRight)
	d.Zoom("editor")
	before := d.Layout()
	calls := 0
	d.OnLayoutChange(func(DockLayout) { calls++ }).Compact(true)
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().H(el.Dp(600)).Child(d.Render(cx)) }), 320, 1)
	clickClass(t, h, "Editor", "Query")
	h.Type("kept query")
	click(t, h, "Files")
	if !shown(h, "file list") {
		t.Fatal("compact panel did not switch")
	}
	click(t, h, "Editor")
	if input.Value() != "kept query" {
		t.Fatal("switch lost input", input.Value())
	}
	d.Compact(false)
	h.Frame()
	if !reflect.DeepEqual(before, d.Layout()) || calls != 0 {
		t.Fatal("presentation changed desktop layout", d.Layout(), calls)
	}
	d.Compact(true)
	h.Frame()
	if input.Value() != "kept query" {
		t.Fatal("resize lost input")
	}
}

func TestDockCompactClosesObsoleteMenus(t *testing.T) {
	d := Dock(nil)
	menu := Menu().Item("Action", "", func() {})
	menu.SetValue(true)
	d.menus[&DockNode{}] = menu
	d.Compact(true)
	if menu.Value() {
		t.Fatal("desktop menu survived presentation change")
	}
	d.compactMenu = Menu().Item("Action", "", func() {})
	d.compactMenu.SetValue(true)
	d.Compact(false)
	if d.compactMenu.Value() {
		t.Fatal("compact menu survived presentation change")
	}
}

func TestDockCompactCloseAndDisabled(t *testing.T) {
	d := Dock(nil).Panel(DockPanel{ID: "one", Title: "One"}, DockCenter).
		Panel(DockPanel{ID: "two", Title: "Two", NoClose: true}, DockRight).Compact(true)
	calls := 0
	d.OnLayoutChange(func(DockLayout) { calls++ })
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().H(el.Dp(600)).Child(d.Render(cx)) }), 320, 1)
	d.SetDisabled(true)
	h.Frame()
	click(t, h, locale.Current().Close)
	if !d.Visible("one") || calls != 0 {
		t.Fatal("disabled close changed layout")
	}
	d.SetDisabled(false)
	h.Frame()
	click(t, h, locale.Current().Close)
	if d.Visible("one") || d.compactActive != "two" || calls != 1 {
		t.Fatal("close fallback", d.compactActive, calls)
	}
	if shown(h, locale.Current().Close) {
		t.Fatal("NoClose ignored")
	}
	d.SetVisible("one", true)
	h.Frame()
	click(t, h, "One")
	if d.compactActive != "one" {
		t.Fatal("reopened panel unreachable")
	}
}
