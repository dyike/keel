package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestRisingGlowSnapshotAndPixels(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(false)
	defer theme.SetReducedMotion(old)
	now := time.Unix(100, 0)
	glow := kit.RisingGlow().Width(200).Height(100).ParticleCount(80).Color(color.NRGBA{B: 255, A: 255})
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Bg(color.NRGBA{A: 255}).Child(el.Text("keel"), glow.Render(cx), kit.Button("Pause", func() { glow.Enabled(false) }).Variant(kit.ButtonSecondary).Render(cx))
	}))
	w := openTest(t, Options{Width: 300, Height: 220, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
	first, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Second)
	mid, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, mid) {
		t.Fatal("particles did not animate")
	}
	im, err := png.Decode(bytes.NewReader(mid))
	if err != nil {
		t.Fatal(err)
	}
	blue := 0
	heading := element(t, w, "keel")
	top := heading.Y + heading.Height
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
			if int(c.B)-int(c.R) > 30 && int(c.B)-int(c.G) > 30 {
				blue++
				if x < heading.X || x >= heading.X+200 || y < top || y >= top+100 {
					t.Fatalf("particle escaped field at %d,%d; top=%d", x, y, top)
				}
			}
		}
	}
	if blue < 20 {
		t.Fatalf("no visible particles: %d", blue)
	}
	count := 0
	for _, e := range w.snapshot() {
		if e.Name == "keel" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("heading duplicated: %d", count)
	}
	w.click(element(t, w, "Pause").center())
	paused, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	after, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(paused, after) {
		t.Fatal("paused pixels changed")
	}
}
