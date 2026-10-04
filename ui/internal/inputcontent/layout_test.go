package inputcontent

import "testing"

func TestLayoutPresentationMapsObjectsToOriginalContent(t *testing.T) {
	c, err := New("中ref🙂 tail", Span{Range: Range{3, 6}, Token: Token{ID: "r", Text: "ref", Label: "长引用标签"}})
	if err != nil {
		t.Fatal(err)
	}
	labels := []string{"\U000f0000"}
	p, err := LayoutPresentation(c, labels)
	if err != nil {
		t.Fatal(err)
	}
	labels[0] = "changed"
	if p.Text != "中\U000f0000🙂 tail" || c.Text() != "中ref🙂 tail" || p.Spans[0].Token.Label != "长引用标签" {
		t.Fatal("projection mutated content", p, c)
	}
	for _, source := range []int{0, 3, 6, 10, 15} {
		display := p.DisplayOffset(source, 0)
		if p.SourceOffset(display, 0) != source {
			t.Fatal("edge did not round trip", source, display)
		}
	}
	displayRange, err := ByteRange(p.Text, Range{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	sourceRange, err := p.SourceRange(displayRange)
	if err != nil || sourceRange != (Range{3, 6}) {
		t.Fatal("object mapping", sourceRange, err)
	}
	var session Session
	session.Set(c)
	if err = session.SelectSource(sourceRange); err != nil {
		t.Fatal(err)
	}
	if session.SelectedText() != "ref" {
		t.Fatal("clipboard exposed placeholder")
	}
	if err = session.ReplaceSource(sourceRange, ""); err != nil {
		t.Fatal(err)
	}
	if session.Snapshot().Content.Text() != "中🙂 tail" || len(session.Snapshot().Content.Tokens()) != 0 {
		t.Fatal("object deletion")
	}
	session.Undo()
	if !Equal(session.Snapshot().Content, c) {
		t.Fatal("undo lost token")
	}
	for _, bad := range [][]string{nil, {""}, {"\n"}, {"\xff"}, {"a", "b"}} {
		if _, err := LayoutPresentation(c, bad); err == nil {
			t.Fatal("accepted invalid replacement", bad)
		}
	}
}

func TestLayoutPresentationReplacementRespectsComposition(t *testing.T) {
	c, _ := New("ref", Span{Range: Range{0, 3}, Token: Token{ID: "r", Text: "ref"}})
	var s Session
	s.Set(c)
	s.BeginComposition()
	if err := s.ReplaceSource(Range{0, 3}, "n"); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceSource(Range{0, 1}, "你"); err != nil {
		t.Fatal(err)
	}
	s.EndComposition()
	s.Undo()
	if !Equal(s.Snapshot().Content, c) {
		t.Fatal("composition lost original reference")
	}
}
