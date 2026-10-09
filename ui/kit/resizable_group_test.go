package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestResizableGroupAdjacentResize(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		for _, scale := range []int{1, 2} {
			calls := 0
			g := ResizableGroup(ResizablePanel{ID: "a", Size: 150, Min: 50, Max: 240}, ResizablePanel{ID: "b", Size: 150, Min: 80, Max: 250}, ResizablePanel{ID: "c", Size: 188, Min: 60, Max: 300}).OnChange(func(s map[string]float32) { calls++; s["a"] = -1 })
			if vertical {
				g.Vertical()
			}
			h := renderView(viewFunc(func(cx *el.Context) el.Element {
				return el.Div().W(el.Dp(500)).H(el.Dp(500)).Items(el.Stretch).Child(g.Render(cx))
			}), 500, scale)
			h.Frame()
			b := bounds(h, locale.Current().Resize+" a / b")
			x, y := center(b)
			dx, dy := float32(40*scale), float32(0)
			if vertical {
				dx, dy = dy, dx
			}
			h.Drag(x, y, x+dx, y+dy)
			h.Frame()
			s := g.Sizes()
			if s["a"] != 190 || s["b"] != 110 || s["c"] != 188 {
				t.Fatal("adjacent drag", s)
			}
			h.Key(key.NameEnd, 0)
			if g.Sizes()["a"] != 220 || g.Sizes()["b"] != 80 {
				t.Fatal("pair limits", g.Sizes())
			}
			h.Key(key.NameHome, 0)
			if g.Sizes()["a"] != 50 || g.Sizes()["b"] != 250 {
				t.Fatal("pair max", g.Sizes())
			}
			before := calls
			g.SetDisabled(true)
			h.Frame()
			h.Key(key.NameEnd, 0)
			if calls != before {
				t.Fatal("disabled callback")
			}
		}
	}
}

func TestResizableGroupFitAndUpdates(t *testing.T) {
	g := ResizableGroup(ResizablePanel{ID: "a", Size: 100, Min: 40, Max: 120}, ResizablePanel{ID: "b", Size: 100, Min: 60, Max: 150}, ResizablePanel{ID: "b"}, ResizablePanel{})
	g.total, g.measured = 500, true
	g.fit()
	if len(g.panels) != 2 || g.sizes["a"] != 120 || g.sizes["b"] != 150 {
		t.Fatal("maximum capacity", g.Sizes())
	}
	g.total = 56
	g.fit()
	if g.sizes["a"] != 20 || g.sizes["b"] != 30 {
		t.Fatal("insufficient minimums", g.Sizes())
	}
	g.SetVisible("b", false)
	g.total = 100
	g.fit()
	if g.sizes["a"] != 100 || g.sizes["b"] != 30 {
		t.Fatal("hidden retained size", g.Sizes())
	}
	g.SetPanels(ResizablePanel{ID: "b", Min: 0}, ResizablePanel{ID: "a", Min: 0})
	if g.sizes["a"] != 100 || g.sizes["b"] != 30 {
		t.Fatal("reorder lost sizes")
	}
	g.SetSizes(map[string]float32{"a": 0, "b": 94})
	g.fit()
	if g.sizes["a"] != 0 {
		t.Fatal("explicit zero reset as auto", g.Sizes())
	}
}

func TestResizableGroupAutoPaneFirstFrame(t *testing.T) {
	content := viewFunc(func(*el.Context) el.Element { return el.Div().Grow().Role("group").Name("auto") })
	g := ResizableGroup(ResizablePanel{ID: "a", Size: 100}, ResizablePanel{ID: "auto", Content: content, Min: 60}, ResizablePanel{ID: "c", Size: 100})
	h := renderView(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(400)).H(el.Dp(100)).Items(el.Stretch).Child(g.Render(cx))
	}), 400, 1)
	if bounds(h, "auto").Dx() != 188 {
		t.Fatal("first frame auto pane", bounds(h, "auto"))
	}
	h.Frame()
	if g.Sizes()["auto"] != 188 || bounds(h, "auto").Dx() != 188 {
		t.Fatal("auto allocation shifted", g.Sizes())
	}
}
