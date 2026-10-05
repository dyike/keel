package theme

import (
	"sync"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
)

// fallbackFaces are the Go fonts used when the system has none of the
// families in Face or MonoFace: regular, bold, italic, mono and mono bold.
// Gio's gofont.Collection has 14 faces (medium, small caps and more) and
// adds about 2 MB to every binary; the shaper synthesizes the rest.
var fallbackFaces = sync.OnceValue(func() []font.FontFace {
	faces := append([]font.FontFace{}, gofont.Regular()...)
	for _, ttf := range [][]byte{gobold.TTF, goitalic.TTF, gomono.TTF, gomonobold.TTF} {
		parsed, err := opentype.ParseCollection(ttf)
		if err != nil {
			panic(err) // the fonts are compiled in
		}
		faces = append(faces, parsed[0])
	}
	return faces[:len(faces):len(faces)]
})
