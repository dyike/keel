package window

import (
	"bytes"
	"image/color"
	"image/png"
	"slices"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestToggleStylePreservesFocusAndSelection(t *testing.T) {
	calls := 0
	v := kit.Toggle("Bold", false).Icon(kit.IconStar).OnChange(func(bool) { calls++ })
	group := kit.ToggleGroup("Left", "Right").Multiple()
	w := openTest(t, kitPage(v, group))
	w.press("Tab")
	for _, variant := range []kit.ToggleVariant{kit.ToggleGhost, kit.ToggleOutline, kit.ToggleDefault} {
		for i, size := range []kit.ToggleSize{kit.ToggleSizeXSmall, kit.ToggleSizeSmall, kit.ToggleSizeMedium, kit.ToggleSizeLarge} {
			v.Variant(variant).Size(size)
			group.Variant(variant).Size(size)
			e := element(t, w, "Bold")
			if e.Height != []int{24, 28, 32, 40}[i] {
				t.Fatalf("size %d height %d", size, e.Height)
			}
			if element(t, w, "Left").Height != e.Height {
				t.Fatal("group size differs")
			}
			old := v.Value()
			w.press("Space")
			if v.Value() == old {
				t.Fatal("style change lost keyboard focus")
			}
		}
	}
	if calls != 12 {
		t.Fatal(calls)
	}
	v.SetDisabled(true)
	w.snapshot()
	w.press("Space")
	if calls != 12 || !element(t, w, "Bold").Disabled {
		t.Fatal("disabled toggle activated")
	}
	w.click(element(t, w, "Left").center())
	w.click(element(t, w, "Right").center())
	if !slices.Equal(group.Value(), []string{"Left", "Right"}) {
		t.Fatal(group.Value())
	}
	group.Variant(kit.ToggleGhost).Size(kit.ToggleSizeSmall)
	w.snapshot()
	w.press("Space")
	if !slices.Equal(group.Value(), []string{"Left"}) {
		t.Fatal("group focus lost", group.Value())
	}
}

func TestToggleGhostAndOutlinePixels(t *testing.T) {
	old := theme.ReducedMotion
	core.Update(func() { theme.SetReducedMotion(true) })
	defer core.Update(func() { theme.SetReducedMotion(old) })
	v := kit.Toggle("Toggle", false).Variant(kit.ToggleGhost)
	backdrop := color.NRGBA{R: 120, G: 40, B: 80, A: 255}
	w := openTest(t, Options{Width: 220, Height: 100, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Bg(backdrop).P(16).Items(el.Start).Child(v.Render(cx))
	}))})
	sample := func(dx, dy int) color.NRGBA {
		t.Helper()
		e := element(t, w, "Toggle")
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(e.X+dx, e.Y+dy)).(color.NRGBA)
	}
	if c := sample(10, 5); c != backdrop {
		t.Fatal("ghost has background", c)
	}
	if c := sample(10, 0); c != backdrop {
		t.Fatal("ghost has border", c)
	}
	v.Variant(kit.ToggleOutline)
	if c := sample(10, 5); c != backdrop {
		t.Fatal("outline has background", c)
	}
	if c := sample(10, 0); c == backdrop {
		t.Fatal("outline border missing")
	}
	v.SetValue(true)
	if c := sample(10, 5); c == backdrop {
		t.Fatal("selection background missing")
	}
}
