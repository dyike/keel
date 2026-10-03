package kit

import (
	"gioui.org/io/key"
	"testing"
)

func TestNumberGroupingEditUndoAndCommit(t *testing.T) {
	n := NumberInput("amount").ThousandsSeparator(',').Decimals(2).Range(-10000, 10000)
	h := renderView(n, 300, 1)
	clickClass(t, h, "Editor", "amount")
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 4}, Text: "１２３４．５"})
	h.Frame()
	if n.text != "1,234.5" || n.Value() != 0 {
		t.Fatal("live grouping", n.text, n.Value())
	}
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 2, End: 3}, Text: "９"})
	h.Frame()
	if n.text != "1,934.5" {
		t.Fatal("middle replacement", n.text)
	}
	h.Key("Z", key.ModShortcut)
	if n.text != "1,234.5" {
		t.Fatal("undo", n.text)
	}
	h.Key("Z", key.ModShortcut|key.ModShift)
	if n.text != "1,934.5" {
		t.Fatal("redo", n.text)
	}
	h.Key(key.NameReturn, 0)
	if n.Value() != 1934.5 || n.text != "1,934.50" {
		t.Fatal("commit", n.Value(), n.text)
	}
	h.Key(key.NameUpArrow, 0)
	if n.Value() != 1935.5 || n.text != "1,935.50" {
		t.Fatal("step", n.Value(), n.text)
	}
}

func TestNumberGroupingPrecisionRangeAndSeparators(t *testing.T) {
	for _, sep := range []rune{',', '\'', ' ', '\u00a0', '\u202f'} {
		n := NumberInput("n").Decimals(2).ThousandsSeparator(sep).Range(-1234.567, 1234.567)
		n.SetValue(-999999)
		if n.text != "-1"+string(sep)+"234.567" {
			t.Fatal("endpoint precision", sep, n.text)
		}
		n.text = "1" + string(sep) + "000.25"
		n.commit()
		if n.Value() != 1000.25 {
			t.Fatal("parse", sep, n.Value())
		}
		n.ThousandsSeparator(0)
		if n.text != "1000.25" {
			t.Fatal("disable grouping", n.text)
		}
	}
}

func TestNumberMalformedEditRejectedAndUndoContinues(t *testing.T) {
	n := NumberInput("n")
	h := renderView(n, 300, 1)
	clickClass(t, h, "Editor", "n")
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 1}, Text: ".5"})
	h.Frame()
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: 1, End: 1}, Text: "+"})
	h.Frame()
	if n.text != ".5" {
		t.Fatal("invalid sign accepted", n.text)
	}
	h.Key("Z", key.ModShortcut)
	if n.text != "0" {
		t.Fatal("rejected edit polluted undo", n.text)
	}
}
