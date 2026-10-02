package kit

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func documentDock() *DockView {
	return Dock(text("welcome")).
		Panel(DockPanel{ID: "files", Title: "Files", View: text("file list")}, DockLeft).
		Panel(DockPanel{ID: "main.go", Title: "main.go", View: text("package main")}, DockCenter).
		Panel(DockPanel{ID: "app.go", Title: "app.go", View: text("package app")}, DockCenter)
}

// Documents fill the center in tab groups that split, save and restore like
// the side regions; with none left, the center view shows again.
func TestDockCenterDocuments(t *testing.T) {
	v := documentDock()
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(900)).H(el.Dp(600)).Child(v.Render(cx)) }), 900, 1)
	h.Frame()
	if shown(h, "welcome") || !shown(h, "package main") {
		t.Fatal("documents did not replace the center view")
	}
	if !v.Split("app.go", "main.go", DockPlacementRight) {
		t.Fatal("split documents")
	}
	h.Frame()
	if !shown(h, "package main") || !shown(h, "package app") || v.layout.CenterTree.First == nil {
		t.Fatal("split documents are not side by side")
	}

	data, err := json.Marshal(v.Layout())
	if err != nil {
		t.Fatal(err)
	}
	var l DockLayout
	if err := json.Unmarshal(data, &l); err != nil {
		t.Fatal(err)
	}
	restored := documentDock()
	if !restored.SetLayout(l) || !reflect.DeepEqual(restored.Layout().CenterTree, v.Layout().CenterTree) {
		t.Fatalf("center tree did not round-trip: %+v", restored.Layout().CenterTree)
	}

	v.Move("files", DockCenter)
	if v.where("files") != int(DockCenter) || !slices.Contains(v.Layout().Center, "files") {
		t.Fatal("a side panel did not move to the center")
	}
	for _, id := range []string{"main.go", "app.go", "files"} {
		v.SetVisible(id, false)
	}
	h.Frame()
	if !shown(h, "welcome") {
		t.Fatal("the center view did not come back")
	}
}

// Dragging a side tab into an empty center opens it as a document.
func TestDockDragIntoEmptyCenter(t *testing.T) {
	v := documentDock()
	v.SetVisible("main.go", false)
	v.SetVisible("app.go", false)
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(900)).H(el.Dp(600)).Child(v.Render(cx)) }), 900, 1)
	x, y := dockPoint(v.tabRects["files"])
	c := v.centerRect
	dockPress(h, x, y)
	dockMove(h, c.x+c.w/2, c.y+c.h/2)
	if v.drag.drop.kind != 3 || v.drag.drop.side != DockCenter {
		t.Fatalf("center target %+v", v.drag.drop)
	}
	dockRelease(h, c.x+c.w/2, c.y+c.h/2)
	if v.where("files") != int(DockCenter) {
		t.Fatal("the tab did not become a document")
	}
}

// A panel detaches into a window of its own from its menu or by dragging
// its tab out of the dock, and reattaches where it was.
func TestDockDetachAndReattach(t *testing.T) {
	v := documentDock()
	var got []string
	var back func()
	v.OnDetach(func(p DockPanel, reattach func()) { got = append(got, p.ID); back = reattach })
	calls := 0
	v.OnLayoutChange(func(DockLayout) { calls++ })
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(900)).H(el.Dp(600)).Child(v.Render(cx)) }), 900, 1)
	h.Frame()

	click(t, h, locale.Current().Name(locale.Current().More, "Files"))
	h.Frame()
	if !shown(h, locale.Current().DockDetach) || !shown(h, locale.Current().DockCenter) {
		t.Fatal("menu lacks detach or move-to-center")
	}
	click(t, h, locale.Current().DockDetach)
	h.Frame()
	if !slices.Equal(got, []string{"files"}) || v.Visible("files") || !slices.Equal(v.Detached(), []string{"files"}) || calls != 1 {
		t.Fatalf("detach: got %v visible %v detached %v calls %d", got, v.Visible("files"), v.Detached(), calls)
	}
	back()
	h.Frame()
	if !v.Visible("files") || v.where("files") != int(DockLeft) || len(v.Detached()) != 0 || calls != 2 {
		t.Fatal("reattach did not return the panel to its region")
	}

	x, y := dockPoint(v.tabRects["main.go"])
	dockPress(h, x, y)
	dockMove(h, 950, 650)
	dockRelease(h, 950, 650)
	if got[len(got)-1] != "main.go" || v.Visible("main.go") {
		t.Fatal("dragging a tab out of the dock did not detach it")
	}

	l := v.Layout()
	fresh := documentDock()
	if !fresh.SetLayout(l) || !fresh.Visible("main.go") {
		t.Fatal("a restored layout kept a panel detached without a window")
	}
}
