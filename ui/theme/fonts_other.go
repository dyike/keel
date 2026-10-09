//go:build !android

package theme

import "github.com/dyike/keel/third_party/gio/font"

func platformFaces() []font.FontFace { return nil }
