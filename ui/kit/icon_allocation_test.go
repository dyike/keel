package kit

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestIconRenderReusesImmutableColorSnapshot(t *testing.T) {
	i := Icon(IconHome)
	i.Render(nil)
	first := i.renderer
	if first.color != theme.Text {
		t.Fatal("default color changed")
	}
	if got := testing.AllocsPerRun(20, func() { iconRenderSink = i.Render(nil) }); got > 2 {
		t.Fatalf("warm Render allocated %.0f times; only independent element/node allocations expected", got)
	}
	if i.renderer != first {
		t.Fatal("unchanged renderer replaced")
	}
	i.Color(color.NRGBA{}).Render(nil)
	if i.renderer == first || i.renderer.color != (color.NRGBA{}) || first.color != theme.Text {
		t.Fatal("saved color snapshot overwritten")
	}
	copy := *i
	copy.Size(27).Render(nil)
	if copy.renderer.view != &copy || copy.renderer == i.renderer || i.size == copy.size {
		t.Fatal("copied IconView retained original renderer")
	}
	// Other view settings remain live, as with the old closure.
	i.Size(25).Rotate(90)
	if first.view.size != 25 || first.view.rotation != 90 {
		t.Fatal("view settings unexpectedly snapshotted")
	}
}

func TestIconCachedElementKeepsColorPixels(t *testing.T) {
	for _, scale := range []int{1, 2} {
		i := Icon(IconHome).Size(24)
		red, blue := color.NRGBA{R: 255, A: 255}, color.NRGBA{B: 255, A: 255}
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			first := cx.Cache("red", func() el.Element { return i.Color(red).Render(cx) })
			return el.Div().Row().Items(el.Start).Child(first, i.Color(blue).Render(cx))
		}))
		gpu, err := headless.NewWindow(48*scale, 24*scale)
		if err != nil {
			t.Fatal(err)
		}
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(48*scale, 24*scale)
			root.Layout(gtx)
			if err := gpu.Frame(gtx.Ops); err != nil {
				t.Fatal(err)
			}
			img := image.NewRGBA(image.Rect(0, 0, 48*scale, 24*scale))
			if err := gpu.Screenshot(img); err != nil {
				t.Fatal(err)
			}
			left, right := img.RGBAAt(12*scale, 12*scale), img.RGBAAt(36*scale, 12*scale)
			if left != (color.RGBA{R: 255, A: 255}) || right != (color.RGBA{B: 255, A: 255}) {
				t.Fatalf("scale%d cached colors: %v %v", scale, left, right)
			}
		})
		for range 4 {
			h.Frame()
		}
		gpu.Release()
	}
}
