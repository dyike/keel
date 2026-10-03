package window

import (
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

func TestKitSelectCustomPresentationAndMenu(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	for _, dark := range []bool{false, true} {
		palette := theme.Light()
		if dark {
			palette = theme.Dark()
		}
		core.Update(func() { theme.Apply(palette) })
		s := kit.Select("Choice", "a", "b", "c", "d", "e", "f", "g", "h").Clearable(true).TitlePrefix("Value:").MenuWidth(360).MenuMaxHeight(160).RowHeight(40)
		s.RenderItem(func(cx *el.Context, row kit.SelectItemContext) el.Element { return el.Text("Item " + row.Option.Label) })
		s.RenderValue(func(cx *el.Context, rows []kit.SelectOption) el.Element { return el.Text("Chosen " + rows[0].Label) })
		w := openTest(t, Options{Width: 600, Height: 400, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Child(el.Div().W(el.Dp(240)).Child(s.Render(cx)))
		}))})
		w.click(element(t, w, "Choice").center())
		panelFound := false
		for _, e := range w.snapshot() {
			if e.Role == "listbox" {
				panelFound = true
				if e.Width < 350 || e.Width > 370 || e.Height > 161 {
					t.Fatalf("menu bounds: %+v", e)
				}
			}
		}
		if !panelFound {
			t.Fatal("missing listbox")
		}
		w.click(element(t, w, "b").center())
		w.snapshot()
		if s.Value() != "b" {
			t.Fatal("custom row selection", s.Value())
		}
		if e := element(t, w, "Choice"); e.Value != "b" {
			t.Fatal("stored value semantics", e)
		}
		w.click(element(t, w, locale.Current().Name(locale.Current().Clear, "Choice")).center())
		w.snapshot()
		if s.Value() != "" {
			t.Fatal("clear")
		}
		for _, e := range w.snapshot() {
			if e.Role == "listbox" {
				t.Fatal("clear opened menu")
			}
		}
		if err := w.press("space"); err != nil {
			t.Fatal(err)
		}
		w.snapshot()
		if _, err := w.screenshot(); err != nil {
			t.Fatal(err)
		}
	}
}
