package theme

import (
	"bytes"
	"fmt"
	"sync"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/text"
	gotext "github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"

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
	addFaces(faces)
	return nil
}

// LoadFontsWhere is LoadFonts for the faces keep accepts, and parses only
// those. The faces may keep referring to files, which must not change. A face parsed holds its tables in memory, megabytes for a CJK one,
// and a collection such as PingFang.ttc holds two dozen faces of which an
// app draws one or two: loading all of it costs hundreds of megabytes.
func LoadFontsWhere(keep func(font.Font) bool, files ...[]byte) error {
	faces, err := parseFaces(keep, files)
	if err != nil {
		return err
	}
	addFaces(faces)
	return nil
}

// parseFaces parses the faces of files keep accepts, a file per goroutine.
// Safe from any goroutine.
func parseFaces(keep func(font.Font) bool, files [][]byte) ([]font.FontFace, error) {
	perFile := make([][]font.FontFace, len(files))
	errs := make([]error, len(files))
	var wg sync.WaitGroup
	for i, data := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lds, err := ot.NewLoaders(fontBytes{bytes.NewReader(data), data})
			if err != nil {
				errs[i] = fmt.Errorf("font file %d: %w", i, err)
				return
			}
			for _, ld := range lds {
				desc, _ := gotext.Describe(ld, nil)
				f := opentype.DescriptionToFont(desc)
				if !keep(f) {
					continue
				}
				parsed, err := gotext.NewFont(ld)
				if err != nil {
					errs[i] = fmt.Errorf("font file %d: %w", i, err)
					return
				}
				perFile[i] = append(perFile[i], font.FontFace{Font: f, Face: parsedFace{parsed}})
			}
		}()
	}
	wg.Wait()
	var faces []font.FontFace
	for i := range files {
		if errs[i] != nil {
			return nil, errs[i]
		}
		faces = append(faces, perFile[i]...)
	}
	return faces, nil
}

// FontSet is fonts parsed, with a text shaper made for them, ready to put to
// use. Parsing fonts and making a shaper, which reads the system's font index,
// takes tens of milliseconds: prepare the set on another goroutine while the
// window opens, and Use it before the first frame.
type FontSet struct {
	faces  []font.FontFace
	shaper *text.Shaper
}

// PrepareFontFiles parses the faces keep accepts of the font files at paths,
// mapped as LoadFontFilesWhere maps them, and makes a shaper for them and the
// Go fonts. It touches no shared state: call it from any goroutine.
func PrepareFontFiles(keep func(font.Font) bool, paths ...string) (*FontSet, error) {
	files := make([][]byte, 0, len(paths))
	for _, path := range paths {
		b, err := mapFile(path)
		if err != nil {
			return nil, err
		}
		files = append(files, b)
	}
	faces, err := parseFaces(keep, files)
	if err != nil {
		return nil, err
	}
	sh := text.NewShaper(text.WithCollection(append(faces[:len(faces):len(faces)], fallbackFaces()...)))
	return &FontSet{faces: faces, shaper: sh}, nil
}

// Use adds the set's fonts, as LoadFonts would, and redraws every window.
// Call it before window.Main or from a callback; from another goroutine wrap
// it in core.Update. Its shaper serves when no fonts were loaded before it;
// otherwise a new one is made with them all.
func (s *FontSet) Use() {
	if len(loaded) > 0 {
		addFaces(s.faces)
		return
	}
	loaded = append(loaded, s.faces...)
	Material.Shaper = s.shaper
	revision++
	loop.InvalidateAll()
}

// LoadFontFilesWhere is LoadFontsWhere for font files by path. A file is
// mapped into memory rather than read where the system allows, and stays
// mapped for the process lifetime. Upstream typesetting v0.3.5 still copies
// font tables and eagerly parses outlines; mapping the source does not eliminate
// those heap allocations.
func LoadFontFilesWhere(keep func(font.Font) bool, paths ...string) error {
	files := make([][]byte, 0, len(paths))
	for _, path := range paths {
		b, err := mapFile(path)
		if err != nil {
			return err
		}
		files = append(files, b)
	}
	return LoadFontsWhere(keep, files...)
}

// fontBytes retains the source bytes alongside its reader. The Bytes method
// is not consumed by upstream typesetting v0.3.5; that loader copies tables.
// LoadFontsWhere's source files must not change afterwards.
type fontBytes struct {
	*bytes.Reader
	data []byte
}

func (f fontBytes) Bytes() []byte { return f.data }

// parsedFace is a face LoadFontsWhere parsed.
type parsedFace struct{ f *gotext.Font }

func (p parsedFace) Face() *gotext.Face { return gotext.NewFace(p.f) }

func addFaces(faces []font.FontFace) {
	loaded = append(loaded, faces...)
	Material.Shaper = text.NewShaper(text.WithCollection(append(loaded[:len(loaded):len(loaded)], fallbackFaces()...)))
	revision++
	loop.InvalidateAll()
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
