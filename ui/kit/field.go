package kit

import (
	"image/color"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// check is the shared row of checkbox-like controls: a focusable,
// clickable row with a mark and a label. Space and Enter toggle it.
func check(id, role, label, name string, selected, disabled bool, mark el.Element, toggle func()) *el.DivEl {
	if name == "" || label != "" {
		name = label
	}
	row := el.Div().ID(id).Role(role).Name(name).Selected(selected).Disabled(disabled).
		Row().Items(el.Center).Gap(8).Py(2).Rounded(theme.RadiusSm).Focusable(true).OnClick(toggle).
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
		box.Child(el.Text(label).TextSize(theme.TextMd).TextColor(theme.Muted))
	}
	box.Child(control)
	if errMsg != "" {
		box.Child(el.Text(errMsg).TextSize(theme.TextSm).TextColor(theme.DangerText))
	}
	return box
}

// fieldFrame is the one box every text field in kit draws: same height,
// padding, radius and border colors (error, then focus, then rest). Change
// how fields look here, not in each component. Frames around a text input
// also call FocusOnPress, so a press anywhere in the frame focuses the text. A field that wraps (tags in a
// multiple Combobox) grows past theme.ControlHeight.
func fieldFrame(id string, focused, invalid, disabled, readOnly bool) *el.DivEl {
	border := theme.Border
	switch {
	case invalid:
		border = theme.Danger
	case focused && !disabled:
		border = theme.Primary
	}
	bg := theme.Surface
	if disabled || readOnly {
		bg = theme.Subtle
	}
	return el.Div().ID(id).WFull().MinH(el.Dp(float32(theme.ControlHeight))).Row().Items(el.Center).Gap(8).Px(10).Py(4).
		Rounded(theme.RadiusMd).Border(1, border).Bg(bg).Disabled(disabled)
}

// fieldText strips el.Input's own box so the text sits inside a fieldFrame
// and shows the frame's background.
func fieldText(in *el.InputEl) *el.InputEl {
	return in.Border(0, color.NRGBA{}).Bg(color.NRGBA{}).P(0).MinH(el.Auto).Grow()
}

// searchField is the search box of Select, Command, Settings and the like:
// a fieldFrame with a search icon in front of the text, whose blank space
// focuses the text (inputID, the input's ID).
func searchField(cx *el.Context, id, inputID string, in *el.InputEl) *el.DivEl {
	return fieldFrame(id, cx.FocusWithin(id), false, false, false).FocusOnPress(inputID).
		Child(Icon(IconSearch).Size(16).Color(theme.Muted).Render(cx), fieldText(in))
}
