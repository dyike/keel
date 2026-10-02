package kit

import (
	"encoding/json"
	"reflect"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
)

func nestedDock() *DockView {
	return Dock(text("center")).Panel(DockPanel{ID: "a", Title: "A", View: Input("edit A")}, DockLeft).
		Panel(DockPanel{ID: "b", Title: "B", View: text("body B")}, DockLeft).
		Panel(DockPanel{ID: "c", Title: "C", View: text("body C")}, DockRight)
}
func TestDockNestedSplitRestoreAndIsolation(t *testing.T) {
	v := nestedDock()
	if !v.Split("b", "a", DockPlacementBottom) || !v.Split("c", "a", DockPlacementRight) {
		t.Fatal("split")
	}
	l := v.Layout()
	if l.LeftTree.Axis != DockAxisVertical || l.LeftTree.First.Axis != DockAxisHorizontal || len(l.Right) != 0 {
		t.Fatalf("nested layout %+v", l)
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatal(err)
	}
	var saved DockLayout
	if json.Unmarshal(data, &saved) != nil {
		t.Fatal("decode")
	}
	other := nestedDock()
	if !other.SetLayout(saved) || !reflect.DeepEqual(l, other.Layout()) {
		t.Fatalf("round trip\n%+v\n%+v", l, other.Layout())
	}
	saved.LeftTree.First.First.Panels[0] = "changed"
	l.LeftTree.First.Second.Active = "changed"
	if other.Layout().LeftTree.First.First.Panels[0] != "a" || other.Layout().LeftTree.First.Second.Active != "c" {
		t.Fatal("tree alias")
	}
	before := v.Layout()
	for _, args := range []struct {
		id, target string
		p          DockPlacement
	}{{"a", "a", 0}, {"missing", "a", 0}, {"a", "b", 99}} {
		if v.Split(args.id, args.target, args.p) || !reflect.DeepEqual(before, v.Layout()) {
			t.Fatal("invalid split changed layout")
		}
	}
	v.Move("c", DockBottom)
	if v.Layout().LeftTree.First.First != nil || v.Layout().BottomTree.Active != "c" {
		t.Fatal("empty split not collapsed")
	}
	v.SetVisible("b", false)
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().H(el.Dp(600)).Child(v.Render(cx)) }), 800, 1)
	if shown(h, "body B") || !shown(h, "A") {
		t.Fatal("hidden group")
	}
	v.SetVisible("b", true)
	h.Frame()
	if !shown(h, "body B") {
		t.Fatal("reopen group")
	}
}
func TestDockTreeRejectsInvalidStructureAtomically(t *testing.T) {
	v := nestedDock()
	before := v.Layout()
	cyclic := &DockNode{Ratio: .5, Second: &DockNode{Panels: []string{"b"}}}
	cyclic.First = cyclic
	for _, tree := range []*DockNode{
		cyclic,
		{First: &DockNode{Panels: []string{"a"}}, Ratio: .5},
		{First: &DockNode{Panels: []string{"a"}}, Second: &DockNode{Panels: []string{"a"}}, Ratio: .5},
		{First: &DockNode{Panels: []string{"a"}}, Second: &DockNode{Panels: []string{"b"}}, Ratio: 1.1},
		{First: &DockNode{Panels: []string{"a"}}, Second: &DockNode{Panels: []string{"b"}}, Ratio: .5, Axis: 99},
		{Panels: []string{"a"}, Ratio: .5},
	} {
		l := v.Layout()
		l.LeftTree = tree
		if v.SetLayout(l) || !reflect.DeepEqual(before, v.Layout()) {
			t.Fatal("invalid tree changed layout")
		}
	}
	// Unknown-only branches collapse, and unmentioned registered panels survive.
	l := DockLayout{Version: 2, LeftTree: &DockNode{Ratio: .5, First: &DockNode{Panels: []string{"a"}}, Second: &DockNode{Panels: []string{"gone"}}}}
	if !v.SetLayout(l) || v.Layout().LeftTree.First != nil || !reflect.DeepEqual(v.Layout().Left, []string{"a", "b"}) {
		t.Fatal("unknown pruning")
	}
}
func TestDockNestedGeometryAndKeyboardResize(t *testing.T) {
	v := nestedDock()
	v.Split("b", "a", DockPlacementBottom)
	calls := 0
	v.OnLayoutChange(func(DockLayout) { calls++ })
	h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { return el.Div().H(el.Dp(600)).Child(v.Render(ctx)) }), 800, 1)
	a, b := bounds(h, "A"), bounds(h, "body B")
	if a.Empty() || b.Empty() || a.Max.Y >= b.Min.Y {
		t.Fatalf("split geometry A=%v B=%v", a, b)
	}
	// The nested separator precedes the outer region separator.
	click(t, h, "调整大小")
	h.Key(key.NameDownArrow, 0)
	if v.layout.LeftTree.Ratio != .55 || calls != 1 {
		t.Fatalf("keyboard ratio=%v calls=%d", v.layout.LeftTree.Ratio, calls)
	}
	v.SetDisabled(true)
	h.Frame()
	h.Key(key.NameDownArrow, 0)
	if v.layout.LeftTree.Ratio != .55 || calls != 1 {
		t.Fatal("disabled split")
	}
}

func TestDockMenuSplitsAndRejoinsGroups(t *testing.T) {
	v := nestedDock()
	calls := 0
	v.OnLayoutChange(func(DockLayout) { calls++ })
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().H(el.Dp(600)).Child(v.Render(cx)) }), 800, 1)
	click(t, h, "更多 A")
	click(t, h, "向下拆分")
	h.Frame()
	if v.layout.LeftTree.First == nil || calls != 1 || !shown(h, "body B") {
		t.Fatal("menu split")
	}
	v.Move("a", DockLeft)
	h.Frame()
	if v.layout.LeftTree.First != nil || v.layout.LeftTree.Active != "a" || shown(h, "body B") {
		t.Fatal("merge activates moved tab")
	}
}

func TestDockNestedDragCancelAndAncestorDisable(t *testing.T) {
	for _, cancel := range []bool{true, false} {
		v := nestedDock()
		v.Split("b", "a", DockPlacementBottom)
		calls := 0
		v.OnLayoutChange(func(DockLayout) { calls++ })
		disabled := false
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().H(el.Dp(600)).Disabled(disabled).Child(v.Render(cx)) }), 800, 1)
		x, y := center(bounds(h, "调整大小"))
		h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, y)}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
		h.Frame()
		h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y+60)})
		h.Frame()
		if v.layout.LeftTree.Ratio <= .5 {
			t.Fatal("split drag did not resize")
		}
		if cancel {
			h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
		} else {
			disabled = true
		}
		h.Frame()
		h.Frame()
		if v.layout.LeftTree.Ratio != .5 || calls != 0 {
			t.Fatal("canceled split committed")
		}
	}
}
