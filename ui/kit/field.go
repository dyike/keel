package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// check is the shared row of checkbox-like controls: a focusable,
// clickable row with a mark and a label. Space and Enter toggle it.
func check(id, role, label string, selected, disabled bool, mark el.Element, toggle func()) *el.DivEl {
	row := el.Div().ID(id).Role(role).Name(label).Selected(selected).Disabled(disabled).
		Row().Items(el.Center).Gap(8).Py(2).Rounded(4).Focusable(true).OnClick(toggle).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		Child(mark)
	if !disabled {
		row.CursorPointer()
	}
	if label != "" {
		row.Child(el.Text(label))
	}
	return row
}

// labelled stacks a field label, the control and an error message, the
// layout every form control shares.
func labelled(label string, control el.Element, errMsg string) el.Element {
	if label == "" && errMsg == "" {
		return control
	}
	box := el.Div().Gap(6).Items(el.Stretch)
	if label != "" {
		box.Child(el.Text(label).TextSize(13).TextColor(theme.Muted))
	}
	box.Child(control)
	if errMsg != "" {
		box.Child(el.Text(errMsg).TextSize(12).TextColor(theme.DangerText))
	}
	return box
}
