package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitEmptySnapshot(t *testing.T) {
	w := openTest(t, Options{Content: el.Embed(kit.Empty("暂无订单").Description("创建订单后将在这里显示"))})
	if e := element(t, w, "暂无订单"); e.Role != "text" {
		t.Fatalf("invalid empty role: %+v", e)
	}
	if e := element(t, w, "创建订单后将在这里显示"); e.Role != "text" {
		t.Fatal("missing description")
	}
}

func TestKitEmptyMediaSnapshot(t *testing.T) {
	calls := 0
	v := kit.Empty("No members").Media(kit.Avatar("Alex").Size(48)).Action(kit.Button("Invite", func() { calls++ }))
	w := openTest(t, kitPage(v))
	media := element(t, w, "Alex")
	if media.Width != 48 || media.Height != 48 {
		t.Fatal("media size lost", media)
	}
	w.click(element(t, w, "Invite").center())
	if calls != 1 {
		t.Fatal("action did not fire")
	}
	v.Media(nil).Icon(kit.IconNone)
	w.render()
	if roleOfName(w, "Alex") != "" {
		t.Fatal("stale avatar")
	}
}
