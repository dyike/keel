//go:build !darwin || ios

// SPDX-License-Identifier: Unlicense OR MIT
package text

func newNativeFontShaper([]FontFace) nativeFontShaper { return nil }
