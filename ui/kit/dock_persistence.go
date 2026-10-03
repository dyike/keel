package kit

import (
	"encoding/json"
	"fmt"
	"slices"
)

// DockPanelState describes one instance. Kind selects the factory; State's
// schema and migrations belong to the application. ID is stable across loads.
type DockPanelState struct {
	ID    string          `json:"id"`
	Kind  string          `json:"kind,omitempty"`
	Title string          `json:"title"`
	State json.RawMessage `json:"state,omitempty"`
}

// DockState persists both the arrangement and all panels, including hidden
// and detached ones. Version is currently 1, independent of DockLayout.Version.
type DockState struct {
	Version int              `json:"version"`
	Layout  DockLayout       `json:"layout"`
	Panels  []DockPanelState `json:"panels"`
}

// DockPanelFactory creates a fresh view from one saved instance, on the UI
// thread. Return its SaveState callback to support subsequent snapshots.
// An empty returned ID/Kind/Title inherits the saved value. ID and Kind must
// otherwise match. Avoid external side effects: Dock can roll back its own
// state on failure, but cannot undo work performed by application factories.
type DockPanelFactory func(DockPanelState) (DockPanel, error)

// RegisterPanel registers a type for this Dock. Empty kinds, nil factories and
// duplicate kinds return errors. It does not create or display any panels.
func (v *DockView) RegisterPanel(kind string, factory DockPanelFactory) error {
	if kind == "" || factory == nil {
		return fmt.Errorf("dock: panel kind and factory are required")
	}
	if v.factories[kind] != nil {
		return fmt.Errorf("dock: panel kind %q already registered", kind)
	}
	if v.factories == nil {
		v.factories = map[string]DockPanelFactory{}
	}
	v.factories[kind] = factory
	return nil
}

// Snapshot captures panel data in ID order and copies the layout and JSON.
// Call on the UI thread. A panel with SaveState must have a Kind; panels with
// no Kind can only be restored by reusing existing stateless views with that ID.
// This does not invoke OnLayoutChange.
func (v *DockView) Snapshot() (DockState, error) {
	state := DockState{Version: 1, Layout: v.Layout()}
	ids := make([]string, 0, len(v.panels))
	for id := range v.panels {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		p := v.panels[id]
		entry := DockPanelState{ID: id, Kind: p.Kind, Title: p.Title}
		if p.SaveState != nil {
			if p.Kind == "" {
				return DockState{}, fmt.Errorf("dock: panel %q has SaveState but no kind", id)
			}
			data, err := p.SaveState()
			if err != nil {
				return DockState{}, fmt.Errorf("dock: save panel %q: %w", id, err)
			}
			if len(data) > 0 && !json.Valid(data) {
				return DockState{}, fmt.Errorf("dock: panel %q returned invalid JSON", id)
			}
			entry.State = slices.Clone(data)
		}
		state.Panels = append(state.Panels, entry)
	}
	return state, nil
}

// Restore replaces the panel collection and arrangement without emitting
// OnLayoutChange. It validates the complete manifest and layout before calling
// factories, then commits only when all succeed. Panels absent from state are
// removed. Unknown kinds, unknown layout IDs and unplaced panels are errors.
// Detached panels return to the Dock; application windows are not opened or
// closed. The center view, skin, registrations and callbacks are retained.
func (v *DockView) Restore(state DockState) error {
	if state.Version != 1 {
		return fmt.Errorf("dock: unsupported state version %d", state.Version)
	}
	next := Dock(v.center)
	entries := slices.Clone(state.Panels)
	factories := make([]DockPanelFactory, len(entries))
	for i, entry := range entries {
		if entry.ID == "" {
			return fmt.Errorf("dock: empty panel ID")
		}
		if _, exists := next.panels[entry.ID]; exists {
			return fmt.Errorf("dock: duplicate panel ID %q", entry.ID)
		}
		if len(entry.State) > 0 && !json.Valid(entry.State) {
			return fmt.Errorf("dock: panel %q has invalid JSON", entry.ID)
		}
		entries[i].State = slices.Clone(entry.State)
		if entry.Kind == "" {
			p, ok := v.panels[entry.ID]
			if !ok || p.Kind != "" || p.SaveState != nil || len(entry.State) > 0 {
				return fmt.Errorf("dock: panel %q needs a registered kind", entry.ID)
			}
			p.Title = entry.Title
			next.panels[entry.ID] = p
		} else {
			factories[i] = v.factories[entry.Kind]
			if factories[i] == nil {
				return fmt.Errorf("dock: unknown panel kind %q for %q", entry.Kind, entry.ID)
			}
			next.panels[entry.ID] = DockPanel{ID: entry.ID, Kind: entry.Kind, Title: entry.Title}
		}
	}
	if err := next.validateStateLayout(state.Layout); err != nil {
		return err
	}
	if !next.SetLayout(state.Layout) {
		return fmt.Errorf("dock: invalid saved layout")
	}
	for _, entry := range entries {
		if next.where(entry.ID) < 0 {
			return fmt.Errorf("dock: panel %q has no layout position", entry.ID)
		}
	}
	for i, entry := range entries {
		if factories[i] == nil {
			continue
		}
		panel, err := factories[i](entry)
		if err != nil {
			return fmt.Errorf("dock: restore panel %q: %w", entry.ID, err)
		}
		if panel.ID == "" {
			panel.ID = entry.ID
		}
		if panel.Kind == "" {
			panel.Kind = entry.Kind
		}
		if panel.Title == "" {
			panel.Title = entry.Title
		}
		if panel.ID != entry.ID || panel.Kind != entry.Kind || panel.View == nil {
			return fmt.Errorf("dock: factory returned invalid panel %q", entry.ID)
		}
		next.panels[entry.ID] = panel
	}
	// Replace transient interaction/geometry state with that of a fresh Dock.
	// A prior detached window's close callback must not affect new instances.
	next.factories, next.skin = v.factories, v.skin
	next.onDetach, next.onLayout, next.disabled = v.onDetach, v.onLayout, v.disabled
	next.documents = v.documents || len(next.layout.Center) > 0
	next.layoutGeneration = v.layoutGeneration + 1
	*v = *next
	return nil
}

// SetLayout intentionally drops unknown panels; full state restores must
// instead catch a missing manifest entry before application factories run.
func (v *DockView) validateStateLayout(l DockLayout) error {
	check := func(ids ...string) error {
		for _, id := range ids {
			if _, ok := v.panels[id]; !ok {
				return fmt.Errorf("dock: layout refers to unknown panel %q", id)
			}
		}
		return nil
	}
	for _, ids := range [][]string{l.Left, l.Right, l.Bottom, l.Center, l.Hidden, l.Detached} {
		if err := check(ids...); err != nil {
			return err
		}
	}
	for _, id := range []string{l.LeftActive, l.RightActive, l.BottomActive, l.CenterActive, l.Zoomed} {
		if id != "" {
			if err := check(id); err != nil {
				return err
			}
		}
	}
	seen := map[*DockNode]bool{}
	var walk func(*DockNode, int) error
	walk = func(n *DockNode, depth int) error {
		if n == nil {
			return nil
		}
		if depth > 32 || seen[n] {
			return fmt.Errorf("dock: invalid saved layout tree")
		}
		seen[n] = true
		if err := check(n.Panels...); err != nil {
			return err
		}
		if n.Active != "" {
			if err := check(n.Active); err != nil {
				return err
			}
		}
		if err := walk(n.First, depth+1); err != nil {
			return err
		}
		return walk(n.Second, depth+1)
	}
	for _, n := range []*DockNode{l.LeftTree, l.RightTree, l.BottomTree, l.CenterTree} {
		if err := walk(n, 0); err != nil {
			return err
		}
	}
	return nil
}
