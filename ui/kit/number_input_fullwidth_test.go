package kit

import (
	"gioui.org/io/key"
	"testing"
)

func TestNumberInputFullWidthNumericBoundaries(t *testing.T) {
	n := NumberInput("number").Range(-100, 100).Decimals(2)
	for _, tc := range []struct {
		text    string
		value   float64
		display string
	}{
		{"＋１２．５", 12.5, "12.50"},
		{"－２。７５", -2.75, "-2.75"},
		{"．５", 0.5, "0.50"},
		{" １２３ ", 100, "100.00"},
	} {
		n.text = tc.text
		if got := n.draftValue(); got != tc.value {
			t.Fatalf("draft %q: %v", tc.text, got)
		}
		n.commit()
		if n.Value() != tc.value || n.text != tc.display {
			t.Fatalf("commit %q: %v %q", tc.text, n.Value(), n.text)
		}
	}
	n.SetValue(8)
	n.text = "１．２．３"
	n.commit()
	if n.Value() != 8 || n.text != "8.00" {
		t.Fatal("invalid decimal accepted")
	}
	n.text = "９．５"
	n.Step(0.1).move(1)
	if n.Value() != 9.6 {
		t.Fatal("full-width draft did not step")
	}
}

func TestNumberInputFullWidthTypingUndoAndCommit(t *testing.T) {
	n := NumberInput("number")
	h := renderView(n, 240, 1)
	clickClass(t, h, "Editor", "number")
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 1}, Text: "１２．５"})
	h.Frame()
	if n.text != "12.5" || n.Value() != 0 {
		t.Fatal("draft changed before commit", n.text, n.Value())
	}
	// Normalization preserves rune offsets for middle insertion.
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 1, End: 1}, Text: "３"})
	h.Frame()
	if n.text != "132.5" {
		t.Fatal("middle insertion", n.text)
	}
	h.Key("Z", key.ModShortcut)
	if n.text != "12.5" {
		t.Fatal("undo normalization loop", n.text)
	}
	h.Key("Z", key.ModShortcut|key.ModShift)
	if n.text != "132.5" {
		t.Fatal("redo", n.text)
	}
	h.Key(key.NameReturn, 0)
	if n.Value() != 132.5 || n.text != "132.5" {
		t.Fatal("commit", n.Value(), n.text)
	}
}
