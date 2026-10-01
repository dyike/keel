package markdown

import (
	"strings"
	"unicode"
)

func (p *mathReader) inner(src []rune) *mathExpr {
	nested := mathReader{src: src, depth: p.depth + 1}
	n := nested.row(0)
	if nested.bad || nested.i != len(src) {
		p.bad = true
	}
	return n
}
func texCommandAt(src []rune, i int) (string, int) {
	if i >= len(src) || src[i] != '\\' {
		return "", i
	}
	i++
	start := i
	for i < len(src) && unicode.IsLetter(src[i]) {
		i++
	}
	if start == i && i < len(src) {
		i++
	}
	return string(src[start:i]), i
}
func texGroupAt(src []rune, i int) (string, int, bool) {
	for i < len(src) && unicode.IsSpace(src[i]) {
		i++
	}
	if i >= len(src) || src[i] != '{' {
		return "", i, false
	}
	i++
	start := i
	for i < len(src) && src[i] != '}' {
		i++
	}
	if i == len(src) {
		return "", i, false
	}
	return string(src[start:i]), i + 1, true
}
func (p *mathReader) matrix(env string) *mathExpr {
	switch env {
	case "matrix", "pmatrix", "bmatrix", "Bmatrix", "vmatrix", "Vmatrix", "cases", "aligned":
	default:
		p.bad = true
		return &mathExpr{}
	}
	n := &mathExpr{kind: "matrix", value: env}
	row := &mathExpr{kind: "row"}
	stack := []string{env}
	braces := 0
	start := p.i
	cell := func(end int) { row.children = append(row.children, p.inner(p.src[start:end])) }
	for p.i < len(p.src) {
		c := p.src[p.i]
		if c == '\\' {
			cmd, end := texCommandAt(p.src, p.i)
			if cmd == "begin" || cmd == "end" {
				name, after, ok := texGroupAt(p.src, end)
				if !ok {
					p.bad = true
					return n
				}
				if cmd == "begin" {
					stack = append(stack, name)
				} else {
					if name != stack[len(stack)-1] {
						p.bad = true
						return n
					}
					stack = stack[:len(stack)-1]
					if len(stack) == 0 {
						if strings.TrimSpace(string(p.src[start:p.i])) != "" || len(row.children) > 0 {
							cell(p.i)
							n.children = append(n.children, row)
						}
						p.i = after
						if len(n.children) == 0 || braces != 0 {
							p.bad = true
						}
						return n
					}
				}
				p.i = after
				continue
			}
			if cmd == "\\" && len(stack) == 1 && braces == 0 {
				cell(p.i)
				n.children = append(n.children, row)
				row = &mathExpr{kind: "row"}
				p.i = end
				start = end
				continue
			}
			p.i = end
			continue
		}
		if len(stack) == 1 {
			if c == '{' {
				braces++
			}
			if c == '}' {
				braces--
				if braces < 0 {
					p.bad = true
					return n
				}
			}
			if c == '&' && braces == 0 {
				cell(p.i)
				p.i++
				start = p.i
				continue
			}
		}
		p.i++
	}
	p.bad = true
	return n
}
func (p *mathReader) delimiter() string {
	p.spaces()
	if p.i >= len(p.src) {
		p.bad = true
		return "."
	}
	c := p.src[p.i]
	p.i++
	if c == '\\' {
		cmd, end := texCommandAt(p.src, p.i-1)
		p.i = end
		if value, ok := map[string]string{"{": "{", "}": "}", "lbrace": "{", "rbrace": "}", "langle": "⟨", "rangle": "⟩", "vert": "|", "Vert": "‖", "|": "‖", "lvert": "|", "rvert": "|", "lVert": "‖", "rVert": "‖", "lbrack": "[", "rbrack": "]"}[cmd]; ok {
			return value
		}
		p.bad = true
		return "."
	}
	if strings.ContainsRune(".()[]|<>/", c) {
		if c == '<' {
			return "⟨"
		}
		if c == '>' {
			return "⟩"
		}
		return string(c)
	}
	p.bad = true
	return "."
}
func (p *mathReader) delimited() *mathExpr {
	left := p.delimiter()
	start := p.i
	level := 1
	for p.i < len(p.src) {
		if p.src[p.i] != '\\' {
			p.i++
			continue
		}
		cmd, end := texCommandAt(p.src, p.i)
		if cmd == "left" {
			level++
		}
		if cmd == "right" {
			level--
			if level == 0 {
				content := p.inner(p.src[start:p.i])
				p.i = end
				right := p.delimiter()
				return &mathExpr{kind: "delimited", value: left + "\x00" + right, children: []*mathExpr{content}}
			}
		}
		p.i = end
	}
	p.bad = true
	return &mathExpr{}
}
