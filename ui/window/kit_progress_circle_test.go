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

func TestProgressCircleSnapshotAndAnimation(t *testing.T) {
	old := theme.ReducedMotion
	defer core.Update(func() { theme.SetReducedMotion(old) })
	now := time.Unix(0, 0)
	p := kit.ProgressCircle("Import").Size(64).Child(text("Center"))
	r := el.Embed(p)
	w := openTest(t, Options{Width: 180, Height: 180, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return r.Layout(gtx) })})
	capture := func() []byte {
		t.Helper()
		b, e := w.screenshot()
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	core.Update(func() { theme.SetReducedMotion(true) })
	var before []byte
	for _, value := range []float32{0, .25, 1} {
		core.Update(func() { p.SetValue(value) })
		b := capture()
		if before != nil && bytes.Equal(before, b) {
			t.Fatal("determinate progress did not change pixels")
		}
		before = b
		if e := element(t, w, "Import"); e.Role != "progressbar" {
			t.Fatalf("role %q", e.Role)
		}
	}
	core.Update(func() { p.SetIndeterminate(true); theme.SetReducedMotion(false) })
	a := capture()
	now = now.Add(250 * time.Millisecond)
	b := capture()
	if bytes.Equal(a, b) {
		t.Fatal("indeterminate ring did not rotate")
	}
	core.Update(func() { theme.SetReducedMotion(true) })
	a = capture()
	now = now.Add(250 * time.Millisecond)
	b = capture()
	if !bytes.Equal(a, b) {
		t.Fatal("reduced motion animated")
	}
}
