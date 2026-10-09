package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestSettingsNarrowLayoutKeyboardAndState(t *testing.T) {
	for _, scale := range []int{1, 2} {
		field := Input("")
		items := []SettingItem{{Label: "姓名 Name 123", Description: "用于公开显示的名称", Control: field}}
		v := Settings().Section("通用", IconUser, items...).Section("通知", IconInbox, SettingItem{Label: "邮件提醒", Control: Switch("", true)})
		items[0].Label = "external mutation"
		root := el.Root(v)
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints = layout.Exact(image.Pt(320*scale, 640*scale))
			root.Layout(gtx)
		})
		for _, name := range []string{"姓名 Name 123", "用于公开显示的名称", "搜索设置"} {
			b := bounds(h, name)
			if b.Empty() || b.Min.X < 0 || b.Max.X > 320*scale || b.Max.Y > 640*scale {
				t.Fatal(name, b)
			}
		}
		// Tab reaches both section buttons and Return selects the second.
		h.Router.MoveFocus(key.FocusForward)
		h.Frame()
		h.Router.MoveFocus(key.FocusForward)
		h.Frame()
		h.Key(key.NameReturn, 0)
		h.Frame()
		if v.Value() != "通知" || !shown(h, "邮件提醒") {
			t.Fatal("section keyboard navigation", v.Value())
		}
		click(t, h, "通用")
		h.Frame()
		clickClass(t, h, "Editor", "姓名 Name 123")
		h.Type("张三 Ada 123")
		h.Frame()
		click(t, h, "通知")
		h.Frame()
		click(t, h, "通用")
		h.Frame()
		if field.Value() != "张三 Ada 123" {
			t.Fatal("section switch lost input", field.Value())
		}
		clickClass(t, h, "Editor", "搜索设置")
		h.Type("邮件")
		h.Frame()
		if !shown(h, "邮件提醒") || shown(h, "姓名 Name 123") {
			t.Fatal("cross-section search")
		}
	}
}
