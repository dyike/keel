package el

import (
	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget"
	"github.com/dyike/keel/gio/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
	"image"
	"image/color"
	"testing"
)

func TestInputKeepsDescenderInk(t *testing.T) {
	for _, face := range []font.Typeface{theme.Material.Face, "Go"} {
		old := theme.Material.Face
		theme.Material.Face = face
		for _, scale := range []int{1, 2} {
			for _, size := range []float32{12, 15, 20, 32} {
				for _, value := range []string{"gypqj", "国Agyp", "Égj"} {
					value := value
					got := redInkScaled(t, Input().Bind(&value).TextSize(size).TextColor(color.NRGBA{R: 255, A: 255}).W(Dp(280)).Border(0, color.NRGBA{}).P(0).MinH(Auto), scale)
					var editor widget.Editor
					editor.SingleLine = true
					editor.SetText(value)
					want := redInkScaled(t, Widget(core.Func(func(gtx core.C) core.D {
						style := material.Editor(theme.Material, &editor, "")
						style.TextSize = unit.Sp(size)
						style.Color = color.NRGBA{R: 255, A: 255}
						return style.Layout(gtx)
					})).W(Dp(280)).H(Dp(110)), scale)
					if got < want*97/100 {
						t.Errorf("face=%s size=%g text=%s ink=%d want=%d", face, size, value, got, want)
					}
				}
			}
		}
		theme.Material.Face = old
	}
}

func TestSingleLineMonoInputAlignsWithLabel(t *testing.T) {
	inkCenterY := func(t *testing.T, elem Element, width, height, scale int) float64 {
		t.Helper()
		gpu, err := headless.NewWindow(width*scale, height*scale)
		if err != nil {
			t.Fatal(err)
		}
		defer gpu.Release()
		root := Root(ViewFunc(func(cx *Context) Element {
			return Div().Bg(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}).Items(Start).Child(elem)
		}))
		img := image.NewRGBA(image.Rect(0, 0, width*scale, height*scale))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(width*scale, height*scale)
			root.Layout(gtx)
			if err := gpu.Frame(gtx.Ops); err != nil {
				t.Fatal(err)
			}
			if err := gpu.Screenshot(img); err != nil {
				t.Fatal(err)
			}
		})
		h.Frame()
		minY, maxY := img.Bounds().Max.Y, img.Bounds().Min.Y-1
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				p := img.RGBAAt(x, y)
				if p.R > 180 && p.G < 80 && p.B < 80 {
					if y < minY {
						minY = y
					}
					if y > maxY {
						maxY = y
					}
				}
			}
		}
		if maxY < minY {
			t.Fatal("missing rendered ink")
		}
		return float64(minY+maxY) / 2
	}
	red := color.NRGBA{R: 255, A: 255}
	for _, scale := range []int{1, 2} {
		labelY := inkCenterY(t, Div().Row().Items(Center).H(Dp(44)).W(Dp(120)).Child(Text("Host").TextSize(15).TextColor(red)), 120, 44, scale)
		for _, textVal := range []string{"mac.local", "192.168.1.10", "中文测试"} {
			v := textVal
			inputY := inkCenterY(t, Div().Row().Items(Center).H(Dp(44)).W(Dp(240)).Child(Input().Mono().Bind(&v).TextSize(15).TextColor(red).Border(0, color.NRGBA{}).Bg(color.NRGBA{}).H(Dp(42))), 240, 44, scale)
			diff := labelY - inputY
			if diff < 0 {
				diff = -diff
			}
			if diff > float64(scale)*1.5 {
				t.Errorf("scale=%dx text=%q: mono input center %.1f misaligned with label %.1f (diff=%.1f)", scale, textVal, inputY, labelY, diff)
			}
		}
	}
}
