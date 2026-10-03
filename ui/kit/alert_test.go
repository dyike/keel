package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestAlertWrapsAndRetainsStatus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		a := Alert("保存 123").Description("中英文 Mixed description that wraps in a narrow container. 多行提示内容。").Tone(ToneSuccess)
		h := renderView(a, 160, scale)
		for i := 0; i < 3; i++ {
			h.Frame()
		}
		n, ok := semanticNode(h, "alert:success")
		if !ok || n.Desc.Label != "保存 123" || n.Desc.Bounds.Dx() > 160*scale || n.Desc.Bounds.Dy() <= 40*scale {
			t.Fatalf("invalid alert: %+v", n)
		}
		a.SetTitle("更新")
		a.SetDescription("")
		a.SetTone(ToneWarning)
		h.Frame()
		if n, ok = semanticNode(h, "alert:warning"); !ok || n.Desc.Label != "更新" {
			t.Fatal("updated alert not rendered")
		}
	}
}

func TestAlertDismissAndRestore(t *testing.T) {
	calls := 0
	a := Alert("失败").Tone(ToneDanger).Description("网络不可用").OnClose(func() { calls++ })
	h := renderView(a, 240, 1)
	click(t, h, "关闭 失败")
	h.Frame()
	if a.Visible() || calls != 1 {
		t.Fatal("dismiss failed")
	}
	a.SetVisible(true)
	h.Frame()
	if _, ok := node(h, "失败"); !ok || calls != 1 {
		t.Fatal("restore invoked callback or stayed hidden")
	}
	a.SetVisible(false)
	h.Frame()
	if calls != 1 {
		t.Fatal("program visibility invoked callback")
	}
}
func TestAlertKindsAndDescriptionHeight(t *testing.T) {
	for _, k := range []Tone{ToneNeutral, ToneInfo, ToneSuccess, ToneWarning, ToneDanger} {
		a := Alert("标题").Tone(k)
		h := renderView(a, 240, 1)
		if _, ok := semanticNode(h, "alert:"+k.name()); !ok {
			t.Fatal(k)
		}
		if _, ok := node(h, "关闭 标题"); ok {
			t.Fatal("unexpected close")
		}
		n, _ := semanticNode(h, "alert:"+k.name())
		before := n.Desc.Bounds.Dy()
		a.Description("说明")
		h.Frame()
		n, _ = semanticNode(h, "alert:"+k.name())
		if n.Desc.Bounds.Dy() <= before {
			t.Fatal("description did not add height")
		}
	}
}

func TestAlertDisabledRestore(t *testing.T) {
	n := 0
	v := Alert("提示").OnClose(func() { n++ })
	v.SetDisabled(true)
	h := renderView(v, 220, 1)
	click(t, h, "关闭 提示")
	if !v.Visible() || n != 0 {
		t.Fatal("disabled closed")
	}
	v.SetDisabled(false)
	h.Frame()
	click(t, h, "关闭 提示")
	if v.Visible() || n != 1 {
		t.Fatal("enabled failed")
	}
}

func TestAlertSizesBannerAndRichActions(t *testing.T) {
	for _, scale := range []int{1, 2} {
		a := Alert("Title").Description("Message")
		h := renderView(a, 300, scale)
		previous := 0
		for _, size := range []AlertSize{AlertSizeXSmall, AlertSizeSmall, AlertSizeMedium, AlertSizeLarge} {
			a.Size(size)
			h.Frame()
			n, _ := semanticNode(h, "alert:info")
			if n.Desc.Bounds.Dy() <= previous {
				t.Fatal("size not increasing", size, n.Desc.Bounds)
			}
			previous = n.Desc.Bounds.Dy()
		}
		normal, _ := semanticNode(h, "alert:info")
		a.Banner(true)
		h.Frame()
		banner, _ := semanticNode(h, "alert:info")
		if banner.Desc.Bounds.Dy() >= normal.Desc.Bounds.Dy() || banner.Desc.Bounds.Dx() != 300*scale {
			t.Fatal("banner layout")
		}
		x := bounds(h, "Message").Min.X
		a.Icon(IconNone)
		h.Frame()
		if bounds(h, "Message").Min.X >= x {
			t.Fatal("hidden icon kept space")
		}
		a.Description("")
		h.Frame()
		if !shown(h, "Title") {
			t.Fatal("empty banner lost fallback")
		}
		actions, closes := 0, 0
		disabled := false
		a.Content(Button("Retry", func() { actions++ })).OnClose(func() { closes++ }).Icon(IconCalendar)
		h = renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(a.Render(cx)) }), 200, scale)
		click(t, h, "Retry")
		if actions != 1 || !a.Visible() {
			t.Fatal("body action closed alert")
		}
		disabled = true
		h.Frame()
		click(t, h, "Retry")
		click(t, h, "关闭 Title")
		if actions != 1 || closes != 0 || !a.Visible() {
			t.Fatal("ancestor disabled")
		}
		disabled = false
		h.Frame()
		click(t, h, "关闭 Title")
		if closes != 1 || a.Visible() {
			t.Fatal("rich alert close")
		}
		a.Content(nil).Description("Restored").Banner(false).Size(AlertSizeMedium)
		a.SetVisible(true)
		h.Frame()
		if !shown(h, "Restored") || shown(h, "Retry") || closes != 1 {
			t.Fatal("reset content")
		}
	}
}
