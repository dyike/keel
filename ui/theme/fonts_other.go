//go:build !android

package theme

import "github.com/dyike/keel/gio/font"

func platformFaces() []font.FontFace { return nil }
