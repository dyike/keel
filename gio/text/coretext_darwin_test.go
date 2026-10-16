//go:build darwin && !ios

// SPDX-License-Identifier: Unlicense OR MIT
package text

import (
	"os"
	"runtime"
	"slices"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	giofont "github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/font/opentype"
	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/typesetting/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

func TestNativeFontShaping(t *testing.T) {
	sh := NewShaper(WithPlatformFonts())
	if sh.shaper.native == nil {
		t.Fatal("native font engine unavailable")
	}
	if sh.shaper.fontMap != nil {
		t.Fatal("native path loaded the Go font index")
	}
	for _, sample := range []string{"", "ABC ligature office", "中文终端 ⠋⠙⠹ ┌─┐ █ ░ ⚡", "مرحبا بالعالم", "שלום 123 abc", "हिन्दी", "á é 🧑‍💻 🇨🇳 👍🏽", "line one\n第二行\n"} {
		for _, width := range []int{58, 1000} {
			for _, maxLines := range []int{0, 1} {
				params := Parameters{Font: giofont.Font{Typeface: "Menlo, Hiragino Sans GB, Apple Symbols, Apple Color Emoji"}, PxPerEm: fixed.I(18), MaxWidth: width, MaxLines: maxLines, Locale: system.Locale{Direction: system.LTR}}
				sh.LayoutString(params, sample)
				count := 0
				var gs []Glyph
				for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
					count += int(g.Runes)
					gs = append(gs, g)
					if g.Ascent <= 0 || g.Descent < 0 {
						t.Fatalf("%q invalid metrics %+v", sample, g)
					}
					if g.Flags&(FlagParagraphBreak|FlagTruncator) == 0 && func() bool { _, _, gid := splitGlyphID(g.ID); return gid == 0 }() && g.Advance > 0 {
						t.Errorf("%q missing glyph", sample)
					}
				}
				if count != utf8.RuneCountInString(sample) {
					t.Errorf("%q width %d lines %d rune accounting: got %d want %d", sample, width, maxLines, count, utf8.RuneCountInString(sample))
				}
				var ops op.Ops
				sh.Shape(gs)
				sh.Bitmaps(gs).Add(&ops)
			}
		}
	}
	for _, f := range sh.shaper.faces {
		if f.Cmap != nil {
			t.Fatal("native path parsed Go font tables")
		}
	}
}

func TestNativeFontMemoryCollectionAndIdentity(t *testing.T) {
	face, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	sh := NewShaper(WithPlatformFonts(), WithCollection([]FontFace{{Font: face.Font(), Face: face}}))
	if sh.shaper.native == nil {
		t.Fatal("memory font forced Go fallback")
	}
	sh.LayoutString(Parameters{Font: face.Font(), PxPerEm: fixed.I(19), MaxWidth: 1000}, "Hello")
	g, ok := sh.NextGlyph()
	if !ok {
		t.Fatal("empty layout")
	}
	native, size, _, ok := sh.GlyphNativeFont(g.ID)
	if !ok || size != fixed.I(19) {
		t.Fatal("missing native identity")
	}
	name := nativeInit.fontName(native)
	defer ct.release(name)
	if got := ctString(name); got != "GoRegular" {
		t.Fatalf("memory font identity %q", got)
	}
	// NoSystemFonts means no native system fallback, even when requested.
	legacy := NewShaper(NoSystemFonts(), WithPlatformFonts(), WithCollection([]FontFace{{Font: face.Font(), Face: face}}))
	if legacy.shaper.native != nil || legacy.shaper.fontMap == nil {
		t.Fatal("NoSystemFonts was ignored")
	}
}

func TestNativeFontFileRoundTrip(t *testing.T) {
	sh := NewShaper(WithPlatformFonts())
	for _, family := range []string{"Menlo", "Helvetica Neue", "Hiragino Sans GB", "Apple Color Emoji"} {
		sh.LayoutString(Parameters{Font: giofont.Font{Typeface: giofont.Typeface(family)}, PxPerEm: fixed.I(22), MaxWidth: 1000}, "A中😀")
		for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
			native, _, _, ok := sh.GlyphNativeFont(g.ID)
			if !ok {
				t.Fatal("missing native font")
			}
			path, index, instance, size, gid, ok := sh.GlyphFileInstance(g.ID)
			if !ok {
				t.Fatalf("native font lost its file: family=%s glyph=%v", family, g)
			}
			if size != fixed.I(22) || gid == 0 || path == "" || index < 0 || instance < 0 {
				t.Fatal("bad file identity")
			}
			_ = native
		}
	}
}

// This measures the allocation site the optimization removes. It is not a
// physical-footprint measurement and does not force GC in a rendering loop.
func BenchmarkSystemFontShaping(b *testing.B) {
	for _, native := range []bool{false, true} {
		name := "Go"
		if native {
			name = "CoreText"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var options []ShaperOption
				if native {
					options = append(options, WithPlatformFonts())
				}
				sh := NewShaper(options...)
				sh.LayoutString(Parameters{Font: giofont.Font{Typeface: "Menlo, Hiragino Sans GB, Apple Symbols, Apple Color Emoji"}, PxPerEm: fixed.I(18), MaxWidth: 1000}, "abc 中文终端 ⠋⠙⠹ ┌─┐ ⚡ 😀")
				for _, ok := sh.NextGlyph(); ok; _, ok = sh.NextGlyph() {
				}
				runtime.KeepAlive(sh)
			}
		})
	}
}

// A custom Face may carry non-default variable coordinates; FontData is an
// explicit immutable-source opt-in. A file/family guess must not replace it.
type opaqueFace struct{ f *font.Face }

func (f opaqueFace) Face() *font.Face { return f.f }
func TestNativeOpaqueFaceKeepsGoShaper(t *testing.T) {
	face, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	custom := opaqueFace{face.Face()}
	sh := NewShaper(WithPlatformFonts(), WithCollection([]FontFace{{Font: face.Font(), Face: custom}}))
	if sh.shaper.native != nil {
		t.Fatal("opaque font identity was guessed")
	}
	sh.LayoutString(Parameters{Font: face.Font(), PxPerEm: fixed.I(18), MaxWidth: 1000}, "custom")
	if _, ok := sh.NextGlyph(); !ok {
		t.Fatal("Go fallback did not shape")
	}
}

func TestNativeSourcesSharedWithIndependentShaperState(t *testing.T) {
	face, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	collection := []FontFace{{Font: face.Font(), Face: face}}
	a := NewShaper(WithPlatformFonts(), WithCollection(collection))
	b := NewShaper(WithPlatformFonts(), WithCollection(collection))
	nativeA, nativeB := a.shaper.native.(*ctShaper), b.shaper.native.(*ctShaper)
	if len(nativeA.sourceOwners) != 1 || nativeA.sourceOwners[0] != nativeB.sourceOwners[0] {
		t.Fatal("immutable font source was copied per shaper")
	}
	for i, sh := range []*Shaper{a, b} {
		sh.LayoutString(Parameters{Font: face.Font(), PxPerEm: fixed.I(16 + i), MaxWidth: 1000}, "shared tables")
	}
	if nativeA.faces[0].face == nativeB.faces[0].face {
		t.Fatal("mutable Face scratch was shared between shapers")
	}
	ga, _ := a.NextGlyph()
	gb, _ := b.NextGlyph()
	fa, sa, _, _ := a.GlyphNativeFont(ga.ID)
	fb, sb, _, _ := b.GlyphNativeFont(gb.ID)
	if fa == fb || sa == sb {
		t.Fatal("sized native font identities were merged")
	}
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
}

func TestNativeSystemFontStyles(t *testing.T) {
	sh := NewShaper(WithPlatformFonts())
	for _, family := range []giofont.Typeface{"system-ui", "monospace"} {
		var regular, bold uintptr
		for i, weight := range []giofont.Weight{giofont.Normal, giofont.Bold} {
			sh.LayoutString(Parameters{Font: giofont.Font{Typeface: family, Weight: weight}, PxPerEm: fixed.I(18), MaxWidth: 1000}, "system weight")
			g, _ := sh.NextGlyph()
			native, _, _, ok := sh.GlyphNativeFont(g.ID)
			if !ok {
				t.Fatal("missing native system font")
			}
			if i == 0 {
				regular = native
			} else {
				bold = native
			}
		}
		if ct.equal(regular, bold) {
			t.Fatalf("%s ignored requested weight", family)
		}
	}
}

type oldFileFace struct {
	face  *font.Face
	path  string
	index int
}

func (f oldFileFace) Face() *font.Face      { return f.face }
func (f oldFileFace) Source() (string, int) { return f.path, f.index }
func TestNativeLegacyFileCollectionReleasesGoTables(t *testing.T) {
	data, err := os.ReadFile("/System/Library/Fonts/Menlo.ttc")
	if err != nil {
		t.Fatal(err)
	}
	faces, err := opentype.ParseCollection(data)
	if err != nil {
		t.Fatal(err)
	}
	parsed := faces[0].Face.Face()
	collection := []FontFace{{Font: faces[0].Font, Face: oldFileFace{parsed, "/System/Library/Fonts/Menlo.ttc", 0}}}
	sh := NewShaper(WithPlatformFonts(), WithCollection(collection))
	if sh.shaper.native == nil {
		t.Fatal("file-backed collection did not use native shaping")
	}
	if sh.config.collection != nil {
		t.Fatal("native shaper retained parsed Go collection")
	}
	if sh.shaper.fontMap != nil {
		t.Fatal("legacy file source loaded the Go index")
	}
	sh.LayoutString(Parameters{Font: faces[0].Font, PxPerEm: fixed.I(20), MaxWidth: 1000}, "Legacy file source")
	if _, ok := sh.NextGlyph(); !ok {
		t.Fatal("native shaping was empty")
	}
	for _, f := range sh.shaper.faces {
		if f.Cmap != nil {
			t.Fatal("parsed Go font tables were retained")
		}
	}
}

// Embedding a regular OpenType source must not accidentally opt a wrapper
// with a different Face implementation into native default coordinates.
type embeddedSource = opentype.Face
type embeddedCustomFace struct{ embeddedSource }

func (f embeddedCustomFace) Face() *font.Face { return f.embeddedSource.Face() }
func TestNativeEmbeddedCustomFaceKeepsGoShaper(t *testing.T) {
	source, err := opentype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	wrapper := embeddedCustomFace{source}
	sh := NewShaper(WithPlatformFonts(), WithCollection([]FontFace{{Font: source.Font(), Face: wrapper}}))
	if sh.shaper.native != nil {
		t.Fatal("inherited FontData discarded the wrapper's Face implementation")
	}
}

// Compare the adapter's actual draw positions with an independent CTLine,
// rather than comparing two renderers that consume the same adapted glyphs.
func TestNativePositionsMatchCoreText(t *testing.T) {
	sh := NewShaper(WithPlatformFonts())
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := ct.poolPush()
	defer ct.poolPop(pool)
	engine := sh.shaper.native.(*ctShaper)
	type positioned struct {
		font uint
		gid  uint16
		x, y fixed.Int26_6
	}
	for _, direction := range []system.TextDirection{system.LTR, system.RTL} {
		for _, value := range []string{"office AV á é", "abc مرحبا 123 שלום 中文", "हिन्दी 🧑‍💻 🇨🇳 á", "abc \u2067مرحبا 123\u2069 שלום 中文", "a(אבג 123)Z"} {
			params := Parameters{Font: giofont.Font{Typeface: "Menlo, Hiragino Sans GB, Apple Color Emoji"}, PxPerEm: fixed.I(23), MaxWidth: 10000, DisableSpaceTrim: true, Locale: system.Locale{Direction: direction}}
			sh.LayoutString(params, value)
			alignment := params.Alignment.Align(direction, sh.txt.lines[0].width, sh.txt.alignWidth)
			var got []positioned
			for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
				native, _, gid, ok := sh.GlyphNativeFont(g.ID)
				if !ok {
					t.Fatal("missing native identity")
				}
				got = append(got, positioned{ct.hash(native), gid, g.X - g.Offset.X - alignment, g.Offset.Y})
			}
			font := engine.primaryFont(params.PxPerEm)
			u16 := utf16.Encode([]rune(value))
			str := ct.stringWithCharacters(0, &u16[0], len(u16))
			dir := 0
			if direction == system.RTL {
				dir = 1
			}
			attrs := ctDict([]uintptr{ct.fontAttributeName, ct.paragraphStyleName}, []uintptr{font, engine.styles[dir]})
			attributed := ct.attributedString(0, str, attrs)
			setter := ct.typesetterCreate(attributed)
			line := ct.typesetterCreateLine(setter, cfRange{0, len(u16)})
			runs := ct.lineGetGlyphRuns(line)
			var want []positioned
			for i := 0; i < ct.arrayGetCount(runs); i++ {
				run := ct.arrayGetValueAtIndex(runs, i)
				n := ct.runGetGlyphCount(run)
				if n == 0 {
					continue
				}
				ids := make([]uint16, n)
				positions := make([]cgPoint, n)
				ct.runGetGlyphs(run, cfRange{}, &ids[0])
				ct.runGetPositions(run, cfRange{}, &positions[0])
				f := ct.dictionaryGetValue(ct.runGetAttributes(run), ct.fontAttributeName)
				for j := range n {
					want = append(want, positioned{ct.hash(f), ids[j], ctFixed(positions[j].x), ctFixed(positions[j].y)})
				}
			}
			for _, p := range []uintptr{line, setter, attributed, attrs, str} {
				ct.release(p)
			}
			less := func(a, b positioned) int {
				if a.font < b.font {
					return -1
				}
				if a.font > b.font {
					return 1
				}
				if a.gid != b.gid {
					return int(a.gid) - int(b.gid)
				}
				if a.x != b.x {
					return int(a.x - b.x)
				}
				return int(a.y - b.y)
			}
			slices.SortFunc(got, less)
			slices.SortFunc(want, less)
			if len(got) != len(want) {
				t.Fatalf("%q glyph count %d want %d", value, len(got), len(want))
			}
			for i, g := range got {
				w := want[i]
				dx, dy := int(g.x-w.x), int(g.y-w.y)
				if g.font != w.font || g.gid != w.gid || max(dx, -dx) > 1 || max(dy, -dy) > 1 {
					t.Fatalf("dir=%v %q glyph %d: got %+v want %+v", direction, value, i, g, w)
				}
			}
		}
	}
	runtime.KeepAlive(sh)
}

func TestNativeZeroSize(t *testing.T) {
	sh := NewShaper(WithPlatformFonts())
	sh.LayoutString(Parameters{MaxWidth: 100}, "zero 中文")
	n := 0
	for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
		n += int(g.Runes)
		if g.Advance != 0 || g.Ascent != 0 || g.Descent != 0 || g.Bounds != (fixed.Rectangle26_6{}) {
			t.Fatalf("zero PPEM took CoreText's default size: %+v", g)
		}
	}
	if n != 7 {
		t.Fatalf("zero-size rune accounting %d", n)
	}
}
