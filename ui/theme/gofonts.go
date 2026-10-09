package theme

import (
	"sync"

	"github.com/dyike/keel/third_party/gio/font"
	"github.com/dyike/keel/third_party/gio/font/gofont"
	"github.com/dyike/keel/third_party/gio/font/opentype"
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
	for _, ttf := range [][]byte{gobold.TTF, goitalic.TTF, gomono.TTF, gomonobold.TTF} {
		parsed, err := opentype.ParseCollectionShared(ttf) // compiled in, never changes
		if err != nil {
			panic(err) // the fonts are compiled in
		}
		faces = append(faces, parsed[0])
	}
	return faces[:len(faces):len(faces)]
})
