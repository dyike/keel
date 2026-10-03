package kit

import (
	"math"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
)

func TestRadioSizeAndRichContentKeepSelectionAndFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		v := RadioGroup("Plan", "A", "B", "C").OnChange(func(string) { calls++ })
		v.SetOptionDisabled("B", true)
		h := renderView(v, 220, scale)
		initial := bounds(h, "A")
		click(t, h, "A")
		v.Size(36).TextSize(24).Content("A", el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Child(el.Text("Premium"), el.Text("Extra storage").TextSize(12))
		}))
		h.Frame()
		if b := bounds(h, "A"); b.Dy() <= initial.Dy() {
			t.Fatal("size/content ignored", initial, b)
		}
		n, ok := node(h, "A")
		if !ok || !n.Desc.Selected {
			t.Fatal("rich option lost selected semantics", n.Desc)
		}
		if _, ok := node(h, "Extra storage"); !ok {
			t.Fatal("rich description absent")
		}
		h.Key(key.NameRightArrow, 0)
		if v.Value() != "C" || calls != 2 {
			t.Fatal("rich content lost focus or disabled skipping", v.Value(), calls)
		}
		click(t, h, "Extra storage")
		if v.Value() != "A" || calls != 3 {
			t.Fatal("rich label click", v.Value(), calls)
		}
		v.Content("A", nil).Size(0).TextSize(0)
		h.Frame()
		if bounds(h, "A").Size() != initial.Size() {
			t.Fatal("default size not restored")
		}
		h.Key(key.NameRightArrow, 0)
		if v.Value() != "C" || calls != 4 {
			t.Fatal("restoring label lost focus")
		}
	}
}

func TestRadioContentOwnershipAndIndependentItems(t *testing.T) {
	v := RadioGroup("Plan", "A", "B")
	rich := el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("Description") })
	v.Content("A", rich).Content("unknown", rich)
	if len(v.content) != 1 {
		t.Fatal("accepted unknown option")
	}
	disabled := false
	h := render(func(cx *el.Context) el.Element {
		return el.Div().Disabled(disabled).Child(v.Item("A").Render(cx), v.Item("B").Render(cx))
	})
	click(t, h, "Description")
	if v.Value() != "A" {
		t.Fatal("independent rich item did not select")
	}
	v.SetOptions("B", "A")
	h.Frame()
	if _, ok := node(h, "Description"); !ok {
		t.Fatal("reorder discarded content")
	}
	v.SetValue("B")
	disabled = true
	h.Frame()
	click(t, h, "Description")
	if v.Value() != "B" {
		t.Fatal("ancestor disabled ignored")
	}
	v.SetOptions("B")
	v.SetOptions("A", "B")
	if len(v.content) != 0 {
		t.Fatal("removed content resurrected")
	}
}

func TestRadioSizeValidation(t *testing.T) {
	v := RadioGroup("Plan", "A").Size(28).TextSize(20)
	for _, bad := range []float32{-1, float32(math.NaN()), float32(math.Inf(1))} {
		v.Size(bad).TextSize(bad)
	}
	if v.size != 28 || v.textSize != 20 {
		t.Fatal("invalid size accepted")
	}
	v.Size(1).TextSize(1)
	if v.size != 12 || v.textSize != 8 {
		t.Fatal("minimum clamp")
	}
	v.Size(1000).TextSize(1000)
	if v.size != 64 || v.textSize != 128 {
		t.Fatal("maximum clamp")
	}
}
