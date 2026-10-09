package theme

import (
	"sync"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
)

// fallbackFaces include Android CJK faces and the Go fonts used when the
// system has none of the families in Face or MonoFace.
// Gio's gofont.Collection has 14 faces (medium, small caps and more) and
// adds about 2 MB to every binary; the shaper synthesizes the rest.
var fallbackFaces = sync.OnceValue(func() []font.FontFace {
	faces := append(platformFaces(), gofont.Regular()...)
	// Compiled in, so their bytes never change: parse them in place.
	parsed, err := parseFaces(func(font.Font) bool { return true }, [][]byte{gobold.TTF, goitalic.TTF, gomono.TTF, gomonobold.TTF})
	if err != nil {
		panic(err) // the fonts are compiled in
	}
	faces = append(faces, parsed...)
	return faces[:len(faces):len(faces)]
})
