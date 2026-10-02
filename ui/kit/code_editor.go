package kit

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/key"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// CodeSeverity is how serious a CodeDiagnostic is; it picks its color.
type CodeSeverity uint8

const (
	CodeSeverityError CodeSeverity = iota
	CodeSeverityWarning
	CodeSeverityInfo
)

// CodeDiagnostic marks a range of a CodeEditor, as a language server reports
// it: a squiggle under the text, a dot in the gutter and the message on
// hover. Lines and columns count from 0; columns count runes.
type CodeDiagnostic struct {
	Line, Col, EndLine, EndCol int
	Severity                   CodeSeverity
	Message                    string
}

// CodeCompletion is one suggestion in the completion list. Insert replaces
// the word before the caret; it defaults to Label.
type CodeCompletion struct {
	Label, Detail, Insert string
}

// CodeEditorView edits source code: line numbers, syntax highlighting,
// selection, undo, clipboard and input methods. Only the visible lines are
// laid out, so files of hundreds of thousands of lines stay responsive, and
// highlighting runs in the background after edits.
//
// It does not speak a language server protocol itself. The app connects one
// through SetDiagnostics, OnComplete and OnHover.
//
// Tab indents; press Esc first to move focus on with Tab.
type CodeEditorView struct {
	buf              *codeBuffer
	lang, name       string
	height           float32
	fill             bool
	readOnly         bool
	disabled         bool
	anchor, caret    codePos
	goalX            int // px, kept across vertical moves
	scrollX, scrollY float32
	reveal           bool // scroll the caret into view on the next frame
	tabExits         bool // Esc pressed: Tab moves focus on
	onChange         func(string)
	onComplete       func(line, col int, prefix string) []CodeCompletion
	onHover          func(line, col int) string
	diagnostics      []CodeDiagnostic

	// Completion list.
	comp     []CodeCompletion
	compSel  int
	compFrom codePos

	// Hover tip.
	hoverAt    f32.Point
	hoverSince time.Time
	hoverDone  bool
	hoverText  string
	hovering   bool

	// Pointer.
	dragging   bool
	lastPress  time.Time
	pressCount int
	pressAt    f32.Point

	// Highlighting.
	hlRev     uint64
	hlStyle   string
	hlRunning bool

	// Input method state last reported to the platform.
	imeLine  int
	imeText  string
	imeRange key.Range
	imeCaret image.Point

	blink     time.Time
	focused   bool
	wantFocus bool
	scroll    gesture.Scroll
	scrollH   gesture.Scroll

	// Layout of the last frame, in px.
	metrics codeMetrics
	xs      map[string][]int // x of every column boundary, by line text
	visible int              // lines that fit
}

func CodeEditor(text string) *CodeEditorView {
	return &CodeEditorView{buf: newCodeBuffer(text), height: 320, xs: map[string][]int{}}
}

// Language picks the syntax highlighting by chroma lexer name, e.g. "go",
// "python", "json". Unknown names show plain text.
func (v *CodeEditorView) Language(name string) *CodeEditorView {
	v.lang, v.hlRev = name, 0
	return v
}

// Name is the editor's accessible name; agents find it by it.
func (v *CodeEditorView) Name(s string) *CodeEditorView { v.name = s; return v }

// Height sets the height in dp, 320 by default; Fill takes the parent's.
func (v *CodeEditorView) Height(dp float32) *CodeEditorView {
	if dp > 0 {
		v.height = dp
	}
	return v
}
func (v *CodeEditorView) Fill() *CodeEditorView { v.fill = true; return v }

func (v *CodeEditorView) OnChange(fn func(text string)) *CodeEditorView { v.onChange = fn; return v }

// OnComplete supplies completions for the word before the caret. It runs
// when the user types a letter or presses Ctrl+Space; returning nothing
// closes the list.
func (v *CodeEditorView) OnComplete(fn func(line, col int, prefix string) []CodeCompletion) *CodeEditorView {
	v.onComplete = fn
	return v
}

// OnHover supplies a tip for the text under the pointer after it rests
// there, e.g. a language server's hover; "" shows none.
func (v *CodeEditorView) OnHover(fn func(line, col int) string) *CodeEditorView {
	v.onHover = fn
	return v
}

// Value returns the whole text. For very large files it builds a large string.
func (v *CodeEditorView) Value() string { return v.buf.text() }

// SetValue replaces the text, clears undo and moves the caret to the start.
// It does not call OnChange.
func (v *CodeEditorView) SetValue(s string) {
	v.buf.set(s)
	v.anchor, v.caret = codePos{}, codePos{}
	v.scrollX, v.scrollY, v.comp, v.hlRev = 0, 0, nil, 0
	clear(v.xs)
}

func (v *CodeEditorView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.comp, v.dragging = nil, false
	}
}
func (v *CodeEditorView) SetReadOnly(on bool) { v.readOnly = on }

// Lines is the number of lines.
func (v *CodeEditorView) Lines() int { return len(v.buf.lines) }

// Cursor returns the caret's line and column (in runes), from 0.
func (v *CodeEditorView) Cursor() (line, col int) { return v.caret.line, v.caret.col }

// SetCursor moves the caret, clears the selection and scrolls to it.
func (v *CodeEditorView) SetCursor(line, col int) {
	v.caret = v.buf.clamp(codePos{line, col})
	v.anchor, v.reveal, v.goalX = v.caret, true, -1
}

// Selection returns the selected text.
func (v *CodeEditorView) Selection() string { return v.buf.slice(v.anchor, v.caret) }

// SetDiagnostics replaces the marked ranges, e.g. after a language server
// publishes diagnostics.
func (v *CodeEditorView) SetDiagnostics(d []CodeDiagnostic) {
	v.diagnostics = append([]CodeDiagnostic(nil), d...)
}

// Focus gives the editor keyboard focus on the next frame.
func (v *CodeEditorView) Focus() { v.wantFocus = true }

func (v *CodeEditorView) Render(cx *el.Context) el.Element {
	id := autoID("code", v)
	if v.disabled || !cx.Enabled(id) {
		v.comp, v.dragging = nil, false
	}
	w := el.Widget(core.Func(v.layout)).ID(id).Disabled(v.disabled)
	if v.fill {
		return w.Grow()
	}
	return w.H(el.Dp(v.height))
}
