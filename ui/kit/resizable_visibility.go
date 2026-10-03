package kit

// Visible controls the two panes independently; both are visible by default.
// With one pane visible it fills the container without a handle or split limits.
// Hidden content remains declared, preserving widget state, but cannot receive input.
// Restoring both panes clamps the stored split size to the current constraints.
// This does not call OnChange.
func (v *ResizableView) Visible(first, second bool) *ResizableView {
	v.hiddenFirst, v.hiddenSecond = !first, !second
	return v
}
