package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestMenuContentClickAndReplacement(t *testing.T) {
	calls := 0
	rich := el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(el.Text("Primary"), el.Text("Description").TextSize(12))
	})
	m := Menu().ContentItem("Action", "", rich, func() { calls++ })
	m.Trigger(Button("Open", m.Toggle))
	h := page(m)
	click(t, h, "Open")
	h.Frame()
	if b := bounds(h, "Action"); b.Dy() <= 30 {
		t.Fatal("custom content clipped", b)
	}
	click(t, h, "Description")
	h.Frame()
	if calls != 1 || m.Value() {
		t.Fatal("rich content click")
	}
	click(t, h, "Open")
	h.Frame()
	m.SetItemContent("Action", nil)
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if calls != 2 || m.Value() {
		t.Fatal("replacement lost focus")
	}
	m.SetItemContent("Action", rich)
	m.SetItemDisabled("Action", true)
	click(t, h, "Open")
	h.Frame()
	click(t, h, "Description")
	if calls != 2 || !m.Value() {
		t.Fatal("disabled custom content activated")
	}
}

func TestMenuVariableRowsEndAndHome(t *testing.T) {
	for _, scale := range []int{1, 2} {
		selected := -1
		m := Menu().Label("Commands")
		for i := 0; i < 20; i++ {
			m.ContentItem(fmt.Sprintf("Action %d", i), "", el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().H(el.Dp(float32(40+i*3))).Child(el.Text(fmt.Sprintf("Title %d", i)), el.Text("Details"))
			}), func() { selected = i })
		}
		m.Trigger(Button("Open", m.Toggle))
		h := renderView(m, 240, scale)
		click(t, h, "Open")
		h.Frame()
		h.Key(key.NameEnd, 0)
		h.Frame()
		h.Frame()
		n, ok := node(h, "Action 19")
		if !ok || n.Desc.Bounds.Min.Y < 0 || n.Desc.Bounds.Max.Y > 1000*scale {
			t.Fatal("variable End not visible", n.Desc.Bounds)
		}
		h.Key(key.NameReturn, 0)
		h.Frame()
		if selected != 19 {
			t.Fatal("End lost focus", selected)
		}
		click(t, h, "Open")
		h.Frame()
		h.Key(key.NameEnd, 0)
		h.Frame()
		h.Frame()
		h.Key(key.NameHome, 0)
		h.Frame()
		h.Frame()
		h.Key(key.NameReturn, 0)
		h.Frame()
		if selected != 0 {
			t.Fatal("Home lost focus", selected)
		}
	}
}
