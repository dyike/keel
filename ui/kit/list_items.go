package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"math"
	"slices"
	"strconv"
)

// ListItem gives an entry a stable identity, label and interaction state.
type ListItem struct {
	ID, Label string
	Disabled  bool
}

// SetEntries copies entries and preserves selection by ID. Empty/duplicate IDs
// panic before changing the list. Removed IDs are discarded from selection.
func (v *ListView) SetEntries(entries ...ListItem) {
	index := make(map[string]int, len(entries))
	for i, item := range entries {
		if item.ID == "" {
			panic("kit.List: empty item ID")
		}
		if _, ok := index[item.ID]; ok {
			panic("kit.List: duplicate item ID")
		}
		index[item.ID] = i
	}
	selected := ""
	if v.selected >= 0 && v.selected < len(v.keys) {
		selected = v.keys[v.selected]
	}
	v.items = make([]string, len(entries))
	v.keys = make([]string, len(entries))
	v.itemDisabled = make([]bool, len(entries))
	for i, item := range entries {
		v.items[i], v.keys[i], v.itemDisabled[i] = item.Label, item.ID, item.Disabled
	}
	v.selected = -1
	if i, ok := index[selected]; ok {
		v.selected = i
	}
	for id := range v.selection {
		if _, ok := index[id]; !ok {
			delete(v.selection, id)
		}
	}
	if _, ok := index[v.anchor]; !ok {
		v.anchor = ""
	}
	v.list.SetCount(len(entries))
}
func (v *ListView) Entries() []ListItem {
	entries := make([]ListItem, len(v.items))
	for i, label := range v.items {
		entries[i] = ListItem{v.keys[i], label, v.itemDisabled[i]}
	}
	return entries
}
func (v *ListView) SetItemDisabled(index int, on bool) {
	if index >= 0 && index < len(v.items) {
		v.itemDisabled[index] = on
	}
}
func (v *ListView) MultiSelect() *ListView {
	v.multi = true
	v.SetSelectedValues([]int{v.selected})
	return v
}
func (v *ListView) OnSelectionChange(fn func([]int)) *ListView { v.onSelection = fn; return v }
func (v *ListView) SelectedValues() []int {
	if !v.multi {
		if v.selected >= 0 {
			return []int{v.selected}
		}
		return nil
	}
	var values []int
	for i, id := range v.keys {
		if v.selection[id] {
			values = append(values, i)
		}
	}
	return values
}
func (v *ListView) SetSelectedValues(values []int) {
	v.reveal = true
	v.selection = make(map[string]bool)
	v.selected = -1
	v.anchor = ""
	for _, i := range values {
		if i >= 0 && i < len(v.items) {
			v.selection[v.keys[i]] = true
			v.selected = i
			v.anchor = v.keys[i]
			if !v.multi {
				break
			}
		}
	}
}
func (v *ListView) selectedItem(i int) bool {
	if v.multi {
		return v.selection[v.keys[i]]
	}
	return i == v.selected
}
func (v *ListView) selectItem(cx *el.Context, i int, mods key.Modifiers) {
	if i < 0 || i >= len(v.items) || v.itemDisabled[i] {
		return
	}
	before := v.SelectedValues()
	if v.multi {
		add := mods.Contain(key.ModShortcut)
		a := slices.Index(v.keys, v.anchor)
		if mods.Contain(key.ModShift) && a >= 0 {
			if !add {
				v.selection = make(map[string]bool)
			}
			for p := min(a, i); p <= max(a, i); p++ {
				if !v.itemDisabled[p] {
					v.selection[v.keys[p]] = true
				}
			}
		} else {
			if !add {
				v.selection = make(map[string]bool)
			}
			id := v.keys[i]
			if add && v.selection[id] {
				delete(v.selection, id)
			} else {
				v.selection[id] = true
			}
			v.anchor = id
		}
	}
	v.choose(cx, i)
	if !slices.Equal(before, v.SelectedValues()) && v.onSelection != nil {
		v.onSelection(v.SelectedValues())
	}
}

// Reorderable enables drag reordering. The callback runs after the list owns
// the new order and receives the old and new indexes. Programmatic Move is silent.
func (v *ListView) Reorderable(fn func(from, to int)) *ListView {
	v.reorderable = true
	v.onReorder = fn
	return v
}
func (v *ListView) Move(from, to int) {
	if from < 0 || from >= len(v.items) || to < 0 || to >= len(v.items) || from == to {
		return
	}
	entries := v.Entries()
	item := entries[from]
	entries = slices.Delete(entries, from, from+1)
	entries = slices.Insert(entries, to, item)
	v.SetEntries(entries...)
}
func (v *ListView) dragItem(i int, e el.DragEvent) {
	switch e.Kind {
	case el.DragStart:
		v.dragID = v.keys[i]
		v.dragY = e.Y
	case el.DragEnd:
		from := slices.Index(v.keys, v.dragID)
		v.dragID = ""
		if e.Canceled || from < 0 || v.disabled || v.itemDisabled[from] {
			return
		}
		to := max(0, min(len(v.items)-1, from+int(math.Round(float64((e.Y-v.dragY)/32)))))
		if to == from {
			return
		}
		v.Move(from, to)
		if v.onReorder != nil {
			v.onReorder(from, to)
		}
	}
}

func listEnabledKey(name string, current, count int, disabled func(int) bool) (int, bool) {
	i, ok := listKeys(name, current, count, 10)
	if !ok || count == 0 {
		return current, ok
	}
	direction := 1
	switch key.Name(name) {
	case key.NameUpArrow, key.NamePageUp, key.NameEnd:
		direction = -1
	}
	if current < 0 && key.Name(name) == key.NameUpArrow {
		i = count - 1
	}
	target := i
	for i >= 0 && i < count {
		if !disabled(i) {
			return i, true
		}
		i += direction
	}
	if key.Name(name) == key.NamePageDown {
		for i := target - 1; i > current; i-- {
			if !disabled(i) {
				return i, true
			}
		}
	}
	if key.Name(name) == key.NamePageUp {
		for i := target + 1; i < current; i++ {
			if !disabled(i) {
				return i, true
			}
		}
	}
	return current, true
}
func indexListItems(items []string) []ListItem {
	entries := make([]ListItem, len(items))
	for i, label := range items {
		entries[i] = ListItem{ID: strconv.Itoa(i), Label: label}
	}
	return entries
}
