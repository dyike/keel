package inputcontent

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func mustContent(t *testing.T, text string, spans ...Span) Content {
	t.Helper()
	c, err := New(text, spans...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestValidation(t *testing.T) {
	for _, text := range []string{"e\u0301", "👩‍💻", "🇨🇳"} {
		t.Run(text, func(t *testing.T) {
			mustContent(t, text, Span{Range{0, len(text)}, Token{"id", text, "显示"}})
			// Every interior rune boundary still lies within this one grapheme.
			for at := range text {
				if at == 0 {
					continue
				}
				_, err := New(text, Span{Range{0, at}, Token{"id", text[:at], ""}})
				if !errors.Is(err, ErrBoundary) {
					t.Fatalf("at %d: %v", at, err)
				}
			}
		})
	}
	for _, tok := range []Token{{"", "x", ""}, {" ", "x", ""}, {"id", "", ""}, {"id", "x\n", ""}, {"id", "x", "a\t"}, {"id", "x", "a\u2028"}, {"id", "\xff", ""}} {
		if tok.Validate() == nil {
			t.Fatalf("accepted %#v", tok)
		}
	}
	if _, err := New("abc", Span{Range{0, 2}, Token{"a", "ab", ""}}, Span{Range{1, 3}, Token{"b", "bc", ""}}); !errors.Is(err, ErrOverlap) {
		t.Fatal(err)
	}
	if _, err := New("abc", Span{Range{0, 2}, Token{"a", "xx", ""}}); !errors.Is(err, ErrText) {
		t.Fatal(err)
	}
	for _, r := range []Range{{-1, 2}, {2, 1}, {0, 4}, {2, 2}} {
		if _, err := New("abc", Span{r, Token{"a", "a", ""}}); !errors.Is(err, ErrBoundary) {
			t.Fatal(r, err)
		}
	}
}
func TestAtomicEdits(t *testing.T) {
	spans := []Span{{Range{1, 5}, Token{"id", "name", "人"}}, {Range{6, 10}, Token{"id", "name", "人"}}}
	c := mustContent(t, " name name!", spans...)
	spans[0].Token.ID = "mutated"
	copy := c.Tokens()
	copy[0].Token.ID = "also mutated"
	if c.Tokens()[0].Token.ID != "id" {
		t.Fatal("mutable input or accessor")
	}
	for _, tc := range []struct{ a, b, x, y int }{{2, 3, 1, 5}, {3, 2, 5, 1}, {-9, -2, 0, 0}, {90, 80, 11, 11}, {4, 8, 1, 10}} {
		a, b := c.Selection(tc.a, tc.b)
		if a != tc.x || b != tc.y {
			t.Fatalf("%+v: %d %d", tc, a, b)
		}
	}
	for _, r := range []Range{{4, 5}, {1, 2}, {2, 4}} {
		next, caret, err := c.Replace(r, "", nil)
		if err != nil || next.Text() != "  name!" || caret != (Range{1, 1}) {
			t.Fatal(next, caret, err)
		}
		if got := next.Tokens(); len(got) != 1 || got[0].Range != (Range{2, 6}) {
			t.Fatal(got)
		}
	}
	if _, _, err := c.Replace(Range{3, 3}, "x", nil); !errors.Is(err, ErrBoundary) {
		t.Fatal(err)
	}
	plain, _, err := c.Replace(Range{1, 5}, "name", nil)
	if err != nil || len(plain.Tokens()) != 1 {
		t.Fatal(plain, err)
	}
	tok := Token{"new", "链接", "Link"}
	next, caret, err := c.Replace(Range{1, 5}, tok.Text, &tok)
	if err != nil || caret.Start != 7 || next.Tokens()[1].Range.Start != 8 {
		t.Fatal(next, caret, err)
	}
	// Rejected edits preserve both content and the caller's original selection.
	before := mustContent(t, "aZ", Span{Range{0, 1}, Token{"a", "a", ""}}, Span{Range{1, 2}, Token{"z", "Z", ""}})
	next, caret, err = before.Replace(Range{1, 2}, "\u0301", nil)
	if !errors.Is(err, ErrBoundary) || !reflect.DeepEqual(next, before) || caret != (Range{1, 2}) {
		t.Fatal(next, caret, err)
	}
	if c.Text() != " name name!" || len(c.Tokens()) != 2 {
		t.Fatal("original changed")
	}
}
func TestPresentationMapping(t *testing.T) {
	c := mustContent(t, "aLONG b短!", Span{Range{1, 5}, Token{"1", "LONG", "人"}}, Span{Range{7, 10}, Token{"2", "短", "label"}})
	p := c.Presentation()
	if p.Text != "a人 blabel!" {
		t.Fatal(p.Text)
	}
	for _, pair := range [][2]int{{0, 0}, {1, 1}, {5, 4}, {6, 5}, {7, 6}, {10, 11}, {11, 12}} {
		if got := p.DisplayOffset(pair[0], 0); got != pair[1] {
			t.Fatal(pair, got)
		}
		if got := p.SourceOffset(pair[1], 0); got != pair[0] {
			t.Fatal(pair, got)
		}
	}
	if p.SourceOffset(2, -1) != 1 || p.SourceOffset(2, 1) != 5 || p.DisplayOffset(2, -1) != 1 || p.DisplayOffset(2, 1) != 4 {
		t.Fatal("interior mapping")
	}
	if p.SourceOffset(-99, 0) != 0 || p.SourceOffset(99, 0) != len(c.Text()) || p.DisplayOffset(99, 0) != len(p.Text) {
		t.Fatal("clamping")
	}
}
func TestHistoryReferencesAndBranching(t *testing.T) {
	a := Snapshot{mustContent(t, "x", Span{Range{0, 1}, Token{"a", "x", "A"}}), Range{0, 1}}
	b := Snapshot{mustContent(t, "x", Span{Range{0, 1}, Token{"b", "x", "B"}}), Range{1, 1}}
	var h History
	h.Set(a)
	h.Commit(b)
	if !h.Undo() || !reflect.DeepEqual(h.Current, a) || !h.Redo() || !reflect.DeepEqual(h.Current, b) {
		t.Fatal("metadata history")
	}
	h.Undo()
	h.Commit(Snapshot{Content: mustContent(t, "plain")})
	if h.Redo() {
		t.Fatal("stale branch")
	}
	h.Set(a)
	if h.Undo() || h.Redo() {
		t.Fatal("set retains history")
	}
	for i := 1; i <= 150; i++ {
		h.Commit(Snapshot{Content: mustContent(t, strings.Repeat("x", i))})
	}
	count := 0
	for h.Undo() {
		count++
	}
	if count != 100 || len(h.Current.Content.Text()) != 50 {
		t.Fatal(count, h.Current)
	}
	for h.Redo() {
		count--
	}
	if count != 0 || len(h.Current.Content.Text()) != 150 {
		t.Fatal(count)
	}
}

// Exercise interleaved plain edits and reference insertion across Unicode text.
// Every accepted edit must remain a valid independent snapshot, and every
// visible token edge must map back to its original byte offset.
func FuzzEdits(f *testing.F) {
	f.Add([]byte{1, 4, 3, 8, 2, 0, 9, 6, 7, 5})
	f.Add([]byte{255, 0, 128, 42, 99, 1, 4, 22})
	f.Fuzz(func(t *testing.T, edits []byte) {
		if len(edits) > 256 {
			edits = edits[:256]
		}
		c := mustContent(t, "hello 世界 e\u0301 👩‍💻")
		replacements := []string{"", "x", "世界", "e\u0301", "👩‍💻", "\u0301", "\n"}
		for i, b := range edits {
			var edges []int
			bs := boundaries(c.Text())
			for at := 0; at <= len(c.Text()); at++ {
				if bs[at] {
					edges = append(edges, at)
				}
			}
			a, z := edges[int(b)%len(edges)], edges[(int(b)+i)%len(edges)]
			if a > z {
				a, z = z, a
			}
			replacement := replacements[int(b)%len(replacements)]
			var token *Token
			if b%3 == 0 {
				token = &Token{ID: "shared", Text: replacement, Label: "显示"}
			}
			oldText, oldSpans := c.Text(), c.Tokens()
			next, caret, err := c.Replace(Range{a, z}, replacement, token)
			if c.Text() != oldText || !reflect.DeepEqual(c.Tokens(), oldSpans) {
				t.Fatal("mutated snapshot")
			}
			if err != nil {
				if !reflect.DeepEqual(next, c) || caret != (Range{a, z}) {
					t.Fatal("non-atomic failure")
				}
				continue
			}
			if _, err := New(next.Text(), next.Tokens()...); err != nil {
				t.Fatal(err)
			}
			if caret.Start != caret.End || caret.Start < 0 || caret.End > len(next.Text()) {
				t.Fatal(caret)
			}
			p := next.Presentation()
			for _, s := range p.Spans {
				if p.SourceOffset(s.Display.Start, 0) != s.Source.Start || p.SourceOffset(s.Display.End, 0) != s.Source.End || p.DisplayOffset(s.Source.Start, 0) != s.Display.Start || p.DisplayOffset(s.Source.End, 0) != s.Display.End {
					t.Fatal("mapping drift", s)
				}
			}
			c = next
		}
	})
}
