//go:build !js

package atspi

// AtspiRole values used here, from at-spi2-core's atspi-constants.h; the
// test checks them against the header where it is installed.
const (
	roleAlert             = 2
	roleCheckBox          = 7
	roleCheckMenuItem     = 8
	roleComboBox          = 11
	roleDialog            = 16
	roleFrame             = 23
	roleImage             = 27
	roleLabel             = 29
	roleList              = 31
	roleListItem          = 32
	roleMenu              = 33
	roleMenuItem          = 35
	rolePageTab           = 37
	rolePageTabList       = 38
	rolePanel             = 39
	roleProgressBar       = 42
	rolePushButton        = 43
	roleRadioButton       = 44
	roleScrollPane        = 49
	roleSeparator         = 50
	roleSlider            = 51
	roleStatusBar         = 54
	roleTable             = 55
	roleTableCell         = 56
	roleTableColumnHeader = 57
	roleToggleButton      = 62
	roleToolBar           = 63
	roleToolTip           = 64
	roleTree              = 65
	roleHeader            = 71
	roleApplication       = 75
	roleEntry             = 79
	roleForm              = 87
	roleLink              = 88
	roleTableRow          = 90
	roleTreeItem          = 91
)

// roleNames are the names GetRoleName reports, as libatspi spells them.
var roleNames = map[uint32]string{
	roleAlert: "alert", roleCheckBox: "check box", roleCheckMenuItem: "check menu item",
	roleComboBox: "combo box", roleDialog: "dialog", roleFrame: "frame", roleImage: "image",
	roleLabel: "label", roleList: "list", roleListItem: "list item", roleMenu: "menu",
	roleMenuItem: "menu item", rolePageTab: "page tab", rolePageTabList: "page tab list",
	rolePanel: "panel", roleProgressBar: "progress bar", rolePushButton: "push button",
	roleRadioButton: "radio button", roleScrollPane: "scroll pane", roleSeparator: "separator",
	roleSlider: "slider", roleStatusBar: "status bar", roleTable: "table", roleTableCell: "table cell",
	roleTableColumnHeader: "table column header", roleToggleButton: "toggle button",
	roleToolBar: "tool bar", roleToolTip: "tool tip", roleTree: "tree", roleHeader: "header",
	roleApplication: "application", roleEntry: "entry", roleForm: "form", roleLink: "link",
	roleTableRow: "table row", roleTreeItem: "tree item",
}

// roleOf maps an element role (see docs/automation.md) to AT-SPI's.
func roleOf(role string) uint32 {
	switch role {
	case "button":
		return rolePushButton
	case "columnheader":
		return roleTableColumnHeader
	case "gridcell":
		return roleTableCell
	case "link":
		return roleLink
	case "textbox":
		return roleEntry
	case "checkbox":
		return roleCheckBox
	case "menuitemcheckbox":
		return roleCheckMenuItem
	case "switch", "toggle", "disclosure":
		return roleToggleButton
	case "radio":
		return roleRadioButton
	case "tab":
		return rolePageTab
	case "tablist":
		return rolePageTabList
	case "select", "combobox":
		return roleComboBox
	case "slider":
		return roleSlider
	case "progressbar":
		return roleProgressBar
	case "image", "img", "avatar":
		return roleImage
	case "menu":
		return roleMenu
	case "menuitem":
		return roleMenuItem
	case "list", "listbox":
		return roleList
	case "option":
		return roleListItem
	case "table", "grid":
		return roleTable
	case "row":
		return roleTableRow
	case "tree":
		return roleTree
	case "treeitem":
		return roleTreeItem
	case "toolbar":
		return roleToolBar
	case "separator":
		return roleSeparator
	case "tooltip":
		return roleToolTip
	case "dialog":
		return roleDialog
	case "alertdialog":
		return roleAlert
	case "status":
		return roleStatusBar
	case "form":
		return roleForm
	case "banner":
		return roleHeader
	case "scrollarea":
		return roleScrollPane
	case "text", "paragraph", "badge", "step", "attachment":
		return roleLabel
	}
	return rolePanel // containers and roles a component declares
}

// AtspiStateType bit numbers.
const (
	stateActive        = 1
	stateCheckable     = 41
	stateChecked       = 4
	stateEditable      = 7
	stateEnabled       = 8
	stateExpandable    = 9
	stateExpanded      = 10
	stateFocusable     = 11
	stateFocused       = 12
	stateIndeterminate = 32
	stateResizable     = 21
	stateSelectable    = 22
	stateSelected      = 23
	stateSensitive     = 24
	stateShowing       = 25
	stateSingleLine    = 26
	stateVisible       = 30
)

// stateSet is AT-SPI's state bitfield, two 32-bit words.
type stateSet [2]uint32

func (s *stateSet) add(states ...uint) {
	for _, st := range states {
		s[st/32] |= 1 << (st % 32)
	}
}
