package window

import (
	"bytes"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestKitStatusMarkerAgentAndMotion(t *testing.T) {
	old := theme.Current()
	motion := theme.ReducedMotion
	defer core.Update(func() { theme.Apply(old); theme.SetReducedMotion(motion) })
	for _, dark := range []bool{false, true} {
		p := theme.Light()
		if dark {
			p = theme.Dark()
		}
		core.Update(func() { theme.Apply(p); theme.SetReducedMotion(false) })
		now := time.Unix(1000, 0)
		calls := 0
		v := kit.StatusMarker("Loading messages").ID("progress").Role("status").Variant(kit.StatusMarkerSeparator).Loading(true).LoadingStyle(kit.StatusMarkerLoadingStyleShimmer).Content(kit.Button("Retry", func() { calls++ }))
		root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(420)).P(16).Bg(theme.Bg).Child(v.Render(cx)) }))
		w := openTest(t, Options{Width: 420, Height: 120, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
		capture := func() []byte {
			t.Helper()
			b, err := w.screenshot()
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		first := capture()
		now = now.Add(600 * time.Millisecond)
		mid := capture()
		if bytes.Equal(first, mid) {
			t.Fatal("text shimmer did not animate", dark)
		}
		if element(t, w, "Loading messages").Role != "text" {
			t.Fatal("readable label")
		}
		w.click(element(t, w, "Retry").center())
		if calls != 1 {
			t.Fatal("child action")
		}
		core.Update(func() { theme.SetReducedMotion(true) })
		still := capture()
		now = now.Add(700 * time.Millisecond)
		if !bytes.Equal(still, capture()) {
			t.Fatal("reduced motion moved", dark)
		}
		core.Update(func() { v.SetText(""); v.Content(kit.Label("Rich status")); theme.SetReducedMotion(false) })
		rich := capture()
		now = now.Add(700 * time.Millisecond)
		if bytes.Equal(rich, capture()) {
			t.Fatal("rich content pulse missing")
		}
		core.Update(func() { v.Loading(false) })
		stopped := capture()
		now = now.Add(700 * time.Millisecond)
		if !bytes.Equal(stopped, capture()) {
			t.Fatal("loading false animated")
		}
	}
}
