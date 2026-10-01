package widget

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/gpu/headless"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/layout"
)

func TestBadgeChildStaysCenteredInRow(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		for _, mode := range []string{"count", "dot", "icon"} {
			for _, size := range []ComponentSize{Small, Medium, Large} {
				t.Run(fmt.Sprintf("%s/%d/%gx", mode, size, scale), func(t *testing.T) {
					badge := Badge(4).Size(size).Child(Button("通知", nil).Secondary())
					if mode == "dot" {
						badge.Dot()
					} else if mode == "icon" {
						badge.Icon(Icon(IconCheck))
					}
					row := layout.Row(badge, Button("增加未读", nil).Secondary())
					h := uitest.NewFunc(func(gtx core.C) {
						gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
						row.Layout(gtx)
					})
					for _, count := range []int{4, 150, 0, 1} {
						badge.SetCount(count)
						h.Frame()
						a, b := namedBounds(h, "通知"), namedBounds(h, "增加未读")
						if a.Empty() || b.Empty() {
							t.Fatal("missing button bounds")
						}
						if delta := a.Min.Y + a.Max.Y - b.Min.Y - b.Max.Y; delta < -1 || delta > 1 {
							t.Fatalf("count %d: button centers differ: %v vs %v", count, a, b)
						}
					}
				})
			}
		}
	}
}

// Check the rendered digits against the pill, rather than comparing metrics
// from a font probe that can select a different fallback face.
func TestBadgeCountInkCentered(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, size := range []ComponentSize{Small, Medium, Large} {
			for _, count := range []int{1, 4, 8, 10, 150} {
				t.Run(fmt.Sprintf("%d/%d/%dx", count, size, scale), func(t *testing.T) {
					viewport := image.Pt(100*scale, 60*scale)
					win, err := headless.NewWindow(viewport.X, viewport.Y)
					if err != nil {
						t.Fatal(err)
					}
					defer win.Release()
					var ops op.Ops
					gtx := core.C{Ops: &ops, Now: time.Now(), Constraints: giolayout.Constraints{Max: viewport}, Metric: unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}}
					paint.Fill(&ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
					Badge(count).Size(size).Color(color.NRGBA{B: 255, A: 255}, color.NRGBA{G: 255, A: 255}).Layout(gtx)
					if err := win.Frame(&ops); err != nil {
						t.Fatal(err)
					}
					img := image.NewRGBA(image.Rectangle{Max: viewport})
					if err := win.Screenshot(img); err != nil {
						t.Fatal(err)
					}
					pillTop, pillBottom, inkTop, inkBottom := viewport.Y, -1, viewport.Y, -1
					for y := 0; y < viewport.Y; y++ {
						for x := 0; x < viewport.X; x++ {
							p := img.RGBAAt(x, y)
							if p.B > 180 && p.R < 100 && p.G < 100 {
								pillTop, pillBottom = min(pillTop, y), max(pillBottom, y)
							}
							if int(p.G) > int(p.R)+30 {
								inkTop, inkBottom = min(inkTop, y), max(inkBottom, y)
							}
						}
					}
					if pillBottom < 0 || inkBottom < 0 {
						t.Fatal("missing pill or count pixels")
					}
					if delta := inkTop + inkBottom - pillTop - pillBottom; delta < -2 || delta > 2 {
						t.Fatalf("painted centers differ: count y=%d..%d pill y=%d..%d", inkTop, inkBottom, pillTop, pillBottom)
					}
				})
			}
		}
	}
}

func TestBadgeVisibilityAndChildInteraction(t *testing.T) {
	clicks := 0
	child := Button("通知", func() { clicks++ })
	badge := Badge(150).Child(child)
	h := uitest.New(badge)
	clickNamed(t, h, "通知")
	if clicks != 1 {
		t.Fatal("badge blocked child")
	}
	if badge.label() != "99+" {
		t.Fatal("count overflow not formatted")
	}
	badge.Max(9)
	if badge.label() != "9+" {
		t.Fatal("custom max ignored")
	}
	badge.SetCount(0)
	h.Frame()
	if !namedBounds(h, "150").Empty() {
		t.Fatal("zero badge still exposed")
	}
	clickNamed(t, h, "通知")
	if clicks != 2 {
		t.Fatal("hidden badge blocked child")
	}
	var plain, hidden image.Point
	uitest.NewFunc(func(gtx core.C) { plain = child.Layout(gtx).Size })
	uitest.NewFunc(func(gtx core.C) { hidden = badge.Layout(gtx).Size })
	if hidden != plain {
		t.Fatalf("hidden badge reserves space: %v != %v", hidden, plain)
	}
	empty := Badge(-1)
	var dims core.D
	uitest.NewFunc(func(gtx core.C) { dims = empty.Layout(gtx) })
	if dims.Size != (image.Point{}) {
		t.Fatal("negative count should be hidden")
	}
	dot := Badge(1).Dot()
	uitest.NewFunc(func(gtx core.C) { dims = dot.Layout(gtx) })
	if dims.Size != image.Pt(8, 8) {
		t.Fatalf("unexpected dot size %v", dims.Size)
	}
	// The badge and child must both stay inside the allocation, even when the
	// parent is too small to accommodate the full overhang.
	for _, width := range []int{0, 4, 20, 100} {
		for _, height := range []int{0, 1, 4, 5, 20, 40} {
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Constraints.Max = image.Pt(width, height)
				dims = Badge(150).Child(Text("x")).Layout(gtx)
			})
			if dims.Size.X > width || dims.Size.Y > height {
				t.Fatalf("badge exceeded constraints %dx%d: %v", width, height, dims.Size)
			}
			for _, name := range []string{"150", "x"} {
				if bounds := namedBounds(h, name); !bounds.Empty() && !bounds.In(image.Rectangle{Max: dims.Size}) {
					t.Fatalf("%s exceeds allocation %v: %v", name, dims.Size, bounds)
				}
			}
		}
	}
}
