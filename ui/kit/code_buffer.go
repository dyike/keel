package kit

import (
	"image/color"
	"strings"
	"unicode"
)

// codePos is a place in a code buffer: a line and a rune column in it.
type codePos struct{ line, col int }

func (a codePos) less(b codePos) bool { return a.line < b.line || a.line == b.line && a.col < b.col }

// ordered returns a and b with the earlier first.
func ordered(a, b codePos) (codePos, codePos) {
	if b.less(a) {
		return b, a
	}
	return a, b
}

// codeSel is a selection: the anchor stays, the caret moves. Equal ends are
// a plain caret.
type codeSel struct{ anchor, caret codePos }

func (s codeSel) span() (codePos, codePos) { return ordered(s.anchor, s.caret) }
func (s codeSel) empty() bool              { return s.anchor == s.caret }

// codeKind is the syntax context of a stretch of text, for editing rules
// that behave differently in strings and comments.
type codeKind uint8

const (
	codeKindCode codeKind = iota
	codeKindString
	codeKindComment
)

// codeSpan colors runes [start, end) of a line. A zero color means plain text.
type codeSpan struct {
	start, end int
	color      color.NRGBA
	kind       codeKind
}

// codeLine is one line and its highlight.
type codeLine struct {
	text  []rune
	spans []codeSpan
}

// lineStore keeps lines in chunks of up to codeChunk lines, so inserting or
// deleting lines moves at most a chunk's worth of headers, not the whole
// file: the role a rope plays for text, at line granularity.
type lineStore struct {
	chunks [][]codeLine
	n      int
}

const codeChunk = 512

func (s *lineStore) len() int { return s.n }

// locate returns the chunk holding line i and the line's index in it; i ==
// len() is the end of the last chunk.
func (s *lineStore) locate(i int) (int, int) {
	for c, ch := range s.chunks {
		if i < len(ch) {
			return c, i
		}
		i -= len(ch)
	}
	last := len(s.chunks) - 1
	return last, len(s.chunks[last])
}

func (s *lineStore) at(i int) *codeLine {
	c, j := s.locate(i)
	return &s.chunks[c][j]
}

// each calls fn for lines from first on until it returns false.
func (s *lineStore) each(first int, fn func(i int, l *codeLine) bool) {
	c, j := s.locate(first)
	i := first
	for ; c < len(s.chunks); c, j = c+1, 0 {
		for ; j < len(s.chunks[c]); j++ {
			if !fn(i, &s.chunks[c][j]) {
				return
			}
			i++
		}
	}
}

// splice replaces lines [from, to) with repl.
func (s *lineStore) splice(from, to int, repl []codeLine) {
	fc, fo := s.locate(from)
	tc, tOff := s.locate(to)
	merged := make([]codeLine, 0, fo+len(repl)+len(s.chunks[tc])-tOff)
	merged = append(merged, s.chunks[fc][:fo]...)
	merged = append(merged, repl...)
	merged = append(merged, s.chunks[tc][tOff:]...)
	var pieces [][]codeLine
	for len(merged) > 2*codeChunk {
		pieces = append(pieces, merged[:codeChunk:codeChunk])
		merged = merged[codeChunk:]
	}
	if len(merged) > 0 {
		pieces = append(pieces, merged)
	}
	rest := append([][]codeLine(nil), s.chunks[tc+1:]...)
	s.chunks = append(append(s.chunks[:fc], pieces...), rest...)
	if len(s.chunks) == 0 {
		s.chunks = [][]codeLine{{{}}}
	}
	s.n += len(repl) - (to - from)
}

func newLineStore(lines []codeLine) lineStore {
	s := lineStore{n: len(lines)}
	for len(lines) > 0 {
		k := min(codeChunk, len(lines))
		s.chunks = append(s.chunks, lines[:k:k])
		lines = lines[k:]
	}
	return s
}

// codeEdit is one replacement in an undo step: text old from from to to was
// replaced by new, which ends at end.
type codeEdit struct {
	from, to, end codePos
	old, new      string
}

// codeStep is what undo reverts at once: the edits of one command, across
// every selection, and the selections around it.
type codeStep struct {
	edits         []codeEdit
	before, after []codeSel
	typing        bool // a run of typed characters, merged while it lasts
}

// codeBuffer holds the text, its highlight and its undo history. Columns
// count runes.
type codeBuffer struct {
	lines      lineStore
	undo, redo []codeStep
	open       *codeStep // the step being recorded between begin and commit
	revision   uint64    // changes on every edit, for highlighting and caches
	// onSplice tells the editor lines changed: from on, removed lines were
	// replaced by added ones; removed < 0 means the whole text was replaced.
	onSplice func(from, removed, added int)
}

func newCodeBuffer(text string) *codeBuffer {
	b := &codeBuffer{}
	b.set(text)
	return b
}

func splitLines(text string) []codeLine {
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]codeLine, len(parts))
	for i, p := range parts {
		out[i].text = []rune(p)
	}
	return out
}

func (b *codeBuffer) set(text string) {
	b.lines = newLineStore(splitLines(text))
	b.undo, b.redo, b.open = nil, nil, nil
	b.revision++
	if b.onSplice != nil {
		b.onSplice(0, -1, 0)
	}
}

func (b *codeBuffer) count() int             { return b.lines.len() }
func (b *codeBuffer) line(i int) []rune      { return b.lines.at(i).text }
func (b *codeBuffer) spans(i int) []codeSpan { return b.lines.at(i).spans }

func (b *codeBuffer) text() string {
	var sb strings.Builder
	b.lines.each(0, func(i int, l *codeLine) bool {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(string(l.text))
		return true
	})
	return sb.String()
}

// clamp moves p inside the text.
func (b *codeBuffer) clamp(p codePos) codePos {
	p.line = min(max(p.line, 0), b.count()-1)
	p.col = min(max(p.col, 0), len(b.line(p.line)))
	return p
}

// slice returns the text between two positions.
func (b *codeBuffer) slice(from, to codePos) string {
	from, to = ordered(b.clamp(from), b.clamp(to))
	if from.line == to.line {
		return string(b.line(from.line)[from.col:to.col])
	}
	var sb strings.Builder
	b.lines.each(from.line, func(i int, l *codeLine) bool {
		switch {
		case i == from.line:
			sb.WriteString(string(l.text[from.col:]))
		case i == to.line:
			sb.WriteByte('\n')
			sb.WriteString(string(l.text[:to.col]))
			return false
		default:
			sb.WriteByte('\n')
			sb.WriteString(string(l.text))
		}
		return true
	})
	return sb.String()
}

// replace swaps the text between from and to for text and returns where the
// new text ends. It records no undo; edit does. Lines it touches lose their
// highlight until the next pass.
func (b *codeBuffer) replace(from, to codePos, text string) codePos {
	from, to = ordered(b.clamp(from), b.clamp(to))
	head := b.line(from.line)[:from.col]
	tail := b.line(to.line)[to.col:]
	repl := splitLines(text)
	end := codePos{from.line + len(repl) - 1, len(repl[len(repl)-1].text)}
	if len(repl) == 1 {
		end.col += from.col
	}
	repl[0].text = append(append([]rune(nil), head...), repl[0].text...)
	last := &repl[len(repl)-1]
	last.text = append(last.text[:len(last.text):len(last.text)], tail...)
	b.lines.splice(from.line, to.line+1, repl)
	b.revision++
	if b.onSplice != nil {
		b.onSplice(from.line, to.line-from.line+1, len(repl))
	}
	return end
}

// begin starts an undo step; edits until commit undo together.
func (b *codeBuffer) begin(before []codeSel) {
	b.open = &codeStep{before: append([]codeSel(nil), before...)}
}

// edit replaces text, inside the open step if any, and returns where the
// new text ends.
func (b *codeBuffer) edit(from, to codePos, text string) codePos {
	from, to = ordered(b.clamp(from), b.clamp(to))
	old := b.slice(from, to)
	end := b.replace(from, to, text)
	if b.open != nil {
		b.open.edits = append(b.open.edits, codeEdit{from: from, to: to, end: end, old: old, new: text})
	}
	return end
}

// commit closes the step. A typed character on one caret merges with the
// typed run before it, so undo removes a word, not a letter.
func (b *codeBuffer) commit(after []codeSel, typing bool) {
	s := b.open
	b.open = nil
	if s == nil || len(s.edits) == 0 {
		return
	}
	s.after, s.typing = append([]codeSel(nil), after...), typing
	if n := len(b.undo); n > 0 && typing && len(s.edits) == 1 && s.edits[0].old == "" && !strings.Contains(s.edits[0].new, "\n") {
		prev := &b.undo[n-1]
		e := s.edits[0]
		if prev.typing && len(prev.edits) == 1 && prev.edits[0].end == e.from && !isWordBreak(e.new) {
			pe := &prev.edits[0]
			pe.new += e.new
			pe.end = e.end
			prev.after = s.after
			b.redo = nil
			return
		}
	}
	b.undo = append(b.undo, *s)
	b.redo = nil
}

// isWordBreak ends a typed run: a space starts a new undo step.
func isWordBreak(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// undoStep reverts the last step and returns the selections before it.
func (b *codeBuffer) undoStep() ([]codeSel, bool) {
	if len(b.undo) == 0 {
		return nil, false
	}
	s := b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]
	for i := len(s.edits) - 1; i >= 0; i-- {
		e := s.edits[i]
		b.replace(e.from, e.end, e.old)
	}
	b.redo = append(b.redo, s)
	return s.before, true
}

// redoStep reapplies the last undone step and returns the selections after it.
func (b *codeBuffer) redoStep() ([]codeSel, bool) {
	if len(b.redo) == 0 {
		return nil, false
	}
	s := b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]
	for i, e := range s.edits {
		s.edits[i].end = b.replace(e.from, e.to, e.new)
	}
	b.undo = append(b.undo, s)
	return s.after, true
}

// endOf is where text inserted at from ends.
func endOf(from codePos, text string) codePos {
	n := strings.Count(text, "\n")
	if n == 0 {
		return codePos{from.line, from.col + len([]rune(text))}
	}
	return codePos{from.line + n, len([]rune(text[strings.LastIndexByte(text, '\n')+1:]))}
}

// shiftPos moves p for an edit that replaced text up to to with text ending
// at end; positions before to stay.
func shiftPos(p, to, end codePos) codePos {
	if p.less(to) {
		return p
	}
	if p.line == to.line {
		return codePos{end.line, end.col + p.col - to.col}
	}
	return codePos{p.line + end.line - to.line, p.col}
}

// wordAt returns the word (letters, digits, underscore) around a position.
func (b *codeBuffer) wordAt(p codePos) (codePos, codePos) {
	p = b.clamp(p)
	l := b.line(p.line)
	start, end := p.col, p.col
	for start > 0 && isIdent(l[start-1]) {
		start--
	}
	for end < len(l) && isIdent(l[end]) {
		end++
	}
	return codePos{p.line, start}, codePos{p.line, end}
}

func isIdent(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }

// indent is the leading whitespace of a line.
func (b *codeBuffer) indent(line int) string {
	l := b.line(line)
	n := 0
	for n < len(l) && (l[n] == ' ' || l[n] == '\t') {
		n++
	}
	return string(l[:n])
}

// blank reports whether a line is only whitespace.
func (b *codeBuffer) blank(line int) bool { return len([]rune(b.indent(line))) == len(b.line(line)) }

// kindAt is the syntax context of the rune before p, code if unknown.
func (b *codeBuffer) kindAt(p codePos) codeKind {
	if p.col == 0 {
		return codeKindCode
	}
	for _, s := range b.spans(p.line) {
		if s.start < p.col && p.col <= s.end {
			return s.kind
		}
	}
	return codeKindCode
}
