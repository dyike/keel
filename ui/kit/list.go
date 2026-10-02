package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"slices"
)

// listKeys moves a selection index for ↑ ↓ Home End PageUp PageDown; it
// reports the new index and whether the key was one of them.
func listKeys(name string, i, n, page int) (int, bool) {
	switch key.Name(name) {
	case key.NameDownArrow:
		i++
	case key.NameUpArrow:
		i--
	case key.NameHome:
		i = 0
	case key.NameEnd:
		i = n - 1
	case key.NamePageDown:
		i += page
	case key.NamePageUp:
		i -= page
	default:
		return i, false
	}
	return min(max(i, 0), n-1), true
}

// ListView is a list of text items with one selection. Click or use ↑ ↓
// Home End PageUp PageDown to select; double-click or Enter activates.
// Only rows near the viewport are built, so it handles long lists.
type ListView struct {
	keys               []string
	itemDisabled       []bool
	multi, reorderable bool
	reveal             bool
	selection          map[string]bool
	anchor, dragID     string
	dragY              float32
	onSelection        func([]int)
	onReorder          func(int, int)
	items              []string
	selected           int
	disabled, plain    bool
	list               *VirtualListView
	onChange, onActive func(int)
}

func List(items ...string) *ListView {
	v := &ListView{items: slices.Clone(items), selected: -1}
	v.list = VirtualList(len(items), 32, v.row).ItemKey(func(i int) string { return v.keys[i] })
	v.SetEntries(indexListItems(items)...)
	return v
}

// Height sets the viewport height in dp, 320 by default; Fill grows instead.
func (v *ListView) Height(dp float32) *ListView { v.list.Height(dp); return v }
func (v *ListView) Fill() *ListView             { v.list.Fill(); return v }

// Plain drops the frame and background, for a list that sits inside a panel
// or sidebar that already frames it. Focus still shows as an outline.
func (v *ListView) Plain() *ListView                        { v.plain = true; return v }
func (v *ListView) OnChange(fn func(index int)) *ListView   { v.onChange = fn; return v }
func (v *ListView) OnActivate(fn func(index int)) *ListView { v.onActive = fn; return v }
func (v *ListView) Value() int                              { return v.selected }
func (v *ListView) SetDisabled(on bool)                     { v.disabled = on }
func (v *ListView) Items() []string                         { return slices.Clone(v.items) }

// SetValue selects item i (-1 clears) without calling OnChange.
func (v *ListView) SetValue(i int) { v.SetSelectedValues([]int{i}) }

// SetItems replaces labels using index identities. Use SetEntries for stable
// application IDs across insertions and reorders.
func (v *ListView) SetItems(items ...string) { v.SetEntries(indexListItems(items)...) }

func (v *ListView) choose(cx *el.Context, i int) {
	v.list.ScrollTo(cx, i)
	if i == v.selected {
		return
	}
	v.selected = i
	if v.onChange != nil {
		v.onChange(i)
	}
}

func (v *ListView) activate(i int) {
	if i >= 0 && i < len(v.items) && !v.itemDisabled[i] && v.onActive != nil {
		v.onActive(i)
	}
}

func (v *ListView) row(cx *el.Context, i int) el.Element {
	on := v.selectedItem(i)
	r := el.Div().Role("option").Name(v.items[i]).Selected(on).Disabled(v.itemDisabled[i]).Row().Items(el.Center).Px(10).Rounded(theme.RadiusSm).Mx(4)
	if on {
		r.Bg(theme.Highlight).TextColor(theme.PrimaryText)
	}
	if !v.disabled && !v.itemDisabled[i] {
		r.CursorPointer().
			OnClick(func() { v.selectItem(cx, i, cx.ClickModifiers()); cx.Focus(autoID("list", v)) }).
			OnDoubleClick(func() { v.activate(i) })
		if !on {
			r.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	if v.reorderable && !v.disabled && !v.itemDisabled[i] {
		r.OnDrag(func(e el.DragEvent) { v.dragItem(i, e) })
	}
	return r.Child(el.Text(v.items[i]).MaxLines(1).Grow())
}

func (v *ListView) Render(cx *el.Context) el.Element {
	if v.reveal {
		v.list.ScrollTo(cx, v.selected)
		v.reveal = false
	}
	// The list, not each row, takes focus: rows scroll out of existence, so
	// keyboard focus could not stay on one.
	return listFrame(el.Div().ID(autoID("list", v)).Role("listbox").Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() }), v.plain).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) == key.NameReturn {
				if e.State == el.KeyPress {
					v.activate(v.selected)
				}
				return true
			}
			if v.multi && key.Name(e.Name) == "A" && e.Modifiers.Contain(key.ModShortcut) {
				if e.State == el.KeyPress {
					before := v.SelectedValues()
					v.selection = make(map[string]bool)
					for i, id := range v.keys {
						if !v.itemDisabled[i] {
							v.selection[id] = true
						}
					}
					if !slices.Equal(before, v.SelectedValues()) && v.onSelection != nil {
						v.onSelection(v.SelectedValues())
					}
				}
				return true
			}
			i, ok := listEnabledKey(e.Name, v.selected, len(v.items), func(i int) bool { return v.itemDisabled[i] })
			if ok && e.State == el.KeyPress && len(v.items) > 0 {
				v.selectItem(cx, i, e.Modifiers)
			}
			return ok
		}).
		Child(v.list.Render(cx))
}

// listFrame styles the focusable box around a List or Tree. A plain one keeps
// a transparent border so the focus outline has somewhere to show.
func listFrame(d *el.DivEl, plain bool) *el.DivEl {
	if plain {
		return d.Rounded(theme.RadiusMd).Border(1, color.NRGBA{}).Py(4).Items(el.Stretch)
	}
	return d.Rounded(theme.RadiusMd).Border(1, theme.Border).Bg(theme.Surface).Py(4).Items(el.Stretch)
}
