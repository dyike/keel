package markdown

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestMathParsingAndLiteralCode(t *testing.T) {
	src := "勾股 $a^2+b^2=c^2$，$x_i$，$\\alpha + \\beta$。\n\n$$\nx = \\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}\n$$\n\n$$\\sum_{i=1}^{n} i = \\frac{n(n+1)}{2}$$\n\n$$A = \\begin{pmatrix} 1 & 2 \\\\ 3 & 4 \\end{pmatrix}$$\n\n`$x_i$`\n\n```text\n$a^2$\n```"
	d := New(src)
	n := 0
	for _, c := range d.chunks {
		for _, b := range c.blocks {
			for _, s := range b.spans {
				if s.math != nil {
					n++
				}
			}
		}
	}
	if n != 6 {
		t.Fatalf("parsed %d formulas, want 6; %#v", n, d.chunks)
	}
	if got := d.chunks[4].blocks[0].spans[0]; !got.code || got.math != nil || got.text != "$x_i$" {
		t.Fatalf("inline code changed: %+v", got)
	}
	if got := d.chunks[5].blocks[0]; got.kind != codeBlock || got.code != "$a^2$" {
		t.Fatalf("code block changed: %+v", got)
	}
	h := uitest.New(el.Root(docView{d}))
	h.Frame()
	// Native math is an atomic selection piece, preserving TeX for copying.
	r := d.chunks[0].blocks[0].view.(*richBlock)
	for _, p := range r.rt.pieces {
		if r.runs[p.run].math != nil && (p.runes < 2 || len(p.glyphs) != 1) {
			t.Fatal("math selection lost source rune range")
		}
	}
}

func TestMathTallBoxesAndBaseline(t *testing.T) {
	d := New("前 $x_i^2$ 后\n\n$$\\frac{1}{\\sqrt{x^2+1}}$$\n\n$$\\sum_{i=1}^n i$$\n\n$$\\begin{pmatrix}1 & 22 \\\\ 333 & 4\\end{pmatrix}$$")
	_ = uitest.New(el.Root(docView{d}))
	r := d.chunks[0].blocks[0].view.(*richBlock)
	if len(r.rt.pieces) != 3 {
		t.Fatalf("inline pieces = %d", len(r.rt.pieces))
	}
	baseline := r.rt.pieces[0].baseline
	for _, p := range r.rt.pieces {
		if p.baseline != baseline {
			t.Fatal("inline math does not share text baseline")
		}
	}
	for _, c := range d.chunks[1:] {
		r := c.blocks[0].view.(*richBlock)
		if len(r.rt.pieces) != 1 {
			t.Fatalf("display math has %d pieces", len(r.rt.pieces))
		}
		p := r.rt.pieces[0]
		if p.ascent+p.descent < 30 {
			t.Fatalf("formula is flat text: %+v", p)
		}
		if p.rect.Dy() < p.ascent+p.descent {
			t.Fatal("formula exceeds line height")
		}
		if p.rect.Min.X <= 0 {
			t.Fatal("display formula not centered")
		}
	}
}

func TestMathFallbackAndStreaming(t *testing.T) {
	for _, src := range []string{`$\unsupported{x}$`, `$\frac{1}$`, `$x^{2$`, `$\begin{unknown}x\end{unknown}$`} {
		b := parse(src)
		if len(b) != 1 || plain(b[0].spans) != src {
			t.Fatalf("fallback lost source %q: %+v", src, b)
		}
		for _, s := range b[0].spans {
			if s.math != nil {
				t.Fatalf("invalid formula rendered: %q", src)
			}
		}
	}
	for _, src := range []string{`price $5 and $10`, `escaped \$x$`, "open $x^2\n\nfollowing paragraph"} {
		for _, b := range parse(src) {
			for _, s := range b.spans {
				if s.math != nil {
					t.Fatalf("literal parsed as math: %q", src)
				}
			}
		}
	}
	d := New("before $\\frac{1}{")
	d.SetStreaming(true)
	if !strings.Contains(plain(d.chunks[0].blocks[0].spans), "before") {
		t.Fatal("streaming swallowed text")
	}
	d.Append("2}$ after")
	spans := d.chunks[0].blocks[0].spans
	if len(spans) != 3 || spans[1].math == nil || spans[2].text != " after" {
		t.Fatalf("completed math not parsed: %+v", spans)
	}
}

func TestDisplayMathBlankLinesAndStreamingHealing(t *testing.T) {
	src := "before\n\n$$\n\\frac{1}\n\n{2}\n$$\n\nafter"
	d := New(src)
	if len(d.chunks) != 3 || d.chunks[1].blocks[0].spans[0].math == nil {
		t.Fatalf("blank lines split display math: %+v", d.chunks)
	}
	d = New("$a*b$ after")
	d.SetStreaming(true)
	if got := plain(d.chunks[0].blocks[0].spans); got != "$a*b$ after" {
		t.Fatalf("healing changed formula: %q", got)
	}
	d = New("$$\n\\frac{1}{")
	d.SetStreaming(true)
	if got := plain(d.chunks[0].blocks[0].spans); got != "$$\n\\frac{1}{" {
		t.Fatalf("unfinished formula lost source: %q", got)
	}
	d.Append("2}\n$$\n\nafter")
	if d.chunks[0].blocks[0].spans[0].math == nil || plain(d.chunks[1].blocks[0].spans) != "after" {
		t.Fatal("streaming math swallowed following paragraph")
	}
	if parseMath(strings.Repeat("{", 100)+"x"+strings.Repeat("}", 100)) != nil {
		t.Fatal("nesting limit not enforced")
	}
}

func TestSelectAndCopyMathSource(t *testing.T) {
	const first = "前 $x^2$ 后"
	const formula = "$$\n\\frac{1}{2}\n$$"
	d := New(first + "\n\n" + formula + "\n\n末尾")
	h := uitest.New(el.Root(docView{d}))
	x0, y0 := textPoint(t, h, d, first, 0)
	x1, y1 := textPoint(t, h, d, "末尾", 2)
	h.Drag(x0+1, y0, x1, y1)
	assertSelectionCopy(t, h, d, first+"\n\n"+formula+"\n\n末尾")
}

// A subscript should hang just below its base, not below the font's full
// descent/leading box. Check both ordinary scripts and simultaneous scripts
// at desktop and Retina scales.
func TestMathSubscriptsStayWithBase(t *testing.T) {
	for _, scale := range []float32{1, 1.5, 2} {
		uitest.NewFunc(func(gtx core.C) {
			gtx.Metric.PxPerSp, gtx.Metric.PxPerDp = scale, scale
			rn := run{size: theme.BodySize}
			em := gtx.Sp(rn.size)
			for _, source := range []string{"x_i", "x_i^2", "y_j", "H_2O"} {
				box := layoutMath(gtx, theme.Material.Shaper, parseMath(source), rn, false)
				if box.d <= 0 || box.d > em/2 {
					t.Errorf("%s at scale %g: subscript extends %dpx below baseline (em=%d)", source, scale, box.d, em)
				}
			}
		})
	}
}
