package el

import (
	"image"
	"strconv"
	"testing"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
)

func TestTextMeasurementMatchesGioLabel(t *testing.T) {
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	th := material.NewTheme()
	th.Shaper = sh
	e := engine{}
	for _, value := range []string{"", "Hello gy", "one\ntwo\n", "a long sentence that wraps to multiple lines", "مرحبا", "中文", "\n"} {
		for _, width := range []int{0, 35, 300} {
			for _, height := range []int{0, 30, inf} {
				for _, lines := range []int{0, 1, 2} {
					for _, align := range []text.Alignment{text.Start, text.Middle, text.End} {
						var ops op.Ops
						ctx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 2, PxPerSp: 2}, Constraints: layout.Constraints{Max: image.Pt(width, height)}}
						lb := material.Label(th, 12, value)
						lb.Font = font.Font{Typeface: "Go"}
						lb.MaxLines = lines
						lb.Alignment = align
						got := e.measureLabel(ctx, lb)
						want := lb.Layout(ctx)
						if got != want {
							t.Fatalf("%q %dx%d lines=%d align=%v: got %v want %v", value, width, height, lines, align, got, want)
						}
						if again := e.measureLabel(ctx, lb); again != want {
							t.Fatal("cached dimensions differ")
						}
					}
				}
			}
		}
	}
	// A distinct locale, metric, width or shaper must not reuse old dimensions.
	var ops op.Ops
	ctx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(60, inf)}, Locale: system.Locale{Direction: system.RTL}}
	lb := material.Label(th, 20, "Hello again")
	lb.Font = font.Font{Typeface: "Go"}
	if got, want := e.measureLabel(ctx, lb), lb.Layout(ctx); got != want {
		t.Fatalf("changed context: %v != %v", got, want)
	}
}

func BenchmarkTextMeasurement(b *testing.B) {
	for _, name := range []string{"GioLayout", "KeelMeasure"} {
		b.Run(name, func(b *testing.B) {
			th := material.NewTheme()
			th.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
			var ops op.Ops
			ctx := layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(180, inf)}}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ops.Reset()
				lb := material.Label(th, 12, "R·"+strconv.Itoa(i))
				lb.Font = font.Font{Typeface: "Go"}
				e := engine{}
				// Flex sizing revisits a label with the same constraints in one frame.
				for pass := 0; pass < 3; pass++ {
					if name == "GioLayout" {
						lb.Layout(ctx)
					} else {
						e.measureLabel(ctx, lb)
					}
				}
			}
		})
	}
}

func TestTextMeasurementCacheDoesNotRetainFrames(t *testing.T) {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	var ops op.Ops
	ctx := layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(180, inf)}}
	e := engine{}
	for frame := 0; frame < 100; frame++ {
		e.beginTextMeasurements()
		if len(e.textMeasurements) != 0 {
			t.Fatal("previous frame retained")
		}
		for i := 0; i < 8; i++ {
			e.measureLabel(ctx, material.Label(th, 12, strconv.Itoa(frame*8+i)))
		}
		if len(e.textMeasurements) != 8 {
			t.Fatal("unexpected cache size", len(e.textMeasurements))
		}
	}
}

func TestRootClearsTextMeasurementCache(t *testing.T) {
	value := "first label"
	root := Root(ViewFunc(func(*Context) Element { return Text(value) }))
	var ops op.Ops
	ctx := core.C{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(200, 100)}}
	root.Layout(ctx)
	value = "second label"
	ops.Reset()
	root.Layout(ctx)
	for key := range root.e.textMeasurements {
		if key.value == "first label" {
			t.Fatal("RootWidget retained previous frame text")
		}
	}
}
