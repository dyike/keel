package kit

import (
	"github.com/dyike/keel/ui/locale"
	"strconv"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type menuItem struct {
	label, shortcut string
	action          func()
	sub             *MenuView
	separator       bool
	disabled        bool
}

// MenuView is a list of commands in a layer next to its trigger:
//
//	more := kit.Menu().Item("复制", "mod+c", copy).Separator().Sub("导出", exportMenu)
//	more.Trigger(kit.Button("更多", more.Toggle).Variant(kit.ButtonGhost))
//
// ↑ ↓ move between enabled items, Home / End jump, Enter / Space run the item,
// → opens a submenu, ← or Esc closes one level. Running any item closes the
// whole menu. Shortcuts are only displayed, never registered.
type MenuView struct {
	trigger el.View
	items   []menuItem
	open    bool
	openSub int // index of the open submenu, -1 none
	parent  *MenuView
	width   float32
}

func Menu() *MenuView { return &MenuView{openSub: -1, width: 220} }

// Trigger sets the view the menu opens from (top-level menus only).
func (v *MenuView) Trigger(t el.View) *MenuView { v.trigger = t; return v }

// Width sets the minimum menu width in dp, 220 by default.
func (v *MenuView) Width(dp float32) *MenuView {
	if dp > 0 {
		v.width = dp
	}
	return v
}

// Item adds a command; shortcut uses core.ParseShortcut syntax and may be empty.
func (v *MenuView) Item(label, shortcut string, action func()) *MenuView {
	v.items = append(v.items, menuItem{label: label, shortcut: shortcut, action: action})
	return v
}
func (v *MenuView) Separator() *MenuView {
	v.items = append(v.items, menuItem{separator: true})
	return v
}

// Sub adds an item that opens sub to its side.
func (v *MenuView) Sub(label string, sub *MenuView) *MenuView {
	sub.parent = v
	v.items = append(v.items, menuItem{label: label, sub: sub})
	return v
}

// SetItemDisabled enables or disables the item with label.
func (v *MenuView) SetItemDisabled(label string, disabled bool) {
	for i := range v.items {
		if v.items[i].label == label {
			v.items[i].disabled = disabled
		}
	}
}
func (v *MenuView) Value() bool        { return v.open }
func (v *MenuView) SetValue(open bool) { v.open = open; v.openSub = -1 }

// Toggle opens or closes the menu; pass it as the trigger's click handler.
func (v *MenuView) Toggle() { v.SetValue(!v.open) }

func (v *MenuView) root() *MenuView {
	for v.parent != nil {
		v = v.parent
	}
	return v
}
func (v *MenuView) itemID(i int) string { return autoID("menu", v) + "/" + strconv.Itoa(i) }

// step returns the next enabled item from i in direction d, wrapping around.
func (v *MenuView) step(i, d int) int {
	n := len(v.items)
	for k := 1; k <= n; k++ {
		j := ((i+d*k)%n + n) % n
		if it := v.items[j]; !it.separator && !it.disabled {
			return j
		}
	}
	return i
}

func (v *MenuView) Render(cx *el.Context) el.Element {
	id := autoID("menu", v)
	if v.open {
		cx.Overlay(id, el.Anchored(id, v.panel(cx)).Placement(el.Bottom, el.Start).Modal().TrapFocus().
			OnDismiss(func() { v.SetValue(false) }))
		v.renderSub(cx)
	}
	return anchor(id, cx, v.trigger)
}

// renderSub declares the open submenu chain after this menu's layer, so each
// level sits above its parent and Esc closes the innermost one first.
func (v *MenuView) renderSub(cx *el.Context) {
	if v.openSub < 0 || v.openSub >= len(v.items) || v.items[v.openSub].sub == nil {
		return
	}
	sub := v.items[v.openSub].sub
	cx.Overlay(autoID("menu", sub), el.Anchored(v.itemID(v.openSub), sub.panel(cx)).Placement(el.Right, el.Start).Offset(2).TrapFocus().
		OnDismiss(func() { v.openSub = -1 }))
	sub.renderSub(cx)
}

func (v *MenuView) panel(cx *el.Context) el.Element {
	list := surface().Role("menu").Name(v.label()).MinW(el.Dp(v.width)).Py(4).Items(el.Stretch)
	for i, it := range v.items {
		if it.separator {
			list.Child(el.Div().H(el.Dp(1)).My(4).Bg(theme.Border))
			continue
		}
		list.Child(v.row(cx, i, it))
	}
	return list
}

func (v *MenuView) label() string {
	if v.parent == nil {
		return locale.Current().Menu
	}
	for _, it := range v.parent.items {
		if it.sub == v {
			return it.label
		}
	}
	return locale.Current().Menu
}

func (v *MenuView) row(cx *el.Context, i int, it menuItem) el.Element {
	run := func() {
		if it.disabled {
			return
		}
		if it.sub != nil {
			v.openSub = i
			cx.Focus(it.sub.itemID(it.sub.step(-1, 1)))
			return
		}
		v.root().SetValue(false)
		if it.action != nil {
			it.action()
		}
	}
	row := el.Div().ID(v.itemID(i)).Role("menuitem").Name(it.label).Row().Items(el.Center).Gap(12).
		Mx(4).Px(8).H(el.Dp(30)).Rounded(4).Focusable(true).Disabled(it.disabled).
		DisabledStyle(func(s *el.Style) { s.TextColor(theme.Muted) }).
		FocusStyle(func(s *el.Style) { s.Bg(theme.Subtle).BorderColor(theme.Subtle) }).
		OnClick(run).
		OnKey(func(e el.KeyEvent) bool { return v.key(cx, i, e) }).
		Child(el.Text(it.label).Grow().MaxLines(1))
	if !it.disabled {
		row.CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	}
	if it.sub != nil {
		row.Value("submenu").Child(Icon(IconChevronRight).Size(14).Color(theme.Muted).Render(cx))
	} else if it.shortcut != "" {
		row.Child(Kbd(it.shortcut).Plain().Render(cx))
	}
	return row
}

func (v *MenuView) key(cx *el.Context, i int, e el.KeyEvent) bool {
	if e.State != el.KeyPress {
		return navKey(e)
	}
	switch key.Name(e.Name) {
	case key.NameDownArrow:
		cx.Focus(v.itemID(v.step(i, 1)))
	case key.NameUpArrow:
		cx.Focus(v.itemID(v.step(i, -1)))
	case key.NameHome:
		cx.Focus(v.itemID(v.step(-1, 1)))
	case key.NameEnd:
		cx.Focus(v.itemID(v.step(len(v.items), -1)))
	case key.NameRightArrow:
		if sub := v.items[i].sub; sub != nil && !v.items[i].disabled {
			v.openSub = i
			cx.Focus(sub.itemID(sub.step(-1, 1)))
		}
	case key.NameLeftArrow:
		if v.parent != nil {
			v.parent.openSub = -1 // focus returns to the parent item
		}
	default:
		return false
	}
	return true
}

func navKey(e el.KeyEvent) bool {
	switch key.Name(e.Name) {
	case key.NameDownArrow, key.NameUpArrow, key.NameHome, key.NameEnd, key.NameRightArrow, key.NameLeftArrow:
		return true
	}
	return false
}
