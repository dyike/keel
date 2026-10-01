package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestTableDisabledBlocksSelectionSortAndActivation(t *testing.T) {
	changes, activations := 0, 0
	tb := Table(Col("列", 1)).Height(160).OnSelect(func(int) { changes++ }).OnActivate(func(int) { activations++ })
	tb.SetRows([][]string{{"乙"}, {"甲"}, {"丙"}})
	tb.SetSelected(0)
	h := uitest.New(tb)
	if changes != 0 {
		t.Fatal("setter notified")
	}
	clickNamed(t, h, "丙")
	if tb.Selected() != 2 || changes != 1 {
		t.Fatal("row did not select")
	}
	tb.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "甲")
	clickNamed(t, h, "列")
	h.Key(key.NameHome, 0)
	h.Key(key.NameReturn, 0)
	if tb.Selected() != 2 || tb.sortCol != -1 || changes != 1 || activations != 0 {
		t.Fatal("disabled table changed")
	}
	if h.Router.Source().Focused(tb) {
		t.Fatal("disabled table retained focus")
	}
	tb.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "甲")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if tb.Selected() != 2 || changes != 3 || activations != 1 {
		t.Fatal("table did not recover")
	}
}
