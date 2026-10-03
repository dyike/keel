package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitGroupBoxSnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.GroupBox("账户").Child(kit.Tag("已验证")))})
	if e := element(t, w, "账户"); e.Role != "group" {
		t.Fatalf("invalid group: %+v", e)
	}
	if e := element(t, w, "已验证"); e.Role != "tag" {
		t.Fatal("group absorbed child")
	}
}

func TestKitGroupBoxFooterOutsideContentStyle(t *testing.T) {
	calls := 0
	body := kit.Button("Body action", func() { t.Error("disabled body activated") })
	footer := kit.Button("Footer action", func() { calls++ })
	v := kit.GroupBox("Settings").Variant(kit.GroupBoxFill).Child(body).Footer(footer).ContentStyle(func(e *el.DivEl) { e.Disabled(true).P(24) })
	w := openTest(t, kitPage(v))
	b := element(t, w, "Body action")
	f := element(t, w, "Footer action")
	if !b.Disabled || f.Disabled || f.Y <= b.Y+b.Height {
		t.Fatalf("footer/body isolation: %+v %+v", b, f)
	}
	w.click(b.center())
	w.click(f.center())
	if calls != 1 {
		t.Fatal("footer action")
	}
	v.ContentStyle(nil).Footer(nil)
	if e := element(t, w, "Body action"); e.Disabled {
		t.Fatal("style reset")
	}
	if roleOfName(w, "Settings") != "group" || roleOfName(w, "Footer action") != "" {
		t.Fatal("group semantics")
	}
}
