package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"time"
)

// DatePickerPreset is a shortcut with a stable nonempty ID and visible label.
// Single-date mode ignores End; in range mode a zero End uses Start.
type DatePickerPreset struct {
	ID, Label  string
	Start, End time.Time
}

// Presets replaces the shortcuts, copying the slice. Empty IDs/labels and
// duplicate IDs are ignored (first wins). No arguments removes all shortcuts.
// Dates are snapshots; refresh them explicitly for moving windows like Today.
func (v *DatePickerView) Presets(items ...DatePickerPreset) *DatePickerView {
	v.presets = nil
	seen := map[string]bool{}
	for _, p := range items {
		if p.ID == "" || p.Label == "" || seen[p.ID] {
			continue
		}
		seen[p.ID] = true
		v.presets = append(v.presets, p)
	}
	return v
}

func (v *DatePickerView) presetDates(p DatePickerPreset) (time.Time, time.Time, bool) {
	a, b := day(p.Start), day(p.End)
	if !v.cal.rangeMode || b.IsZero() {
		b = a
	}
	if a.IsZero() {
		return a, b, false
	}
	if b.Before(a) {
		a, b = b, a
	}
	if v.disabled || !v.cal.allowed(a) || !v.cal.allowed(b) {
		return a, b, false
	}
	// Endpoints suffice for Bounds; a custom matcher also applies inside ranges.
	if v.cal.blocked != nil {
		for d := a.AddDate(0, 0, 1); d.Before(b); {
			if !v.cal.allowed(d) {
				return a, b, false
			}
			next := d.AddDate(0, 0, 1)
			if !next.After(d) {
				return a, b, false
			}
			d = next
		}
	}
	return a, b, true
}

func (v *DatePickerView) selectPreset(id string) bool {
	for _, p := range v.presets {
		if p.ID != id {
			continue
		}
		a, b, ok := v.presetDates(p)
		if !ok {
			return false
		}
		v.close()
		v.cal.SetValue(a, b)
		v.err = ""
		if v.onChange != nil {
			v.onChange(a, b)
		}
		return true
	}
	return false
}

func (v *DatePickerView) renderPresets(cx *el.Context) el.Element {
	root := el.Div().ID(v.FocusID() + "/presets").Wrap().Gap(theme.SpaceXs).MaxW(el.Dp(252)).Hidden(len(v.presets) == 0)
	for _, p := range v.presets {
		_, _, allowed := v.presetDates(p)
		button := Button(p.Label, func() {
			if v.selectPreset(p.ID) {
				cx.Focus(v.FocusID())
			}
		}).ID(v.FocusID() + "/preset/" + p.ID).Variant(ButtonGhost).Size(28)
		button.SetDisabled(!allowed)
		root.Child(button.Render(cx))
	}
	return root
}
