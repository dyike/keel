package markdown

import (
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// The wider vocabulary parses natively, so none of it falls back to source.
var widerTeX = []string{
	`\mathbb{R}^n \to \mathcal{L}(\mathfrak{g}) \subsetneq \mathsf{A}\mathtt{x}`,
	`\hat{x} + \bar{y} + \vec{v} + \dot{q} + \ddot{q} + \tilde{n} + \overline{AB} + \underline{c} + \widehat{xyz}`,
	`\overrightarrow{AB} \cdot \overleftarrow{CD}`,
	`\binom{n}{k} = \dbinom{n}{n-k}`,
	`\underbrace{a + \cdots + a}_{n} \overbrace{b+b}^{2}`,
	`\overset{\text{def}}{=} \underset{x}{\arg\max} \stackrel{?}{=}`,
	`\lim_{x \to 0} \frac{\sin x}{x} = 1, \max_{i} a_i, \limsup_{n\to\infty} x_n`,
	`\bigoplus_{i=1}^n V_i \otimes \bigwedge^k W, \iint_D f \, dA`,
	`a \not= b, x \not\in S, p \nmid q, A \not\subset B`,
	`\color{red} x + y`,
	`\textcolor{blue}{z} + \colorbox{yellow}{w} + \boxed{E = mc^2}`,
	`a \phantom{xx} b \hphantom{y} \vphantom{\frac{1}{2}}`,
	`\begin{align} a &= b + c \\ &= d \end{align}`,
	`\begin{gather} x = 1 \\ y = 2 \end{gather}`,
	`\begin{array}{lcr} 1 & 2 & 3 \\ 4 & 5 & 6 \end{array}`,
	`\begin{equation} E = mc^2 \tag{1} \end{equation}`,
	`f'(x) = \lim_{h \to 0} \frac{f(x+h) - f(x)}{h} \pmod{7}`,
	`\displaystyle \sum\limits_{k} \left( \frac{1}{k} \right) \quad \enspace \thinspace \hspace{1em}`,
	`\hbar \ell \Re \Im \aleph \nabla \angle \therefore \lfloor x \rfloor \lceil y \rceil`,
	`x \mapsto y \implies z \iff w \hookrightarrow v`,
	`\textbf{bold} \textit{it} \texttt{mono} \boldsymbol{\alpha}`,
	`\{ a \} \| b \|`,
}

func TestWiderTeXParses(t *testing.T) {
	for _, src := range widerTeX {
		if parseMath(src) == nil {
			t.Errorf("fell back to source: %s", src)
		}
	}
	if got := leafText(parseMath(`\mathbb{RZ}\mathcal{L}`)); got != "ℝℤℒ" {
		t.Errorf("alphabets = %q", got)
	}
	if got := leafText(parseMath(`\not= \not\in`)); got != "≠∉" {
		t.Errorf("negations = %q", got)
	}
	for _, bad := range []string{`\color{nocolor} x`, `\begin{tabular}x\end{tabular}`, `\notacommand`} {
		if parseMath(bad) != nil {
			t.Errorf("accepted %s", bad)
		}
	}
}

func leafText(n *mathExpr) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	if n.kind == "text" || n.kind == "variable" || n.kind == "operator" {
		b.WriteString(n.value)
	}
	for _, c := range n.children {
		b.WriteString(leafText(c))
	}
	return b.String()
}

// Every formula lays out, inline and in display, without panicking and
// with a visible size.
func TestWiderTeXLaysOut(t *testing.T) {
	var src strings.Builder
	for _, f := range widerTeX {
		src.WriteString("Inline $" + f + "$ text.\n\n$$" + f + "$$\n\n")
	}
	d := New(src.String())
	h := uitest.New(el.Root(docView{d}))
	h.Frame()
	for _, c := range d.chunks {
		for _, b := range c.blocks {
			r, ok := b.view.(*richBlock)
			if !ok {
				continue
			}
			for _, p := range r.rt.pieces {
				if r.runs[p.run].math != nil && (p.rect.Dx() <= 0 || p.rect.Dy() <= 0) {
					t.Errorf("empty formula box for %q", r.runs[p.run].text)
				}
			}
		}
	}
}
