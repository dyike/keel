// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	gotext "github.com/go-text/typesetting/font"
)

// sourcedFace is a Go font face reporting a file, as Keel's file-backed
// faces do.
type sourcedFace struct{ inner font.Face }

func (f sourcedFace) Face() *gotext.Face  { return f.inner.Face() }
func (sourcedFace) Source() (string, int) { return "/fonts/go.ttc", 2 }

// Keel patch: a raster hook draws solid labels; GlyphFile identifies the file.
func TestLabelRasterHookAndGlyphFile(t *testing.T) {
	face := gofont.Regular()[0]
	face.Face = sourcedFace{face.Face}
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection([]font.FontFace{face}))
	var calls int
	var got color.NRGBA
	var frac float32
	var file string
	sh.SetRasterHook(func(ops *op.Ops, gs []text.Glyph, col color.NRGBA, fracX float32) bool {
		calls++
		got, frac = col, fracX
		path, index, ppem, _, ok := sh.GlyphFile(gs[0].ID)
		if ok && index == 2 && ppem > 0 {
			file = path
		}
		return true
	})
	col := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Exact(image.Pt(300, 50)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	Label{Alignment: text.Middle, Color: &col}.Layout(gtx, sh, face.Font, 13, "Hello", op.CallOp{})
	if calls == 0 || got != col {
		t.Fatalf("hook calls=%d color=%v", calls, got)
	}
	if frac < 0 || frac >= 1 {
		t.Fatalf("the hook got %v pixels of fraction; lines must translate by whole pixels", frac)
	}
	if file != "/fonts/go.ttc" {
		t.Fatalf("GlyphFile reported %q", file)
	}
	// Without a color the material is opaque to the hook: outlines are used.
	calls = 0
	Label{}.Layout(gtx, sh, face.Font, 13, "Hello", op.CallOp{})
	if calls != 0 {
		t.Fatal("a label without a solid color used the hook")
	}
	// A shaper without a hook reports no file for built-in faces.
	plain := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Regular()))
	plain.LayoutString(text.Parameters{Font: face.Font, PxPerEm: 13 << 6, MaxWidth: 1000}, "H")
	g, _ := plain.NextGlyph()
	if _, _, _, _, ok := plain.GlyphFile(g.ID); ok {
		t.Fatal("a built-in face reported a file")
	}
}
