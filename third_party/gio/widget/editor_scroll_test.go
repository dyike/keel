// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image"
	"strings"
	"testing"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
)

func TestEditorViewportPreservesSelectionAndResumesCaretScrolling(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		ed := new(Editor)
		source := strings.Repeat("line of text\n", 40)
		ed.SetText(source)
		sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
		gtx := layout.Context{Ops: new(op.Ops), Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale}, Constraints: layout.Exact(image.Pt(int(200*scale), int(80*scale)))}
		frame := func() { gtx.Ops.Reset(); ed.Layout(gtx, sh, font.Font{}, 14, op.CallOp{}, op.CallOp{}) }
		frame()
		bound := ed.ScrollBounds().Max.Y
		if bound <= 0 {
			t.Fatal("missing overflow")
		}
		ed.SetCaret(0, 3)
		ed.ScrollTo(image.Pt(0, bound+100))
		frame()
		start, end := ed.Selection()
		if ed.ScrollOffset().Y != bound || start != 0 || end != 3 || ed.Text() != source {
			t.Fatal("manual scroll changed editor state")
		}
		ed.SetCaret(0, 0)
		frame()
		if ed.ScrollOffset().Y != 0 {
			t.Fatal("caret no longer scrolls into view")
		}
		ed.SetCaret(ed.Len(), ed.Len())
		frame()
		if ed.ScrollOffset().Y < bound-2 {
			t.Fatal("end caret is offscreen")
		}
		ed.SetText("short")
		frame()
		if ed.ScrollOffset().Y != 0 {
			t.Fatal("short text retained old offset")
		}
	}
}
