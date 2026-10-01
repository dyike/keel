package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitAlertSnapshot(t *testing.T) {
	a := kit.Alert("保存成功", "订单 SO-123 已保存").Tone(kit.Success)
	w := openTest(t, Options{Content: el.Embed(a)})
	e := element(t, w, "保存成功")
	if e.Role != "alert" || e.Value != "success" {
		t.Fatalf("alert semantics: %+v", e)
	}
	if e := element(t, w, "订单 SO-123 已保存"); e.Role != "text" {
		t.Fatal("missing description")
	}
	a.SetTitle("保存失败")
	a.Tone(kit.Danger)
	if e := element(t, w, "保存失败"); e.Value != "danger" {
		t.Fatal("stale alert")
	}
}

func TestKitAlertKeyboardDismiss(t *testing.T) {
	n := 0
	a := kit.Alert("提示").OnClose(func() { n++ })
	w := openTest(t, Options{Content: el.Embed(a)})
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if a.Visible() || n != 1 {
		t.Fatal("keyboard close failed")
	}
}
