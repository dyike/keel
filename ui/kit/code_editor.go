package kit

import (
	"image"
	"time"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/gesture"
	"github.com/dyike/keel/third_party/gio/io/key"

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
// several carets and column selection, find and replace, folding, bracket
// pairs, undo, clipboard and input methods. Only the visible lines are laid
// out, and lines are stored in chunks, so files of hundreds of thousands of
// lines stay responsive; highlighting is redone near an edit at once and for
// the whole file in the background.
//
// It does not speak a language server protocol itself. The app connects one
// through SetDiagnostics, OnComplete, OnHover and OnDefinition.
//
// Tab indents; press Esc first to move focus on with Tab.
type CodeEditorView struct {
	onPaste       func(core.ClipboardData) bool
	pasteReader   core.ClipboardReader
	onPasteError  func(error)
	pendingPaste  *codePasteRequest
	buf           *codeBuffer
	lang, name    string
	height        float32
	fill          bool
	readOnly      bool
	disabled      bool
	autoClose     bool
	noSmartIndent bool
	editingRules  *codeLanguageRules
	syntaxContext func(int, int) CodeSyntaxContext
	ruleRev       uint64
	ruleLine      int
	ruleLang      string
	whitespace    bool
	tabSize       int
	hardTabs      int // 0 detect, 1 tabs, 2 spaces
	sels          []codeSel
	prim          int // the primary selection, the one the caret APIs report
	goalX         int // px from the row start, kept across vertical moves of the primary caret
	wrap          bool
	vrows         []codeVRow // visual rows with soft wrap; nil without
	vrowsKey      codeWrapKey
	rowsGen       uint64 // counts fold row rebuilds
	scrollX       float32
	scrollY       float32 // px from the first display row
	reveal        bool    // scroll the primary caret into view on the next frame
	tabExits      bool    // Esc pressed: Tab moves focus on
	onChange      func(string)
	onComplete    func(line, col int, prefix string) []CodeCompletion
	onHover       func(line, col int) string
	onDefinition  func(line, col int)
	diagnostics   []CodeDiagnostic
	decorations   []*CodeDecorationCollection

	// Completion list.
	comp     []CodeCompletion
	compSel  int
	compFrom codePos

	// Hover tip and go-to-definition underline.
	hoverAt    f32.Point
	hoverSince time.Time
	hoverDone  bool
	hoverText  string
	hovering   bool
	linkFrom   codePos // the word under the pointer while Cmd/Ctrl is held
	linkTo     codePos

	// Pointer.
	dragging   bool
	column     bool // an Alt drag selects a block
	columnFrom f32.Point
	lastPress  time.Time
	pressCount int
	pressAt    f32.Point

	// Folding.
	folds    map[int]bool // folded regions by their first line
	foldEnds map[int]int  // cached region end by first line, -1 for none
	foldRev  uint64
	rows     []int // display row to line, nil when nothing is folded
	rowsRev  uint64
	rowsKey  int

	// Find and replace.
	search searchState

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
	visible int              // rows that fit
}

func CodeEditor(text string) *CodeEditorView {
	v := &CodeEditorView{buf: newCodeBuffer(text), height: 320, xs: map[string][]int{}, autoClose: true, tabSize: 4,
		sels: []codeSel{{}}, folds: map[int]bool{}, foldEnds: map[int]int{}, imeLine: -1}
	v.buf.onSplice = v.spliced
	v.search.searchable = true
	v.search.current = -1
	return v
}

// Language picks the syntax highlighting and bracket rules by chroma lexer
// name, e.g. "go", "python", "json". Unknown names show plain text.
func (v *CodeEditorView) Language(name string) *CodeEditorView {
	if v.lang == name {
		return v
	}
	v.lang, v.hlRev = name, 0
	v.buf.lines.each(0, func(_ int, line *codeLine) bool {
		line.spans = nil
		return true
	})
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

// TabSize sets the width of a tab stop in spaces, and whether Tab inserts a
// tab character (hard) or that many spaces. By default the width is 4 and
// Tab follows the file: tabs if its lines are indented with tabs.
func (v *CodeEditorView) TabSize(size int, hard bool) *CodeEditorView {
	v.tabSize = max(1, size)
	v.hardTabs = 2
	if hard {
		v.hardTabs = 1
	}
	clear(v.xs)
	return v
}

// AutoClose types the closing bracket or quote with the opening one, steps
// over a closing one typed next to it, and deletes an empty pair together.
// On by default; brackets are not paired inside strings and comments.
func (v *CodeEditorView) AutoClose(on bool) *CodeEditorView { v.autoClose = on; return v }

// ShowWhitespace marks spaces with dots and tabs with arrows.
func (v *CodeEditorView) ShowWhitespace(on bool) *CodeEditorView { v.whitespace = on; return v }

// Searchable turns the find panel (Cmd/Ctrl+F) on or off; on by default.
func (v *CodeEditorView) Searchable(on bool) *CodeEditorView {
	v.search.searchable = on
	if !on && v.search.open {
		v.CloseSearch()
	}
	return v
}

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

// OnDefinition runs on Cmd/Ctrl+click or F12 with the position of the word
// to look up, e.g. to ask a language server and SetCursor to the result.
// While Cmd/Ctrl is held the word under the pointer is underlined.
func (v *CodeEditorView) OnDefinition(fn func(line, col int)) *CodeEditorView {
	v.onDefinition = fn
	return v
}

// Value returns the whole text. For very large files it builds a large string.
func (v *CodeEditorView) Value() string { return v.buf.text() }

// SetValue replaces the text, clears undo, folds and extra carets, and moves
// the caret to the start. It does not call OnChange.
func (v *CodeEditorView) SetValue(s string) {
	v.buf.set(s)
	v.sels, v.prim = []codeSel{{}}, 0
	v.scrollX, v.scrollY, v.comp, v.hlRev = 0, 0, nil, 0
	v.linkFrom, v.linkTo, v.hoverText = codePos{}, codePos{}, ""
	clear(v.xs)
}

func (v *CodeEditorView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.comp, v.dragging = nil, false
		v.pendingPaste = nil
	}
}
func (v *CodeEditorView) SetReadOnly(on bool) {
	v.readOnly = on
	if on {
		v.pendingPaste = nil
	}
}

// Lines is the number of lines.
func (v *CodeEditorView) Lines() int { return v.buf.count() }

func (v *CodeEditorView) primary() codeSel { return v.sels[v.prim] }

// Cursor returns the primary caret's line and column (in runes), from 0.
func (v *CodeEditorView) Cursor() (line, col int) { c := v.primary().caret; return c.line, c.col }

// SetCursor keeps one caret, moves it there, unfolds around it and scrolls
// to it.
func (v *CodeEditorView) SetCursor(line, col int) {
	p := v.buf.clamp(codePos{line, col})
	v.sels, v.prim = []codeSel{{p, p}}, 0
	v.reveal, v.goalX = true, -1
	v.keepCaretsVisible()
}

// Cursors is how many carets there are; Esc keeps only the primary one.
func (v *CodeEditorView) Cursors() int { return len(v.sels) }

// Selection returns the primary selection's text.
func (v *CodeEditorView) Selection() string {
	s := v.primary()
	return v.buf.slice(s.anchor, s.caret)
}

// SetDiagnostics replaces the marked ranges, e.g. after a language server
// publishes diagnostics.
func (v *CodeEditorView) SetDiagnostics(d []CodeDiagnostic) {
	v.diagnostics = append([]CodeDiagnostic(nil), d...)
}

// Fold folds the region starting at a line (its more-indented lines below);
// Unfold opens it. FoldAll and UnfoldAll act on every region.
func (v *CodeEditorView) Fold(line int) {
	if v.foldEnd(line) > line {
		v.folds[line] = true
		v.keepCaretsVisible()
	}
}
func (v *CodeEditorView) Unfold(line int) { delete(v.folds, line) }
func (v *CodeEditorView) FoldAll() {
	for i := range v.buf.count() {
		if v.foldEnd(i) > i {
			v.folds[i] = true
		}
	}
	v.keepCaretsVisible()
}
func (v *CodeEditorView) UnfoldAll() { clear(v.folds) }

// Folded reports whether the region starting at a line is folded.
func (v *CodeEditorView) Folded(line int) bool { return v.folds[line] }

// Focus gives the editor keyboard focus on the next frame.
func (v *CodeEditorView) Focus() { v.wantFocus = true }

func (v *CodeEditorView) Render(cx *el.Context) el.Element {
	id := autoID("code", v)
	if v.disabled || !cx.Enabled(id) {
		v.comp, v.dragging = nil, false
		v.pendingPaste = nil
	}
	w := el.Widget(core.Func(v.layout)).ID(id).Disabled(v.disabled).Grow()
	box := el.Div().Items(el.Stretch).Child(w)
	if v.search.open {
		box.Child(v.searchPanel(cx))
	}
	if v.fill {
		return box.Grow()
	}
	return box.H(el.Dp(v.height))
}
