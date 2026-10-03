package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestEmptyContentAndConstraints(t *testing.T) {
	v := Empty("暂无结果 0").Description("尝试其他关键词 Search again，或者调整筛选条件。")
	for _, scale := range []int{1, 2} {
		h := renderView(v, 140, scale)
		n, ok := node(h, "暂无结果 0")
		if !ok || n.Desc.Label != "暂无结果 0" || n.Desc.Bounds.Dx() > 140*scale {
			t.Fatalf("invalid empty state: %+v", n)
		}
		v.SetTitle("无数据")
		v.SetDescription("")
		h.Frame()
		if n, ok = node(h, "无数据"); !ok || n.Desc.Label != "无数据" {
			t.Fatal("stale empty state")
		}
		v.SetTitle("暂无结果 0")
	}
}

func TestEmptyAction(t *testing.T) {
	calls := 0
	v := Empty("空").Action(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Name("创建").OnClick(func() { calls++ }).Child(el.Text("创建"))
	}))
	h := renderView(v, 240, 1)
	click(t, h, "创建")
	if calls != 1 {
		t.Fatal("action failed")
	}
}

func TestEmptyMediaReplacementKeepsActionFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		input := Input("Search")
		v := Empty("Nothing").Description("Try again").Action(input).Media(Avatar("Alex").Size(48))
		h := renderView(v, 200, scale)
		if bounds(h, "Alex").Max.Y > bounds(h, "Nothing").Min.Y {
			t.Fatal("media order")
		}
		clickClass(t, h, "Editor", "Search")
		h.Type("a")
		h.Key(key.NameRightArrow, 0)
		v.Media(nil).Icon(IconNone)
		v.SetTitle("")
		v.SetDescription("")
		h.Frame()
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "" {
			t.Fatal("slot replacement lost action focus", input.Value())
		}
		v.Media(text("Custom illustration"))
		h.Frame()
		if !shown(h, "Custom illustration") {
			t.Fatal("replacement missing")
		}
		b := bounds(h, "Custom illustration")
		if b.Min.X < 0 || b.Max.X > 200*scale {
			t.Fatal("media overflow", b)
		}
		v.Media(nil)
		h.Frame()
		if shown(h, "Custom illustration") {
			t.Fatal("stale media")
		}
		v.Icon(IconInbox)
		h.Frame()
		if bounds(h, "Search").Min.Y <= 0 {
			t.Fatal("fallback missing")
		}
	}
}
