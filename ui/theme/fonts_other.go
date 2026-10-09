//go:build !android

package theme

import "gioui.org/font"

func platformFaces() []font.FontFace { return nil }
