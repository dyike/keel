package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestResizableVisibilityLayout(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, scale := range []int{1, 2} {
			pane := func(name string) el.View {
				return viewFunc(func(*el.Context) el.Element { return el.Div().Grow().Name(name).Role("group") })
			}
			calls := 0
			v := Resizable(pane("first"), pane("second")).Min(40, 40).Max(180, 220).OnChange(func(float32) { calls++ })
			v.SetValue(170)
			if vertical {
				v.Vertical()
			}
			length := float32(350)
			h := renderView(viewFunc(func(cx *el.Context) el.Element {
				return el.Div().W(el.Dp(length)).H(el.Dp(length)).Items(el.Stretch).Child(v.Render(cx))
			}), 500, scale)
			h.Frame()
			for _, show := range [][2]bool{{false, true}, {true, false}, {false, false}, {true, true}} {
				v.Visible(show[0], show[1])
				h.Frame()
				h.Frame()
				for i, name := range []string{"first", "second"} {
					b := bounds(h, name)
					if show[i] && b.Empty() {
						t.Fatal("visible pane missing", name)
					}
					if !show[i] && !b.Empty() {
						t.Fatal("hidden pane present", name, b)
					}
					if show[i] && !show[1-i] {
						size := b.Dx()
						if vertical {
							size = b.Dy()
						}
						if size != 350*scale {
							t.Fatal("solo pane not filled", b)
						}
					}
				}
				if v.Value() != 170 || calls != 0 {
					t.Fatal("visibility changed split", v.Value(), calls)
				}
			}
			v.Visible(false, true)
			length = 250
			h.Frame()
			h.Frame()
			if v.Value() != 170 {
				t.Fatal("hidden resize lost stored value")
			}
			v.Visible(true, true)
			h.Frame()
			h.Frame()
			if v.Value() != 170 {
				t.Fatal("restore changed feasible value")
			}
		}
	}
}
