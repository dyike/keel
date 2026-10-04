package inputcontent

import (
	"errors"
	"reflect"
	"testing"
)

func tokenSession(t *testing.T) *Session {
	t.Helper()
	var s Session
	s.Set(mustContent(t, "a/ref/z", Span{Range{1, 6}, Token{"resource", "/ref/", "引用"}}))
	return &s
}
func TestSessionDisplayEditing(t *testing.T) {
	s := tokenSession(t)
	// Reverse-select only the last label character. Copy returns the entire path.
	if err := s.SelectDisplay(Range{7, 4}); err != nil {
		t.Fatal(err)
	}
	if s.SelectedText() != "/ref/" || s.Snapshot().Selection != (Range{6, 1}) {
		t.Fatal(s.Snapshot(), s.SelectedText())
	}
	before := s.Snapshot()
	if err := s.ReplaceDisplay(Range{7, 4}, "引用"); err != nil {
		t.Fatal(err)
	}
	if s.Snapshot().Content.Text() != "a引用z" || len(s.Snapshot().Content.Tokens()) != 0 {
		t.Fatal(s.Snapshot())
	}
	if !s.Undo() || !reflect.DeepEqual(s.Snapshot(), before) {
		t.Fatal("undo must restore metadata and reversed selection")
	}
	if !s.Redo() || s.SelectedText() != "" {
		t.Fatal("redo caret")
	}
}
func TestSessionSameTextReplacement(t *testing.T) {
	var s Session
	s.Set(mustContent(t, "ref", Span{Range{0, 3}, Token{"id", "ref", ""}}))
	if err := s.ReplaceDisplay(Range{0, 3}, "ref"); err != nil {
		t.Fatal(err)
	}
	if s.Snapshot().Content.Text() != "ref" || len(s.Snapshot().Content.Tokens()) != 0 {
		t.Fatal("same text paste retained reference")
	}
	if !s.Undo() || len(s.Snapshot().Content.Tokens()) != 1 {
		t.Fatal("undo lost reference")
	}
	if err := s.SetText("ref"); err != nil {
		t.Fatal(err)
	}
	if len(s.Snapshot().Content.Tokens()) != 0 || s.Undo() || s.Redo() {
		t.Fatal("SetText retained reference or history")
	}
}
func TestSessionComposition(t *testing.T) {
	s := tokenSession(t)
	s.SelectDisplay(Range{1, 7})
	before := s.Snapshot()
	s.BeginComposition()
	if err := s.ReplaceDisplay(Range{1, 7}, "n"); err != nil {
		t.Fatal(err)
	}
	s.BeginComposition() // native IMEs may update the composing range repeatedly
	if err := s.ReplaceDisplay(Range{1, 2}, "ni"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceToken(Token{"new", "path", "名称"}); !errors.Is(err, ErrCompositionActive) {
		t.Fatal(err)
	}
	if s.Undo() || s.Redo() {
		t.Fatal("history changed during composition")
	}
	if err := s.ReplaceDisplay(Range{1, 3}, "你"); err != nil {
		t.Fatal(err)
	}
	after := s.Snapshot()
	s.EndComposition()
	if s.Composing() || s.Snapshot().Content.Text() != "a你z" {
		t.Fatal(s.Snapshot())
	}
	if !s.Undo() || !reflect.DeepEqual(s.Snapshot(), before) || s.Undo() {
		t.Fatal("composition not one transaction")
	}
	if !s.Redo() || !reflect.DeepEqual(s.Snapshot(), after) {
		t.Fatal("composition redo")
	}
}
func TestSessionCompositionCancelAndReset(t *testing.T) {
	s := tokenSession(t)
	s.SelectDisplay(Range{1, 7})
	before := s.Snapshot()
	s.BeginComposition()
	s.ReplaceDisplay(Range{1, 7}, "draft")
	s.CancelComposition()
	if !reflect.DeepEqual(s.Snapshot(), before) || s.Undo() {
		t.Fatal("cancel failed")
	}
	// Cancelling or committing an empty composition leaves the redo branch intact.
	s.ReplaceDisplay(Range{1, 7}, "changed")
	s.Undo()
	s.BeginComposition()
	s.EndComposition()
	if !s.Redo() {
		t.Fatal("empty composition removed redo")
	}
	s.BeginComposition()
	s.ReplaceDisplay(Range{0, 1}, "x")
	if err := s.SetText("new draft"); err != nil {
		t.Fatal(err)
	}
	s.CancelComposition()
	if s.Composing() || s.Snapshot().Content.Text() != "new draft" || s.Undo() {
		t.Fatal("reset restored old composition")
	}
}
func TestSessionTokenInsertionAndFailures(t *testing.T) {
	s := tokenSession(t)
	s.SelectDisplay(Range{7, 7})
	before := s.Snapshot()
	tok := Token{"new", "路径", "新引用"}
	if err := s.ReplaceToken(tok); err != nil {
		t.Fatal(err)
	}
	if got := s.Snapshot(); got.Content.Text() != "a/ref/路径z" || len(got.Content.Tokens()) != 2 || got.Selection != (Range{12, 12}) {
		t.Fatal(got)
	}
	if !s.Undo() || !reflect.DeepEqual(s.Snapshot(), before) {
		t.Fatal("insert undo")
	}
	for _, r := range []Range{{-1, 0}, {0, 99}, {2, 4}} {
		if err := s.ReplaceDisplay(r, "bad"); !errors.Is(err, ErrBoundary) {
			t.Fatal(r, err)
		}
		if !reflect.DeepEqual(s.Snapshot(), before) {
			t.Fatal("failed edit changed state")
		}
	}
	if !s.Redo() {
		t.Fatal("failed edit cleared redo")
	}
}
func TestCoordinateBridge(t *testing.T) {
	text := "a👩‍💻e\u0301中"
	for a := 0; a <= 7; a++ {
		for b := 0; b <= 7; b++ {
			r := Range{a, b}
			bytes, err := ByteRange(text, r)
			if err != nil {
				t.Fatal(r, err)
			}
			runes, err := RuneRange(text, bytes)
			if err != nil || runes != r {
				t.Fatal(r, bytes, runes, err)
			}
		}
	}
	for _, r := range []Range{{-1, 0}, {0, 8}} {
		if _, err := ByteRange(text, r); !errors.Is(err, ErrBoundary) {
			t.Fatal(r, err)
		}
	}
	for _, r := range []Range{{2, 3}, {-1, 0}, {0, len(text) + 1}} {
		if _, err := RuneRange(text, r); !errors.Is(err, ErrBoundary) {
			t.Fatal(r, err)
		}
	}
}

func TestCompositionSameTextDropsReference(t *testing.T) {
	var s Session
	s.Set(mustContent(t, "name", Span{Range{0, 4}, Token{"id", "name", ""}}))
	s.SelectSource(Range{0, 4})
	before := s.Snapshot()
	s.BeginComposition()
	if err := s.ReplaceDisplay(Range{0, 4}, "name"); err != nil {
		t.Fatal(err)
	}
	s.EndComposition()
	if len(s.Snapshot().Content.Tokens()) != 0 {
		t.Fatal("composition kept replaced reference")
	}
	if !s.Undo() || !reflect.DeepEqual(s.Snapshot(), before) {
		t.Fatal("same-text composition lost history")
	}
}

func TestSessionSnapshotRestoresIndependentDraft(t *testing.T) {
	original := tokenSession(t)
	original.SelectSource(Range{1, 6})
	draft := original.Snapshot().Content
	var restored Session
	restored.Set(draft)
	if err := original.ReplaceToken(Token{"other", "/other/", "不同"}); err != nil {
		t.Fatal(err)
	}
	if restored.Snapshot().Content.Text() != "a/ref/z" || restored.Snapshot().Content.Tokens()[0].Token.ID != "resource" {
		t.Fatal("draft changed with original session")
	}
	if restored.Undo() {
		t.Fatal("draft inherited unrelated undo history")
	}
}
