package kit

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func dockInputFactory(s DockPanelState) (DockPanel, error) {
	var value string
	if len(s.State) != 0 {
		if err := json.Unmarshal(s.State, &value); err != nil {
			return DockPanel{}, err
		}
	}
	input := Input(s.Title)
	input.SetValue(value)
	return DockPanel{ID: s.ID, Kind: s.Kind, Title: s.Title, View: input,
		SaveState: func() (json.RawMessage, error) { return json.Marshal(input.Value()) }}, nil
}

func dockWithSavedInputs(t *testing.T) *DockView {
	t.Helper()
	v := Dock(text("welcome"))
	if err := v.RegisterPanel("input", dockInputFactory); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		id, title, value string
		side             DockSide
	}{{"a", "Search", "世界", DockLeft}, {"b", "Notes", "draft", DockLeft}, {"c", "Document", "package main", DockCenter}} {
		data, _ := json.Marshal(item.value)
		panel, err := dockInputFactory(DockPanelState{ID: item.id, Kind: "input", Title: item.title, State: data})
		if err != nil {
			t.Fatal(err)
		}
		v.Panel(panel, item.side)
	}
	v.Split("b", "a", DockPlacementBottom)
	return v
}

func TestDockStateJSONRebuildsPanelInstances(t *testing.T) {
	source := dockWithSavedInputs(t)
	source.SetVisible("b", false)
	source.Zoom("c")
	state, err := source.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DockState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	view := Dock(text("empty"))
	if err := view.RegisterPanel("input", func(s DockPanelState) (DockPanel, error) {
		p, err := dockInputFactory(s)
		p.ID, p.Kind, p.Title = "", "", ""
		return p, err
	}); err != nil {
		t.Fatal(err)
	}
	callbacks := 0
	view.OnLayoutChange(func(DockLayout) { callbacks++ })
	if err := view.Restore(decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(source.Layout(), view.Layout()) || view.Visible("b") || !view.documents || callbacks != 0 {
		t.Fatal("restore lost arrangement or emitted a callback")
	}
	for id, old := range source.panels {
		got := view.panels[id]
		if got.ID != id || got.Kind != old.Kind || got.Title != old.Title || got.View == old.View || got.View.(*InputView).Value() != old.View.(*InputView).Value() {
			t.Fatalf("panel %s was not reconstructed", id)
		}
	}
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(900)).H(el.Dp(600)).Child(view.Render(cx))
	}), 900, 1)
	if !shown(h, "Document") || shown(h, "Search") {
		t.Fatal("zoomed restore did not render")
	}
	view.Zoom("")
	view.SetVisible("b", true)
	h.Frame()
	if !shown(h, "Search") || !shown(h, "Notes") {
		t.Fatal("hidden split panel did not reopen")
	}
	view.panels["a"].View.(*InputView).SetValue("edited")
	resaved, err := view.Snapshot()
	if err != nil || string(resaved.Panels[0].State) != `"edited"` {
		t.Fatal("restored panel cannot save new values", err)
	}
	decoded.Layout.LeftTree.First.Panels[0] = "changed"
	decoded.Panels[0].State[1] = 'x'
	if view.Layout().Left[0] != "a" || source.panels["a"].View.(*InputView).Value() != "世界" {
		t.Fatal("restore input or source view was mutated")
	}
}

func TestDockStateRejectsInvalidInputBeforeFactories(t *testing.T) {
	view := dockWithSavedInputs(t)
	state, err := view.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	baseline, _ := json.Marshal(state)
	calls := 0
	view.factories["input"] = func(s DockPanelState) (DockPanel, error) { calls++; return dockInputFactory(s) }
	before := view.Layout()
	oldView := view.panels["a"].View
	for name, change := range map[string]func(*DockState){
		"version":             func(s *DockState) { s.Version++ },
		"empty ID":            func(s *DockState) { s.Panels[0].ID = "" },
		"duplicate":           func(s *DockState) { s.Panels = append(s.Panels, s.Panels[0]) },
		"bad JSON":            func(s *DockState) { s.Panels[2].State = json.RawMessage("{") },
		"unknown kind":        func(s *DockState) { s.Panels[2].Kind = "missing" },
		"missing kind":        func(s *DockState) { s.Panels[0].Kind = "" },
		"unknown tree ID":     func(s *DockState) { s.Layout.LeftTree.First.Panels[0] = "missing" },
		"unknown hidden ID":   func(s *DockState) { s.Layout.Hidden = []string{"missing"} },
		"unknown active ID":   func(s *DockState) { s.Layout.CenterActive = "missing" },
		"unplaced":            func(s *DockState) { s.Layout.CenterTree, s.Layout.Center, s.Layout.CenterActive = nil, nil, "" },
		"duplicate placement": func(s *DockState) { s.Layout.Right = []string{"a"} },
		"invalid ratio":       func(s *DockState) { s.Layout.LeftTree.Ratio = 2 },
		"cycle":               func(s *DockState) { s.Layout.LeftTree.First = s.Layout.LeftTree },
	} {
		t.Run(name, func(t *testing.T) {
			var bad DockState
			if err := json.Unmarshal(baseline, &bad); err != nil {
				t.Fatal(err)
			}
			change(&bad)
			if err := view.Restore(bad); err == nil {
				t.Fatal("invalid state accepted")
			}
			if calls != 0 || !reflect.DeepEqual(before, view.Layout()) || view.panels["a"].View != oldView {
				t.Fatal("invalid state invoked factory or modified live Dock")
			}
		})
	}
}

func TestDockStateFactoryFailureIsAtomicAndOwnsBytes(t *testing.T) {
	view := dockWithSavedInputs(t)
	state, err := view.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	before := view.Layout()
	oldView := view.panels["a"].View
	want := errors.New("cannot restore document")
	for _, mode := range []string{"error", "wrong ID", "wrong kind", "no view"} {
		view.factories["input"] = func(s DockPanelState) (DockPanel, error) {
			p, err := dockInputFactory(s)
			s.State[1] = 'x' // A factory may keep or modify its own payload.
			if s.ID == "c" {
				switch mode {
				case "error":
					return DockPanel{}, want
				case "wrong ID":
					p.ID = "other"
				case "wrong kind":
					p.Kind = "other"
				case "no view":
					p.View = nil
				}
			}
			return p, err
		}
		err := view.Restore(state)
		if err == nil || (mode == "error" && !errors.Is(err, want)) {
			t.Fatal("factory failure not returned", mode, err)
		}
		if !reflect.DeepEqual(before, view.Layout()) || view.panels["a"].View != oldView || string(state.Panels[0].State) != `"世界"` {
			t.Fatal("failed factory changed Dock or borrowed input bytes")
		}
	}
}

func TestDockSnapshotValidationAndOwnership(t *testing.T) {
	view := Dock(nil)
	factory := DockPanelFactory(dockInputFactory)
	if view.RegisterPanel("", factory) == nil || view.RegisterPanel("input", nil) == nil {
		t.Fatal("invalid registration")
	}
	if err := view.RegisterPanel("input", factory); err != nil {
		t.Fatal(err)
	}
	if view.RegisterPanel("input", factory) == nil {
		t.Fatal("duplicate registration")
	}
	data := json.RawMessage(`{"value":1}`)
	panel := DockPanel{ID: "a", Kind: "input", SaveState: func() (json.RawMessage, error) { return data, nil }}
	view.Panel(panel, DockLeft)
	state, err := view.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	data[9] = '2'
	if string(state.Panels[0].State) != `{"value":1}` {
		t.Fatal("snapshot borrowed application bytes")
	}
	for _, mode := range []string{"kind", "JSON", "error"} {
		p := panel
		switch mode {
		case "kind":
			p.Kind = ""
		case "JSON":
			p.SaveState = func() (json.RawMessage, error) { return json.RawMessage("no"), nil }
		case "error":
			p.SaveState = func() (json.RawMessage, error) { return nil, errors.New("save failed") }
		}
		view.Panel(p, DockLeft)
		if _, err := view.Snapshot(); err == nil || !strings.Contains(err.Error(), `"a"`) {
			t.Fatal("invalid save not identified", mode, err)
		}
	}
}

func TestDockRestoreReusesStaticViewsAndInvalidatesOldWindows(t *testing.T) {
	view := dockWithSavedInputs(t)
	static := Input("Static input")
	view.Panel(DockPanel{ID: "static", Title: "Static", View: static}, DockRight)
	var back func()
	view.OnDetach(func(_ DockPanel, reattach func()) { back = reattach })
	view.Detach("a")
	oldBack := back
	state, err := view.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	view.Panel(DockPanel{ID: "extra", View: text("extra")}, DockBottom)
	skin := &DockSkin{}
	view.Skin(skin)
	if err := view.Restore(state); err != nil {
		t.Fatal(err)
	}
	if view.panels["static"].View != static || len(view.panels) != 4 || !view.Visible("a") || len(view.Detached()) != 0 || view.skin != skin {
		t.Fatal("restore did not reuse static view, remove extra or return detached panel")
	}
	view.Detach("a")
	oldBack()
	if len(view.Detached()) != 1 {
		t.Fatal("old window close affected new panel instance")
	}
	back()
	if len(view.Detached()) != 0 {
		t.Fatal("new window close did not reattach panel")
	}
	view.SetDisabled(true)
	if err := view.Restore(state); err != nil || !view.disabled {
		t.Fatal("restore lost disabled state", err)
	}
}
