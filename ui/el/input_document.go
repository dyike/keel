package el

import "github.com/dyike/keel/ui/internal/inputcontent"

// InputToken identifies an atomic reference. Text is the submitted/copied value;
// Label, when nonempty, is its visible name.
type InputToken = inputcontent.Token
type InputTokenSpan = inputcontent.Span
type InputRange = inputcontent.Range
type InputContent = inputcontent.Content

// NewInputContent validates UTF-8 byte ranges and returns an immutable draft.
func NewInputContent(text string, tokens ...InputTokenSpan) (InputContent, error) {
	return inputcontent.New(text, tokens...)
}

// InputDocument retains reference metadata, selection and undo history for one
// input. Do not attach the same document to multiple live editors.
type InputDocument struct {
	session  inputcontent.Session
	revision uint64
}

func (d *InputDocument) Content() InputContent     { return d.session.Snapshot().Content }
func (d *InputDocument) SetContent(c InputContent) { d.session.Set(c); d.revision++ }
func (d *InputDocument) SetText(text string) error {
	if err := d.session.SetText(text); err != nil {
		return err
	}
	d.revision++
	return nil
}
func (d *InputDocument) ReplaceWithToken(token InputToken) error {
	if err := d.session.ReplaceToken(token); err != nil {
		return err
	}
	d.revision++
	return nil
}
func (d *InputDocument) Selection() InputRange { return d.session.Snapshot().Selection }
func (d *InputDocument) Select(r InputRange) error {
	if err := d.session.SelectSource(r); err != nil {
		return err
	}
	d.revision++
	return nil
}

// Document binds an atomic-reference draft. It owns text and undo; Bind,
// Transform, password masks, Filter and MaxLen are not applied in this mode.
func (e *InputEl) Document(d *InputDocument) *InputEl { e.n.input.document = d; return e }

// OnTokenActivate runs for an unmodified token click, including in read-only
// fields. Dragging a selection and disabled fields do not activate references.
func (e *InputEl) OnTokenActivate(fn func(InputToken)) *InputEl {
	e.n.input.onTokenActivate = fn
	return e
}

// SelectedToken returns a completely selected reference.
func (d *InputDocument) SelectedToken() (InputToken, bool) {
	r := d.Selection()
	for _, s := range d.Content().Tokens() {
		if min(r.Start, r.End) == s.Range.Start && max(r.Start, r.End) == s.Range.End {
			return s.Token, true
		}
	}
	return InputToken{}, false
}
