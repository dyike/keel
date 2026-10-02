package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestToolbarFirstFrameKeepsActionsAccessibleBesideSlots(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		v := Toolbar(ToolbarItem{Label: "Long command 中文 123", Action: func() { calls++ }}, ToolbarItem{Label: "Another command"}).Leading(Button("文件", nil)).Trailing(Button("帮助", nil))
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(224)).Child(v.Render(cx)) }), 224, scale)
		left, more, right := bounds(h, "文件"), bounds(h, "更多"), bounds(h, "帮助")
		if left.Empty() || more.Empty() || right.Empty() || left.Max.X > more.Min.X || more.Max.X > right.Min.X || right.Max.X > 224*scale {
			t.Fatal(left, more, right)
		}
		click(t, h, "更多")
		h.Frame()
		click(t, h, "Long command 中文 123")
		h.Frame()
		if calls != 1 {
			t.Fatal("first-frame overflow action lost", calls)
		}
	}
}
