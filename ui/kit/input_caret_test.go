package kit

import (
	"image"
	"testing"
	"time"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// The caret is 1.5dp wide and centered where Chinese text is centered: one
// spanning 国 to g sat 1.5dp low and reached below the characters.
func TestCaretWidthAndCentering(t *testing.T) {
	const scale = 2
	in := Input("")
	in.SetValue("国国")
	W, H := 200*scale, 60*scale
	gpu, err := headless.NewWindow(W, H)
	if err != nil {
		t.Skip("no headless GPU:", err)
	}
	defer gpu.Release()
	root := el.Root(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(10).W(el.Dp(200)).Items(el.Stretch).Child(in.Render(cx))
	}))
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	now := time.Unix(1000, 0) // a fixed clock keeps the caret in its visible phase
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
		gtx.Constraints = layout.Exact(image.Pt(W, H))
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
	})
	h.Frame()
	h.Click(150*scale, 28*scale) // past the text: the caret goes to the end
	h.Frame()
	h.Frame()
	dark := func(x, y int) bool { c := img.RGBAAt(x, y); return (int(c.R)+int(c.G)+int(c.B))/3 < 0x80 }
	inked := func(x int) bool {
		for y := 0; y < H; y++ {
			if dark(x, y) {
				return true
			}
		}
		return false
	}
	// The caret is the rightmost ink; the text is to its left.
	x1 := W - 1
	for x1 > 0 && !inked(x1) {
		x1--
	}
	x0 := x1
	for x0 > 0 && inked(x0-1) {
		x0--
	}
	extent := func(a, b int) (int, int) {
		top, bottom := -1, -1
		for y := 0; y < H; y++ {
			for x := a; x <= b; x++ {
				if dark(x, y) {
					if top < 0 {
						top = y
					}
					bottom = y + 1
					break
				}
			}
		}
		return top, bottom
	}
	ct, cb := extent(x0, x1)
	tt, tb := extent(0, x0-2)
	if w := x1 - x0 + 1; w < 3 {
		t.Fatalf("caret %dpx wide at 2x, want 1.5dp", w)
	}
	if d := (ct + cb) - (tt + tb); d < -2 || d > 2 {
		t.Fatalf("caret y %d..%d, text y %d..%d: centers %.1fpx apart", ct, cb, tt, tb, float64(d)/2)
	}
	if ct >= tt || cb <= tb {
		t.Fatalf("caret y %d..%d does not span the text y %d..%d", ct, cb, tt, tb)
	}
}
