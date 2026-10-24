//go:build windows

package uia

// controlType maps an element role (see docs/automation.md) to UI
// Automation's control type.
func controlType(role string) int32 {
	switch role {
	case "button", "disclosure":
		return ctButton
	case "link":
		return ctHyperlink
	case "textbox":
		return ctEdit
	case "checkbox", "switch", "toggle":
		return ctCheckBox
	case "radio":
		return ctRadioButton
	case "tab":
		return ctTabItem
	case "tablist":
		return ctTab
	case "select", "combobox":
		return ctComboBox
	case "slider":
		return ctSlider
	case "progressbar":
		return ctProgressBar
	case "image", "img", "avatar":
		return ctImage
	case "menu":
		return ctMenu
	case "menuitem", "menuitemcheckbox":
		return ctMenuItem
	case "list", "listbox":
		return ctList
	case "option":
		return ctListItem
	case "table", "grid":
		return ctTable
	case "row", "gridcell":
		return ctDataItem
	case "columnheader":
		return ctHeaderItem
	case "tree":
		return ctTree
	case "treeitem":
		return ctTreeItem
	case "toolbar":
		return ctToolBar
	case "separator":
		return ctSeparator
	case "tooltip":
		return ctToolTip
	case "status":
		return ctStatusBar
	case "scrollarea", "dialog", "alertdialog", "tabpanel", "region":
		return ctPane
	case "text", "paragraph", "badge", "step", "attachment":
		return ctText
	}
	if container(role) {
		return ctGroup
	}
	return ctText
}

// localizedType names the roles a control type alone does not tell.
func localizedType(role string) string {
	switch role {
	case "dialog", "alertdialog", "switch", "toggle":
		return role
	}
	return ""
}

func container(role string) bool {
	switch role {
	case "tooltip", "dialog", "figure", "banner", "region", "toolbar", "tablist", "tabpanel", "log",
		"article", "tree", "navigation", "attachment", "menu", "alertdialog", "grid", "radiogroup",
		"listbox", "list", "form", "combobox", "table", "code", "footnotes", "accordion", "alert",
		"group", "status", "tag", "search", "scrollarea":
		return true
	}
	return false
}

// pressable reports whether a press on an element can do something.
func pressable(role string) bool {
	switch controlType(role) {
	case ctText, ctGroup, ctPane, ctImage, ctList, ctTable, ctTree, ctToolBar, ctTab, ctMenu,
		ctToolTip, ctProgressBar, ctStatusBar, ctSeparator:
		return false
	}
	return true
}

// hasValue reports whether an element's value is its Value pattern's.
func hasValue(role string) bool {
	switch role {
	case "textbox", "select", "combobox", "slider", "progressbar":
		return true
	}
	return false
}
