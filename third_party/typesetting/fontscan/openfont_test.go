//go:build unix && !js && !wasip1

package fontscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dyike/keel/third_party/typesetting/di"
	"github.com/dyike/keel/third_party/typesetting/font"
	ot "github.com/dyike/keel/third_party/typesetting/font/opentype"
	"github.com/dyike/keel/third_party/typesetting/language"
	"github.com/dyike/keel/third_party/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

// Keel patch test: system fonts parse from read-only mappings, their tables
// are slices of the mapping, and describing, parsing, shaping and glyph
// outlines never write into them (a write would fault on the read-only pages).
func TestMappedSystemFontsAreReadOnly(t *testing.T) {
	dirs, err := DefaultFontDirectories(nil)
	if err != nil {
		t.Skip("no system font directories:", err)
	}
	var files []string
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				switch strings.ToLower(filepath.Ext(path)) {
				case ".ttf", ".otf", ".ttc", ".otc":
					files = append(files, path)
				}
			}
			return nil
		})
	}
	if len(files) == 0 {
		t.Skip("no system fonts")
	}
	if testing.Short() && len(files) > 40 {
		files = files[:40]
	}
	text := []rune("Keel ff fi 你好世界 مرحبا 1234 —…")
	var shaper shaping.HarfbuzzShaper
	faces := 0
	for _, path := range files {
		res, release, err := openFont(path)
		if err != nil {
			t.Fatal(err)
		}
		release()
		shared, ok := res.(ot.Shared)
		if !ok {
			t.Fatalf("%s: not mapped", path)
		}
		loaders, err := ot.NewLoaders(shared)
		if err != nil {
			continue // not a font this parser reads
		}
		for _, ld := range loaders {
			// Describe reuses each table it reads as the buffer for the next.
			font.Describe(ld, nil)
			ft, err := font.NewFont(ld)
			if err != nil {
				continue
			}
			face := font.NewFace(ft)
			faces++
			out := shaper.Shape(shaping.Input{Text: text, RunEnd: len(text), Direction: di.DirectionLTR,
				Face: face, Size: fixed.I(16), Script: language.Latin, Language: language.NewLanguage("en")})
			for _, g := range out.Glyphs {
				face.GlyphData(g.GlyphID)
			}
		}
	}
	if faces == 0 {
		t.Skip("no faces parsed")
	}
	t.Logf("%d faces from %d files parsed, shaped and outlined from read-only mappings", faces, len(files))
}

func TestSharedTablesAreNotCopied(t *testing.T) {
	dirs, _ := DefaultFontDirectories(nil)
	for _, dir := range dirs {
		matches, _ := filepath.Glob(filepath.Join(dir, "*.tt[fc]"))
		for _, path := range matches {
			res, release, err := openFont(path)
			if err != nil {
				continue
			}
			release()
			shared, ok := res.(ot.Shared)
			if !ok {
				continue
			}
			loaders, err := ot.NewLoaders(shared)
			if err != nil || !loaders[0].HasTable(ot.MustNewTag("cmap")) {
				continue
			}
			table, err := loaders[0].RawTable(ot.MustNewTag("cmap"))
			if err != nil || len(table) == 0 {
				continue
			}
			all := shared.Bytes()
			start := &all[0]
			if p := &table[0]; uintptr(unsafePointer(p)) < uintptr(unsafePointer(start)) || uintptr(unsafePointer(p)) >= uintptr(unsafePointer(start))+uintptr(len(all)) {
				t.Fatalf("%s: cmap table was copied", path)
			}
			if cap(table) != len(table) {
				t.Fatalf("%s: table slice can grow into the mapping", path)
			}
			return
		}
	}
	t.Skip("no TrueType system font found")
}
