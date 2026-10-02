package kit

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"gioui.org/io/key"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// codeRange is a stretch of text from from to to.
type codeRange struct{ from, to codePos }

// searchState is the find panel and its results.
type searchState struct {
	open, replace, searchable bool
	query, replacement        string
	matchCase, wholeWord      bool
	regex                     bool
	badRegex                  bool
	matches                   []codeRange
	current                   int // index into matches, -1 if none chosen
	rev                       uint64
	key                       string
	focusQuery                bool
}

// codeSearchLimit stops counting matches in very large files.
const codeSearchLimit = 10000

// OpenSearch shows the find panel, with the replace row if replace is set
// and the editor is editable, starting from the selected text.
func (v *CodeEditorView) OpenSearch(replace bool) {
	if !v.search.searchable {
		return
	}
	s := &v.search
	s.open, s.replace, s.focusQuery = true, replace && !v.readOnly, true
	if sel := v.primary(); !sel.empty() && sel.anchor.line == sel.caret.line {
		s.query = v.buf.slice(sel.anchor, sel.caret)
	}
	s.key = ""
}

// CloseSearch hides the find panel and returns focus to the text.
func (v *CodeEditorView) CloseSearch() {
	if v.search.open {
		v.search.open = false
		v.search.matches = nil
		v.Focus()
	}
}

// SearchMatches is how many matches the find panel found, up to 10,000.
func (v *CodeEditorView) SearchMatches() int { v.refreshMatches(); return len(v.search.matches) }

// refreshMatches recomputes matches when the text or the query changed.
func (v *CodeEditorView) refreshMatches() {
	s := &v.search
	key := s.query + "\x00" + strconv.FormatBool(s.matchCase) + strconv.FormatBool(s.wholeWord) + strconv.FormatBool(s.regex)
	if !s.open || s.rev == v.buf.revision && s.key == key {
		return
	}
	s.rev, s.key = v.buf.revision, key
	s.matches, s.badRegex = nil, false
	if s.query == "" {
		s.current = -1
		return
	}
	find := v.matcher()
	if find == nil {
		s.badRegex = true
		return
	}
	v.buf.lines.each(0, func(i int, l *codeLine) bool {
		for _, m := range find(l.text) {
			s.matches = append(s.matches, codeRange{codePos{i, m[0]}, codePos{i, m[1]}})
		}
		return len(s.matches) < codeSearchLimit
	})
	s.current = v.matchAt(v.primary().span())
}

// matcher returns a function listing match column ranges in a line, or nil
// for an invalid regular expression.
func (v *CodeEditorView) matcher() func([]rune) [][2]int {
	s := v.search
	var re *regexp.Regexp
	if s.regex {
		expr := s.query
		if !s.matchCase {
			expr = "(?i)" + expr
		}
		var err error
		if re, err = regexp.Compile(expr); err != nil {
			return nil
		}
	}
	needle := []rune(s.query)
	if !s.matchCase {
		needle = []rune(strings.ToLower(s.query))
	}
	word := func(l []rune, a, b int) bool {
		return !s.wholeWord || (a == 0 || !isIdent(l[a-1])) && (b == len(l) || !isIdent(l[b]))
	}
	return func(l []rune) [][2]int {
		var out [][2]int
		if re != nil {
			str := string(l)
			for _, m := range re.FindAllStringIndex(str, -1) {
				if m[0] == m[1] {
					continue
				}
				a, b := utf8.RuneCountInString(str[:m[0]]), utf8.RuneCountInString(str[:m[1]])
				if word(l, a, b) {
					out = append(out, [2]int{a, b})
				}
			}
			return out
		}
		for i := 0; i+len(needle) <= len(l); i++ {
			ok := true
			for j, r := range needle {
				c := l[i+j]
				if !s.matchCase {
					c = unicode.ToLower(c)
				}
				if c != r {
					ok = false
					break
				}
			}
			if ok && word(l, i, i+len(needle)) {
				out = append(out, [2]int{i, i + len(needle)})
				i += len(needle) - 1
			}
		}
		return out
	}
}

// matchAt is the index of the match equal to from..to, or -1.
func (v *CodeEditorView) matchAt(from, to codePos) int {
	ms := v.search.matches
	i := sort.Search(len(ms), func(i int) bool { return !ms[i].from.less(from) })
	if i < len(ms) && ms[i].from == from && ms[i].to == to {
		return i
	}
	return -1
}

// findNext selects the next match after the primary caret (dir 1) or the one
// before it (dir -1), wrapping around.
func (v *CodeEditorView) findNext(dir int) {
	v.refreshMatches()
	ms := v.search.matches
	if len(ms) == 0 {
		return
	}
	from, to := v.primary().span()
	var i int
	switch cur := v.matchAt(from, to); {
	case cur >= 0:
		i = (cur + dir + len(ms)) % len(ms)
	case dir > 0:
		i = sort.Search(len(ms), func(i int) bool { return !ms[i].from.less(to) }) % len(ms)
	default:
		i = sort.Search(len(ms), func(i int) bool { return !ms[i].from.less(from) }) - 1
		if i < 0 {
			i = len(ms) - 1
		}
	}
	v.search.current = i
	v.sels, v.prim = []codeSel{{ms[i].from, ms[i].to}}, 0
	v.reveal = true
	v.keepCaretsVisible()
}

// replacementFor is the text replacing a match: with a regular expression,
// $1 and the like expand.
func (v *CodeEditorView) replacementFor(m codeRange) string {
	s := v.search
	if !s.regex {
		return s.replacement
	}
	expr := s.query
	if !s.matchCase {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return s.replacement
	}
	src := v.buf.slice(m.from, m.to)
	return re.ReplaceAllString(src, s.replacement)
}

// replaceOne replaces the selected match and moves to the next one.
func (v *CodeEditorView) replaceOne() {
	if v.readOnly {
		return
	}
	v.refreshMatches()
	from, to := v.primary().span()
	i := v.matchAt(from, to)
	if i < 0 {
		v.findNext(1)
		return
	}
	m := v.search.matches[i]
	v.buf.begin(v.sels)
	end := v.buf.edit(m.from, m.to, v.replacementFor(m))
	v.sels, v.prim = []codeSel{{end, end}}, 0
	v.buf.commit(v.sels, false)
	v.changedNoCall()
	v.findNext(1)
}

// replaceAll replaces every match as one undo step.
func (v *CodeEditorView) replaceAll() {
	if v.readOnly {
		return
	}
	v.refreshMatches()
	ms := append([]codeRange(nil), v.search.matches...)
	if len(ms) == 0 {
		return
	}
	reps := make([]codeReplace, len(ms))
	for i, m := range ms {
		reps[i] = codeReplace{from: m.from, to: m.to, text: v.replacementFor(m), caret: -1}
	}
	v.applyReplaces(reps, false)
	v.changedNoCall()
}

func (v *CodeEditorView) searchPanel(cx *el.Context) el.Element {
	text := locale.Current()
	s := &v.search
	v.refreshMatches()
	base := autoID("code-search", v)
	queryID, replID := base+"/query", base+"/replace"
	if s.focusQuery {
		cx.Focus(queryID)
		s.focusQuery = false
	}
	count := text.NoMatches
	switch {
	case s.badRegex:
		count = text.InvalidPattern
	case len(s.matches) > 0:
		cur := "?"
		if s.current >= 0 {
			cur = strconv.Itoa(s.current + 1)
		}
		count = cur + "/" + strconv.Itoa(len(s.matches))
		if len(s.matches) >= codeSearchLimit {
			count += "+"
		}
	case s.query == "":
		count = ""
	}
	toggle := func(label, name string, on *bool) el.Element {
		b := Button(label, func() { *on = !*on }).Name(name).Size(24)
		if *on {
			b.Variant(ButtonSecondary)
		} else {
			b.Variant(ButtonGhost)
		}
		return b.Render(cx)
	}
	icon := func(name string, ic IconName, fn func()) el.Element {
		return Button("", fn).Name(name).Icon(ic).Variant(ButtonGhost).Size(24).Render(cx)
	}
	query := el.Input().ID(queryID).Name(text.Find).Placeholder(text.Find).Bind(&s.query).
		OnChange(func(string) { s.key = "" }).
		OnSubmit(func(string) { v.findNext(1) }).
		OnKey(func(e el.KeyEvent) bool {
			if e.State != el.KeyPress {
				return true
			}
			switch key.Name(e.Name) {
			case key.NameUpArrow:
				v.findNext(-1)
			case key.NameDownArrow:
				v.findNext(1)
			default:
				return false
			}
			return true
		})
	row := el.Div().Row().Items(el.Center).Gap(2).Child(
		el.Div().W(el.Dp(220)).Child(searchField(cx, base+"/queryframe", queryID, query)),
		toggle("Aa", text.MatchCase, &s.matchCase),
		toggle("W", text.WholeWord, &s.wholeWord),
		toggle(".*", text.RegularExpression, &s.regex),
		el.Div().MinW(el.Dp(56)).Px(4).Child(el.Text(count).TextSize(theme.TextSm).TextColor(theme.Muted)),
		icon(text.PrevMatch, IconChevronLeft, func() { v.findNext(-1) }),
		icon(text.NextMatch, IconChevronRight, func() { v.findNext(1) }),
	)
	if !v.readOnly {
		row.Child(icon(text.Replace, IconEdit, func() { s.replace = !s.replace }))
	}
	row.Child(icon(text.Close, IconClose, v.CloseSearch))
	panel := floating(theme.ElevationMd).Role("search").Name(text.Find).P(6).Gap(6).Items(el.Start).Child(row)
	if s.replace && !v.readOnly {
		repl := fieldText(el.Input().ID(replID).Name(text.Replace).Placeholder(text.Replace).Bind(&s.replacement).
			OnSubmit(func(string) { v.replaceOne() }))
		panel.Child(el.Div().Row().Items(el.Center).Gap(4).Child(
			el.Div().W(el.Dp(220)).Child(fieldFrame(base+"/replaceframe", cx.FocusWithin(base+"/replaceframe"), false, false, false).FocusOnPress(replID).Child(repl)),
			Button(text.Replace, v.replaceOne).Variant(ButtonSecondary).Size(24).Render(cx),
			Button(text.ReplaceAll, v.replaceAll).Variant(ButtonSecondary).Size(24).Render(cx),
		))
	}
	return el.Div().Absolute().Top(8).Right(16).OnKey(func(e el.KeyEvent) bool {
		if key.Name(e.Name) != key.NameEscape {
			return false
		}
		if e.State == el.KeyPress {
			v.CloseSearch()
		}
		return true
	}).Child(panel)
}
