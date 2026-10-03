package kit

// Clearable shows a clear button when a selection exists. Clearing closes the
// popup, invalidates pending searches and discards draft text and errors. It
// reports an empty selection once; programmatic setters remain silent.
func (v *ComboboxView) Clearable(on bool) *ComboboxView { v.clearable = on; return v }

func (v *ComboboxView) clearSelection() {
	if v.disabled || len(v.Values()) == 0 {
		return
	}
	old, wasOpen := v.value, v.open
	v.SetValues(nil)
	v.active = -1
	v.err = ""
	if v.multiple && v.onValues != nil {
		v.onValues(nil)
	}
	if old != "" && v.onChange != nil {
		v.onChange("")
	}
	v.emitConfirm(wasOpen, nil)
}
