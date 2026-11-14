package kit

import (
	"image"
	"strings"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/gesture"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// TerminalView displays a bounded ANSI screen and sends input bytes to an
// application-owned session. It never starts a process or echoes input itself.
// Call all methods under the UI frame lock (core.Update from other goroutines).
type TerminalView struct {
	semanticText                                                                          string
	semanticVersion                                                                       uint64
	semanticFirst, semanticLast                                                           int
	semanticKnown                                                                         bool
	draftSelection, imeRange                                                              key.Range
	touch                                                                                 bool
	touchStart                                                                            f32.Point
	touchID                                                                               pointer.ID
	model                                                                                 *terminalModel
	painter                                                                               *CodeEditorView
	name                                                                                  string
	height                                                                                float32
	fill, disabled, wantFocus, focused, follow, dragging, selected, tabExit, pastePending bool
	scroll                                                                                gesture.Scroll
	scrollY                                                                               float32
	anchor, end                                                                           codePos
	draft                                                                                 []rune
	composition                                                                           key.Range
	imeText                                                                               string
	imePos                                                                                codePos
	onData                                                                                func(string)
	onResize                                                                              func(int, int)
	lastCols, lastRows                                                                    int
}

func Terminal() *TerminalView {
	return &TerminalView{model: newTerminalModel(80, 24, 1000), painter: CodeEditor(""), height: 320, follow: true, composition: key.Range{Start: -1, End: -1}}
}
func (v *TerminalView) Name(s string) *TerminalView { v.name = s; return v }
func (v *TerminalView) Height(dp float32) *TerminalView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
		v.fill = false
	}
	return v
}
func (v *TerminalView) Fill() *TerminalView                            { v.fill = true; return v }
func (v *TerminalView) OnData(fn func(string)) *TerminalView           { v.onData = fn; return v }
func (v *TerminalView) OnResize(fn func(cols, rows int)) *TerminalView { v.onResize = fn; return v }
func (v *TerminalView) Focus()                                         { v.wantFocus = true }
func (v *TerminalView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.draft = nil
		v.draftSelection = key.Range{}
		v.composition = key.Range{Start: -1, End: -1}
		v.dragging = false
		v.pastePending = false
	}
}
func (v *TerminalView) Size() (cols, rows int) { return v.model.cols, v.model.rows }

// Scrollback sets the maximum retained history lines (0..10000).
func (v *TerminalView) Scrollback(lines int) *TerminalView {
	m := v.model
	limit := min(max(lines, 0), 10000)
	if limit == m.historyLimit {
		return v
	}
	var history [][]terminalCell
	for i := max(0, len(m.history)-limit); i < len(m.history); i++ {
		history = append(history, m.historyLine(i))
	}
	m.history, m.historyStart, m.historyLimit = history, 0, limit
	v.selected = false
	v.scrollY = 0
	v.follow = true
	return v
}

// Write consumes a chunk, retaining incomplete UTF-8 and escape sequences.
// Disabled terminals still accept output. Write does not invoke OnData.
func (v *TerminalView) Write(data []byte) (int, error) {
	m := v.model
	selection := v.Selection()
	before, alt := m.discarded, m.normal != nil
	m.write(data)
	if m.discarded > before && !v.follow {
		v.scrollY = max(0, v.scrollY-float32(m.discarded-before)*float32(v.painter.metrics.lh))
	}
	if v.selected {
		dropped := int(m.discarded - before)
		v.anchor.line -= dropped
		v.end.line -= dropped
		if v.anchor.line < 0 || v.end.line < 0 || v.anchor.line >= v.lineCount() || v.end.line >= v.lineCount() || v.Selection() != selection {
			v.selected = false
		}
	}
	if alt != (m.normal != nil) {
		v.follow = true
		v.selected = false
		v.scrollY = 0
	}
	return len(data), nil
}
func (v *TerminalView) WriteString(s string) (int, error) { return v.Write([]byte(s)) }

// Reset clears output, history, modes and selection without notifying OnData.
func (v *TerminalView) Reset() {
	v.model = newTerminalModel(v.model.cols, v.model.rows, v.model.historyLimit)
	v.semanticKnown = false
	v.scrollY = 0
	v.follow = true
	v.selected = false
	v.draft = nil
	v.draftSelection = key.Range{}
	v.composition = key.Range{Start: -1, End: -1}
}
func (v *TerminalView) Text() string { return v.model.text() }
func (v *TerminalView) lineCount() int {
	if v.model.normal != nil {
		return v.model.rows
	}
	return len(v.model.history) + v.model.rows
}
func (v *TerminalView) line(row int) []terminalCell {
	m := v.model
	row = min(max(row, 0), v.lineCount()-1)
	if m.normal == nil {
		if row < len(m.history) {
			return m.historyLine(row)
		}
		row -= len(m.history)
	}
	return m.screen.lines[row]
}

// Selection returns selected plain text, preserving complete wide graphemes.
func (v *TerminalView) Selection() string {
	if !v.selected {
		return ""
	}
	a, z := v.anchor, v.end
	if z.less(a) {
		a, z = z, a
	}
	var out strings.Builder
	for row := a.line; row <= z.line; row++ {
		line := v.line(row)
		start, end := 0, len(line)
		if row == a.line {
			start = a.col
		}
		if row == z.line {
			end = z.col
		}
		start, end = min(max(start, 0), len(line)), min(max(end, 0), len(line))
		for col := start; col < end; col++ {
			cell := line[col]
			if cell.width == 0 {
				continue
			}
			if cell.text == "" {
				out.WriteByte(' ')
			} else {
				out.WriteString(cell.text)
			}
		}
		if row < z.line {
			out.WriteByte('\n')
		}
	}
	return out.String()
}
func (v *TerminalView) Render(cx *el.Context) el.Element {
	box := el.Div().Disabled(v.disabled).Child(el.Widget(core.ViewportFunc(v.layoutViewport)).Grow())
	if v.fill {
		return box.Grow()
	}
	return box.H(el.Dp(v.height))
}

func (v *TerminalView) layoutViewport(gtx core.C, visible image.Rectangle) core.D {
	if visible.Empty() {
		return core.D{Size: gtx.Constraints.Max}
	}
	return v.layout(gtx)
}
