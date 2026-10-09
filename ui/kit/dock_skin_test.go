package kit

import (
	"reflect"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

func TestDockSkinPreservesPanelFocusAndLayout(t *testing.T) {
	for _, scale := range []int{1, 2} {
		input := Input("Query")
		view := Dock(text("center")).Panel(DockPanel{ID: "a", Title: "A", View: input}, DockLeft).
			Panel(DockPanel{ID: "b", Title: "B", View: text("B body")}, DockLeft)
		view.Split("b", "a", DockPlacementBottom)
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(900)).H(el.Dp(600)).Child(view.Render(cx))
		}), 900, scale)
		clickClass(t, h, "Editor", "Query")
		h.Type("a")
		original := bounds(h, "Query")
		layout := view.Layout()
		var panels, headers, bodies, tabs, separators int
		view.Skin(&DockSkin{
			Panel:  func(e *el.DivEl) { panels++; e.Bg(theme.Bg) },
			Header: func(e *el.DivEl) { headers++; e.Bg(theme.Surface) },
			Body:   func(e *el.DivEl) { bodies++; e.P(24) },
			Tab: func(e *el.DivEl, selected bool) {
				tabs++
				if selected {
					e.Bold()
				}
			},
			Separator: func(e *el.DivEl) { separators++; e.Bg(theme.Primary) },
		})
		h.Frame()
		if panels != 2 || headers != 2 || bodies != 2 || tabs != 2 || separators != 2 {
			t.Fatalf("skin missed groups or split/outer separators: %d %d %d %d %d", panels, headers, bodies, tabs, separators)
		}
		if bounds(h, "Query").Min.X <= original.Min.X || !reflect.DeepEqual(layout, view.Layout()) {
			t.Fatal("skin padding did not apply, or changed persisted layout")
		}
		h.Key(key.NameRightArrow, 0)
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "" {
			t.Fatal("skin change lost editor focus")
		}
		view.Skin(nil)
		h.Frame()
		h.Type("b")
		if input.Value() != "b" || bounds(h, "Query") != original {
			t.Fatal("reset lost state or default padding")
		}
		click(t, h, "调整大小")
		h.Key(key.NameRightArrow, 0)
		if reflect.DeepEqual(layout, view.Layout()) {
			t.Fatal("separator stopped responding")
		}
	}
}
