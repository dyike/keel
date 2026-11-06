package editorstyle

import (
	"bytes"
	"image"
	"testing"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/font/gofont"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget"
	"github.com/dyike/keel/typesetting/font/opentype"
	"github.com/dyike/keel/typesetting/font/opentype/tables"
)

func TestInlineFontMetricsAndOwnership(t *testing.T) {
	glyphs := []InlineGlyph{{0xf0000, 1200}, {0xe000, 600}}
	face, err := InlineFont("Keel Inline", InlineFontMetrics{1000, 800, 200}, glyphs)
	if err != nil {
		t.Fatal(err)
	}
	glyphs[0].Advance = 1
	f := face.Face.Face()
	for _, tc := range []struct {
		r     rune
		width float32
	}{{0xe000, 600}, {0xf0000, 1200}} {
		g, ok := f.NominalGlyph(tc.r)
		if !ok || f.HorizontalAdvance(g) != tc.width {
			t.Fatal("invalid glyph metrics", tc)
		}
	}
	data, index := face.Face.(interface{ NativeFontData() ([]byte, int) }).NativeFontData()
	if index != 0 || inlineChecksum(data) != 0xb1b0afba {
		t.Fatal("invalid native font data")
	}
	loader, err := opentype.NewLoader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := loader.RawTable(opentype.MustNewTag("name"))
	if err != nil {
		t.Fatal(err)
	}
	names, _, err := tables.ParseName(raw)
	if err != nil {
		t.Fatal(err)
	}
	if ps := names.Name(6); ps == "" || len(ps) > 63 {
		t.Fatal("invalid PostScript name", ps)
	}
	if _, ok := f.NominalGlyph('A'); ok {
		t.Fatal("font claimed ordinary text")
	}
	if inlineChecksum(inlineFontData(InlineFontMetrics{1000, 800, 200}, []InlineGlyph{{0xe000, 600}})) != 0xb1b0afba {
		t.Fatal("bad sfnt checksum")
	}
	for _, bad := range [][]InlineGlyph{nil, {{'A', 10}}, {{0xe000, 0}}, {{0xe000, 32768}}, {{0xe000, 10}, {0xe000, 20}}, {{0xffffe, 10}}} {
		if _, err := InlineFont("Keel Inline", InlineFontMetrics{1000, 800, 200}, bad); err == nil {
			t.Fatal("accepted invalid glyphs", bad)
		}
	}
}

func TestInlineFontUsesEditorWrappingAndSelectionGeometry(t *testing.T) {
	testInlineFontEditorGeometry(t, false)
}

func testInlineFontEditorGeometry(t *testing.T, platform bool) {
	t.Helper()
	for _, scale := range []int{1, 2} {
		face, err := InlineFont("Keel Inline", InlineFontMetrics{1000, 800, 200}, []InlineGlyph{{0xe000, 3750}}) // 60px at 16sp
		if err != nil {
			t.Fatal(err)
		}
		option := text.NoSystemFonts()
		if platform {
			option = text.WithPlatformFonts()
		}
		shaper := text.NewShaper(option, text.WithCollection(append([]font.FontFace{face}, gofont.Collection()...)))
		var ed widget.Editor
		ed.SetText("AA \ue000")
		var ops op.Ops
		gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}, Constraints: layout.Constraints{Max: image.Pt(80*scale, 300*scale)}}
		rec := op.Record(&ops)
		ink := rec.Stop()
		ed.Layout(gtx, shaper, font.Font{Typeface: "Keel Inline, Go"}, 16, ink, ink)
		before := ed.Regions(0, 2, nil)
		obj := ed.Regions(3, 4, nil)
		if len(before) != 1 || len(obj) != 1 {
			t.Fatal("object split", before, obj)
		}
		if obj[0].Bounds.Dx() != 60*scale || obj[0].Bounds.Min.Y <= before[0].Bounds.Min.Y {
			t.Fatal("object did not wrap as one unit", scale, before, obj)
		}
		ed.SetCaret(3, 3)
		ed.MoveCaret(1, 1)
		a, b := ed.Selection()
		if a != 4 || b != 4 {
			t.Fatal("object was not a single caret step", a, b)
		}
	}
}
