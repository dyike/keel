package core

import (
	"fmt"
	"strings"
	"unicode"
)

// A key context predicate chooses where a contextual binding applies, as in
// GPUI keymaps:
//
//	Editor               the focused level is an Editor
//	Editor && !ReadOnly  an Editor that is not read-only
//	Pane > Editor        an Editor inside a Pane
//	Terminal || Editor   either
//
// Each level of the focus path is one el KeyContext, whose name may hold
// several space-separated identifiers ("Editor ReadOnly"). Identifiers test
// the level being matched; a > b matches b there and a at some outer level.
// ! binds tightest, then >, then &&, then ||; parentheses group.

type predicate interface {
	match(levels [][]string, at int) bool
}

type (
	predIdent string
	predNot   struct{ p predicate }
	predAnd   struct{ a, b predicate }
	predOr    struct{ a, b predicate }
	predChild struct{ outer, inner predicate }
)

func (p predIdent) match(levels [][]string, at int) bool {
	for _, id := range levels[at] {
		if id == string(p) {
			return true
		}
	}
	return false
}
func (p predNot) match(l [][]string, at int) bool { return !p.p.match(l, at) }
func (p predAnd) match(l [][]string, at int) bool { return p.a.match(l, at) && p.b.match(l, at) }
func (p predOr) match(l [][]string, at int) bool  { return p.a.match(l, at) || p.b.match(l, at) }
func (p predChild) match(l [][]string, at int) bool {
	if !p.inner.match(l, at) {
		return false
	}
	for outer := at + 1; outer < len(l); outer++ { // levels run inner to outer
		if p.outer.match(l, outer) {
			return true
		}
	}
	return false
}

// parsePredicate reads a context predicate. A bare name, or several names
// separated by spaces, is a plain context: all of the names at one level.
func parsePredicate(s string) (predicate, error) {
	p := &predParser{toks: predTokens(s)}
	if len(p.toks) == 0 {
		return nil, fmt.Errorf("keymap: empty context")
	}
	e, err := p.or()
	if err != nil {
		return nil, fmt.Errorf("keymap: context %q: %w", s, err)
	}
	if p.i < len(p.toks) {
		return nil, fmt.Errorf("keymap: context %q: unexpected %q", s, p.toks[p.i])
	}
	return e, nil
}

func predTokens(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		r := rune(s[i])
		switch {
		case unicode.IsSpace(r):
			i++
		case strings.HasPrefix(s[i:], "&&"), strings.HasPrefix(s[i:], "||"):
			out = append(out, s[i:i+2])
			i += 2
		case strings.ContainsRune("!()>", r):
			out = append(out, string(r))
			i++
		default:
			j := i
			for j < len(s) && !unicode.IsSpace(rune(s[j])) && !strings.ContainsRune("!()>&|", rune(s[j])) {
				j++
			}
			if j == i {
				out = append(out, string(r)) // a lone & or |
				i++
				continue
			}
			out = append(out, s[i:j])
			i = j
		}
	}
	return out
}

type predParser struct {
	toks []string
	i    int
}

func (p *predParser) peek() string {
	if p.i < len(p.toks) {
		return p.toks[p.i]
	}
	return ""
}

func (p *predParser) or() (predicate, error) {
	a, err := p.and()
	for err == nil && p.peek() == "||" {
		p.i++
		var b predicate
		if b, err = p.and(); err == nil {
			a = predOr{a, b}
		}
	}
	return a, err
}

func (p *predParser) and() (predicate, error) {
	a, err := p.child()
	for err == nil {
		switch t := p.peek(); {
		case t == "&&":
			p.i++
		case t != "" && t != "||" && t != ")" && t != ">":
			// Adjacent names ("Editor ReadOnly") all hold at one level.
		default:
			return a, nil
		}
		var b predicate
		if b, err = p.child(); err == nil {
			a = predAnd{a, b}
		}
	}
	return a, err
}

func (p *predParser) child() (predicate, error) {
	a, err := p.unary()
	for err == nil && p.peek() == ">" {
		p.i++
		var b predicate
		if b, err = p.unary(); err == nil {
			a = predChild{outer: a, inner: b}
		}
	}
	return a, err
}

func (p *predParser) unary() (predicate, error) {
	switch t := p.peek(); t {
	case "":
		return nil, fmt.Errorf("unexpected end")
	case "!":
		p.i++
		e, err := p.unary()
		return predNot{e}, err
	case "(":
		p.i++
		e, err := p.or()
		if err != nil {
			return nil, err
		}
		if p.peek() != ")" {
			return nil, fmt.Errorf("missing )")
		}
		p.i++
		return e, nil
	case ")", ">", "&&", "||", "&", "|":
		return nil, fmt.Errorf("unexpected %q", t)
	default:
		p.i++
		return predIdent(t), nil
	}
}

// contextLevels splits each level's KeyContext name into its identifiers.
func contextLevels(contexts []string) [][]string {
	out := make([][]string, len(contexts))
	for i, c := range contexts {
		out[i] = strings.Fields(c)
	}
	return out
}
