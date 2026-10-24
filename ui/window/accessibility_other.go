//go:build (!darwin || ios) && (!linux || android)

package window

import "image"

func accessibilityOpen(*Window)               {}
func accessibilityInput(*Window)              {}
func accessibilityFrame(*Window, image.Point) {}
func accessibilityFocus(*Window, bool)        {}
func accessibilityClose(*Window)              {}
