package kit

// DisableOption controls whether the user may choose a value. Disabled entries
// remain visible, and configuration survives filtering and async result updates.
// Existing selections are retained and may still be removed; SetValue/SetValues
// remain application-controlled. False removes the override, including for
// values not currently offered. AllowCustom does not bypass this restriction.
func (v *ComboboxView) DisableOption(value string, on bool) *ComboboxView {
	if v.disabledOptions[value] == on {
		return v
	}
	if on {
		if v.disabledOptions == nil {
			v.disabledOptions = make(map[string]bool)
		}
		v.disabledOptions[value] = true
	} else {
		delete(v.disabledOptions, value)
	}
	matches := v.matches()
	if v.active >= 0 && v.active < len(matches) && v.optionDisabled(matches[v.active]) {
		v.active = v.enabledOption(matches, v.active, 1)
	}
	return v
}

func (v *ComboboxView) enabledOption(matches []string, start, direction int) int {
	if len(matches) == 0 {
		return -1
	}
	start = (start%len(matches) + len(matches)) % len(matches)
	for n := 0; n < len(matches); n++ {
		if !v.optionDisabled(matches[start]) {
			return start
		}
		start = (start + direction + len(matches)) % len(matches)
	}
	return -1
}
