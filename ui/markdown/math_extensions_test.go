package markdown

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestMathMacroExpansion(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`\newcommand{\sq}[1]{#1^2}\sq{x}`, `x^2`},
		{`\newcommand\ratio[2]{\frac{#1}{#2}}\ratio{a+b}{c}`, `\frac{a+b}{c}`},
		{`\def\sq#1{#1^2}\sq x`, `x^2`},
		{`\newcommand\power[2][2]{#2^{#1}}\power{x}+\power[3]{y}`, `x^{2}+y^{3}`},
		{`\newcommand\v{x}\renewcommand\v{y}\v`, `y`},
		{`\newcommand\v{x}\providecommand\v{y}\v`, `x`},
		{`\newcommand\a[1]{\sqrt{#1}}\newcommand\b[1]{\a{#1}}\b{x}`, `\sqrt{x}`},
		{`\newcommand\pref[1]{\alpha#1}\pref{x}`, `\alpha x`},
		{`\newcommand\id[1]{#1}\id{\alpha}x`, `\alpha x`},
		{`\newcommand\v{x}\alpha\v`, `\alpha x`},
	} {
		got := parseMath(tc.src)
		want := parseMath(tc.want)
		if got == nil || !reflect.DeepEqual(got, want) {
			t.Errorf("macro %s did not expand like %s", tc.src, tc.want)
		}
	}
	for _, src := range []string{`\newcommand\loop{\loop}\loop`, `\newcommand\bad[2]{#3}\bad{x}{y}`, `\def\bad#2{x}\bad{y}`, `\renewcommand\unknown{x}`, `\newcommand\missing[1]{#1}\missing`} {
		if parseMath(src) != nil {
			t.Errorf("invalid/recursive macro was accepted: %s", src)
		}
	}
}
func TestMathMacrosAcrossDocumentAndStreaming(t *testing.T) {
	d := New("$$\\newcommand{\\sqr}[1]{#1^2}$$\n\nUse $\\sqr{x}$.")
	bs := documentBlocks(d)
	if !d.contextual || len(bs) != 2 || bs[1].spans[1].math == nil {
		t.Fatal("document macro not visible to later formula")
	}
	d.SetStreaming(true)
	d.Append("\n\n$\\sqr{y")
	if got := plain(documentBlocks(d)[2].spans); got != "$\\sqr{y" {
		t.Fatalf("unfinished macro swallowed source: %q", got)
	}
	d.Append("}$")
	if documentBlocks(d)[2].spans[0].math == nil {
		t.Fatal("completed macro not rendered")
	}
	d.SetSource("$\\sqr{x}$")
	if documentBlocks(d)[0].spans[0].math != nil {
		t.Fatal("macro leaked into replacement document")
	}
}
func TestNestedMathStructures(t *testing.T) {
	src := `\begin{pmatrix}1 & \begin{bmatrix}a & b\\ c & d\end{bmatrix}\\ \frac{1}{2} & 4\end{pmatrix}`
	n := parseMath(src)
	if n == nil {
		t.Fatal("nested matrix rejected")
	}
	outer := n.children[0]
	if len(outer.children) != 2 || len(outer.children[0].children) != 2 {
		t.Fatal("nested separators split outer cells")
	}
	inner := outer.children[0].children[1].children[0]
	if inner.kind != "matrix" || len(inner.children) != 2 {
		t.Fatal("lost inner rows")
	}
	for _, src := range []string{`\left(\frac{1}{x}\right)`, `\left\{\begin{matrix}x\\ y\end{matrix}\right.`, `\left[\left(\frac12\right)^2\right]`, `\left\langle x\right\rangle`, `\left\Vert x\right\Vert`} {
		if parseMath(src) == nil {
			t.Errorf("valid delimiters rejected: %s", src)
		}
	}
	for _, src := range []string{`\left(x`, `x\right)`, `\begin{matrix}x\end{pmatrix}`, strings.Repeat(`\left(`, 80) + "x" + strings.Repeat(`\right)`, 80)} {
		if parseMath(src) != nil {
			t.Errorf("invalid/deep structure accepted: %s", src)
		}
	}
}
func TestStretchDelimitersEncloseTallContent(t *testing.T) {
	uitest.NewFunc(func(gtx core.C) {
		rn := run{size: theme.BodySize}
		inner := layoutMath(gtx, theme.Material.Shaper, parseMath(`\frac{1}{\sqrt{x}}`), rn, true)
		outer := layoutMath(gtx, theme.Material.Shaper, parseMath(`\left(\frac{1}{\sqrt{x}}\right)`), rn, true)
		if outer.a < inner.a || outer.d < inner.d || outer.w <= inner.w {
			t.Fatal("delimiters did not enclose fraction")
		}
	})
	d := New("$$\\left(\\frac{1}{x}\\right)$$")
	_ = uitest.New(el.Root(docView{d}))
	if len(d.chunks[0].blocks[0].view.(*richBlock).rt.pieces) != 1 {
		t.Fatal("stretch formula did not remain atomic")
	}
}
