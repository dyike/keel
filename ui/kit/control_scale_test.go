package kit

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestControlsAdaptToIndependentFontScale(t *testing.T) {
	for _, display := range []float32{1, 1.25, 1.5, 2} {
		button := Button("Continue", nil)
		selectView := Select("Choice", "One", "Two")
		combo := Combobox("Search", "One", "Two")
		input := Input("Input")
		fontScale := float32(1)
		root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Items(el.Start).Child(button.Render(cx), selectView.Render(cx), combo.Render(cx), input.Render(cx))
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: display, PxPerSp: display * fontScale}
			gtx.Constraints.Max = image.Pt(1000, 1000)
			root.Layout(gtx)
		})
		before := map[string]int{}
		for _, name := range []string{"Continue", "Choice", "Search", "Input"} {
			before[name] = bounds(h, name).Dy()
		}
		normalRow := selectView.optionHeight()
		fontScale = 2
		h.Frame()
		for name, height := range before {
			if after := bounds(h, name).Dy(); after <= height {
				t.Fatalf("display=%v: %s did not grow: %d -> %d", display, name, height, after)
			}
		}
		if selectView.optionHeight() != normalRow*2 || combo.optionHeight() != normalRow*2 {
			t.Fatal("virtual row ignored font scale")
		}
		fontScale = 1
		h.Frame()
		if bounds(h, "Continue").Dy() != before["Continue"] || selectView.optionHeight() != normalRow {
			t.Fatal("scale change did not restore sizes")
		}
	}
}
