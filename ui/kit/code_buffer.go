package kit

import (
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

// codeEdit is one undoable change: text old at from was replaced by text new,
// ending at to. Caret positions before and after restore the selection.
type codeEdit struct {
	from, to      codePos
	old, new      string
	before, after [2]codePos // anchor, caret
	typing        bool       // a run of typed characters, merged while it lasts
}

// codeBuffer holds text as lines of runes: an edit inside a line touches only
// that line, and inserting lines into a 200,000-line file moves a slice of
// line headers, not the text. Columns count runes.
type codeBuffer struct {
	lines      [][]rune
	undo, redo []codeEdit
	revision   uint64 // changes on every edit, for highlighting and caches
	// spans are the highlight colors per line, from a background pass. Edits
	// splice them along with the lines, so colors stay on their text until
	// the next pass; an edited line is plain until then.
	spans [][]codeSpan
}

func newCodeBuffer(text string) *codeBuffer {
	b := &codeBuffer{}
	b.set(text)
	return b
}

func (b *codeBuffer) set(text string) {
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	b.lines = make([][]rune, len(parts))
	for i, p := range parts {
		b.lines[i] = []rune(p)
	}
	b.undo, b.redo, b.spans = nil, nil, nil
	b.revision++
}

func (b *codeBuffer) text() string {
	var sb strings.Builder
	for i, l := range b.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(string(l))
	}
	return sb.String()
}

// clamp moves p inside the text.
func (b *codeBuffer) clamp(p codePos) codePos {
	p.line = min(max(p.line, 0), len(b.lines)-1)
	p.col = min(max(p.col, 0), len(b.lines[p.line]))
	return p
}

// slice returns the text between two positions.
func (b *codeBuffer) slice(from, to codePos) string {
	from, to = ordered(b.clamp(from), b.clamp(to))
	if from.line == to.line {
		return string(b.lines[from.line][from.col:to.col])
	}
	var sb strings.Builder
	sb.WriteString(string(b.lines[from.line][from.col:]))
	for l := from.line + 1; l < to.line; l++ {
		sb.WriteByte('\n')
		sb.WriteString(string(b.lines[l]))
	}
	sb.WriteByte('\n')
	sb.WriteString(string(b.lines[to.line][:to.col]))
	return sb.String()
}

// replace swaps the text between from and to for text and returns where the
// new text ends. It records no undo; edit does.
func (b *codeBuffer) replace(from, to codePos, text string) codePos {
	from, to = ordered(b.clamp(from), b.clamp(to))
	head := b.lines[from.line][:from.col:from.col]
	tail := b.lines[to.line][to.col:]
	parts := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	repl := make([][]rune, len(parts))
	for i, p := range parts {
		repl[i] = []rune(p)
	}
	end := codePos{from.line + len(repl) - 1, len(repl[len(repl)-1])}
	if len(repl) == 1 {
		end.col += from.col
	}
	first := append(append([]rune(nil), head...), repl[0]...)
	last := repl[len(repl)-1]
	if len(repl) == 1 {
		last = first
	}
	last = append(last[:len(last):len(last)], tail...)
	repl[0] = first
	repl[len(repl)-1] = last
	if len(b.spans) == len(b.lines) {
		b.spans = append(b.spans[:from.line], append(make([][]codeSpan, len(repl)), b.spans[to.line+1:]...)...)
	}
	b.lines = append(b.lines[:from.line], append(repl, b.lines[to.line+1:]...)...)
	b.revision++
	return end
}

// edit replaces text and records it for undo. A typed run on one line merges
// with the previous typed run, so undo removes a word, not a letter.
func (b *codeBuffer) edit(from, to codePos, text string, sel [2]codePos, typing bool) codePos {
	from, to = ordered(b.clamp(from), b.clamp(to))
	old := b.slice(from, to)
	end := b.replace(from, to, text)
	e := codeEdit{from: from, to: end, old: old, new: text, before: sel, after: [2]codePos{end, end}, typing: typing}
	if n := len(b.undo); n > 0 && typing && old == "" && !strings.Contains(text, "\n") {
		prev := &b.undo[n-1]
		if prev.typing && prev.to == from && !isWordBreak(text) {
			prev.new += text
			prev.to = end
			prev.after = e.after
			b.redo = nil
			return end
		}
	}
	b.undo = append(b.undo, e)
	b.redo = nil
	return end
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

// undoOne reverts the last edit and returns the selection before it.
func (b *codeBuffer) undoOne() ([2]codePos, bool) {
	if len(b.undo) == 0 {
		return [2]codePos{}, false
	}
	e := b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]
	b.replace(e.from, e.to, e.old)
	b.redo = append(b.redo, e)
	return e.before, true
}

// redoOne reapplies the last undone edit and returns the selection after it.
func (b *codeBuffer) redoOne() ([2]codePos, bool) {
	if len(b.redo) == 0 {
		return [2]codePos{}, false
	}
	e := b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]
	end := b.replace(e.from, b.endOf(e.from, e.old), e.new)
	e.to = end
	b.undo = append(b.undo, e)
	return e.after, true
}

// endOf is where text inserted at from ends.
func (b *codeBuffer) endOf(from codePos, text string) codePos {
	n := strings.Count(text, "\n")
	if n == 0 {
		return codePos{from.line, from.col + len([]rune(text))}
	}
	return codePos{from.line + n, len([]rune(text[strings.LastIndexByte(text, '\n')+1:]))}
}

// wordAt returns the word (letters, digits, underscore) around a position.
func (b *codeBuffer) wordAt(p codePos) (codePos, codePos) {
	p = b.clamp(p)
	l := b.lines[p.line]
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

// indent is the leading whitespace of a line, kept on Enter.
func (b *codeBuffer) indent(line int) string {
	l := b.lines[line]
	n := 0
	for n < len(l) && (l[n] == ' ' || l[n] == '\t') {
		n++
	}
	return string(l[:n])
}
