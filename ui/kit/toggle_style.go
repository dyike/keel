package kit

import "github.com/dyike/keel/ui/theme"

// ToggleVariant selects the unpressed surface. The default preserves the original filled outline.
type ToggleVariant uint8

const (
	ToggleDefault ToggleVariant = iota
	ToggleGhost
	ToggleOutline
)

// ToggleSize selects coordinated button height, text, icon and padding.
type ToggleSize uint8

const (
	ToggleSizeMedium ToggleSize = iota
	ToggleSizeXSmall
	ToggleSizeSmall
	ToggleSizeLarge
)

type toggleAppearance struct {
	variant             ToggleVariant
	size                ToggleSize
	variantSet, sizeSet bool
}

func (v *ToggleView) Variant(variant ToggleVariant) *ToggleView {
	if variant <= ToggleOutline {
		v.appearance.variant = variant
		v.appearance.variantSet = true
	}
	return v
}
func (v *ToggleGroupView) Variant(variant ToggleVariant) *ToggleGroupView {
	if variant <= ToggleOutline {
		v.appearance.variant = variant
		v.appearance.variantSet = true
	}
	return v
}
func (v *ToggleView) Size(size ToggleSize) *ToggleView {
	if size <= ToggleSizeLarge {
		v.appearance.size = size
		v.appearance.sizeSet = true
	}
	return v
}
func (v *ToggleGroupView) Size(size ToggleSize) *ToggleGroupView {
	if size <= ToggleSizeLarge {
		v.appearance.size = size
		v.appearance.sizeSet = true
	}
	return v
}
func (a toggleAppearance) metrics() (height, font, icon, padding float32) {
	switch a.size {
	case ToggleSizeXSmall:
		return 24, theme.TextXs, 12, theme.SpaceSm
	case ToggleSizeSmall:
		return 28, theme.TextSm, 14, theme.SpaceMd
	case ToggleSizeLarge:
		return 40, theme.TextBody, 20, theme.SpaceXl
	default:
		return 32, theme.TextControl, 16, theme.SpaceLg
	}
}
