package kit

import (
	"math"
	"strconv"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
)

func TestNumberInputFinitePrecisionAndRange(t *testing.T) {
	calls := 0
	n := NumberInput("price").Decimals(2).OnChange(func(float64) { calls++ })
	n.SetValue(1.236)
	if n.Value() != 1.24 || n.text != "1.24" {
		t.Fatalf("precision: %v %q", n.Value(), n.text)
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		n.SetValue(bad)
		if n.Value() != 1.24 {
			t.Fatal("nonfinite assignment changed value")
		}
	}
	for _, bad := range []string{"NaN", "+Inf", "-Inf", "1e999", "--2", ""} {
		n.text = bad
		n.commit()
		if n.Value() != 1.24 || n.text != "1.24" {
			t.Fatalf("bad draft %q: %v %q", bad, n.Value(), n.text)
		}
	}
	n.Decimals(1000000).Decimals(-2)
	if n.decimals != 2 {
		t.Fatal("invalid precision accepted")
	}
	n.Range(math.NaN(), 3)
	if n.hi != math.Inf(1) {
		t.Fatal("NaN range accepted")
	}
	n.Range(math.Inf(1), math.Inf(1))
	if n.lo != math.Inf(-1) {
		t.Fatal("nonfinite singleton range accepted")
	}
	n.Range(0.001, 0.009)
	for _, x := range []float64{-1, 0.004, 1} {
		n.SetValue(x)
		shown, err := strconv.ParseFloat(n.text, 64)
		if err != nil || shown != n.Value() || shown < 0.001 || shown > 0.009 {
			t.Fatalf("boundary: %v %q", n.Value(), n.text)
		}
	}
	if calls != 0 {
		t.Fatalf("programmatic changes or rejected input fired %d callbacks", calls)
	}
}

func TestNumberInputDecimalStepAndSingleCommit(t *testing.T) {
	var changes []float64
	n := NumberInput("price").Step(0.1).OnChange(func(x float64) { changes = append(changes, x) })
	for range 3 {
		n.move(1)
	}
	if n.Value() != 0.3 || n.text != "0.3" {
		t.Fatalf("decimal addition: %v %q", n.Value(), n.text)
	}
	changes = nil
	n.text = "1.2"
	n.move(1)
	if n.Value() != 1.3 || len(changes) != 1 || changes[0] != 1.3 {
		t.Fatalf("single commit: %v", changes)
	}
	n.SetValue(math.MaxFloat64)
	n.Step(math.MaxFloat64)
	n.move(10)
	if !finiteNumber(n.Value()) {
		t.Fatal("step overflow")
	}
}

func TestNumberInputDisabledDraftAndNormalBlur(t *testing.T) {
	for _, ancestor := range []bool{false, true} {
		t.Run(strconv.FormatBool(ancestor), func(t *testing.T) {
			disabled := false
			calls := 0
			n := NumberInput("price").OnChange(func(float64) { calls++ })
			h := render(func(cx *el.Context) el.Element {
				return el.Div().Child(el.Div().Disabled(disabled).Child(n.Render(cx)), Button("outside", func() {}).Render(cx))
			})
			clickClass(t, h, "Editor", "price")
			h.Key("A", key.ModShortcut)
			h.Type("5")
			if ancestor {
				disabled = true
			} else {
				n.SetDisabled(true)
			}
			for range 4 {
				h.Frame()
			}
			disabled = false
			n.SetDisabled(false)
			for range 4 {
				h.Frame()
			}
			if n.Value() != 0 || n.text != "0" || calls != 0 {
				t.Fatalf("disabled committed: %v %q %d", n.Value(), n.text, calls)
			}
			clickClass(t, h, "Editor", "price")
			h.Key("A", key.ModShortcut)
			h.Type("7")
			want, err := strconv.ParseFloat(n.text, 64)
			if err != nil || want == 0 {
				t.Fatalf("draft not entered: %q", n.text)
			}
			click(t, h, "outside")
			for range 4 {
				h.Frame()
			}
			if n.Value() != want || calls != 1 {
				t.Fatalf("normal blur: %v %d", n.Value(), calls)
			}
		})
	}
}

func TestNumberInputDraftButtonCommitsOnce(t *testing.T) {
	var changes []float64
	n := NumberInput("price").Step(1).OnChange(func(x float64) { changes = append(changes, x) })
	h := page(n)
	clickClass(t, h, "Editor", "price")
	h.Type("2")
	draft, err := strconv.ParseFloat(n.text, 64)
	if err != nil || draft == 0 {
		t.Fatalf("draft: %q", n.text)
	}
	click(t, h, "增加 price")
	for range 3 {
		h.Frame()
	}
	if n.Value() != draft+1 || len(changes) != 1 {
		t.Fatalf("button committed %v, value %v", changes, n.Value())
	}
}

func TestNumberInputButtonsUseDraftBoundary(t *testing.T) {
	n := NumberInput("price").Range(0, 10)
	n.SetValue(10)
	h := page(n)
	n.text = "2"
	h.Frame()
	click(t, h, "增加 price")
	if n.Value() != 3 {
		t.Fatalf("upper boundary disabled draft increment: %v", n.Value())
	}
	n.SetValue(0)
	n.text = "8"
	h.Frame()
	click(t, h, "减少 price")
	if n.Value() != 7 {
		t.Fatalf("lower boundary disabled draft decrement: %v", n.Value())
	}
}
