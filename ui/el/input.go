package el

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// InputEl is a text box. Its editing state (content, caret, selection) is
// kept per element, so give it an ID when siblings may change.
type InputEl struct{ Styled[InputEl] }

type inputSpec struct {
	tokenRenderer    InputTokenRenderer
	onTokenActivate  func(InputToken)
	document         *InputDocument
	onPaste          func(core.ClipboardData) bool
	pasteReader      core.ClipboardReader
	onPasteError     func(error)
	placeholder      string
	captureKeys      []string
	selectOnFocus    bool
	bind             *string
	multiline        bool
	password         bool
	onChange         func(string)
	transform        func(InputEdit) InputEdit
	transformEdit    func(InputEdit, InputEdit) InputEdit
	onSubmit         func(string)
	maxLen           int
	filter           string
	readOnly         bool
	minRows, maxRows int
	line             int // measured line height, to center single-line text
}

// Input creates a single-line text box. Enter triggers OnSubmit.
func Input() *InputEl {
	e := &InputEl{}
	e.n, e.self = newNode(), e
	e.n.input = &inputSpec{}
	// A bordered box by default, the same as kit's fields: theme.ControlHeight
	// tall with the line centered. Every part can be restyled; the border
	// turns Primary while the box has focus.
	e.Border(1, theme.Border).Rounded(6).Bg(theme.Surface).Px(10).MinH(Dp(float32(theme.ControlHeight)))
	return e
}

// TextArea creates a multi-line text box; Enter inserts a newline.
func TextArea() *InputEl {
	e := Input()
	e.n.input.multiline = true
	e.Py(8).MinH(Auto)
	return e
}

// Placeholder is shown while the box is empty; agents also see it as the
// name when Name is not set.
func (e *InputEl) Placeholder(s string) *InputEl { e.n.input.placeholder = s; return e }

// Bind keeps *p and the box in sync: typing writes *p, and a program change
// to *p shows in the box on the next frame.
func (e *InputEl) Bind(p *string) *InputEl { e.n.input.bind = p; return e }

// Password masks the content.
func (e *InputEl) Password() *InputEl { e.n.input.password = true; return e }

// OnChange runs after every edit by the user, with the new content.
func (e *InputEl) OnChange(fn func(string)) *InputEl { e.n.input.onChange = fn; return e }

// MaxLen limits the content to n runes; 0 means no limit.
func (e *InputEl) MaxLen(n int) *InputEl { e.n.input.maxLen = n; return e }

// Filter accepts only the runes in chars as typed or pasted input; "" accepts all.
func (e *InputEl) Filter(chars string) *InputEl { e.n.input.filter = chars; return e }

// ReadOnly lets the user select and copy but not edit.
func (e *InputEl) ReadOnly(on bool) *InputEl { e.n.input.readOnly = on; return e }

// OnSubmit runs when Enter is pressed in a single-line box.
func (e *InputEl) OnSubmit(fn func(string)) *InputEl { e.n.input.onSubmit = fn; return e }

// AutoGrow sizes a multiline input to its wrapped text, between minRows and
// maxRows lines. Overflow scrolls inside the editor. Invalid ranges are ignored.
// Passing (0, 0) restores the default height. Single-line inputs ignore this.
func (e *InputEl) AutoGrow(minRows, maxRows int) *InputEl {
	if minRows == 0 && maxRows == 0 || minRows > 0 && maxRows >= minRows {
		e.n.input.minRows, e.n.input.maxRows = minRows, maxRows
	}
	return e
}

// InputEdit describes text and rune-based selection endpoints after an edit.
type InputEdit struct {
	Text       string
	Start, End int
}

// Transform normalizes user edits before Bind and OnChange. Return mapped rune
// selection endpoints with the new text. Programmatic Bind changes are unchanged.
// Transformed inputs retain up to 100 user edits for undo/redo; programmatic changes reset this history.
func (e *InputEl) Transform(fn func(InputEdit) InputEdit) *InputEl {
	e.n.input.transform = fn
	e.n.input.transformEdit = nil
	return e
}

// CaptureKeys reserves named, unmodified keys for OnKey before the editor
// handles them. The handler owns these keys even when it returns false.
// Ordinary editing shortcuts with modifiers remain with the editor, so a
// TextArea that captures "⏎" to send still breaks lines with Shift+Enter.
func (e *InputEl) CaptureKeys(names ...string) *InputEl {
	e.n.input.captureKeys = append([]string(nil), names...)
	return e
}

// SelectOnFocus selects the complete value when the input gains focus.
func (e *InputEl) SelectOnFocus(on bool) *InputEl { e.n.input.selectOnFocus = on; return e }

// SelectInput schedules rune-based selection in an input after its next Bind
// synchronization. It does not change text or focus. Missing or disabled inputs
// are ignored, and the editor clamps the endpoints to its content.
func (cx *Context) SelectInput(id string, start, end int) {
	if !cx.root.e.gtx.Enabled() {
		return
	}
	for _, st := range cx.root.store.states {
		if st.id == id && st.edInit && !st.disabled {
			st.inputSelection = &[2]int{start, end}
			return
		}
	}
}

// TransformEdit normalizes an edit with access to the previous text and rune
// selection. It replaces Transform; undo/redo restores normalized snapshots.
func (e *InputEl) TransformEdit(fn func(before, after InputEdit) InputEdit) *InputEl {
	e.n.input.transformEdit, e.n.input.transform = fn, nil
	return e
}
