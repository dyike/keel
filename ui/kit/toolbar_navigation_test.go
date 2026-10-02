package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestToolbarKeyboardReachesOverflowAndReturns(t *testing.T) {
	v := Toolbar(ToolbarItem{Label: "First"}, ToolbarItem{Label: "Second"}, ToolbarItem{Label: "Third"}, ToolbarItem{Label: "Fourth"})
	h := sized(200, v)
	for range 3 {
		h.Frame()
	}
	click(t, h, "First")
	h.Key(key.NameEnd, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !v.more.Value() {
		t.Fatal("End/Enter did not open overflow")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Key(key.NameHome, 0)
	h.Key(key.NameReturn, 0)
	if v.active != 0 || v.more.Value() {
		t.Fatal("overflow did not restore keyboard navigation")
	}
}

func TestToolbarSlotsSizeAndDisabled(t *testing.T) {
	actions := 0
	v := Toolbar(ToolbarItem{Label: "First", Action: func() { actions++ }}, ToolbarItem{Label: "Second"}, ToolbarItem{Label: "Third"}).
		Leading(text("File")).Trailing(Button("Help", func() { actions++ })).Size(40)
	h := sized(260, v)
	for range 3 {
		h.Frame()
	}
	if !shown(h, "File") || !shown(h, "Help") || !shown(h, "更多") {
		t.Fatal("fixed slots or overflow missing")
	}
	if bounds(h, "First").Dy() != 40 {
		t.Fatal("size not applied")
	}
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "First")
	click(t, h, "Help")
	if actions != 0 {
		t.Fatal("disabled actions")
	}
	v.SetDisabled(false)
	h.Frame()
	click(t, h, "First")
	if actions != 1 {
		t.Fatal("restore")
	}
}

func TestToolbarCopiesItemsAndKeepsDuplicateDisablingSeparate(t *testing.T) {
	input := []ToolbarItem{{Label: "Same", Disabled: true}, {Label: "Same"}}
	v := Toolbar(input...)
	input[0].Disabled = false
	if !v.Items()[0].Disabled {
		t.Fatal("constructor alias")
	}
	copy := v.Items()
	copy[1].Disabled = true
	if v.Items()[1].Disabled {
		t.Fatal("getter alias")
	}
	h := sized(40, v)
	for range 3 {
		h.Frame()
	}
	if len(v.more.items) != 2 || !v.more.items[0].disabled || v.more.items[1].disabled {
		t.Fatal("duplicate labels share disabled state")
	}
	v.SetItems(input...)
	input[1].Label = "changed"
	if v.Items()[1].Label != "Same" {
		t.Fatal("setter alias")
	}
}

func TestToolbarIconHasKeyboardTooltip(t *testing.T) {
	v := Toolbar(ToolbarItem{Label: "Search", Icon: IconSearch, HasIcon: true, IconOnly: true})
	c := &clock{now: time.Now()}
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(200)).Child(v.Render(cx)) })
	h.Frame() // Command width is measured on the first frame.
	click(t, h, "Search")
	c.advance(h, time.Second)
	h.Frame()
	if _, ok := semanticNode(h, "tooltip"); !ok {
		t.Fatal("icon tooltip missing")
	}
}
