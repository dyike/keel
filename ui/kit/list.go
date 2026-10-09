package kit

import (
	"image/color"
	"slices"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// ListView is a list of text items with one selection. Click or use ↑ ↓
// Home End PageUp PageDown to select; double-click or Enter activates.
// Only rows near the viewport are built, so it handles long lists.
type ListView struct {
	keys                            []string
	itemDisabled                    []bool
	multi, reorderable              bool
	reveal                          bool
	selection                       base.Selection[string]
	typeahead                       base.Typeahead
	dragID                          string
	dragY                           float32
	onSelection                     func([]int)
	onReorder                       func(int, int)
	items                           []string
	selected                        int
	disabled, plain                 bool
	list                            *VirtualListView
	onChange, onActive              func(int)
	entries                         []ListItem
	display                         []listDisplayRow
	positions                       []int
	query                           string
	searchable                      bool
	renderItem                      func(*el.Context, ListItemContext) el.Element
	renderHeader, renderFooter      func(*el.Context, string) el.Element
	onSearch                        func(string)
	onLoadMore                      func()
	loading, hasMore, loadRequested bool
	loadError                       string
	slots                           listSlots
}

func List(items ...string) *ListView {
	v := &ListView{items: slices.Clone(items), selected: -1}
	v.list = VirtualList(len(items), 32, v.row).ItemKey(func(i int) string { row := v.display[i]; return v.keys[row.index] + "/" + string(rune('0'+row.kind)) })
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
	if i >= 0 && i < len(v.positions) && v.positions[i] >= 0 {
		v.list.ScrollTo(cx, v.positions[i])
	}
	if i == v.selected {
		return
	}
	v.selected = i
	if v.onChange != nil {
		v.onChange(i)
	}
}

func (v *ListView) activate(i int) {
	if i >= 0 && i < len(v.items) && !v.itemDisabled[i] && v.positions[i] >= 0 && v.onActive != nil {
		v.onActive(i)
	}
}

func (v *ListView) row(cx *el.Context, p int) el.Element {
	d := v.display[p]
	i := d.index
	if d.kind != 0 {
		group := v.entries[i].Group
		render := v.renderHeader
		if d.kind == 2 {
			render = v.renderFooter
		}
		if render != nil {
			if element := render(cx, group); element != nil {
				return el.Div().Role("heading").Name(group).Items(el.Stretch).Child(element)
			}
		}
		return el.Div().Role("heading").Name(group).Px(theme.SpaceLg).Justify(el.Center).Child(el.Text(group).TextColor(theme.Muted).TextSize(theme.TextSm))
	}

	on := v.selectedItem(i)
	r := el.Div().Role("option").Name(v.items[i]).Selected(on).Disabled(v.itemDisabled[i]).Row().Items(el.Center).Px(10).Rounded(theme.RadiusSm).Mx(4)
	if v.list.horizontal {
		r.Mb(scrollbarGutter)
	} else {
		r.Mr(scrollbarGutter)
	}
	if on {
		r.Bg(theme.Highlight).TextColor(theme.PrimaryText)
	}
	hit := r
	var custom el.Element
	if v.renderItem != nil {
		item := v.entries[i]
		item.Disabled = v.itemDisabled[i]
		item.Keywords = slices.Clone(item.Keywords)
		custom = v.renderItem(cx, ListItemContext{Item: item, Index: i, Selected: on, Disabled: v.disabled || v.itemDisabled[i]})
		if custom != nil {
			hit = el.Div().Absolute().Top(0).Left(0).Right(0).Bottom(0)
			r.Child(hit)
		}
	}
	if !v.disabled && !v.itemDisabled[i] {
		hit.CursorPointer().
			OnClick(func() { v.selectItem(cx, i, cx.ClickModifiers()); cx.Focus(autoID("list", v)) }).
			OnDoubleClick(func() { v.activate(i) })
		if !on {
			r.Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
		}
	}
	if v.reorderable && !v.disabled && !v.itemDisabled[i] {
		hit.OnDrag(func(e el.DragEvent) { v.dragItem(i, e) })
	}
	if custom != nil {
		return r.Child(custom)
	}
	if v.entries[i].Icon != IconNone {
		r.Gap(theme.SpaceSm).Child(Icon(v.entries[i].Icon).Render(cx))
	}
	return r.Child(el.Text(v.items[i]).MaxLines(1).Grow())
}

func (v *ListView) Render(cx *el.Context) el.Element {
	if v.reveal {
		if v.selected >= 0 && v.selected < len(v.positions) && v.positions[v.selected] >= 0 {
			v.list.ScrollTo(cx, v.positions[v.selected])
		}
		v.reveal = false
	}
	// The list, not each row, takes focus: rows scroll out of existence, so
	// keyboard focus could not stay on one.
	v.loadNearEnd(cx)
	frame := listFrame(el.Div().ID(autoID("list", v)).Role("listbox").Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() }), v.plain).
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
					var all []string
					for i, id := range v.keys {
						if !v.itemDisabled[i] && v.positions[i] >= 0 {
							all = append(all, id)
						}
					}
					v.selection.Set(all...)
					if !slices.Equal(before, v.SelectedValues()) && v.onSelection != nil {
						v.onSelection(v.SelectedValues())
					}
				}
				return true
			}
			var visible []int
			for _, row := range v.display {
				if row.kind == 0 {
					visible = append(visible, row.index)
				}
			}
			position := slices.Index(visible, v.selected)
			nav := base.List{Count: len(visible), Disabled: func(i int) bool { return v.itemDisabled[visible[i]] }}
			if s, ok := base.Text(e.Name, e.Modifiers&typeaheadBlockers != 0); ok {
				if e.State == el.KeyPress {
					if i, found := v.typeahead.Find(time.Now(), s, position, nav, func(i int) string { return v.items[visible[i]] }); found {
						v.selectItem(cx, visible[i], 0)
					}
				}
				return true
			}
			i, ok := nav.Key(e.Name, position)
			if ok && e.State == el.KeyPress && len(visible) > 0 {
				v.selectItem(cx, visible[i], e.Modifiers)
			}
			return ok
		}).
		Child(v.list.Render(cx))
	root := el.Div().Items(el.Stretch).Disabled(v.disabled).When(v.list.fill, func(d *el.DivEl) { d.Grow() })
	if v.searchable {
		root.Child(el.Input().ID(autoID("list", v) + "/search").Name(locale.Current().Search).Bind(&v.query).OnChange(func(s string) { v.SetQuery(s) }))
	}
	if v.slots.initial != nil && v.searchable && v.query == "" {
		// Nothing asked yet: the initial view stands in for the rows.
		return root.Child(el.Div().ID(autoID("list", v) + "/initial").Items(el.Stretch).Child(v.slots.initial.Render(cx)))
	}
	root.Child(frame)
	if state := v.stateContent(cx); state != nil {
		root.Child(state)
	}
	return root
}

// typeaheadBlockers are the modifiers that make a letter key a shortcut
// rather than typed text.
const typeaheadBlockers = key.ModCtrl | key.ModCommand | key.ModAlt | key.ModSuper

// listFrame styles the focusable box around a List or Tree. A plain one keeps
// a transparent border so the focus outline has somewhere to show.
func listFrame(d *el.DivEl, plain bool) *el.DivEl {
	if plain {
		return d.Rounded(theme.RadiusMd).Border(1, color.NRGBA{}).Py(theme.SpaceXs).Items(el.Stretch)
	}
	return d.Rounded(theme.RadiusMd).Border(1, theme.Border).Bg(theme.Surface).Py(theme.SpaceXs).Items(el.Stretch)
}
