package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

// ContextMenu replaces the built-in editing menu. Nil restores it. The menu
// belongs to this input; its Trigger is unused. Enablement still controls both.
func (v *InputView) ContextMenu(menu *MenuView) *InputView {
	if v.customMenu != menu {
		if v.customMenu != nil {
			v.customMenu.SetValue(false)
		}
		if v.editMenu != nil {
			v.editMenu.SetValue(false)
		}
		v.customMenu = menu
	}
	return v
}

// ContextMenuEnabled controls the built-in or custom right-click menu, enabled
// by default. Disabling closes any open menu without changing the input value.
func (v *InputView) ContextMenuEnabled(on bool) *InputView {
	v.contextMenuDisabled = !on
	if !on {
		if v.customMenu != nil {
			v.customMenu.SetValue(false)
		}
		if v.editMenu != nil {
			v.editMenu.SetValue(false)
		}
	}
	return v
}

func (v *InputView) renderContextMenu(cx *el.Context, field *el.InputEl) {
	menu := v.customMenu
	if menu == nil {
		if v.editMenu == nil {
			v.editMenu = Menu()
		}
		menu = v.editMenu
		menu.items = nil
		text := locale.Current()
		edit, available := cx.InputSelection(v.FocusID())
		selected := available && edit.Start != edit.End
		for _, entry := range []struct {
			label, shortcut string
			action          el.InputAction
			disabled        bool
		}{
			{text.Copy, "mod+c", el.InputCopy, !selected || v.password},
			{text.Cut, "mod+x", el.InputCut, !selected || v.password || v.readOnly},
			{text.Paste, "mod+v", el.InputPaste, v.readOnly},
			{text.SelectAll, "mod+a", el.InputSelectAll, !available || edit.Text == ""},
		} {
			action := entry.action
			menu.Item(entry.label, entry.shortcut, func() { cx.InputAction(v.FocusID(), action); cx.Focus(v.FocusID()) })
			menu.items[len(menu.items)-1].disabled = entry.disabled
		}
	}
	if v.disabled || v.contextMenuDisabled {
		menu.SetValue(false)
		return
	}
	field.OnContextMenu(func() { menu.SetValue(true) })
	if menu.open {
		context := menu.actionContext
		if context == "" {
			menu.actionContext = v.FocusID()
		}
		defer func() { menu.actionContext = context }()
		id := autoID("menu", menu)
		cx.Overlay(id, el.Anchored(v.FocusID(), menu.panel(cx)).Owner(v.FocusID()).
			Placement(menu.side, menu.align).Offset(menu.offset).Modal().TrapFocus().
			OnDismiss(func() { menu.SetValue(false) }))
		menu.renderSub(cx)
	}
}
