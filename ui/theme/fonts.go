package theme

import (
	"fmt"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"

	"github.com/dyike/keel/ui/internal/loop"
)

// loaded holds the faces added with LoadFonts, ahead of the Go fonts.
var loaded []font.FontFace

// LoadFonts adds font files (TTF, OTF or TTC collections) to the text
// shaper and redraws every window. Text picks them by family name from Face,
// so load a family listed there, such as Noto Sans SC. A web build needs
// this: the browser gives a WebAssembly app no system fonts, and without a
// CJK font Chinese shows as boxes. Desktop apps can use it to ship a font.
//
// Call it before window.Main or from a callback; from another goroutine wrap
// it in core.Update, like Apply.
func LoadFonts(files ...[]byte) error {
	var faces []font.FontFace
	for i, data := range files {
		f, err := opentype.ParseCollection(data)
		if err != nil {
			return fmt.Errorf("font file %d: %w", i, err)
		}
		faces = append(faces, f...)
	}
	loaded = append(loaded, faces...)
	Material.Shaper = text.NewShaper(text.WithCollection(append(loaded[:len(loaded):len(loaded)], fallbackFaces()...)))
	revision++
	loop.InvalidateAll()
	return nil
}

// NewShaper builds an independent shaper with extra faces ahead of the fonts
// loaded through LoadFonts and the Go fallback fonts. System font fallback
// follows the same platform defaults as Material.Shaper.
func NewShaper(extra ...font.FontFace) *text.Shaper {
	faces := append([]font.FontFace{}, extra...)
	faces = append(faces, loaded...)
	faces = append(faces, fallbackFaces()...)
	return text.NewShaper(text.WithCollection(faces))
}
