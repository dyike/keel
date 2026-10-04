package inputcontent

import (
	"errors"
	"slices"
	"unicode/utf8"
)

var ErrCompositionActive = errors.New("inputcontent: composition is active")

// Session owns editing transactions. UI adapters must pass explicit edit ranges:
// a text diff loses reference removal when an edit leaves the characters intact.
// All ranges are UTF-8 bytes. The platform's rune offsets must be converted before
// calling these methods. UI focus and read-only policies belong to the adapter.
type Session struct {
	history           History
	composing         bool
	compositionBefore Snapshot
}

func (s *Session) Snapshot() Snapshot { return s.history.Current }
func (s *Session) Composing() bool    { return s.composing }

// Set restores a draft silently and discards undo and active composition state.
func (s *Session) Set(content Content) {
	s.history.Set(Snapshot{Content: content})
	s.composing = false
	s.compositionBefore = Snapshot{}
}

// SetText always clears references and history, even when text is unchanged.
func (s *Session) SetText(text string) error {
	c, err := New(text)
	if err != nil {
		return err
	}
	s.Set(c)
	return nil
}

func (s *Session) SelectSource(r Range) error {
	c := s.history.Current.Content
	if !byteOffset(c.Text(), r.Start) || !byteOffset(c.Text(), r.End) {
		return ErrBoundary
	}
	a, b := c.Selection(r.Start, r.End)
	s.history.Current.Selection = Range{a, b}
	return nil
}

func (s *Session) SelectDisplay(r Range) error {
	p := s.history.Current.Content.Presentation()
	mapped, err := p.SourceRange(r)
	if err != nil {
		return err
	}
	return s.SelectSource(mapped)
}

func (s *Session) SelectedText() string {
	snapshot := s.Snapshot()
	a, b := snapshot.Content.Selection(snapshot.Selection.Start, snapshot.Selection.End)
	return snapshot.Content.Text()[min(a, b):max(a, b)]
}

// ReplaceDisplay handles typing, plain paste and deletion. Selecting a label
// copies/removes the whole reference, not its display name. No text matching is
// used to infer references. Replacing a label with itself still removes its ID.
func (s *Session) ReplaceDisplay(r Range, text string) error {
	source, err := s.history.Current.Content.Presentation().SourceRange(r)
	if err != nil {
		return err
	}
	return s.replace(source, text, nil)
}

// ReplaceToken replaces the current source selection with an atomic reference.
// Applications must not insert references in a platform's active composition.
func (s *Session) ReplaceToken(token Token) error {
	if s.composing {
		return ErrCompositionActive
	}
	return s.replace(s.history.Current.Selection, token.Text, &token)
}

func (s *Session) replace(r Range, text string, token *Token) error {
	if r.Start > r.End {
		r.Start, r.End = r.End, r.Start
	}
	next, caret, err := s.history.Current.Content.Replace(r, text, token)
	if err != nil {
		return err
	}
	snapshot := Snapshot{next, caret}
	if s.composing {
		s.history.Current = snapshot
	} else {
		s.history.Commit(snapshot)
	}
	return nil
}

// BeginComposition groups all subsequent plain edits into one undo transaction.
// Repeated begin notifications from a platform do not reset the saved snapshot.
func (s *Session) BeginComposition() {
	if s.composing {
		return
	}
	s.composing = true
	s.compositionBefore = s.history.Current
}

func (s *Session) EndComposition() {
	if !s.composing {
		return
	}
	before, after := s.compositionBefore, s.history.Current
	s.composing = false
	s.compositionBefore = Snapshot{}
	if before.Content.Text() == after.Content.Text() && slices.Equal(before.Content.spans, after.Content.spans) {
		return
	}
	s.history.Current = before
	s.history.Commit(after)
}

// CancelComposition restores both the reference metadata and original selection.
func (s *Session) CancelComposition() {
	if !s.composing {
		return
	}
	s.history.Current = s.compositionBefore
	s.compositionBefore = Snapshot{}
	s.composing = false
}

func (s *Session) Undo() bool {
	if s.composing {
		return false
	}
	return s.history.Undo()
}
func (s *Session) Redo() bool {
	if s.composing {
		return false
	}
	return s.history.Redo()
}

func byteOffset(text string, at int) bool {
	return at >= 0 && at <= len(text) && (at == len(text) || utf8.RuneStart(text[at]))
}

// SourceRange expands a nonempty display selection to whole references while
// preserving direction. A caret inside a display label snaps to its nearest edge.
func (p Presentation) SourceRange(r Range) (Range, error) {
	if !byteOffset(p.Text, r.Start) || !byteOffset(p.Text, r.End) {
		return Range{}, ErrBoundary
	}
	if r.Start == r.End {
		at := p.SourceOffset(r.Start, 0)
		return Range{at, at}, nil
	}
	if r.Start < r.End {
		return Range{p.SourceOffset(r.Start, -1), p.SourceOffset(r.End, 1)}, nil
	}
	return Range{p.SourceOffset(r.Start, 1), p.SourceOffset(r.End, -1)}, nil
}

// RuneRange and ByteRange bridge platform/editor rune coordinates and the
// content API's byte coordinates without silently accepting out-of-range values.
func RuneRange(text string, r Range) (Range, error) {
	if !byteOffset(text, r.Start) || !byteOffset(text, r.End) {
		return Range{}, ErrBoundary
	}
	return Range{utf8.RuneCountInString(text[:r.Start]), utf8.RuneCountInString(text[:r.End])}, nil
}
func ByteRange(text string, r Range) (Range, error) {
	if r.Start < 0 || r.End < 0 {
		return Range{}, ErrBoundary
	}
	a, b := -1, -1
	index := 0
	for at := range text {
		if index == r.Start {
			a = at
		}
		if index == r.End {
			b = at
		}
		index++
	}
	if index == r.Start {
		a = len(text)
	}
	if index == r.End {
		b = len(text)
	}
	if a < 0 || b < 0 {
		return Range{}, ErrBoundary
	}
	return Range{a, b}, nil
}
