package kit

import (
	"strconv"
	"time"

	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/locale"

	"gioui.org/io/key"
	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

type menuItem struct {
	label, shortcut    string
	keymapAction       string // shows this action's binding instead of shortcut
	action             func()
	sub                *MenuView
	separator          bool
	heading            bool
	disabled           bool
	icon               IconName
	checkable, checked bool
	onCheck            func(bool)
	content            el.View
	url                string
	link               bool
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
	trigger       el.View
	items         []menuItem
	open          bool
	openSub       int // index of the open submenu, -1 none
	parent        *MenuView
	width         float32
	disabled      bool
	reveal        int
	find          base.Typeahead
	side          el.Side
	align         el.Align
	offset        float32
	checkRight    bool
	focusPending  bool
	onLink        func(string)
	onLinkError   func(error)
	hideLinkIcon  bool
	actionContext string
}

func Menu() *MenuView { return &MenuView{openSub: -1, width: 220, reveal: -1, offset: 4} }

// Trigger sets the view the menu opens from (top-level menus only).
func (v *MenuView) Trigger(t el.View) *MenuView { v.trigger = t; return v }

// Placement sets the top-level menu's direction and alignment. Submenus
// retain their right/start placement and automatic edge flipping.
func (v *MenuView) Placement(side el.Side, align el.Align) *MenuView {
	if side <= el.Right && align <= el.End {
		v.side, v.align = side, align
	}
	return v
}

// Offset sets the top-level anchor gap in dp, including zero or overlap.
// Non-finite values are ignored; the default is 4dp.
func (v *MenuView) Offset(dp float32) *MenuView {
	if finiteNumber(float64(dp)) {
		v.offset = dp
	}
	return v
}

// Width sets the minimum menu width in dp, 220 by default.
func (v *MenuView) Width(dp float32) *MenuView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.width = dp
	}
	return v
}

// ActionItem adds a command that shows the key bound to a keymap action
// (core.Bind/BindIn), following rebinding and the trigger or ActionContext.
// With fn nil, choosing it runs the action's handler where the menu acts,
// as its key would: the innermost cx.ActionAt enclosing the trigger or
// ActionContext, else cx.Action. Otherwise only fn runs.
func (v *MenuView) ActionItem(label, keymapAction string, fn func()) *MenuView {
	v.items = append(v.items, menuItem{label: label, keymapAction: keymapAction, action: fn})
	return v
}

// ActionContext chooses the element ID whose KeyContext ancestry resolves
// ActionItem hints. Empty restores the trigger context (inherited by submenus).
// It also picks the handler an ActionItem without a callback runs.
func (v *MenuView) ActionContext(id string) *MenuView { v.actionContext = id; return v }

func (v *MenuView) keyTarget() string {
	if v.actionContext != "" {
		return v.actionContext
	}
	if v.parent != nil {
		return v.parent.keyTarget()
	}
	return autoID("menu", v)
}

// Item adds a command; shortcut uses core.ParseShortcut syntax and may be empty.
func (v *MenuView) Item(label, shortcut string, action func()) *MenuView {
	v.items = append(v.items, menuItem{label: label, shortcut: shortcut, action: action})
	return v
}

// Link adds an external link command. It closes the menu before opening the URL.
func (v *MenuView) Link(label, url string) *MenuView {
	v.items = append(v.items, menuItem{label: label, url: url, link: true})
	return v
}

// OnLink replaces system URL opening. Submenus inherit the nearest configured
// ancestor handler. Nil restores inheritance/default opening.
func (v *MenuView) OnLink(fn func(string)) *MenuView { v.onLink = fn; return v }

// OnLinkError receives default URL validation/launch errors, with ancestor fallback.
func (v *MenuView) OnLinkError(fn func(error)) *MenuView { v.onLinkError = fn; return v }

func (v *MenuView) ExternalLinkIcon(on bool) *MenuView { v.hideLinkIcon = !on; return v }

func (v *MenuView) openLink(url string) {
	for menu := v; menu != nil; menu = menu.parent {
		if menu.onLink != nil {
			menu.onLink(url)
			return
		}
	}
	if err := core.OpenURL(url); err != nil {
		for menu := v; menu != nil; menu = menu.parent {
			if menu.onLinkError != nil {
				menu.onLinkError(err)
				return
			}
		}
	}
}

// ContentItem adds a command with arbitrary display-only content. Label remains
// its accessible name and typeahead key. Nil content falls back to the label.
func (v *MenuView) ContentItem(label, shortcut string, content el.View, action func()) *MenuView {
	v.items = append(v.items, menuItem{label: label, shortcut: shortcut, content: content, action: action})
	return v
}

// SetItemContent changes matching actionable rows without changing their IDs.
func (v *MenuView) SetItemContent(label string, content el.View) {
	for i := range v.items {
		if v.items[i].label == label && !v.items[i].separator && !v.items[i].heading {
			v.items[i].content = content
		}
	}
}

// IconItem adds an ordinary command with a leading icon.
func (v *MenuView) IconItem(label, shortcut string, icon IconName, action func()) *MenuView {
	v.items = append(v.items, menuItem{label: label, shortcut: shortcut, icon: icon, action: action})
	return v
}

// CheckItem toggles stored state, closes the menu, then calls onChange.
func (v *MenuView) CheckItem(label, shortcut string, checked bool, onChange func(bool)) *MenuView {
	v.items = append(v.items, menuItem{label: label, shortcut: shortcut, checkable: true, checked: checked, onCheck: onChange})
	return v
}

// SetItemChecked updates matching check items without calling onChange.
func (v *MenuView) SetItemChecked(label string, checked bool) {
	for i := range v.items {
		if v.items[i].label == label && v.items[i].checkable {
			v.items[i].checked = checked
		}
	}
}

// ItemChecked returns the first matching check item's state and whether it exists.
func (v *MenuView) ItemChecked(label string) (bool, bool) {
	for _, it := range v.items {
		if it.label == label && it.checkable {
			return it.checked, true
		}
	}
	return false, false
}

// SetItemIcon updates all matching non-separator items, including submenus.
func (v *MenuView) SetItemIcon(label string, icon IconName) {
	for i := range v.items {
		if v.items[i].label == label && !v.items[i].separator && !v.items[i].heading {
			v.items[i].icon = icon
		}
	}
}

// CheckSide chooses Left (default) or Right. Left checks replace item icons.
func (v *MenuView) CheckSide(side el.Side) *MenuView {
	if side == el.Left || side == el.Right {
		v.checkRight = side == el.Right
	}
	return v
}

// Label adds a non-interactive section heading. Empty labels are ignored.
// Headings retain a row of space but are skipped by keyboard navigation/search.
func (v *MenuView) Label(label string) *MenuView {
	if label != "" {
		v.items = append(v.items, menuItem{label: label, heading: true})
	}
	return v
}

func (v *MenuView) Separator() *MenuView {
	v.items = append(v.items, menuItem{separator: true})
	return v
}

// Sub adds an item that opens sub to its side.
func (v *MenuView) Sub(label string, sub *MenuView) *MenuView {
	if sub == nil || sub.parent != nil {
		return v
	}
	for p := v; p != nil; p = p.parent {
		if p == sub {
			return v
		}
	}
	sub.parent = v
	v.items = append(v.items, menuItem{label: label, sub: sub})
	return v
}

// SetItemDisabled enables or disables the item with label.
func (v *MenuView) SetItemDisabled(label string, disabled bool) {
	for i := range v.items {
		if v.items[i].label == label {
			v.items[i].disabled = disabled
			if disabled && v.openSub == i {
				v.closeSub()
			}
		}
	}
}
func (v *MenuView) Value() bool { return v.open }
func (v *MenuView) SetValue(open bool) {
	v.open = open && !v.disabled
	v.closeSub()
	v.reveal = -1
	v.focusPending = false
	v.find.Reset()
	if v.open {
		v.reveal = v.step(-1, 1)
	}
}
func (v *MenuView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.SetValue(false)
		if p := v.parent; p != nil && p.openSub >= 0 && p.openSub < len(p.items) && p.items[p.openSub].sub == v {
			p.closeSub()
		}
	}
}
func (v *MenuView) closeSub() {
	if v.openSub >= 0 && v.openSub < len(v.items) {
		if sub := v.items[v.openSub].sub; sub != nil {
			sub.closeSub()
			sub.open = false
			sub.reveal = -1
		}
	}
	v.openSub = -1
}
func (v *MenuView) itemDisabled(i int) bool {
	it := v.items[i]
	return it.heading || it.disabled || it.sub != nil && it.sub.disabled
}

// Toggle opens or closes the menu; pass it as the trigger's click handler.
func (v *MenuView) Toggle() { v.SetValue(!v.open) }

func (v *MenuView) root() *MenuView {
	for v.parent != nil {
		v = v.parent
	}
	return v
}
func (v *MenuView) itemID(i int) string { return autoID("menu", v) + "/" + strconv.Itoa(i) }

// nav navigates the items, skipping separators and disabled items and
// wrapping around.
func (v *MenuView) nav() base.List {
	return base.List{Count: len(v.items), Wrap: true, Disabled: func(j int) bool { return v.items[j].separator || v.itemDisabled(j) }}
}

// step returns the next enabled item from i in direction d.
func (v *MenuView) step(i, d int) int { return v.nav().Next(i, d) }

func (v *MenuView) Render(cx *el.Context) el.Element {
	id := autoID("menu", v)
	if v.open {
		cx.Overlay(id, el.Anchored(id, v.panel(cx)).Placement(v.side, v.align).Offset(v.offset).Modal().TrapFocus().
			OnDismiss(func() { v.SetValue(false) }))
		v.renderSub(cx)
	}
	return el.Div().Disabled(v.disabled).Child(anchor(id, cx, v.trigger))
}

// renderSub declares the open submenu chain after this menu's layer, so each
// level sits above its parent and Esc closes the innermost one first.
func (v *MenuView) renderSub(cx *el.Context) {
	if v.openSub < 0 || v.openSub >= len(v.items) || v.items[v.openSub].sub == nil || v.itemDisabled(v.openSub) {
		return
	}
	sub := v.items[v.openSub].sub
	cx.Overlay(autoID("menu", sub), el.Anchored(v.itemID(v.openSub), sub.panel(cx)).Placement(el.Right, el.Start).Offset(2).TrapFocus().
		OnDismiss(v.closeSub))
	sub.renderSub(cx)
}

func (v *MenuView) panel(cx *el.Context) el.Element {
	w, h := cx.ViewportSize()
	list := floating(theme.ElevationMd).ID(autoID("menu-scroll", v)).Role("menu").Name(v.label()).MinW(el.Dp(min(v.width, max(0, w-16)))).MaxW(el.Dp(max(0, w-16))).MaxH(el.Dp(max(0, h-16))).ScrollY().Py(theme.SpaceXs).Items(el.Stretch)
	rows := make([]el.Element, len(v.items))
	leading := false
	for _, it := range v.items {
		leading = leading || it.icon != IconNone || it.checkable && !v.checkRight
	}
	for i, it := range v.items {
		if it.separator {
			rows[i] = el.Div().NoShrink().H(el.Dp(1)).My(4).Bg(theme.Border)
			list.Child(rows[i])
			continue
		}
		if it.heading {
			rows[i] = el.Div().ID(v.itemID(i)).Role("heading").Name(it.label).NoShrink().H(el.Dp(30)).
				Mx(4).Px(theme.SpaceMd).Justify(el.Center).
				Child(el.Text(it.label).TextSize(theme.TextSm).TextColor(theme.Muted).Bold().MaxLines(1))
			list.Child(rows[i])
			continue
		}
		rows[i] = v.row(cx, i, it, leading)
		list.Child(rows[i])
	}
	return list.Decorate(func(gtx core.C, draw func()) {
		draw()
		if v.reveal < 0 || v.reveal >= len(rows) || !gtx.Enabled() {
			return
		}
		top := float32(theme.SpaceXs)
		for i, row := range rows {
			_, height := cx.LayoutSize(row)
			if v.items[i].separator {
				height += 8
			}
			if i == v.reveal {
				id := autoID("menu-scroll", v)
				before, viewport, _ := cx.ScrollState(id)
				if viewport <= 0 {
					gtx.Execute(op.InvalidateCmd{})
					return
				}
				cx.ScrollIntoView(id, top, top+height)
				after, _, _ := cx.ScrollState(id)
				if before != after {
					gtx.Execute(op.InvalidateCmd{})
					return
				}
				if v.focusPending {
					cx.Focus(v.itemID(i))
					gtx.Execute(op.InvalidateCmd{})
				}
				v.reveal, v.focusPending = -1, false
				return
			}
			top += height
		}
	})
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

func (v *MenuView) row(cx *el.Context, i int, it menuItem, leading bool) el.Element {
	run := func() {
		if v.itemDisabled(i) {
			return
		}
		if it.sub != nil {
			v.openSub = i
			it.sub.reveal = it.sub.step(-1, 1)
			cx.Focus(it.sub.itemID(it.sub.reveal))
			return
		}
		if it.checkable {
			v.items[i].checked = !v.items[i].checked
			checked := v.items[i].checked
			v.root().SetValue(false)
			if it.onCheck != nil {
				it.onCheck(checked)
			}
			return
		}
		v.root().SetValue(false)
		if it.link {
			v.openLink(it.url)
			return
		}
		if it.action != nil {
			it.action()
		} else if it.keymapAction != "" {
			cx.Perform(v.keyTarget(), it.keymapAction)
		}
	}
	row := el.Div().ID(v.itemID(i)).NoShrink().Role("menuitem").Name(it.label).Row().Items(el.Center).Gap(theme.SpaceLg).
		Ml(4).Mr(scrollbarGutter).Px(theme.SpaceMd).H(el.Dp(30)).Rounded(theme.RadiusSm).Focusable(true).Disabled(v.itemDisabled(i)).
		DisabledStyle(func(s *el.Style) { s.TextColor(theme.Muted) }).
		FocusStyle(func(s *el.Style) { s.Bg(theme.Subtle).BorderColor(theme.Subtle) }).
		OnClick(run).
		OnKey(func(e el.KeyEvent) bool { return v.key(cx, i, e) })
	if it.link {
		row.Value(it.url)
	}
	if it.checkable {
		row.Role("menuitemcheckbox").Selected(it.checked)
	}
	color := theme.Text
	if v.itemDisabled(i) {
		color = theme.Muted
	}
	if leading {
		icon := it.icon
		if it.checkable && !v.checkRight {
			icon = IconNone
			if it.checked {
				icon = IconDone
			}
		}
		slot := el.Div().ID("mark").Size(el.Dp(16)).NoShrink().Center()
		if icon != IconNone {
			slot.Child(Icon(icon).Size(16).Color(color).Render(cx))
		}
		row.Child(slot)
	}
	if it.content != nil {
		row.H(el.Auto).MinH(el.Dp(30)).Py(theme.SpaceXs)
		row.Child(el.Div().ID("label").Grow().MinW(el.Dp(0)).Child(it.content.Render(cx)))
	} else {
		row.Child(el.Text(it.label).ID("label").Grow().MaxLines(1))
	}
	if it.checkable && v.checkRight {
		slot := el.Div().ID("check").Size(el.Dp(16)).NoShrink().Center()
		if it.checked {
			slot.Child(Icon(IconDone).Size(16).Color(color).Render(cx))
		}
		row.Child(slot)
	}
	if !v.itemDisabled(i) {
		row.CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) })
	}
	if it.sub != nil {
		row.Value("submenu").Child(Icon(IconChevronRight).Size(14).Color(theme.Muted).Render(cx))
	} else if it.link && !v.hideLinkIcon {
		row.Child(Icon(IconExternalLink).Size(14).Color(theme.Muted).Render(cx))
	} else if it.keymapAction != "" {
		row.Child(el.KeyHint(it.keymapAction, v.keyTarget(), func(chord string) el.Element {
			return Kbd(chord).Plain().Render(cx)
		}))
	} else if it.shortcut != "" {
		row.Child(Kbd(it.shortcut).Plain().Render(cx))
	}
	return row
}

func (v *MenuView) key(cx *el.Context, i int, e el.KeyEvent) bool {
	if s, ok := base.Text(e.Name, e.Modifiers&typeaheadBlockers != 0); ok {
		if e.State == el.KeyPress {
			if j, found := v.find.Find(time.Now(), s, i, v.nav(), func(j int) string { return v.items[j].label }); found {
				v.focusItem(cx, j)
			}
		}
		return true
	}
	if e.State != el.KeyPress {
		return navKey(e)
	}
	switch key.Name(e.Name) {
	case key.NameDownArrow:
		v.focusItem(cx, v.step(i, 1))
	case key.NameUpArrow:
		v.focusItem(cx, v.step(i, -1))
	case key.NameHome:
		v.focusItem(cx, v.step(-1, 1))
	case key.NameEnd:
		v.focusItem(cx, v.step(len(v.items), -1))
	case key.NameRightArrow:
		if sub := v.items[i].sub; sub != nil && !v.itemDisabled(i) {
			v.openSub = i
			sub.reveal = sub.step(-1, 1)
			cx.Focus(sub.itemID(sub.reveal))
		}
	case key.NameLeftArrow:
		if p := v.parent; p != nil && p.openSub >= 0 && p.openSub < len(p.items) && p.items[p.openSub].sub == v {
			p.closeSub() // focus returns to the parent item
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

func (v *MenuView) focusItem(cx *el.Context, i int) {
	v.reveal, v.focusPending = i, true
}
