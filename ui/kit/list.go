package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
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
	items              []string
	selected           int
	disabled           bool
	list               *VirtualListView
	onChange, onActive func(int)
}

func List(items ...string) *ListView {
	v := &ListView{items: slices.Clone(items), selected: -1}
	v.list = VirtualList(len(items), 32, v.row)
	return v
}

// Height sets the viewport height in dp, 320 by default; Fill grows instead.
func (v *ListView) Height(dp float32) *ListView             { v.list.Height(dp); return v }
func (v *ListView) Fill() *ListView                         { v.list.Fill(); return v }
func (v *ListView) OnChange(fn func(index int)) *ListView   { v.onChange = fn; return v }
func (v *ListView) OnActivate(fn func(index int)) *ListView { v.onActive = fn; return v }
func (v *ListView) Value() int                              { return v.selected }
func (v *ListView) SetDisabled(on bool)                     { v.disabled = on }
func (v *ListView) Items() []string                         { return slices.Clone(v.items) }

// SetValue selects item i (-1 clears) without calling OnChange.
func (v *ListView) SetValue(i int) {
	if i < -1 || i >= len(v.items) {
		i = -1
	}
	v.selected = i
}

// SetItems replaces the items and clears the selection if it no longer exists.
func (v *ListView) SetItems(items ...string) {
	v.items = slices.Clone(items)
	v.list.SetCount(len(items))
	if v.selected >= len(items) {
		v.selected = -1
	}
}

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
	if i >= 0 && v.onActive != nil {
		v.onActive(i)
	}
}

func (v *ListView) row(cx *el.Context, i int) el.Element {
	on := i == v.selected
	r := el.Div().Role("option").Name(v.items[i]).Selected(on).Row().Items(el.Center).Px(10).Rounded(4).Mx(4)
	if on {
		r.Bg(theme.Highlight).TextColor(theme.PrimaryText)
	}
	if !v.disabled {
		r.CursorPointer().
			OnClick(func() { v.choose(cx, i); cx.Focus(autoID("list", v)) }).
			OnDoubleClick(func() { v.activate(i) })
		if !on {
			r.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	return r.Child(el.Text(v.items[i]).MaxLines(1).Grow())
}

func (v *ListView) Render(cx *el.Context) el.Element {
	// The list, not each row, takes focus: rows scroll out of existence, so
	// keyboard focus could not stay on one.
	return el.Div().ID(autoID("list", v)).Role("listbox").Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() }).
		Rounded(6).Border(1, theme.Border).Bg(theme.Surface).Py(4).Items(el.Stretch).
		Focusable(true).FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			if key.Name(e.Name) == key.NameReturn {
				if e.State == el.KeyPress {
					v.activate(v.selected)
				}
				return true
			}
			i, ok := listKeys(e.Name, v.selected, len(v.items), 10)
			if ok && e.State == el.KeyPress && len(v.items) > 0 {
				v.choose(cx, i)
			}
			return ok
		}).
		Child(v.list.Render(cx))
}
