package kit

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2/lexers"
)

type CodeSyntaxContext uint8

const (
	CodeSyntaxCode CodeSyntaxContext = iota
	CodeSyntaxString
	CodeSyntaxComment
)

// CodePair supports single- and multi-rune delimiters on one line.
type CodePair struct {
	Open, Close string
	NotIn       []CodeSyntaxContext
}

// CodeLanguageRules controls editing independently of highlighting. Nil
// AutoClosingPairs uses Brackets; an explicit empty slice disables auto pairs.
// Increase/Decrease are Go regex patterns, tested before/after the selection on
// Enter. Empty patterns use structural brackets (and Python's colon default).
type CodeLanguageRules struct {
	Brackets, AutoClosingPairs []CodePair
	AutoCloseBefore            string
	Increase, Decrease         string
}
type codeLanguageRules struct {
	config             CodeLanguageRules
	increase, decrease *regexp.Regexp
}

var codeRuleRegistry = struct {
	sync.RWMutex
	rules map[string]*codeLanguageRules
}{rules: map[string]*codeLanguageRules{}}

func codeLanguageKey(name string) string {
	if lexer := lexers.Get(name); lexer != nil {
		return strings.ToLower(lexer.Config().Name)
	}
	return name
}
func compileCodeRules(rules CodeLanguageRules) (*codeLanguageRules, error) {
	copyPairs := func(in []CodePair) ([]CodePair, error) {
		out := slices.Clone(in)
		for i, p := range out {
			if p.Open == "" || p.Close == "" || strings.ContainsAny(p.Open+p.Close, "\r\n") || utf8.RuneCountInString(p.Open) > 64 || utf8.RuneCountInString(p.Close) > 64 {
				return nil, fmt.Errorf("code rules: invalid pair %d", i)
			}
			out[i].NotIn = slices.Clone(p.NotIn)
			for _, c := range p.NotIn {
				if c > CodeSyntaxComment {
					return nil, fmt.Errorf("code rules: invalid syntax context")
				}
			}
		}
		slices.SortStableFunc(out, func(a, b CodePair) int { return utf8.RuneCountInString(b.Open) - utf8.RuneCountInString(a.Open) })
		return out, nil
	}
	r := &codeLanguageRules{config: rules}
	var err error
	if r.config.Brackets, err = copyPairs(rules.Brackets); err != nil {
		return nil, err
	}
	if r.config.AutoClosingPairs, err = copyPairs(rules.AutoClosingPairs); err != nil {
		return nil, err
	}
	if rules.AutoClosingPairs != nil && r.config.AutoClosingPairs == nil {
		r.config.AutoClosingPairs = []CodePair{}
	}
	if rules.Increase != "" {
		if r.increase, err = regexp.Compile(rules.Increase); err != nil {
			return nil, fmt.Errorf("code rules increase: %w", err)
		}
	}
	if rules.Decrease != "" {
		if r.decrease, err = regexp.Compile(rules.Decrease); err != nil {
			return nil, fmt.Errorf("code rules decrease: %w", err)
		}
	}
	return r, nil
}

// SetCodeLanguageRules replaces language defaults. Existing editors see it on
// their next edit. Chroma-recognized language names share their canonical key;
// custom names retain their spelling. Invalid input preserves the old rules.
func SetCodeLanguageRules(language string, rules CodeLanguageRules) error {
	if language == "" {
		return fmt.Errorf("code rules: empty language")
	}
	r, err := compileCodeRules(rules)
	if err != nil {
		return err
	}
	codeRuleRegistry.Lock()
	codeRuleRegistry.rules[codeLanguageKey(language)] = r
	codeRuleRegistry.Unlock()
	return nil
}
func ClearCodeLanguageRules(language string) {
	codeRuleRegistry.Lock()
	delete(codeRuleRegistry.rules, codeLanguageKey(language))
	codeRuleRegistry.Unlock()
}

// SetEditingRules installs an editor-local override; nil restores the registry.
// It preserves AutoClose and SmartIndent preferences, the document and history.
func (v *CodeEditorView) SetEditingRules(rules *CodeLanguageRules) error {
	if rules == nil {
		v.editingRules = nil
		return nil
	}
	r, err := compileCodeRules(*rules)
	if err == nil {
		v.editingRules = r
	}
	return err
}

// SmartIndent controls structural/regex indentation on Enter. False still keeps
// the current line's leading whitespace, but does not add or split indentation.
func (v *CodeEditorView) SmartIndent(on bool) *CodeEditorView { v.noSmartIndent = !on; return v }

// SyntaxContext overrides Chroma's code/string/comment classification for rules.
func (v *CodeEditorView) SyntaxContext(fn func(line, col int) CodeSyntaxContext) *CodeEditorView {
	v.syntaxContext = fn
	return v
}

var defaultCodeRules = func() *codeLanguageRules {
	brackets := []CodePair{{Open: "(", Close: ")"}, {Open: "[", Close: "]"}, {Open: "{", Close: "}"}}
	pairs := slices.Clone(brackets)
	pairs = append(pairs, CodePair{Open: "\"", Close: "\""}, CodePair{Open: "'", Close: "'"}, CodePair{Open: "`", Close: "`"})
	for i := range pairs {
		pairs[i].NotIn = []CodeSyntaxContext{CodeSyntaxString, CodeSyntaxComment}
	}
	r, _ := compileCodeRules(CodeLanguageRules{Brackets: brackets, AutoClosingPairs: pairs, AutoCloseBefore: ")]};:.,=>"})
	return r
}()

func (v *CodeEditorView) languageRules() *codeLanguageRules {
	if v.editingRules != nil {
		return v.editingRules
	}
	codeRuleRegistry.RLock()
	r := codeRuleRegistry.rules[codeLanguageKey(v.lang)]
	codeRuleRegistry.RUnlock()
	if r != nil {
		return r
	}
	return defaultCodeRules
}
func (r *codeLanguageRules) pairs() []CodePair {
	if r.config.AutoClosingPairs != nil {
		return r.config.AutoClosingPairs
	}
	return r.config.Brackets
}
func (v *CodeEditorView) syntaxAt(p codePos) CodeSyntaxContext {
	if v.syntaxContext != nil {
		kind := v.syntaxContext(p.line, p.col)
		if kind <= CodeSyntaxComment {
			return kind
		}
		return CodeSyntaxCode
	}
	if v.ruleRev != v.buf.revision || v.ruleLang != v.lang || v.ruleLine != p.line {
		v.buf.highlightNear(v.lang, codeStyleName(), p.line, p.line)
		v.ruleRev, v.ruleLang, v.ruleLine = v.buf.revision, v.lang, p.line
	}
	return CodeSyntaxContext(v.buf.kindAt(p))
}
func (v *CodeEditorView) typed(s codeSel, text string, _ rune, _ bool) codeReplace {
	from, to := s.span()
	rep := codeReplace{from: from, to: to, text: text, caret: -1}
	if !v.autoClose || text == "" || strings.ContainsAny(text, "\r\n") {
		return rep
	}
	rules := v.languageRules()
	before := string(v.buf.line(from.line)[:from.col])
	after := string(v.buf.line(to.line)[to.col:])
	kind := v.syntaxAt(from)
	for _, p := range rules.pairs() {
		if !s.empty() && from.line == to.line && text == p.Open {
			rep.text = p.Open + v.buf.slice(from, to) + p.Close
			rep.keep = utf8.RuneCountInString(p.Open)
			rep.keepEnd = utf8.RuneCountInString(p.Close)
			return rep
		}
		if !s.empty() {
			continue
		}
		// Skip a closer one rune at a time, including the middle of a multi-rune close.
		closeRunes := []rune(p.Close)
		for offset := 0; offset < len(closeRunes); offset++ {
			tail := string(closeRunes[offset:])
			if strings.HasPrefix(tail, text) && strings.HasPrefix(after, tail) && strings.HasSuffix(before, string(closeRunes[:offset])) && strings.Contains(before, p.Open) {
				rep.to = endOf(from, text)
				return rep
			}
		}
		if strings.HasSuffix(before+text, p.Open) && utf8.RuneCountInString(text) <= utf8.RuneCountInString(p.Open) && !slices.Contains(p.NotIn, kind) {
			next, _ := utf8.DecodeRuneInString(after)
			allowed := after == "" || unicode.IsSpace(next) || strings.ContainsRune(rules.config.AutoCloseBefore, next)
			if p.Open == p.Close {
				prefix := []rune(strings.TrimSuffix(before+text, p.Open))
				if len(prefix) > 0 && isIdent(prefix[len(prefix)-1]) {
					allowed = false
				}
			}
			if allowed {
				rep.text = text + p.Close
				rep.caret = utf8.RuneCountInString(text)
				return rep
			}
		}
	}
	if s.empty() {
		for _, p := range rules.config.Brackets {
			if text == p.Close && strings.TrimSpace(before) == "" {
				if open := v.matchingRuleOpen(from, p); open.line >= 0 {
					rep.from = codePos{from.line, 0}
					rep.text = v.buf.indent(open.line) + text
					return rep
				}
			}
		}
	}
	return rep
}
func (v *CodeEditorView) matchingRuleOpen(pos codePos, pair CodePair) codePos {
	depth := 0
	open, close := []rune(pair.Open), []rune(pair.Close)
	for line := pos.line; line >= 0 && pos.line-line < 2000; line-- {
		r := v.buf.line(line)
		end := len(r)
		if line == pos.line {
			end = pos.col
		}
		for i := end; i > 0; {
			if i >= len(close) && string(r[i-len(close):i]) == pair.Close && v.syntaxAt(codePos{line, i - len(close) + 1}) == CodeSyntaxCode {
				depth++
				i -= len(close)
				continue
			}
			if i >= len(open) && string(r[i-len(open):i]) == pair.Open && v.syntaxAt(codePos{line, i - len(open) + 1}) == CodeSyntaxCode {
				if depth == 0 {
					return codePos{line, i - len(open)}
				}
				depth--
				i -= len(open)
				continue
			}
			i--
		}
	}
	return codePos{-1, 0}
}
func (v *CodeEditorView) emptyPair(p codePos) (codePos, codePos, bool) {
	if !v.autoClose {
		return p, p, false
	}
	l := v.buf.line(p.line)
	before, after := string(l[:p.col]), string(l[p.col:])
	for _, pair := range v.languageRules().pairs() {
		if strings.HasSuffix(before, pair.Open) && strings.HasPrefix(after, pair.Close) {
			return codePos{p.line, p.col - utf8.RuneCountInString(pair.Open)}, codePos{p.line, p.col + utf8.RuneCountInString(pair.Close)}, true
		}
	}
	return p, p, false
}
func (v *CodeEditorView) newline(s codeSel) codeReplace {
	from, to := s.span()
	indent := v.buf.indent(from.line)
	rep := codeReplace{from: from, to: to, text: "\n" + indent, caret: -1}
	if v.noSmartIndent {
		return rep
	}
	rules := v.languageRules()
	before, after := string(v.buf.line(from.line)[:from.col]), string(v.buf.line(to.line)[to.col:])
	increase, decrease := false, false
	if rules.increase != nil {
		increase = rules.increase.MatchString(before)
	} else {
		for _, pair := range rules.config.Brackets {
			if strings.HasSuffix(strings.TrimSpace(before), pair.Open) {
				increase = true
				if strings.HasPrefix(strings.TrimSpace(after), pair.Close) {
					decrease = true
				}
			}
		}
		if rules == defaultCodeRules && codeLanguageKey(v.lang) == "python" && strings.HasSuffix(strings.TrimSpace(before), ":") {
			increase = true
		}
	}
	if rules.decrease != nil {
		decrease = rules.decrease.MatchString(after)
	} else {
		for _, pair := range rules.config.Brackets {
			if strings.HasPrefix(strings.TrimSpace(after), pair.Close) {
				decrease = true
				break
			}
		}
	}
	if increase {
		rep.text += v.indentUnit()
		if decrease {
			rep.caret = utf8.RuneCountInString(rep.text)
			rep.text += "\n" + indent
		}
	} else if decrease {
		rep.text = "\n" + v.outdentText(indent)
	}
	return rep
}
func (v *CodeEditorView) outdentText(indent string) string {
	if strings.HasSuffix(indent, "\t") {
		return strings.TrimSuffix(indent, "\t")
	}
	n := 0
	for n < len(indent) && n < v.tabSize && indent[len(indent)-1-n] == ' ' {
		n++
	}
	return indent[:len(indent)-n]
}
