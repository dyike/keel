//go:build (!darwin || ios) && (!linux || android) && !windows

package window

import "image"

func accessibilityOpen(*Window)               {}
func accessibilityFrame(*Window, image.Point) {}
func accessibilityFocus(*Window, bool)        {}
func accessibilityClose(*Window)              {}
