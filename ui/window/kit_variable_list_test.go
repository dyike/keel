package window

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestVariableListAgentSnapshotAndClick(t *testing.T) {
	keys := make([]string, 10000)
	for i := range keys {
		keys[i] = fmt.Sprint(i)
	}
	selected := -1
	var cx *el.Context
	list := kit.VariableList(keys, 40, func(cx *el.Context, i int) el.Element {
		return el.Div().H(el.Dp(float32(30 + i%3*20))).Child(kit.Button(fmt.Sprintf("row-%d", i), func() { selected = i }).Render(cx))
	}).Height(180)
	w := openTest(t, Options{Width: 320, Height: 300, Content: el.Root(el.ViewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().Child(list.Render(c)) }))})
	for range 8 {
		w.render()
	}
	if roleOfName(w, "row-5000") != "" {
		t.Fatal("offscreen row in Agent snapshot")
	}
	list.ScrollTo(cx, 5000)
	for range 8 {
		w.render()
	}
	e := element(t, w, "row-5000")
	if e.Role != "button" || e.Y < 0 || e.Y >= 180 {
		t.Fatalf("revealed row %+v", e)
	}
	w.click(e.center())
	if selected != 5000 {
		t.Fatalf("selected %d", selected)
	}
	if roleOfName(w, "row-0") != "" {
		t.Fatal("old row remained in snapshot")
	}
}
