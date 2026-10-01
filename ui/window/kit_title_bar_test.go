package window

import (
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestKitTitleBarFramelessControls(t *testing.T) {
	bar := kit.TitleBar("编辑器").Trailing(kit.Button("分享", nil))
	w := openTest(t, Options{Width: 640, Height: 300, Frameless: true, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Items(el.Stretch).Child(bar.Render(cx))
	}))})
	if got := roleOfName(w, "编辑器"); got != "banner" {
		t.Fatalf("title bar: %q", got)
	}
	for _, name := range []string{"关闭", "最小化", "最大化", "分享"} {
		if roleOfName(w, name) != "button" {
			t.Errorf("missing button %q", name)
		}
	}
	w.click(element(t, w, "最大化").center())
	if !w.Maximized() || roleOfName(w, "还原") != "button" {
		t.Fatal("maximize did not toggle")
	}
}

func TestKitTitleBarWithSystemFrame(t *testing.T) {
	bar := kit.TitleBar("编辑器")
	w := openTest(t, Options{Width: 640, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Items(el.Stretch).Child(bar.Render(cx))
	}))})
	if roleOfName(w, "关闭") != "" {
		t.Fatal("a window with a system title bar should not get window buttons")
	}
}
