package theme

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/font/opentype"
	"github.com/dyike/keel/gio/text"
	gotext "github.com/dyike/keel/typesetting/font"
	ot "github.com/dyike/keel/typesetting/font/opentype"

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
	// LoadFonts owns immutable copies, unlike LoadFontsWhere's borrowed data.
	owned := make([][]byte, len(files))
	for i, data := range files {
		owned[i] = bytes.Clone(data)
	}
	faces, err := parseFaces(nil, owned)
	if err != nil {
		return err
	}

	addFaces(faces)
	return nil
}

// LoadFontsWhere is LoadFonts for the faces keep accepts, and parses only
// those. The faces refer to files instead of copying their tables, so files
// must not change afterwards; mapped from disk (LoadFontFilesWhere), those
// tables stay clean pages the system can drop and read back. A collection
// such as PingFang.ttc holds two dozen faces of which an app draws one or
// two, so parsing only the faces kept also saves their parsed glyph data.
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
	return parseFontFaces(keep, files, nil)
}

func parseFontFaces(keep func(font.Font) bool, files [][]byte, paths []string) ([]font.FontFace, error) {
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
			for index, ld := range lds {
				desc, _ := gotext.Describe(ld, nil)
				f := opentype.DescriptionToFont(desc)
				if keep != nil && !keep(f) {
					continue
				}
				face := &parsedFace{loader: ld, data: data, index: index}
				if len(paths) > i {
					face.path = paths[i]
				}
				if platformFontPreparation {
					err = gotext.Validate(ld)
				} else {
					face.f, err = gotext.NewFont(ld)
				}
				if err != nil {
					errs[i] = fmt.Errorf("font file %d: %w", i, err)
					return
				}
				perFile[i] = append(perFile[i], font.FontFace{Font: f, Face: face})
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
	faces, err := parseFontFaces(keep, files, paths)
	if err != nil {
		return nil, err
	}
	sh := text.NewShaper(text.WithPlatformFonts(), text.WithCollection(append(faces[:len(faces):len(faces)], fallbackFaces()...)))
	installPlatformText(sh)
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
// mapped for the process lifetime. On macOS, preparation reads metadata;
// CoreText owns glyph and shaping tables. The Go fallback parses on first use.
func LoadFontFilesWhere(keep func(font.Font) bool, paths ...string) error {
	files := make([][]byte, 0, len(paths))
	for _, path := range paths {
		b, err := mapFile(path)
		if err != nil {
			return err
		}
		files = append(files, b)
	}
	faces, err := parseFontFaces(keep, files, paths)
	if err != nil {
		return err
	}
	addFaces(faces)
	return nil
}

// fontBytes is a font file's bytes with a reader over them. Keel's go-text
// (typesetting) reads the tables of a resource with a Bytes
// method in place instead of copying them; upstream go-text ignores it.
// The bytes must not change afterwards.
type fontBytes struct {
	*bytes.Reader
	data []byte
}

func (f fontBytes) Bytes() []byte { return f.data }

// parsedFace is a face LoadFontsWhere parsed.
type parsedFace struct {
	once   sync.Once
	f      *gotext.Font
	loader *ot.Loader
	data   []byte
	path   string
	index  int
}

func (p *parsedFace) Face() *gotext.Face {
	p.once.Do(func() {
		if p.f == nil {
			p.f, _ = gotext.NewFont(p.loader)
		}
	})
	if p.f == nil {
		return nil
	}
	return gotext.NewFace(p.f)
}
func (p *parsedFace) NativeFontData() ([]byte, int) { return p.data, p.index }
func (p *parsedFace) Source() (string, int)         { return p.path, p.index }

func addFaces(faces []font.FontFace) {
	loaded = append(loaded, faces...)
	Material.Shaper = text.NewShaper(text.WithPlatformFonts(), text.WithCollection(append(loaded[:len(loaded):len(loaded)], fallbackFaces()...)))
	installPlatformText(Material.Shaper)
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
	sh := text.NewShaper(text.WithPlatformFonts(), text.WithCollection(faces))
	installPlatformText(sh)
	return sh
}
