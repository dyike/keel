package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

// Size sets the minimum field height in dp, scaling text, spacing, controls and
// candidate rows. Recommended values are 28, 36 and 48. Zero restores defaults.
// Invalid values are ignored. Footer content keeps its own size configuration.
func (v *ComboboxView) Size(dp float32) *ComboboxView {
	if dp >= 0 && finiteNumber(float64(dp)) && v.height != dp {
		v.height = dp
		if v.active >= 0 {
			v.virtual.reveal = v.active
		}
	}
	return v
}
func (v *ComboboxView) sizeRatio() float32 {
	if v.height > 0 {
		return v.height / float32(theme.ControlHeight)
	}
	return 1
}

// CheckIcon replaces the selected-row icon with a snapshot of a built-in or
// vector icon. Its configured size scales with Size; color is preserved.
// Nil restores the default check. Icon(IconNone) hides it but reserves its slot.
func (v *ComboboxView) CheckIcon(icon *IconView) *ComboboxView {
	v.checkIcon = nil
	if icon != nil {
		copy := *icon
		v.checkIcon = &copy
	}
	return v
}
func (v *ComboboxView) renderCheck(cx *el.Context, on bool) el.Element {
	icon := Icon(IconDone).Size(14).Color(theme.PrimaryText)
	if v.checkIcon != nil {
		copy := *v.checkIcon
		icon = &copy
	}
	size := icon.size * v.sizeRatio()
	if !on || icon.icon == nil {
		return el.Div().Size(el.Dp(size))
	}
	return icon.Size(size).Render(cx)
}
