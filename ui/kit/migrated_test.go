package kit

import (
	"image"
	"slices"
	"strconv"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func text(s string) el.View { return viewFunc(func(*el.Context) el.Element { return el.Text(s) }) }

func TestTabsSwitchKeysAndKeepPages(t *testing.T) {
	in := Input("名字")
	tabs := Tabs().Add("基本", in).Add("高级", text("高级设置")).Add("关于", text("版本 1.0"))
	h := page(tabs)
	clickClass(t, h, "Editor", "名字")
	h.Type("小明")
	click(t, h, "高级")
	if tabs.Value() != 1 || !shown(h, "高级设置") || shown(h, "名字") {
		t.Fatal("click did not switch")
	}
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameRightArrow, 0) // wraps to the first
	h.Frame()
	if tabs.Value() != 0 || in.Value() != "小明" {
		t.Fatalf("keys: tab %d, kept %q", tabs.Value(), in.Value())
	}
}

func TestAccordionSingleMultipleKeys(t *testing.T) {
	a := Accordion().Add("一", text("内容一")).Add("二", text("内容二")).Add("三", text("内容三"))
	a.SetItemDisabled(1, true)
	h := page(a)
	click(t, h, "一")
	click(t, h, "三")
	if !slices.Equal(a.Value(), []int{2}) || shown(h, "内容一") {
		t.Fatalf("single: %v", a.Value())
	}
	h.Key(key.NameUpArrow, 0) // skips the disabled header to 一
	h.Key(key.NameReturn, 0)
	if !slices.Equal(a.Value(), []int{0}) {
		t.Fatalf("keys: %v", a.Value())
	}
	m := Accordion().Multiple().Add("A", text("a")).Add("B", text("b"))
	h = page(m)
	click(t, h, "A")
	click(t, h, "B")
	if !slices.Equal(m.Value(), []int{0, 1}) {
		t.Fatalf("multiple: %v", m.Value())
	}
}

func TestBadgeKeepsChildLayout(t *testing.T) {
	b := Badge(0).Child(Button("通知", nil))
	h := uitest.New(el.Embed(b))
	before := bounds(h, "通知")
	b.SetValue(150)
	h.Frame()
	if bounds(h, "通知") != before {
		t.Fatal("badge moved its child")
	}
	if n, ok := semanticNode(h, "badge:99+"); !ok || n.Desc.Label != "150" {
		t.Fatal("badge semantics")
	}
	b.SetValue(0)
	h.Frame()
	if _, ok := semanticNode(h, "badge:99+"); ok {
		t.Fatal("zero count still shown")
	}
}

func TestProgressLinkImage(t *testing.T) {
	p := Progress("导入")
	p.SetValue(0.42)
	clicks := 0
	l := Link("查看详情", func() { clicks++ })
	img := image.NewNRGBA(image.Rect(0, 0, 200, 100))
	pic := Image(img, "示意图")
	empty := Image(nil, "加载中的图")
	h := page(p, l, pic, empty)
	if _, ok := semanticNode(h, "progressbar:42%"); !ok {
		t.Fatal("progress")
	}
	click(t, h, "查看详情")
	h.Key(key.NameReturn, 0) // focused by the click
	if clicks != 2 {
		t.Fatalf("link clicks %d", clicks)
	}
	if b := bounds(h, "示意图"); b.Dx() != 200 || b.Dy() != 100 {
		t.Fatalf("image bounds %v", b)
	}
	if _, ok := semanticNode(h, "image:loading"); !ok {
		t.Fatal("placeholder")
	}
}

func TestMessageBubbleScrollerAttachment(t *testing.T) {
	var msgs []string
	for i := range 30 {
		msgs = append(msgs, "消息 "+strconv.Itoa(i))
	}
	loads := 0
	var sc *MessageScrollerView
	sc = MessageScroller(func(cx *el.Context) []el.Element {
		var out []el.Element
		for i, m := range msgs {
			msg := Message("AI", text(m))
			if i%2 == 0 {
				msg.User()
			}
			out = append(out, msg.Render(cx))
		}
		return out
	}).OnReachTop(func() {
		loads++
		msgs = append([]string{"更早的消息"}, msgs...)
		sc.HistoryPrepended()
	})
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().Child(sc.Render(cx)) })))
	h.Frame()
	if !shown(h, "消息 29") || shown(h, "消息 0") {
		t.Fatal("should start at the newest message")
	}
	for range 40 {
		h.Scroll(200, 150, -200)
	}
	h.Frame()
	if loads != 1 || !shown(h, "回到最新") {
		t.Fatalf("top: loads=%d latest button=%v", loads, shown(h, "回到最新"))
	}
	click(t, h, "回到最新")
	h.Frame()
	h.Frame()
	if !shown(h, "消息 29") {
		t.Fatal("jump to latest")
	}
	removed := 0
	a := Attachment("报价单.pdf", 1536*1024).OnRemove(func() { removed++ })
	a.SetProgress(0.5)
	h = page(a)
	if _, ok := semanticNode(h, "attachment:上传中 50%"); !ok || !shown(h, "上传中 50%") {
		t.Fatal("attachment progress")
	}
	click(t, h, "移除 报价单.pdf")
	a.SetProgress(-1)
	h.Frame()
	if removed != 1 || !shown(h, "1.5 MB") {
		t.Fatalf("remove %d / size", removed)
	}
}
