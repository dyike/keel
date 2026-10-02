package markdown

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	gmparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Math is parsed before Markdown emphasis and escapes, so TeX underscores,
// braces and matrix row separators reach the typesetter unchanged.
var mathKind = ast.NewNodeKind("Math")

type mathInline struct {
	ast.BaseInline
	raw, source string
	display     bool
}

func (n *mathInline) Kind() ast.NodeKind         { return mathKind }
func (n *mathInline) Dump(src []byte, level int) { ast.DumpHelper(n, src, level, nil, nil) }
func (n *mathInline) IsRaw() bool                { return true }

type mathInlineParser struct{}

func (mathInlineParser) Trigger() []byte { return []byte{'$'} }
func (mathInlineParser) Parse(parent ast.Node, r text.Reader, pc gmparser.Context) ast.Node {
	line, _ := r.PeekLine()
	n := 1
	if len(line) > 1 && line[1] == '$' {
		n = 2
	}
	if n == 1 && (len(line) < 2 || unicode.IsSpace(rune(line[1]))) {
		return nil
	}
	ln, pos := r.Position()
	r.Advance(n)
	var b strings.Builder
	for {
		line, _ = r.PeekLine()
		if line == nil || b.Len() > 16384 {
			break
		}
		for i := 0; i < len(line); i++ {
			if line[i] == '\\' {
				i++
				continue
			}
			if line[i] != '$' {
				continue
			}
			if n == 2 && (i+1 >= len(line) || line[i+1] != '$') {
				continue
			}
			if n == 1 && (i == 0 || unicode.IsSpace(rune(line[i-1])) || i+1 < len(line) && line[i+1] >= '0' && line[i+1] <= '9') {
				continue
			}
			b.Write(line[:i])
			source := b.String()
			if strings.TrimSpace(source) == "" {
				break
			}
			r.Advance(i + n)
			delim := strings.Repeat("$", n)
			return &mathInline{raw: delim + source + delim, source: source, display: n == 2}
		}
		if n == 1 {
			break
		} // unfinished inline math never swallows following paragraphs
		b.Write(line)
		r.AdvanceLine()
	}
	r.SetPosition(ln, pos)
	return nil
}

// This bounded native subset covers common chat mathematics. Unsupported or
// malformed TeX is shown verbatim, never silently stripped or partially drawn.
type mathExpr struct {
	kind, value string
	children    []*mathExpr
	sup, sub    *mathExpr
}
type mathReader struct {
	src      []rune
	i, depth int
	bad      bool
}

func parseMath(src string) *mathExpr { return parseMathIn(src, mathMacros{}) }
func (p *mathReader) spaces() {
	for p.i < len(p.src) && unicode.IsSpace(p.src[p.i]) {
		p.i++
	}
}
func (p *mathReader) row(end rune) *mathExpr {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 64 {
		p.bad = true
		p.i = len(p.src)
		return &mathExpr{}
	}
	n := &mathExpr{kind: "row"}
	for {
		p.spaces()
		if p.i == len(p.src) {
			if end != 0 {
				p.bad = true
			}
			return n
		}
		if p.src[p.i] == end && end != 0 {
			p.i++
			return n
		}
		if p.src[p.i] == '}' {
			p.bad = true
			p.i++
			return n
		}
		if p.src[p.i] == '^' || p.src[p.i] == '_' {
			which := p.src[p.i]
			p.i++
			arg := p.arg()
			if len(n.children) == 0 {
				p.bad = true
				continue
			}
			last := n.children[len(n.children)-1]
			if last.kind != "scripts" {
				last = &mathExpr{kind: "scripts", children: []*mathExpr{last}}
				n.children[len(n.children)-1] = last
			}
			if which == '^' {
				if last.sup != nil {
					p.bad = true
				}
				last.sup = arg
			} else {
				if last.sub != nil {
					p.bad = true
				}
				last.sub = arg
			}
			continue
		}
		a := p.atom()
		if a.kind == "colorswitch" { // \color{c} colors the rest of the group
			n.children = append(n.children, &mathExpr{kind: "color", value: a.value, children: []*mathExpr{p.row(end)}})
			return n
		}
		n.children = append(n.children, a)
	}
}
func (p *mathReader) arg() *mathExpr {
	p.spaces()
	if p.i >= len(p.src) {
		p.bad = true
		return &mathExpr{}
	}
	if p.src[p.i] == '{' {
		p.i++
		return p.row('}')
	}
	return p.atom()
}
func (p *mathReader) groupText() string {
	p.spaces()
	if p.i >= len(p.src) || p.src[p.i] != '{' {
		p.bad = true
		return ""
	}
	p.i++
	start := p.i
	for p.i < len(p.src) && p.src[p.i] != '}' {
		p.i++
	}
	if p.i == len(p.src) {
		p.bad = true
		return ""
	}
	s := string(p.src[start:p.i])
	p.i++
	return s
}
func (p *mathReader) atom() *mathExpr {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 64 {
		p.bad = true
		p.i = len(p.src)
		return &mathExpr{}
	}
	p.spaces()
	if p.i >= len(p.src) {
		p.bad = true
		return &mathExpr{}
	}
	c := p.src[p.i]
	p.i++
	if c == '{' {
		return p.row('}')
	}
	if c != '\\' {
		kind := "text"
		if unicode.IsLetter(c) {
			kind = "variable"
		}
		if strings.ContainsRune("=+-<>×÷±", c) {
			kind = "operator"
		}
		if c == '-' {
			c = '−'
		}
		if c == '\'' {
			c = '′'
		}
		if c == '&' || c == '^' || c == '_' {
			p.bad = true
		}
		return &mathExpr{kind: kind, value: string(c)}
	}
	start := p.i
	for p.i < len(p.src) && unicode.IsLetter(p.src[p.i]) {
		p.i++
	}
	if start == p.i && p.i < len(p.src) {
		p.i++
	}
	cmd := string(p.src[start:p.i])
	if e, ok := p.moreCommand(cmd); ok {
		return e
	}
	switch cmd {
	case "frac", "dfrac", "tfrac":
		return &mathExpr{kind: "fraction", children: []*mathExpr{p.arg(), p.arg()}}
	case "sqrt":
		n := &mathExpr{kind: "sqrt"}
		p.spaces()
		if p.i < len(p.src) && p.src[p.i] == '[' {
			p.i++
			n.sup = p.row(']')
		}
		n.children = []*mathExpr{p.arg()}
		return n
	case "sum", "prod", "int", "oint", "bigcup", "bigcap":
		return &mathExpr{kind: "large", value: map[string]string{"sum": "∑", "prod": "∏", "int": "∫", "oint": "∮", "bigcup": "⋃", "bigcap": "⋂"}[cmd]}
	case "text", "mathrm", "operatorname":
		return &mathExpr{kind: "text", value: p.groupText()}
	case "mathbf", "mathit":
		return &mathExpr{kind: cmd, children: []*mathExpr{p.arg()}}
	case "left":
		return p.delimited()
	case "right", "end":
		p.bad = true
		return &mathExpr{}
	case ",", ":", ";", " ", "quad", "qquad":
		return &mathExpr{kind: "space", value: cmd}
	case "!":
		return &mathExpr{kind: "row"}
	case "begin":
		return p.matrix(p.groupText())
	}
	if s, ok := mathSymbols[cmd]; ok {
		if largeSymbols[s] {
			return &mathExpr{kind: "large", value: s}
		}
		kind := "operator"
		if utf8.RuneCountInString(s) == 1 && unicode.IsLetter([]rune(s)[0]) {
			kind = "variable"
		}
		return &mathExpr{kind: kind, value: s}
	}
	if strings.Contains(" sin cos tan cot sec csc log ln exp lim max min det gcd ", " "+cmd+" ") {
		return &mathExpr{kind: "operator", value: cmd}
	}
	if len([]rune(cmd)) == 1 && strings.ContainsAny(cmd, "{}%$#_|\\") {
		return &mathExpr{kind: "text", value: cmd}
	}
	p.bad = true
	return &mathExpr{}
}

var mathSymbols = map[string]string{
	"alpha": "α", "beta": "β", "gamma": "γ", "delta": "δ", "epsilon": "ϵ", "varepsilon": "ε", "zeta": "ζ", "eta": "η", "theta": "θ", "vartheta": "ϑ", "iota": "ι", "kappa": "κ", "lambda": "λ", "mu": "μ", "nu": "ν", "xi": "ξ", "pi": "π", "rho": "ρ", "sigma": "σ", "tau": "τ", "upsilon": "υ", "phi": "ϕ", "varphi": "φ", "chi": "χ", "psi": "ψ", "omega": "ω",
	"Gamma": "Γ", "Delta": "Δ", "Theta": "Θ", "Lambda": "Λ", "Xi": "Ξ", "Pi": "Π", "Sigma": "Σ", "Upsilon": "Υ", "Phi": "Φ", "Psi": "Ψ", "Omega": "Ω",
	"pm": "±", "mp": "∓", "times": "×", "div": "÷", "cdot": "⋅", "le": "≤", "leq": "≤", "ge": "≥", "geq": "≥", "ne": "≠", "neq": "≠", "approx": "≈", "equiv": "≡", "infty": "∞", "partial": "∂", "nabla": "∇", "in": "∈", "notin": "∉", "subset": "⊂", "subseteq": "⊆", "cup": "∪", "cap": "∩", "emptyset": "∅", "forall": "∀", "exists": "∃", "to": "→", "rightarrow": "→", "leftarrow": "←", "Rightarrow": "⇒", "Leftrightarrow": "⇔", "ldots": "…", "cdots": "⋯", "vdots": "⋮", "ddots": "⋱", "langle": "⟨", "rangle": "⟩", "vert": "|", "Vert": "‖",
}

// Standalone delimiters form a raw block, allowing blank lines in display
// math and protecting their contents from Markdown's block parsers.
var mathBlockKind = ast.NewNodeKind("DisplayMath")

type displayMath struct {
	ast.BaseBlock
	closed bool
}

func (n *displayMath) Kind() ast.NodeKind         { return mathBlockKind }
func (n *displayMath) Dump(src []byte, level int) { ast.DumpHelper(n, src, level, nil, nil) }
func (n *displayMath) IsRaw() bool                { return true }

type mathBlockParser struct{}

func (mathBlockParser) Trigger() []byte { return []byte{'$'} }
func (mathBlockParser) Open(parent ast.Node, r text.Reader, pc gmparser.Context) (ast.Node, gmparser.State) {
	line, _ := r.PeekLine()
	if strings.TrimSpace(string(line)) != "$$" {
		return nil, gmparser.NoChildren
	}
	r.AdvanceToEOL()
	return &displayMath{}, gmparser.NoChildren
}
func (mathBlockParser) Continue(n ast.Node, r text.Reader, pc gmparser.Context) gmparser.State {
	line, seg := r.PeekLine()
	if strings.TrimSpace(string(line)) == "$$" {
		n.(*displayMath).closed = true
		r.AdvanceToEOL()
		return gmparser.Close
	}
	n.Lines().Append(seg)
	r.AdvanceToEOL()
	return gmparser.Continue | gmparser.NoChildren
}
func (mathBlockParser) Close(ast.Node, text.Reader, gmparser.Context) {}
func (mathBlockParser) CanInterruptParagraph() bool                   { return true }
func (mathBlockParser) CanAcceptIndentedLine() bool                   { return false }

// Ignore math when healing Markdown markers. A multiplication sign inside
// $a*b$ must not synthesize an emphasis closer after the formula.
func stripMath(src string) string {
	var out strings.Builder
	for i := 0; i < len(src); {
		if src[i] == '\\' && i+1 < len(src) {
			out.WriteString(src[i : i+2])
			i += 2
			continue
		}
		if src[i] != '$' {
			out.WriteByte(src[i])
			i++
			continue
		}
		n := 1
		if i+1 < len(src) && src[i+1] == '$' {
			n = 2
		}
		if n == 1 && (i+1 == len(src) || unicode.IsSpace(rune(src[i+1])) || src[i+1] >= '0' && src[i+1] <= '9') {
			out.WriteByte(src[i])
			i++
			continue
		}
		i += n
		for i < len(src) {
			if src[i] == '\\' && i+1 < len(src) {
				i += 2
				continue
			}
			if src[i] == '$' && (n == 1 || i+1 < len(src) && src[i+1] == '$') {
				i += n
				break
			}
			i++
		}
	}
	return out.String()
}
