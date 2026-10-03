package kit

import "slices"

// OnConfirm runs once after the user closes an open popup, after any selection
// change callbacks. Values are a snapshot of that completed interaction.
// Programmatic setters, disabling, removal and Searchable changes are silent.
// Esc/outside clicks use the existing draft-settling behavior; confirmation
// describes completion and does not imply a selection changed.
func (v *ComboboxView) OnConfirm(fn func([]string)) *ComboboxView { v.onConfirm = fn; return v }
func (v *ComboboxView) emitConfirm(wasOpen bool, values []string) {
	if wasOpen && v.onConfirm != nil {
		v.onConfirm(slices.Clone(values))
	}
}
func (v *ComboboxView) confirmClose() {
	wasOpen := v.open
	values := v.Values()
	v.close()
	v.emitConfirm(wasOpen, values)
}
func (v *ComboboxView) cancelAndConfirm() {
	wasOpen := v.open
	values := v.Values()
	v.cancelDraft()
	v.emitConfirm(wasOpen, values)
}
