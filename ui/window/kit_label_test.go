package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestLabelHighlightAndMask(t *testing.T) {
	v := kit.Label("MMMM MMMM").Highlights("MMMM").HighlightColor(color.NRGBA{R: 255, A: 255}).Style(func(e *el.TextEl) { e.TextSize(28).TextColor(color.NRGBA{G: 255, A: 255}) })
	w := openTest(t, kitPage(v))
	red := func() int {
		t.Helper()
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
			for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
				c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if c.R > 180 && c.G < 80 && c.B < 80 {
					count++
				}
			}
		}
		return count
	}
	all := red()
	if all < 30 {
		t.Fatal("highlight missing", all)
	}
	v.HighlightPrefix("MMMM")
	prefix := red()
	if prefix < 10 || prefix >= all {
		t.Fatal("prefix did not isolate first match", all, prefix)
	}
	v.Masked(true).Secondary("private")
	if red() != 0 {
		t.Fatal("masked text still highlighted")
	}
	seen := false
	for _, e := range w.snapshot() {
		if strings.Contains(e.Name, "MMMM") {
			t.Fatal("unmasked semantics", e)
		}
		if strings.Contains(e.Name, "••••••••• private") {
			seen = true
		}
	}
	if !seen {
		t.Fatal("masked label missing")
	}
	v.Masked(false).Highlights("")
	if red() != 0 {
		t.Fatal("cleared highlight remains")
	}
}

func TestLabelRangesPreserveWrapping(t *testing.T) {
	for _, text := range []string{"office é 中文自动换行文字 repeated repeated", "Hello שלום مرحبا world"} {
		t.Run(text, func(t *testing.T) {
			tinted := false
			c := color.NRGBA{R: 80, G: 100, B: 120, A: 255}
			root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
				v := el.Text(text).TextSize(24).TextColor(c).W(el.Dp(130))
				if tinted {
					v.Ranges(el.TextRange{Start: 0, End: 1000, Color: c})
				}
				return el.Div().Child(v)
			}))
			w := openTest(t, Options{Width: 160, Height: 300, Content: core.Func(func(gtx core.C) core.D { return root.Layout(gtx) })})
			first, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			tinted = true
			second, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			a, err := png.Decode(bytes.NewReader(first))
			if err != nil {
				t.Fatal(err)
			}
			b, err := png.Decode(bytes.NewReader(second))
			if err != nil {
				t.Fatal(err)
			}
			for y := 0; y < a.Bounds().Max.Y; y++ {
				for x := 0; x < a.Bounds().Max.X; x++ {
					if a.At(x, y) != b.At(x, y) {
						t.Fatalf("range painter changed layout at %d,%d", x, y)
					}
				}
			}
		})
	}
}
