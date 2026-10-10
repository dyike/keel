package kit

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
)

func TestMenuScrollbarConfigurationInheritance(t *testing.T) {
	child := Menu()
	root := Menu().Scrollbars(el.ScrollbarAlways).Sub("Child", child)
	if mode, set := child.scrollbarSetting(); !set || mode != el.ScrollbarAlways {
		t.Fatal("child did not inherit")
	}
	child.Scrollbars(el.ScrollbarHover)
	root.Scrollbars(el.ScrollbarScrolling)
	if mode, _ := child.scrollbarSetting(); mode != el.ScrollbarHover {
		t.Fatal("parent overwrote child")
	}
	child.Scrollbars(el.ScrollbarMode(255))
	if mode, _ := child.scrollbarSetting(); mode != el.ScrollbarHover {
		t.Fatal("invalid mode changed configuration")
	}
	if _, set := Menu().scrollbarSetting(); set {
		t.Fatal("default overrides global setting")
	}
}

func TestMenuScrollbarGutterOnlyWhenOverflowing(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, count := range []int{3, 60} {
			t.Run(fmt.Sprintf("%d/%d", scale, count), func(t *testing.T) {
				menu := Menu().Scrollbars(el.ScrollbarAlways)
				for i := 0; i < count; i++ {
					menu.Item(fmt.Sprintf("Row %d", i), "", nil)
				}
				h := renderView(viewFunc(func(cx *el.Context) el.Element { return menu.panel(cx) }), 320, scale)
				settle(h)
				container, ok := semanticNode(h, "menu")
				if !ok {
					t.Fatal("missing menu")
				}
				row := bounds(h, "Row 0")
				left := row.Min.X - container.Desc.Bounds.Min.X
				right := container.Desc.Bounds.Max.X - row.Max.X
				expected := 5 * scale // 4dp row margin plus the panel border
				if count > 3 {
					expected = (scrollbarGutter + 1) * scale
				}
				if left != 5*scale || right != expected {
					t.Fatalf("margins left=%d right=%d expected=%d", left, right, expected)
				}
			})
		}
	}
}
