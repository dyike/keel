package kit

// CodeRange is half-open, zero-based line/rune-column coordinates.
type CodeRange struct{ Line, Col, EndLine, EndCol int }

func (r CodeRange) positions() (codePos, codePos) {
	return codePos{r.Line, r.Col}, codePos{r.EndLine, r.EndCol}
}
func publicCodeRange(a, b codePos) CodeRange { return CodeRange{a.line, a.col, b.line, b.col} }

type CodeSearchOptions struct{ MatchCase, WholeWord, Regex bool }

// CodeSearchSession is an owned snapshot; Current is -1 without a selected
// match. Truncated means only the first 10,000 matches are listed. ReplaceAll
// still visits the entire document. Plain queries match within a line;
// regular expressions search the whole text and may span lines. Empty hits
// are skipped.
type CodeSearchSession struct {
	Query, Replacement                           string
	Options                                      CodeSearchOptions
	Active, PanelOpen, InvalidPattern, Truncated bool
	Current                                      int
	Matches                                      []CodeRange
}

// SetSearchQuery starts a search without opening or focusing the built-in panel.
// It also works with Searchable(false) and on read-only/disabled editors.
func (v *CodeEditorView) SetSearchQuery(query string, options CodeSearchOptions) {
	s := &v.search
	s.query = query
	s.matchCase, s.wholeWord, s.regex = options.MatchCase, options.WholeWord, options.Regex
	s.active = true
	s.key = ""
	v.refreshMatches()
}
func (v *CodeEditorView) SearchSession() CodeSearchSession {
	v.refreshMatches()
	s := v.search
	out := CodeSearchSession{Query: s.query, Replacement: s.replacement, Options: CodeSearchOptions{s.matchCase, s.wholeWord, s.regex}, Active: s.active || s.open, PanelOpen: s.open, InvalidPattern: s.badRegex, Truncated: s.truncated, Current: s.current}
	for _, m := range s.matches {
		out.Matches = append(out.Matches, publicCodeRange(m.from, m.to))
	}
	return out
}
func (v *CodeEditorView) NextSearchMatch()     { v.findNext(1) }
func (v *CodeEditorView) PreviousSearchMatch() { v.findNext(-1) }

// SelectSearchMatch selects a listed match, unfolds it and scrolls it into view.
func (v *CodeEditorView) SelectSearchMatch(index int) bool {
	v.refreshMatches()
	if index < 0 || index >= len(v.search.matches) {
		return false
	}
	m := v.search.matches[index]
	v.search.current = index
	v.sels, v.prim = []codeSel{{m.from, m.to}}, 0
	v.reveal = true
	v.keepCaretsVisible()
	return true
}

// ReplaceCurrentSearchMatch replaces only an exactly selected match. It returns
// false when no match is selected or the editor is read-only/disabled.
func (v *CodeEditorView) ReplaceCurrentSearchMatch(replacement string) bool {
	if v.disabled || v.readOnly {
		return false
	}
	v.refreshMatches()
	if v.matchAt(v.primary().span()) < 0 {
		return false
	}
	v.search.replacement = replacement
	v.replaceOne()
	return true
}

// ReplaceAllSearchMatches replaces every match as a single undo step and
// emits one OnChange. Its count is not limited by the displayed result cap.
func (v *CodeEditorView) ReplaceAllSearchMatches(replacement string) int {
	v.search.replacement = replacement
	return v.replaceAllSearch()
}
