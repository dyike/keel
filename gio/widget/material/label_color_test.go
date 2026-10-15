package material

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/font/gofont"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget"
)

func TestLabelStyleColorAndSelectableLifetime(t *testing.T) {
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Regular()))
	var got color.NRGBA
	sh.SetBorrowedRasterHook(func(_ *op.Ops, _ []text.Glyph, col color.NRGBA, _ float32) bool { got = col; return true })
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(300, 50)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	for _, selectable := range []bool{false, true} {
		l := LabelStyle{Text: "Hello", TextSize: 13, Shaper: sh}
		if selectable {
			l.State = new(widget.Selectable)
		}
		for _, col := range []color.NRGBA{{R: 42, A: 255}, {}} {
			ops.Reset()
			l.Color = col
			got = color.NRGBA{A: 123}
			l.Layout(gtx)
			if got != col {
				t.Fatalf("selectable=%v color=%v want %v", selectable, got, col)
			}
			if selectable {
				if l.State.Color == nil || *l.State.Color != col {
					t.Fatal("public color pointer lost")
				}
				l.Color = color.NRGBA{B: 12}
				if *l.State.Color != col {
					t.Fatal("selectable retained temporary LabelStyle")
				}
				l.State.SetCaret(0, 5)
				if l.State.SelectedText() != "Hello" {
					t.Fatal("selection changed")
				}
			}
		}
		limit := float64(0)
		if selectable {
			limit = 1
		} // Existing textView string-copy allocation.
		if n := testing.AllocsPerRun(50, func() { ops.Reset(); l.Layout(gtx) }); n > limit {
			t.Fatalf("selectable=%v warmed allocations=%g, limit=%g", selectable, n, limit)
		}
	}
}

func BenchmarkLabelStyleColor(b *testing.B) {
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Regular()))
	sh.SetBorrowedRasterHook(func(_ *op.Ops, _ []text.Glyph, _ color.NRGBA, _ float32) bool { return true })
	for _, selectable := range []bool{false, true} {
		name := "label"
		if selectable {
			name = "selectable"
		}
		b.Run(name, func(b *testing.B) {
			var ops op.Ops
			gtx := layout.Context{Ops: &ops, Constraints: layout.Exact(image.Pt(300, 50)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
			l := LabelStyle{Text: "Hello", TextSize: 13, Shaper: sh, Color: color.NRGBA{A: 255}}
			if selectable {
				l.State = new(widget.Selectable)
			}
			for range 10 {
				ops.Reset()
				l.Layout(gtx)
			}
			b.ReportAllocs()
			for b.Loop() {
				ops.Reset()
				l.Layout(gtx)
			}
		})
	}
}
