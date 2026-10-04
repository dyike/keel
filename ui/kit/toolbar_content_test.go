package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestToolbarCustomOverflowLifecycle(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		selectView := Select("zoom", "100%", "150%")
		v := Toolbar(ToolbarItem{Label: "command"}, ToolbarItem{Label: "custom group", Width: 180, Content: Button("inline action", func() { calls++ }), OverflowContent: selectView})
		width := 100
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(float32(width))).Child(v.Render(cx)) }), 600, scale)
		for range 3 {
			h.Frame()
		}
		click(t, h, "更多")
		h.Frame()
		click(t, h, "custom group")
		h.Frame()
		if _, ok := semanticNode(h, "dialog"); !ok {
			t.Fatal("missing custom overflow")
		}
		clickRole(t, h, "select", "zoom")
		h.Frame()
		click(t, h, "150%")
		h.Frame()
		if selectView.Value() != "150%" {
			t.Fatal("nested select did not work", selectView.Value())
		}
		h.Key(key.NameEscape, 0)
		h.Frame()
		if v.customOpen != -1 {
			t.Fatal("escape did not dismiss")
		}
		click(t, h, "更多")
		h.Frame()
		click(t, h, "custom group")
		h.Frame()
		v.SetItemDisabled(1, true)
		h.Frame()
		if _, ok := semanticNode(h, "dialog"); ok {
			t.Fatal("disabled group retained overlay")
		}
		v.SetItemDisabled(1, false)
		width = 600
		for range 4 {
			h.Frame()
		}
		click(t, h, "inline action")
		h.Frame()
		if calls != 1 {
			t.Fatal("inline action unavailable", calls)
		}
		v.SetDisabled(true)
		h.Frame()
		click(t, h, "inline action")
		h.Frame()
		if calls != 1 {
			t.Fatal("disabled parent allowed action")
		}
		v.SetDisabled(false)
		width = 100
		for range 4 {
			h.Frame()
		}
		click(t, h, "更多")
		h.Frame()
		click(t, h, "custom group")
		h.Frame()
		width = 600
		for range 4 {
			h.Frame()
		}
		if _, ok := semanticNode(h, "dialog"); ok {
			t.Fatal("wide toolbar kept overflow")
		}
		v.SetItems(ToolbarItem{Label: "replacement"})
		h.Frame()
		if shown(h, "inline action") || v.customOpen != -1 {
			t.Fatal("replacement kept old content")
		}
	}
}
