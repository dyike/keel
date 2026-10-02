package el

import (
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestRevealKeepsNaturalChildrenAndClipsInput(t *testing.T) {
	fraction := float32(.5)
	calls := 0
	var body, first, last, after Element
	h := uitest.New(Root(viewFunc(func(*Context) Element {
		first = Div().W(Dp(100)).H(Dp(40)).Name("first").OnClick(func() { calls++ })
		last = Div().W(Dp(100)).H(Dp(40)).Name("last").Focusable(true).OnClick(func() { calls += 10 })
		body = Div().Reveal(fraction).Child(first, last)
		after = Div().H(Dp(20)).Name("after")
		return Div().W(Dp(100)).Items(Start).Child(body, after)
	})))
	if rect(body).Dy() != 40 || rect(last).Dy() != 40 || rect(after).Min.Y != 40 {
		t.Fatalf("natural geometry body=%v last=%v after=%v", rect(body), rect(last), rect(after))
	}
	h.Click(10, 50)
	if calls != 0 {
		t.Fatal("clipped content was clickable")
	}
	h.Click(10, 10)
	if calls != 1 {
		t.Fatal("visible content not clickable")
	}
	fraction = 1
	h.Frame()
	h.Click(10, 50)
	if calls != 11 {
		t.Fatal("revealed input missing")
	}
	fraction = 0
	h.Frame()
	if rect(body).Dy() != 0 || rect(after).Min.Y != 0 {
		t.Fatal("closed layout consumes height")
	}
}
