package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"math"
	"testing"
)

func TestNumberInputDynamicStep(t *testing.T) {
	calls, strategies := 0, 0
	n := NumberInput("price").Range(0, 20).Step(0.25).OnChange(func(float64) { calls++ })
	n.StepBy(func(value float64, action NumberStepAction) float64 {
		strategies++
		if value < 1 || value == 1 && action == NumberStepActionDecrement {
			return 0.1
		}
		return 0.5
	})
	n.SetValue(1)
	n.move(-1)
	if n.Value() != 0.9 {
		t.Fatal("downward boundary", n.Value())
	}
	n.move(1)
	n.move(1)
	if n.Value() != 1.5 || strategies != 3 || calls != 3 {
		t.Fatal("upward boundary", n.Value(), strategies, calls)
	}
	n.text = "2.0"
	n.move(10)
	if n.Value() != 7 || calls != 4 || strategies != 4 {
		t.Fatal("draft/page action", n.Value(), calls, strategies)
	}
	for _, bad := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		n.StepBy(func(float64, NumberStepAction) float64 { return bad })
		n.text = "3.25"
		n.move(1)
		if n.Value() != 7 || n.text != "3.25" || calls != 4 {
			t.Fatal("invalid strategy committed draft")
		}
	}
	n.StepBy(nil)
	n.move(1)
	if n.Value() != 3.5 {
		t.Fatal("fixed fallback", n.Value())
	}
	n.StepBy(func(float64, NumberStepAction) float64 { return 10 }).Step(0.1)
	n.move(1)
	if n.Value() != 3.6 {
		t.Fatal("Step did not replace strategy")
	}
	n.SetDisabled(true)
	n.move(1)
	if n.Value() != 3.6 {
		t.Fatal("disabled action")
	}
}

func TestNumberInputSlotsKeyboardAndFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		strategies, actions := 0, 0
		n := NumberInput("price").StepBy(func(float64, NumberStepAction) float64 { strategies++; return 0.5 })
		n.Prefix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("USD") })).Suffix(Button("Info", func() { actions++ }).Size(24))
		h := renderView(n, 280, scale)
		if strategies != 0 {
			t.Fatal("strategy called during layout")
		}
		clickClass(t, h, "Editor", "price")
		h.Key(key.NameUpArrow, 0)
		if n.Value() != 0.5 || strategies != 1 {
			t.Fatal("keyboard step")
		}
		n.Prefix(nil)
		h.Frame()
		h.Key(key.NamePageUp, 0)
		if n.Value() != 5.5 || strategies != 2 {
			t.Fatal("slot removal lost focus")
		}
		click(t, h, "Info")
		if actions != 1 || n.Value() != 5.5 {
			t.Fatal("suffix action changed number")
		}
		loc := locale.Current()
		click(t, h, loc.Name(loc.Increase, "price"))
		if n.Value() != 6 || strategies != 3 {
			t.Fatal("increment button ignored strategy")
		}
		click(t, h, loc.Name(loc.Decrease, "price"))
		if n.Value() != 5.5 || strategies != 4 {
			t.Fatal("decrement button ignored strategy")
		}
		n.SetDisabled(true)
		h.Frame()
		click(t, h, "Info")
		if actions != 1 || strategies != 4 {
			t.Fatal("disabled suffix")
		}
	}
}
