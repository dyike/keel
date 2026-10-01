package markdown

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type mathMacro struct {
	body     string
	args     int
	optional *string
}
type mathMacros map[string]mathMacro

type macroReader struct {
	src string
	i   int
}

func (r *macroReader) spaces() {
	for r.i < len(r.src) {
		c, n := utf8.DecodeRuneInString(r.src[r.i:])
		if !unicode.IsSpace(c) {
			break
		}
		r.i += n
	}
}
func (r *macroReader) command() string {
	if r.i >= len(r.src) || r.src[r.i] != '\\' {
		return ""
	}
	r.i++
	start := r.i
	for r.i < len(r.src) && (r.src[r.i] >= 'a' && r.src[r.i] <= 'z' || r.src[r.i] >= 'A' && r.src[r.i] <= 'Z') {
		r.i++
	}
	if r.i == start && r.i < len(r.src) {
		_, n := utf8.DecodeRuneInString(r.src[r.i:])
		r.i += n
	}
	return r.src[start:r.i]
}
func (r *macroReader) group(open, close byte) (string, bool) {
	r.spaces()
	if r.i >= len(r.src) || r.src[r.i] != open {
		return "", false
	}
	r.i++
	start := r.i
	depth := 1
	for r.i < len(r.src) {
		c := r.src[r.i]
		if c == '\\' {
			r.command()
			continue
		}
		if c == open {
			depth++
		}
		if c == close {
			depth--
			if depth == 0 {
				s := r.src[start:r.i]
				r.i++
				return s, true
			}
		}
		r.i++
	}
	return "", false
}
func (r *macroReader) arg() (string, bool) {
	r.spaces()
	if r.i >= len(r.src) {
		return "", false
	}
	if r.src[r.i] == '{' {
		return r.group('{', '}')
	}
	start := r.i
	if r.src[r.i] == '\\' {
		r.command()
	} else {
		_, n := utf8.DecodeRuneInString(r.src[r.i:])
		r.i += n
	}
	return r.src[start:r.i], true
}
func hasMathMacros(src string) bool {
	for _, cmd := range []string{`\newcommand`, `\renewcommand`, `\providecommand`, `\def`} {
		if strings.Contains(src, cmd) {
			return true
		}
	}
	return false
}

// Keep TeX control words separate when replacing parameter or macro tokens.
// For example, substituting x into \alpha#1 must not create \alphax.
func appendMathTokens(out *strings.Builder, part, rest string) {
	letter := func(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
	controlWord := func(s string) bool {
		i := len(s) - 1
		for i >= 0 && letter(s[i]) {
			i--
		}
		return i >= 0 && i < len(s)-1 && s[i] == '\\'
	}
	if len(part) > 0 && letter(part[0]) && controlWord(out.String()) {
		out.WriteByte(' ')
	}
	out.WriteString(part)
	if len(rest) > 0 && letter(rest[0]) && controlWord(part) {
		out.WriteByte(' ')
	}
}

func expandMathMacros(src string, env mathMacros, depth int, budget *int) (string, bool) {
	if depth > 32 || len(src) > 16384 {
		return "", false
	}
	var out strings.Builder
	r := macroReader{src: src}
	for r.i < len(src) {
		if out.Len() > 16384 {
			return "", false
		}
		if src[r.i] != '\\' {
			out.WriteByte(src[r.i])
			r.i++
			continue
		}
		start := r.i
		cmd := r.command()
		switch cmd {
		case "newcommand", "renewcommand", "providecommand", "def":
			r.spaces()
			if r.i < len(src) && src[r.i] == '*' && cmd != "def" {
				r.i++
				r.spaces()
			}
			name := ""
			if r.i < len(src) && src[r.i] == '{' {
				value, ok := r.group('{', '}')
				if !ok {
					return "", false
				}
				nr := macroReader{src: strings.TrimSpace(value)}
				name = nr.command()
				if nr.i != len(nr.src) {
					return "", false
				}
			} else {
				name = r.command()
			}
			if name == "" {
				return "", false
			}
			macro := mathMacro{}
			r.spaces()
			if cmd == "def" {
				for r.i < len(src) && src[r.i] == '#' {
					if r.i+1 >= len(src) || src[r.i+1] != byte('1'+macro.args) || macro.args >= 9 {
						return "", false
					}
					macro.args++
					r.i += 2
					r.spaces()
				}
			} else if r.i < len(src) && src[r.i] == '[' {
				count, ok := r.group('[', ']')
				if !ok {
					return "", false
				}
				n, err := strconv.Atoi(count)
				if err != nil || n < 0 || n > 9 {
					return "", false
				}
				macro.args = n
				r.spaces()
				if r.i < len(src) && src[r.i] == '[' {
					value, ok := r.group('[', ']')
					if !ok || n == 0 {
						return "", false
					}
					macro.optional = &value
				}
			}
			body, ok := r.group('{', '}')
			if !ok {
				return "", false
			}
			macro.body = body
			_, exists := env[name]
			if cmd == "newcommand" && exists || cmd == "renewcommand" && !exists {
				return "", false
			}
			if cmd != "providecommand" || !exists {
				env[name] = macro
			}
			continue
		}
		macro, ok := env[cmd]
		if !ok {
			out.WriteString(src[start:r.i])
			continue
		}
		*budget--
		if *budget < 0 {
			return "", false
		}
		args := make([]string, macro.args)
		first := 0
		if macro.optional != nil {
			r.spaces()
			args[0] = *macro.optional
			if r.i < len(src) && src[r.i] == '[' {
				value, ok := r.group('[', ']')
				if !ok {
					return "", false
				}
				args[0] = value
			}
			first = 1
		}
		for i := first; i < macro.args; i++ {
			value, ok := r.arg()
			if !ok {
				return "", false
			}
			args[i] = value
		}
		var body strings.Builder
		for i := 0; i < len(macro.body); i++ {
			if macro.body[i] != '#' {
				body.WriteByte(macro.body[i])
				continue
			}
			i++
			if i >= len(macro.body) {
				return "", false
			}
			if macro.body[i] == '#' {
				body.WriteByte('#')
				continue
			}
			index := int(macro.body[i] - '1')
			if index < 0 || index >= len(args) {
				return "", false
			}
			appendMathTokens(&body, args[index], macro.body[i+1:])
		}
		expanded, ok := expandMathMacros(body.String(), env, depth+1, budget)
		if !ok {
			return "", false
		}
		appendMathTokens(&out, expanded, src[r.i:])
	}
	return out.String(), out.Len() <= 16384
}

func parseMathIn(src string, env mathMacros) *mathExpr {
	working := make(mathMacros, len(env))
	for k, v := range env {
		working[k] = v
	}
	budget := 1024
	expanded, ok := expandMathMacros(src, working, 0, &budget)
	if !ok {
		return nil
	}
	p := mathReader{src: []rune(expanded)}
	n := p.row(0)
	if p.bad || p.i != len(p.src) {
		return nil
	}
	for k, v := range working {
		env[k] = v
	}
	return n
}
