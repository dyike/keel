package sys

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf16"
)

type fakeShell struct {
	shown  []string
	hidden int
	fail   error
}

func (f *fakeShell) show(title, body string) error {
	if f.fail != nil {
		return f.fail
	}
	f.shown = append(f.shown, title+"|"+body)
	return nil
}
func (f *fakeShell) hide() error { f.hidden++; return nil }

func TestBalloonCenterReplacesRemovesAndClicks(t *testing.T) {
	shell := &fakeShell{}
	c := &balloonCenter{shell: shell}
	clicks := 0
	if err := c.post("a", "A", "first", func() { clicks += 1 }); err != nil {
		t.Fatal(err)
	}
	seqA := c.sequence()
	if err := c.post("b", "B", "second", func() { clicks += 10 }); err != nil {
		t.Fatal(err)
	}
	// The old balloon's late close does not clear the new one.
	if fn := c.event(balloonClosed, seqA); fn != nil || c.current != "b" {
		t.Fatal("stale close")
	}
	// Removing a replaced ID is harmless; removing the shown one hides it.
	if c.remove("a"); shell.hidden != 0 {
		t.Fatal("removed the wrong balloon")
	}
	if fn := c.event(balloonClicked, c.sequence()); fn == nil {
		t.Fatal("click handler")
	} else {
		fn()
	}
	if clicks != 10 || shell.hidden != 1 || c.current != "" {
		t.Fatal("click runs the current handler once and hides", clicks, shell.hidden)
	}
	if fn := c.event(balloonClicked, c.sequence()); fn != nil {
		t.Fatal("a second click after close")
	}
	c.post("c", "C", "third", nil)
	c.remove("c")
	if shell.hidden != 2 || c.current != "" {
		t.Fatal("remove the shown balloon")
	}
	shell.fail = errors.New("shell down")
	if err := c.post("d", "D", "x", nil); err == nil || c.current != "" {
		t.Fatal("a failed post keeps the old state")
	}
}

func TestBalloonTextFitsFields(t *testing.T) {
	if got := balloonText("short", 63); got != "short" {
		t.Fatal(got)
	}
	long := strings.Repeat("字", 100)
	if got := balloonText(long, 63); len(utf16.Encode([]rune(got))) != 63 || !strings.HasSuffix(got, "…") {
		t.Fatal("cut to the field", len(utf16.Encode([]rune(got))))
	}
	// An emoji is a surrogate pair; the cut never splits it.
	emoji := strings.Repeat("a", 61) + "😀😀"
	got := balloonText(emoji, 63)
	if strings.ContainsRune(got, '�') || len(utf16.Encode([]rune(got))) > 63 {
		t.Fatalf("split surrogate: %q", got)
	}
}
