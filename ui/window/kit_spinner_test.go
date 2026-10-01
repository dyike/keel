package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"testing"
	"time"
)

func TestSpinnerFrameTimeAndReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer core.Update(func() { theme.SetReducedMotion(old) })
	now := time.Unix(1000, 0)
	root := el.Embed(kit.Spinner().Size(32).Label("加载中"))
	w := openTest(t, Options{Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
	capture := func() []byte {
		t.Helper()
		b, e := w.screenshot()
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	core.Update(func() { theme.SetReducedMotion(false) })
	a := capture()
	now = now.Add(250 * time.Millisecond)
	b := capture()
	if bytes.Equal(a, b) {
		t.Fatal("frame time did not animate")
	}
	core.Update(func() { theme.SetReducedMotion(true) })
	a = capture()
	now = now.Add(250 * time.Millisecond)
	b = capture()
	if !bytes.Equal(a, b) {
		t.Fatal("reduced motion still animated")
	}
	e := element(t, w, "加载中")
	if e.Role != "progressbar" || e.Value != "indeterminate" {
		t.Fatal(e)
	}
}
