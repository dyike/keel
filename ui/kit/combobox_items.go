package kit

// ComboboxItem separates a stable application value from its visible label.
// Empty Value entries are ignored; empty Label falls back to Value.
type ComboboxItem struct {
	Value, Label string
	Disabled     bool
}

// SetItems replaces candidate metadata, copying the entries and taking the
// first occurrence of each value. Selections and independent DisableOption
// overrides are retained. A selected value absent from the new results keeps
// its prior label. No selection callback runs.
func (v *ComboboxView) SetItems(items ...ComboboxItem) {
	v.groupFor, v.groupLabels = nil, nil
	previous := v.itemLabels
	oldDisplay := v.optionLabel(v.value)
	v.itemLabels = make(map[string]string)
	v.itemDisabled = make(map[string]bool)
	v.options = nil
	for _, item := range items {
		if item.Value == "" {
			continue
		}
		if _, seen := v.itemLabels[item.Value]; seen {
			continue
		}
		label := item.Label
		if label == "" {
			label = item.Value
		}
		v.options = append(v.options, item.Value)
		v.itemLabels[item.Value] = label
		v.itemDisabled[item.Value] = item.Disabled
	}
	for _, value := range v.Values() {
		if _, ok := v.itemLabels[value]; !ok && previous[value] != "" {
			v.itemLabels[value] = previous[value]
		}
	}
	if !v.multiple && v.text == oldDisplay {
		v.text = v.optionLabel(v.value)
	}
	v.revision++
	v.active = -1
}

// SetItemResults delivers structured async results. Like SetResults it must run
// on the UI loop, and returns false for stale, closed or disabled requests.
func (v *ComboboxView) SetItemResults(token uint64, items ...ComboboxItem) bool {
	if !v.open || v.disabled || token != v.request || v.onSearch == nil {
		return false
	}
	v.SetItems(items...)
	v.loading = false
	v.searchError = ""
	v.active = v.enabledOption(v.matches(), 0, 1)
	return true
}
func (v *ComboboxView) optionLabel(value string) string {
	if label, ok := v.itemLabels[value]; ok {
		return label
	}
	return value
}
func (v *ComboboxView) optionDisabled(value string) bool {
	return v.disabledOptions[value] || v.itemDisabled[value]
}
func (v *ComboboxView) inputValue(text string) (string, bool) {
	// Exact stable values take priority if another item's label collides.
	for _, value := range v.options {
		if value == text {
			return value, true
		}
	}
	for _, value := range v.options {
		if v.optionLabel(value) == text {
			return value, true
		}
	}
	return text, false
}
