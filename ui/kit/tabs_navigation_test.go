package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestTabsOverflowKeepsKeyboardSelectionVisible(t *testing.T) {
	v := Tabs()
	for _, s := range []string{"one.go", "two.go", "three.go", "four.go", "last.go"} {
		v.Add(s, text("body "+s))
	}
	h := sized(240, v)
	h.Frame()
	h.Frame()
	click(t, h, "one.go")
	h.Key(key.NameEnd, 0)
	h.Frame()
	if v.Value() != 4 || !shown(h, "last.go") {
		t.Fatal("selected overflow tab missing")
	}
	h.Key(key.NameLeftArrow, 0)
	h.Frame()
	if v.Value() != 3 || !shown(h, "four.go") {
		t.Fatal("overflow selection lost keyboard focus")
	}
}

func TestTabsCloseCurrentRestoresFocus(t *testing.T) {
	v := Tabs().Add("A", text("page A")).Add("B", text("page B")).Add("C", text("page C"))
	v.Closable(v.Remove)
	h := sized(400, v)
	click(t, h, "B")
	h.Key(key.NameDeleteForward, 0)
	h.Frame()
	h.Key(key.NameLeftArrow, 0)
	if v.Value() != 0 || !shown(h, "page A") {
		t.Fatal("close lost focus")
	}
}

func TestTabsMovePreservesPageAndEditor(t *testing.T) {
	in := Input("Editor")
	v := Tabs().Add("A", text("A body")).Add("B", in).Add("C", text("C body"))
	v.SetValue(1)
	h := sized(400, v)
	clickClass(t, h, "Editor", "Editor")
	h.Type("a")
	id := v.pages[1].id
	v.Move(1, 0)
	h.Frame()
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	if v.Value() != 0 || v.pages[0].id != id || in.Value() != "" {
		t.Fatal("move reset selected editor or focus")
	}
}

func TestTabsDragAndCancellation(t *testing.T) {
	calls := 0
	v := Tabs().Add("AAA", text("a")).Add("BBB", text("b")).Add("CCC", text("c")).Reorderable(func(from, to int) {
		calls++
		if from != 0 || to != 2 {
			t.Errorf("move %d %d", from, to)
		}
	})
	h := sized(500, v)
	ax, ay := center(bounds(h, "AAA"))
	cx, cy := center(bounds(h, "CCC"))
	h.Drag(ax, ay, cx, cy)
	if v.pages[2].Title != "AAA" || v.Value() != 2 || calls != 1 {
		t.Fatalf("drag order %#v current %d calls %d", v.pages, v.Value(), calls)
	}
	var context *el.Context
	h = render(func(c *el.Context) el.Element { context = c; return el.Div().W(el.Dp(500)).Child(v.Render(c)) })
	v.dragTab(context, 0, []int{0, 1, 2}, el.DragEvent{Kind: el.DragStart})
	v.dragTab(context, 0, []int{0, 1, 2}, el.DragEvent{Kind: el.DragEnd, X: 400, Canceled: true})
	if calls != 1 {
		t.Fatal("canceled drag committed")
	}
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "BBB")
	if v.Value() != 2 {
		t.Fatal("disabled tab selected")
	}
}

func TestTabsLongTitleAndFixedAreas(t *testing.T) {
	v := Tabs().Add("A very long tab title which must be clipped", text("body")).Add("B", text("other")).Add("C", text("third")).Add("D", text("fourth")).Trailing(Button("Add", func() {})).Size(32)
	h := sized(200, v)
	for range 4 {
		h.Frame()
	}
	if !shown(h, "Add") || !shown(h, "更多") {
		t.Fatalf("fixed area or overflow clipped avail=%v widths=%v visible=%v Add=%v More=%v", v.avail, v.widths, v.visible(), bounds(h, "Add"), bounds(h, "更多"))
	}
	n, ok := node(h, "A very long tab title which must be clipped")
	if !ok || n.Desc.Bounds.Max.X > bounds(h, "Add").Min.X {
		t.Fatalf("tab overlaps fixed slot %v", n.Desc.Bounds)
	}
}
