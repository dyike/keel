package markdown

import (
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/text"
)

// The wider TeX vocabulary: more symbols and functions, math alphabets,
// accents, binomials, braces, stacked scripts, color, boxes, phantoms and
// more environments. Anything outside it still falls back to the source.

func init() {
	for k, v := range moreSymbols {
		mathSymbols[k] = v
	}
}

var moreSymbols = map[string]string{
	// Greek variants.
	"varpi": "ϖ", "varrho": "ϱ", "varsigma": "ς", "varkappa": "ϰ", "digamma": "ϝ", "omicron": "ο",
	// Relations.
	"ll": "≪", "gg": "≫", "sim": "∼", "simeq": "≃", "cong": "≅", "propto": "∝", "perp": "⊥", "parallel": "∥",
	"mid": "∣", "nmid": "∤", "prec": "≺", "succ": "≻", "preceq": "⪯", "succeq": "⪰", "supset": "⊃",
	"supseteq": "⊇", "ni": "∋", "subsetneq": "⊊", "supsetneq": "⊋", "models": "⊨", "vdash": "⊢", "dashv": "⊣",
	"asymp": "≍", "doteq": "≐", "approxeq": "≊", "lesssim": "≲", "gtrsim": "≳", "leqslant": "⩽", "geqslant": "⩾",
	"nleq": "≰", "ngeq": "≱", "nsubseteq": "⊈", "triangleq": "≜", "coloneqq": "≔", "lt": "<", "gt": ">",
	// Arrows.
	"leftrightarrow": "↔", "longrightarrow": "⟶", "longleftarrow": "⟵", "Longrightarrow": "⟹", "Longleftarrow": "⟸",
	"Leftarrow": "⇐", "longleftrightarrow": "⟷", "Longleftrightarrow": "⟺", "mapsto": "↦", "longmapsto": "⟼",
	"uparrow": "↑", "downarrow": "↓", "Uparrow": "⇑", "Downarrow": "⇓", "updownarrow": "↕", "hookrightarrow": "↪",
	"hookleftarrow": "↩", "rightharpoonup": "⇀", "leftharpoonup": "↼", "rightleftharpoons": "⇌", "nearrow": "↗",
	"searrow": "↘", "nwarrow": "↖", "swarrow": "↙", "iff": "⟺", "implies": "⟹", "impliedby": "⟸", "gets": "←",
	"leadsto": "⇝", "rightsquigarrow": "⇝", "twoheadrightarrow": "↠",
	// Binary operators.
	"oplus": "⊕", "otimes": "⊗", "odot": "⊙", "ominus": "⊖", "oslash": "⊘", "circ": "∘", "bullet": "∙", "star": "⋆",
	"ast": "∗", "wedge": "∧", "vee": "∨", "land": "∧", "lor": "∨", "setminus": "∖", "dagger": "†", "ddagger": "‡",
	"sqcup": "⊔", "sqcap": "⊓", "uplus": "⊎", "wr": "≀", "diamond": "⋄", "bigtriangleup": "△", "bigtriangledown": "▽",
	"triangleleft": "◁", "triangleright": "▷", "boxplus": "⊞", "boxtimes": "⊠", "cdotp": "⋅", "ltimes": "⋉", "rtimes": "⋊",
	// Large operators.
	"bigoplus": "⨁", "bigotimes": "⨂", "bigodot": "⨀", "bigvee": "⋁", "bigwedge": "⋀", "biguplus": "⨄",
	"bigsqcup": "⨆", "coprod": "∐", "iint": "∬", "iiint": "∭", "oiint": "∯",
	// Miscellaneous.
	"hbar": "ℏ", "hslash": "ℏ", "ell": "ℓ", "Re": "ℜ", "Im": "ℑ", "aleph": "ℵ", "beth": "ℶ", "wp": "℘", "angle": "∠",
	"measuredangle": "∡", "triangle": "△", "square": "□", "Box": "□", "blacksquare": "■", "lozenge": "◊",
	"clubsuit": "♣", "diamondsuit": "♢", "heartsuit": "♡", "spadesuit": "♠", "neg": "¬", "lnot": "¬", "top": "⊤",
	"bot": "⊥", "prime": "′", "backslash": "∖", "surd": "√", "checkmark": "✓", "degree": "°", "complement": "∁",
	"nexists": "∄", "varnothing": "∅", "therefore": "∴", "because": "∵", "flat": "♭", "sharp": "♯", "natural": "♮",
	"dots": "…", "dotsc": "…", "dotsb": "⋯", "dotsm": "⋯", "dotsi": "⋯", "dotso": "…", "colon": ":", "S": "§", "P": "¶",
	"lfloor": "⌊", "rfloor": "⌋", "lceil": "⌈", "rceil": "⌉", "lbrace": "{", "rbrace": "}", "lbrack": "[", "rbrack": "]",
	"lvert": "|", "rvert": "|", "lVert": "‖", "rVert": "‖", "ulcorner": "⌜", "urcorner": "⌝", "llcorner": "⌞", "lrcorner": "⌟",
	"infin": "∞", "pounds": "£", "euro": "€", "copyright": "©",
}

// largeSymbols take limits above and below in display math, like \sum.
var largeSymbols = map[string]bool{"⨁": true, "⨂": true, "⨀": true, "⋁": true, "⋀": true, "⨄": true, "⨆": true, "∐": true, "∬": true, "∭": true, "∯": true}

// moreFunctions are set upright like \sin; limitFunctions also take limits
// above and below in display math, like \lim.
var moreFunctions = " arcsin arccos arctan arccot sinh cosh tanh coth arg dim hom ker deg Pr sup inf liminf limsup lg sgn tr rank span diag Var Cov "
var limitFunctions = map[string]bool{"lim": true, "max": true, "min": true, "sup": true, "inf": true, "liminf": true, "limsup": true, "det": true, "gcd": true, "Pr": true, "argmax": true, "argmin": true}

// mathAlphabets maps letters and digits for \mathbb and its kin. Unicode
// leaves holes for letters that existed before, filled from the letterlike block.
var mathAlphabets = map[string]struct {
	upper, lower, digit rune
	holes               map[rune]rune
}{
	"mathbb":   {0x1D538, 0x1D552, 0x1D7D8, map[rune]rune{'C': 'ℂ', 'H': 'ℍ', 'N': 'ℕ', 'P': 'ℙ', 'Q': 'ℚ', 'R': 'ℝ', 'Z': 'ℤ'}},
	"mathcal":  {0x1D49C, 0x1D4B6, 0, map[rune]rune{'B': 'ℬ', 'E': 'ℰ', 'F': 'ℱ', 'H': 'ℋ', 'I': 'ℐ', 'L': 'ℒ', 'M': 'ℳ', 'R': 'ℛ', 'e': 'ℯ', 'g': 'ℊ', 'o': 'ℴ'}},
	"mathfrak": {0x1D504, 0x1D51E, 0, map[rune]rune{'C': 'ℭ', 'H': 'ℌ', 'I': 'ℑ', 'R': 'ℜ', 'Z': 'ℨ'}},
	"mathsf":   {0x1D5A0, 0x1D5BA, 0x1D7E2, nil},
	"mathtt":   {0x1D670, 0x1D68A, 0x1D7F6, nil},
}

func init() {
	mathAlphabets["mathscr"] = mathAlphabets["mathcal"]
	mathAlphabets["Bbb"] = mathAlphabets["mathbb"]
}

func alphabetRune(name string, r rune) rune {
	a := mathAlphabets[name]
	if h, ok := a.holes[r]; ok {
		return h
	}
	switch {
	case r >= 'A' && r <= 'Z':
		return a.upper + r - 'A'
	case r >= 'a' && r <= 'z':
		return a.lower + r - 'a'
	case r >= '0' && r <= '9' && a.digit != 0:
		return a.digit + r - '0'
	}
	return r
}

// restyle maps the letters of every leaf under n into an alphabet, or
// sets them upright when name is "".
func restyle(n *mathExpr, name string) *mathExpr {
	if n == nil {
		return nil
	}
	if n.kind == "variable" || n.kind == "text" {
		if name != "" {
			var b strings.Builder
			for _, r := range n.value {
				b.WriteRune(alphabetRune(name, r))
			}
			n.value = b.String()
		}
		n.kind = "text"
	}
	for _, c := range n.children {
		restyle(c, name)
	}
	restyle(n.sup, name)
	restyle(n.sub, name)
	return n
}

// accents are all drawn, not set from a font: accent glyphs carry their own
// height above the baseline, which would float them away from the letter.
var accents = map[string]bool{
	"hat": true, "widehat": true, "tilde": true, "widetilde": true, "dot": true, "ddot": true, "dddot": true,
	"check": true, "breve": true, "acute": true, "grave": true, "mathring": true,
	"bar": true, "overline": true, "underline": true, "vec": true, "overrightarrow": true, "overleftarrow": true,
	"overleftrightarrow": true, "overbrace": true, "underbrace": true, "widebar": true,
}

var notForms = map[string]string{"=": "≠", "∈": "∉", "<": "≮", ">": "≯", "≤": "≰", "≥": "≱", "⊂": "⊄", "⊃": "⊅", "⊆": "⊈", "⊇": "⊉", "∼": "≁", "≡": "≢", "≈": "≉", "∃": "∄", "∣": "∤", "∥": "∦", "≅": "≇", "≃": "≄"}

var texColors = map[string]color.NRGBA{
	"red": {0xd3, 0x2f, 0x2f, 0xff}, "blue": {0x1e, 0x55, 0xd6, 0xff}, "green": {0x2e, 0x7d, 0x32, 0xff},
	"orange": {0xef, 0x6c, 0x00, 0xff}, "purple": {0x7b, 0x1f, 0xa2, 0xff}, "violet": {0x7b, 0x1f, 0xa2, 0xff},
	"magenta": {0xc2, 0x18, 0x5b, 0xff}, "cyan": {0x00, 0x83, 0x8f, 0xff}, "teal": {0x00, 0x79, 0x6b, 0xff},
	"brown": {0x79, 0x55, 0x48, 0xff}, "gray": {0x75, 0x75, 0x75, 0xff}, "grey": {0x75, 0x75, 0x75, 0xff},
	"black": {0, 0, 0, 0xff}, "white": {0xff, 0xff, 0xff, 0xff}, "yellow": {0xf9, 0xa8, 0x25, 0xff},
	"pink": {0xd8, 0x1b, 0x60, 0xff}, "olive": {0x82, 0x77, 0x17, 0xff}, "lime": {0x7c, 0xb3, 0x42, 0xff},
}

func parseTexColor(s string) (color.NRGBA, bool) {
	s = strings.TrimSpace(s)
	if c, ok := texColors[strings.ToLower(s)]; ok {
		return c, true
	}
	h := strings.TrimPrefix(s, "#")
	if len(h) == 6 {
		if n, err := strconv.ParseUint(h, 16, 32); err == nil {
			return color.NRGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 0xff}, true
		}
	}
	return color.NRGBA{}, false
}

// moreCommand parses the commands of the wider vocabulary; ok is false for
// the others.
func (p *mathReader) moreCommand(cmd string) (*mathExpr, bool) {
	if _, ok := mathAlphabets[cmd]; ok {
		return restyle(p.arg(), cmd), true
	}
	if _, ok := accents[cmd]; ok {
		return &mathExpr{kind: "accent", value: cmd, children: []*mathExpr{p.arg()}}, true
	}
	if strings.Contains(moreFunctions, " "+cmd+" ") {
		return &mathExpr{kind: "operator", value: cmd}, true
	}
	switch cmd {
	case "boldsymbol", "bm", "pmb":
		return &mathExpr{kind: "mathbf", children: []*mathExpr{p.arg()}}, true
	case "mathup", "textrm", "textnormal", "textup":
		return restyle(p.arg(), ""), true
	case "textbf":
		return &mathExpr{kind: "mathbf", children: []*mathExpr{{kind: "text", value: p.groupText()}}}, true
	case "textit", "emph":
		return &mathExpr{kind: "mathit", children: []*mathExpr{{kind: "text", value: p.groupText()}}}, true
	case "texttt":
		return restyle(&mathExpr{kind: "text", value: p.groupText()}, "mathtt"), true
	case "binom", "dbinom", "tbinom", "choose":
		return &mathExpr{kind: "binom", children: []*mathExpr{p.arg(), p.arg()}}, true
	case "overset", "stackrel":
		over := p.arg()
		return &mathExpr{kind: "stack", children: []*mathExpr{p.arg()}, sup: over}, true
	case "underset":
		under := p.arg()
		return &mathExpr{kind: "stack", children: []*mathExpr{p.arg()}, sub: under}, true
	case "textcolor", "colorbox":
		c, ok := parseTexColor(p.groupText())
		if !ok {
			p.bad = true
		}
		kind := "color"
		if cmd == "colorbox" {
			kind = "colorbox"
		}
		return &mathExpr{kind: kind, value: hexColor(c), children: []*mathExpr{p.arg()}}, true
	case "color":
		c, ok := parseTexColor(p.groupText())
		if !ok {
			p.bad = true
		}
		return &mathExpr{kind: "colorswitch", value: hexColor(c)}, true
	case "boxed", "fbox", "framebox":
		return &mathExpr{kind: "boxed", children: []*mathExpr{p.arg()}}, true
	case "phantom", "hphantom", "vphantom":
		return &mathExpr{kind: "phantom", value: cmd, children: []*mathExpr{p.arg()}}, true
	case "not":
		next := p.atom()
		if v, ok := notForms[next.value]; ok {
			next.value = v
		} else {
			next.value += "̸"
		}
		return next, true
	case "displaystyle", "textstyle", "scriptstyle", "scriptscriptstyle", "limits", "nolimits", "nonumber", "notag",
		"big", "Big", "bigg", "Bigg", "bigl", "bigr", "Bigl", "Bigr", "biggl", "biggr", "middle", "relax", "strut":
		return &mathExpr{kind: "row"}, true
	case "tag", "tag*":
		label := p.groupText()
		if cmd == "tag" {
			label = "(" + label + ")"
		}
		return &mathExpr{kind: "row", children: []*mathExpr{{kind: "space", value: "qquad"}, {kind: "text", value: label}}}, true
	case "hspace", "hskip", "kern", "mkern", "mskip":
		if p.i < len(p.src) && p.src[p.i] == '{' {
			p.groupText()
		} else {
			for p.i < len(p.src) && (p.src[p.i] == '.' || p.src[p.i] == '-' || p.src[p.i] >= '0' && p.src[p.i] <= '9' || p.src[p.i] >= 'a' && p.src[p.i] <= 'z') {
				p.i++
			}
		}
		return &mathExpr{kind: "space", value: ","}, true
	case "enspace", "thinspace", "medspace", "thickspace", ">", "negthinspace", "negmedspace":
		return &mathExpr{kind: "space", value: cmd}, true
	case "bmod", "mod":
		return &mathExpr{kind: "operator", value: "mod"}, true
	case "pmod":
		return &mathExpr{kind: "row", children: []*mathExpr{{kind: "text", value: " (mod "}, p.arg(), {kind: "text", value: ")"}}}, true
	case "argmax", "argmin":
		return &mathExpr{kind: "operator", value: "arg" + cmd[3:]}, true
	case "{", "}", "|", "_", "#", "%", "&", "$":
		v := cmd
		if cmd == "|" {
			v = "‖"
		}
		return &mathExpr{kind: "text", value: v}, true
	}
	return nil, false
}

func hexColor(c color.NRGBA) string {
	return strconv.FormatUint(uint64(c.R)<<16|uint64(c.G)<<8|uint64(c.B), 16)
}

func colorOf(hex string) color.NRGBA {
	n, _ := strconv.ParseUint(hex, 16, 32)
	return color.NRGBA{uint8(n >> 16), uint8(n >> 8), uint8(n), 0xff}
}

// environment normalizes an environment name to the layout it uses, or
// reports that it is unknown.
func environment(name string) (string, bool) {
	switch name {
	case "matrix", "pmatrix", "bmatrix", "Bmatrix", "vmatrix", "Vmatrix", "cases", "aligned":
		return name, true
	case "align", "align*", "alignat", "alignat*", "split", "eqnarray", "eqnarray*", "alignedat", "flalign", "flalign*":
		return "aligned", true
	case "gather", "gather*", "gathered", "equation", "equation*", "multline", "multline*", "displaymath":
		return "gathered", true
	case "array", "smallmatrix", "subarray":
		return "matrix", true
	case "dcases", "rcases":
		return "cases", true
	}
	return "", false
}

// limitsBase reports whether scripts on n go above and below in display math.
func limitsBase(n *mathExpr) bool {
	switch n.kind {
	case "large":
		return n.value != "∫" && n.value != "∮" && n.value != "∬" && n.value != "∭" && n.value != "∯"
	case "operator":
		return limitFunctions[n.value]
	}
	return false
}

// layoutMore lays out the wider vocabulary's kinds; ok is false for others.
func layoutMore(gtx layout.Context, shaper *text.Shaper, n *mathExpr, rn run, display bool) (mathBox, bool) {
	em := gtx.Sp(rn.size)
	gap := max(2, em/8)
	stroke := float32(max(1, em/18))
	child := func(n *mathExpr) mathBox { return layoutMath(gtx, shaper, n, rn, display) }
	small := rn
	small.size *= 0.72
	switch n.kind {
	case "accent":
		base := child(n.children[0])
		w := base.w
		switch n.value {
		case "underline":
			y := base.d + gap
			return mathCompose(gtx, w, base.a, y+int(stroke), []mathPlacement{{base, 0, 0}}, func() {
				mathRule(gtx, image.Rect(0, y, w, y+int(stroke)))
			}), true
		case "bar", "overline", "widebar":
			y := -base.a - gap - int(stroke)
			return mathCompose(gtx, w, -y, base.d, []mathPlacement{{base, 0, 0}}, func() {
				mathRule(gtx, image.Rect(0, y, w, y+int(stroke)))
			}), true
		case "vec", "overrightarrow", "overleftarrow", "overleftrightarrow":
			h := max(4, em/3)
			y := float32(-base.a - gap - h/2)
			x0, x1 := float32(0), float32(w)
			if n.value == "vec" {
				x0, x1 = float32(w)/2-float32(em)/4, float32(w)/2+float32(em)/3
			}
			head := float32(h) / 2
			return mathCompose(gtx, w, base.a+gap+h, base.d, []mathPlacement{{base, 0, 0}}, func() {
				mathStroke(gtx, []f32.Point{f32.Pt(x0, y), f32.Pt(x1, y)}, stroke)
				if n.value != "overleftarrow" {
					mathStroke(gtx, []f32.Point{f32.Pt(x1-head, y-head), f32.Pt(x1, y), f32.Pt(x1-head, y+head)}, stroke)
				}
				if n.value == "overleftarrow" || n.value == "overleftrightarrow" {
					mathStroke(gtx, []f32.Point{f32.Pt(x0+head, y-head), f32.Pt(x0, y), f32.Pt(x0+head, y+head)}, stroke)
				}
			}), true
		case "overbrace", "underbrace":
			h := max(5, em/3)
			if n.value == "overbrace" {
				top := float32(-base.a - gap - h)
				return mathCompose(gtx, w, base.a+2*gap+h, base.d, []mathPlacement{{base, 0, 0}}, func() {
					brace(gtx, float32(w), top+float32(h), top, stroke)
				}), true
			}
			y := float32(base.d + gap)
			return mathCompose(gtx, w, base.a, base.d+2*gap+h, []mathPlacement{{base, 0, 0}}, func() {
				brace(gtx, float32(w), y, y+float32(h), stroke)
			}), true
		}
		// Small marks: drawn in a box h high, centered over the letter
		// (italic letters lean right, so nudge), or across a wide base.
		h := max(3, em/4)
		y0 := float32(-base.a - gap/2) // the mark's bottom
		cx := float32(w)/2 + float32(em)/16
		half := float32(min(w, em)) * 0.3
		if n.value == "widehat" || n.value == "widetilde" {
			half = float32(w)/2 - float32(gap)
			cx = float32(w) / 2
		}
		top := y0 - float32(h)
		return mathCompose(gtx, w, base.a+gap/2+h, base.d, []mathPlacement{{base, 0, 0}}, func() {
			drawAccent(gtx, n.value, cx, half, top, y0, stroke)
		}), true
	case "binom":
		inner := rn
		if !display {
			inner.size *= 0.85
		}
		top := layoutMath(gtx, shaper, n.children[0], inner, false)
		bottom := layoutMath(gtx, shaper, n.children[1], inner, false)
		axis := em / 4
		w := max(top.w, bottom.w) + 2*gap
		ty := -axis - gap - top.d
		by := -axis + gap + bottom.a
		stack := mathCompose(gtx, w, top.a-ty, by+bottom.d, []mathPlacement{{top, (w - top.w) / 2, ty}, {bottom, (w - bottom.w) / 2, by}}, nil)
		return encloseMath(gtx, stack, "(\x00)", em), true
	case "stack":
		base := child(n.children[0])
		w := base.w
		parts := []mathPlacement{}
		a, d := base.a, base.d
		var over, under mathBox
		if n.sup != nil {
			over = layoutMath(gtx, shaper, n.sup, small, false)
			w = max(w, over.w)
		}
		if n.sub != nil {
			under = layoutMath(gtx, shaper, n.sub, small, false)
			w = max(w, under.w)
		}
		parts = append(parts, mathPlacement{base, (w - base.w) / 2, 0})
		if n.sup != nil {
			y := -base.a - gap/2 - over.d
			parts = append(parts, mathPlacement{over, (w - over.w) / 2, y})
			a = over.a - y
		}
		if n.sub != nil {
			y := base.d + gap/2 + under.a
			parts = append(parts, mathPlacement{under, (w - under.w) / 2, y})
			d = y + under.d
		}
		return mathCompose(gtx, w, a, d, parts, nil), true
	case "color":
		// Glyphs paint in the current color; set it around the content.
		outer := rn.color
		rn.color = colorOf(n.value)
		b := layoutMath(gtx, shaper, n.children[0], rn, display)
		m := op.Record(gtx.Ops)
		paint.ColorOp{Color: rn.color}.Add(gtx.Ops)
		b.call.Add(gtx.Ops)
		paint.ColorOp{Color: outer}.Add(gtx.Ops)
		b.call = m.Stop()
		return b, true
	case "colorbox", "boxed":
		b := child(n.children[0])
		pad := gap * 2
		w, a, d := b.w+2*pad, b.a+pad, b.d+pad
		fill := n.kind == "colorbox"
		c := colorOf(n.value)
		m := mathCompose(gtx, w, a, d, nil, func() {
			if fill {
				c.A = 0x40
				paint.ColorOp{Color: c}.Add(gtx.Ops)
				mathRule(gtx, image.Rect(0, -a, w, d))
				paint.ColorOp{Color: rn.color}.Add(gtx.Ops)
			} else {
				s := int(stroke)
				mathRule(gtx, image.Rect(0, -a, w, -a+s))
				mathRule(gtx, image.Rect(0, d-s, w, d))
				mathRule(gtx, image.Rect(0, -a, s, d))
				mathRule(gtx, image.Rect(w-s, -a, w, d))
			}
		})
		return mathCompose(gtx, w, a, d, []mathPlacement{{m, 0, 0}, {b, pad, 0}}, nil), true
	case "phantom":
		b := child(n.children[0])
		switch n.value {
		case "hphantom":
			return mathCompose(gtx, b.w, 0, 0, nil, nil), true
		case "vphantom":
			return mathCompose(gtx, 0, b.a, b.d, nil, nil), true
		}
		return mathCompose(gtx, b.w, b.a, b.d, nil, nil), true
	case "space":
		w := map[string]int{"enspace": em / 2, "thinspace": em / 6, ",": em / 6, "medspace": em * 2 / 9, ":": em * 2 / 9, ">": em * 2 / 9,
			"thickspace": em * 5 / 18, ";": em * 5 / 18, " ": em / 4, "quad": em, "qquad": 2 * em, "negthinspace": 0, "negmedspace": 0}
		if v, ok := w[n.value]; ok {
			return mathCompose(gtx, v, 0, 0, nil, nil), true
		}
	}
	return mathBox{}, false
}

// drawAccent draws a small mark between top and bottom, centered at cx and
// half wide.
func drawAccent(gtx layout.Context, kind string, cx, half, top, bottom, stroke float32) {
	mid := (top + bottom) / 2
	dot := func(x float32) {
		r := max(stroke, (bottom-top)/4)
		s := clip.Ellipse{Min: image.Pt(int(x-r), int(mid-r)), Max: image.Pt(int(x+r+0.5), int(mid+r+0.5))}.Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		s.Pop()
	}
	switch kind {
	case "hat", "widehat":
		mathStroke(gtx, []f32.Point{f32.Pt(cx-half, bottom), f32.Pt(cx, top), f32.Pt(cx+half, bottom)}, stroke)
	case "check":
		mathStroke(gtx, []f32.Point{f32.Pt(cx-half, top), f32.Pt(cx, bottom), f32.Pt(cx+half, top)}, stroke)
	case "tilde", "widetilde":
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(cx-half, bottom))
		p.CubeTo(f32.Pt(cx-half/2, top-(bottom-top)/2), f32.Pt(cx+half/2, bottom+(bottom-top)/2), f32.Pt(cx+half, top))
		s := clip.Stroke{Path: p.End(), Width: stroke}.Op().Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		s.Pop()
	case "breve":
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(cx-half, top))
		p.QuadTo(f32.Pt(cx, bottom+(bottom-top)), f32.Pt(cx+half, top))
		s := clip.Stroke{Path: p.End(), Width: stroke}.Op().Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		s.Pop()
	case "acute":
		mathStroke(gtx, []f32.Point{f32.Pt(cx-half/3, bottom), f32.Pt(cx+half/2, top)}, stroke)
	case "grave":
		mathStroke(gtx, []f32.Point{f32.Pt(cx+half/3, bottom), f32.Pt(cx-half/2, top)}, stroke)
	case "dot":
		dot(cx)
	case "ddot":
		dot(cx - half/2)
		dot(cx + half/2)
	case "dddot":
		dot(cx - half/1.5)
		dot(cx)
		dot(cx + half/1.5)
	case "mathring":
		r := (bottom - top) / 2
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(cx+r, mid))
		p.ArcTo(f32.Pt(cx, mid), f32.Pt(cx, mid), 6.2831)
		s := clip.Stroke{Path: p.End(), Width: stroke}.Op().Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		s.Pop()
	}
}

// brace draws a horizontal curly brace across w, its arms at y and its
// point at tip.
func brace(gtx layout.Context, w, y, tip, stroke float32) {
	mid := w / 2
	q := (tip - y) / 2
	mathStroke(gtx, []f32.Point{f32.Pt(0, y), f32.Pt(stroke, y+q), f32.Pt(mid-stroke, y+q), f32.Pt(mid, tip), f32.Pt(mid+stroke, y+q), f32.Pt(w-stroke, y+q), f32.Pt(w, y)}, stroke)
}
