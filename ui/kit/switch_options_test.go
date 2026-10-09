package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestSwitchSizesLabelSideAndInteraction(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		s := Switch("Notifications", false).OnChange(func(bool) { calls++ })
		disabled := false
		h := renderView(viewFunc(func(cx *el.Context) el.Element {
			return el.Div().Disabled(disabled).Items(el.Start).Child(s.Render(cx))
		}), 220, scale)
		original := bounds(h, "Notifications")
		s.Size(SwitchSmall).LabelSide(el.Left)
		h.Frame()
		small := bounds(h, "Notifications")
		if small.Dx() >= original.Dx() || small.Max.X > 220*scale {
			t.Fatal("small layout", small, original)
		}
		// Left end is the label after reordering; the whole row is one control.
		h.Click(float32(small.Min.X+4*scale), float32(small.Min.Y+small.Dy()/2))
		h.Frame()
		if !s.Value() || calls != 1 {
			t.Fatal("label click")
		}
		s.LabelSide(el.Right).Size(SwitchMedium)
		h.Frame()
		h.Key(key.NameSpace, 0)
		h.Frame()
		if s.Value() || calls != 2 {
			t.Fatal("style change lost focus")
		}
		s.SetValue(true)
		h.Frame()
		if calls != 2 {
			t.Fatal("programmatic callback")
		}
		disabled = true
		h.Frame()
		click(t, h, "Notifications")
		if !s.Value() || calls != 2 {
			t.Fatal("ancestor disable")
		}
		disabled = false
		s.SetDisabled(true)
		h.Frame()
		click(t, h, "Notifications")
		if calls != 2 {
			t.Fatal("own disable")
		}
		s.Size(SwitchSmall).Size(SwitchSize(255)).LabelSide(el.Left).LabelSide(el.Top)
		if s.size != SwitchSmall || !s.labelLeft {
			t.Fatal("invalid option accepted")
		}
	}
}
