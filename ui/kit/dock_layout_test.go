package kit

import (
	"encoding/json"
	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
	"math"
	"reflect"
	"testing"
)

func TestDockLayoutVersionsDuplicatesAndOwnership(t *testing.T) {
	v := Dock(text("Center")).Panel(DockPanel{ID: "a", Title: "A"}, DockLeft).Panel(DockPanel{ID: "b", Title: "B"}, DockRight)
	v.SetVisible("a", false)
	before := v.Layout()
	for _, bad := range []DockLayout{
		{Version: 3}, {Version: -1}, {Version: 1, Left: []string{"a"}, Right: []string{"a"}},
		{Version: 1, LeftSize: float32(math.NaN())}, {Version: 1, BottomSize: float32(math.Inf(1))},
	} {
		if v.SetLayout(bad) || !reflect.DeepEqual(before, v.Layout()) {
			t.Fatal("invalid restore was not atomic")
		}
	}
	data, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DockLayout
	if err = json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	decoded.Version = 0
	if !v.SetLayout(decoded) || v.Layout().Version != 2 || v.Visible("a") {
		t.Fatal("legacy migration")
	}
	decoded.Left[0] = "mutated"
	if v.Layout().Left[0] != "a" {
		t.Fatal("restore input aliased")
	}
	out := v.Layout()
	out.Right[0] = "mutated"
	if v.Layout().Right[0] != "b" {
		t.Fatal("snapshot aliased")
	}
	if !v.SetLayout(DockLayout{Version: 1, Hidden: []string{"a", "unknown"}}) || v.where("a") != int(DockLeft) {
		t.Fatal("hidden-only restore lost region")
	}
	if v.Visible("unknown") {
		t.Fatal("unknown panel visible")
	}
}

func TestDockPanelRegistrationAndInvalidMove(t *testing.T) {
	v := Dock(nil).Panel(DockPanel{ID: "a", Title: "A"}, DockLeft)
	v.Panel(DockPanel{ID: "a", Title: "Updated"}, DockRight)
	if len(v.layout.Left) != 1 || len(v.layout.Right) != 0 || v.panels["a"].Title != "Updated" {
		t.Fatal("duplicate registration")
	}
	v.Panel(DockPanel{ID: "", Title: "Empty"}, DockLeft).Panel(DockPanel{ID: "bad"}, DockSide(99))
	v.Move("a", DockSide(99))
	if len(v.panels) != 1 || v.where("a") != int(DockLeft) {
		t.Fatal("invalid panel or destination")
	}
}

func TestDockKeyboardResizeAndDisabled(t *testing.T) {
	calls := 0
	v := Dock(text("center")).Panel(DockPanel{ID: "a", Title: "A", View: text("panel")}, DockLeft).OnLayoutChange(func(DockLayout) { calls++ })
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return v.Render(cx) }), 640, 1)
	click(t, h, "调整大小")
	h.Key(key.NameRightArrow, 0)
	if v.Layout().LeftSize != 250 || calls != 1 {
		t.Fatalf("keyboard resize %v %d", v.Layout().LeftSize, calls)
	}
	v.SetDisabled(true)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	if v.Layout().LeftSize != 250 || calls != 1 {
		t.Fatal("disabled resize")
	}
}

func TestDockCanceledResizeRestoresSize(t *testing.T) {
	calls := 0
	v := Dock(text("center")).Panel(DockPanel{ID: "a", Title: "A"}, DockLeft).OnLayoutChange(func(DockLayout) { calls++ })
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return v.Render(cx) }), 640, 1)
	x, y := center(bounds(h, "调整大小"))
	start := v.Layout().LeftSize
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, y)}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x+50, y)})
	h.Frame()
	if v.Layout().LeftSize == start {
		t.Fatal("drag never resized")
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
	h.Frame()
	if v.Layout().LeftSize != start || calls != 0 {
		t.Fatal("cancel committed resize")
	}
}
