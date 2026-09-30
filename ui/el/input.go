package el

import "github.com/dyike/keel/ui/theme"

// InputEl is a text box. Its editing state (content, caret, selection) is
// kept per element, so give it an ID when siblings may change.
type InputEl struct{ Styled[InputEl] }

type inputSpec struct {
	placeholder string
	bind        *string
	multiline   bool
	password    bool
	onChange    func(string)
	onSubmit    func(string)
}

// Input creates a single-line text box. Enter triggers OnSubmit.
func Input() *InputEl {
	e := &InputEl{}
	e.n, e.self = newNode(), e
	e.n.input = &inputSpec{}
	// A bordered box by default; every part can be restyled. The border turns
	// Primary while the box has focus.
	e.Border(1, theme.Border).Rounded(6).Bg(theme.Surface).Px(10).Py(8)
	return e
}

// TextArea creates a multi-line text box; Enter inserts a newline.
func TextArea() *InputEl {
	e := Input()
	e.n.input.multiline = true
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

// OnSubmit runs when Enter is pressed in a single-line box.
func (e *InputEl) OnSubmit(fn func(string)) *InputEl { e.n.input.onSubmit = fn; return e }
