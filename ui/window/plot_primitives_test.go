package window

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"

	"gioui.org/f32"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/plot"
	"github.com/dyike/keel/ui/theme"
)

func TestPlotPrimitivesPaintClippedCanvas(t *testing.T) {
	red := color.NRGBA{R: 230, G: 30, B: 60, A: 255}
	cases := map[string]func(plot.Canvas){
		"bar": func(c plot.Canvas) {
			plot.Bar{Rect: plot.Rectangle{Min: f32.Pt(-10, -10), Max: f32.Pt(70, 45)}, Fill: red, Radius: 5}.Paint(c)
		},
		"line": func(c plot.Canvas) {
			plot.Line{Points: []f32.Point{f32.Pt(0, 50), f32.Pt(40, 10), f32.Pt(80, 40), f32.Pt(150, 10)}, Stroke: red, Width: 3, DotRadius: 4, Curve: plot.CurveSmooth}.Paint(c)
		},
		"area": func(c plot.Canvas) {
			plot.Area{Upper: []f32.Point{f32.Pt(0, 40), f32.Pt(60, 10), f32.Pt(120, 30)}, Lower: []f32.Point{f32.Pt(0, 70), f32.Pt(60, 70), f32.Pt(120, 70)}, Fill: red}.Paint(c)
		},
		"arc": func(c plot.Canvas) {
			plot.Arc{Center: f32.Pt(70, 40), Inner: 15, Outer: 35, Start: 0, End: 2 * math.Pi, Fill: red}.Paint(c)
		},
		"dot": func(c plot.Canvas) { plot.Dot{At: f32.Pt(70, 40), Radius: 12, Fill: red}.Paint(c) },
		"cross": func(c plot.Canvas) {
			plot.CrossLine{At: f32.Pt(70, 40), Horizontal: true, Vertical: true, Stroke: red, Width: 2}.Paint(c)
		},
		"axis": func(c plot.Canvas) {
			plot.Axis{Side: plot.AxisBottom, At: 45, From: 0, To: 140, TickSize: 5, Width: 2, Stroke: red, TextColor: red, Ticks: []plot.AxisText{{Position: 30, Text: "A"}, {Position: 100, Text: "B"}}}.Paint(c)
		},
	}
	for _, dark := range []bool{false, true} {
		bg := color.NRGBA{R: 250, G: 250, B: 250, A: 255}
		if dark {
			bg = color.NRGBA{R: 20, G: 20, B: 20, A: 255}
		}
		for name, draw := range cases {
			t.Run(name+map[bool]string{false: "-light", true: "-dark"}[dark], func(t *testing.T) {
				bounds := image.Rect(20, 10, 180, 100)
				w := openTest(t, Options{Width: 200, Height: 120, Content: plotPrimitiveRoot{core.Func(func(gtx core.C) core.D {
					paint.Fill(gtx.Ops, bg)
					draw(plot.Canvas{Context: gtx, Bounds: bounds})
					return core.D{Size: gtx.Constraints.Max}
				})}})
				data, err := w.screenshot()
				if err != nil {
					t.Fatal(err)
				}
				im, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				colored := 0
				for y := 0; y < 120; y++ {
					for x := 0; x < 200; x++ {
						r, g, _, _ := im.At(x, y).RGBA()
						if r > g+4000 {
							colored++
							if !image.Pt(x, y).In(bounds) {
								t.Fatal("escaped clip", x, y)
							}
						}
					}
				}
				if colored < 30 {
					t.Fatal("missing primitive", colored)
				}
				if name == "arc" {
					c := color.NRGBAModel.Convert(im.At(90, 50)).(color.NRGBA)
					if plotPixelAbs(int(c.R)-int(bg.R)) > 5 || plotPixelAbs(int(c.G)-int(bg.G)) > 5 || plotPixelAbs(int(c.B)-int(bg.B)) > 5 {
						t.Fatal("donut hole filled", c)
					}
				}
			})
		}
	}
}

func TestPlotInvalidGeometryDoesNotPaint(t *testing.T) {
	bg := theme.Surface
	w := openTest(t, Options{Width: 100, Height: 100, Content: plotPrimitiveRoot{core.Func(func(gtx core.C) core.D {
		paint.Fill(gtx.Ops, bg)
		c := plot.Canvas{Context: gtx, Bounds: image.Rect(0, 0, 100, 100)}
		red := color.NRGBA{R: 255, A: 255}
		plot.Bar{Rect: plot.Rectangle{Min: f32.Pt(float32(math.NaN()), 0), Max: f32.Pt(90, 90)}, Fill: red}.Paint(c)
		plot.Arc{Center: f32.Pt(50, 50), Inner: 40, Outer: 20, End: 6, Fill: red}.Paint(c)
		plot.Line{Points: []f32.Point{f32.Pt(10, 10), f32.Pt(float32(math.Inf(1)), 20), f32.Pt(90, 90)}, Width: 3, Stroke: red}.Paint(c)
		return core.D{Size: gtx.Constraints.Max}
	})}})
	data, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA); c != bg {
				t.Fatal("invalid geometry painted", x, y, c)
			}
		}
	}
}

type plotPrimitiveRoot struct{ core.Widget }

func (plotPrimitiveRoot) FillsWindow() bool { return true }

func plotPixelAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
