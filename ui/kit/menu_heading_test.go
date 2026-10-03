package kit

import (
	"fmt"
	"gioui.org/io/key"
	"testing"
)

func TestMenuHeadingsAreNotActions(t *testing.T) {
	chosen := ""
	m := Menu().Label("").Label("Alpha heading").Item("First", "", func() { chosen = "first" }).Label("Between").Item("Alpha action", "", func() { chosen = "alpha" }).Label("Tail")
	m.Trigger(Button("Open", m.Toggle))
	h := page(m)
	click(t, h, "Open")
	h.Frame()
	n, ok := node(h, "Alpha heading")
	if !ok || n.Desc.Description != "heading" {
		t.Fatal("heading semantics", n.Desc)
	}
	click(t, h, "Between")
	if !m.Value() || chosen != "" {
		t.Fatal("heading activated")
	}
	h.Key(key.NameHome, 0)
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if chosen != "alpha" || m.Value() {
		t.Fatal("navigation did not skip heading", chosen)
	}
	click(t, h, "Open")
	h.Frame()
	h.Key("A", 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if chosen != "alpha" || m.Value() {
		t.Fatal("typeahead matched heading")
	}
}

func TestMenuHeadingLongListEnd(t *testing.T) {
	selected := -1
	m := Menu()
	for i := 0; i < 30; i++ {
		m.Label(fmt.Sprintf("Section %d", i)).Item(fmt.Sprintf("Action %d", i), "", func() { selected = i })
	}
	m.Label("Trailing heading")
	m.Trigger(Button("Open", m.Toggle))
	h := page(m)
	click(t, h, "Open")
	h.Frame()
	h.Key(key.NameEnd, 0)
	h.Frame()
	h.Frame()
	n, ok := node(h, "Action 29")
	if !ok || n.Desc.Bounds.Min.Y < 0 || n.Desc.Bounds.Max.Y > 300 {
		t.Fatal("End did not reveal last action", n.Desc.Bounds)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if selected != 29 {
		t.Fatal("End targeted heading", selected)
	}
}
