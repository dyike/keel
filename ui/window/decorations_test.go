package window

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/gio/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/theme"
)

// On a Wayland compositor that leaves decorations to the app, Keel draws
// the title bar with its own theme and fonts, puts the content below it,
// and its buttons act on the window.
func TestKeelDrawsTitleBarWithoutPlatformDecorations(t *testing.T) {
	saved := clientDecorations
	clientDecorations = true
	t.Cleanup(func() { clientDecorations = saved })

	if askDecorations(false) || askDecorations(true) {
		t.Fatal("Linux windows must turn Gio's own title bar off")
	}
	w := openTest(t, Options{Title: "订单 Orders", Width: 400, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(8).Child(el.Text("内容 content"))
	}))})
	loop.Lock()
	if !w.updateDecorations(false, false) {
		t.Fatal("no platform decorations should make Keel draw the bar")
	}
	loop.Unlock()
	w.render()
	if y := element(t, w, "内容 content").Y; y < titleBarHeight {
		t.Fatalf("content at y=%v overlaps the %vdp title bar", y, titleBarHeight)
	}
	shot, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(shot))
	if err != nil {
		t.Fatal(err)
	}
	if path := os.Getenv("KEEL_TITLEBAR_PNG"); path != "" {
		os.WriteFile(path, shot, 0o644)
	}
	r, g, b, _ := img.At(img.Bounds().Dx()/2, 4).RGBA()
	if s := theme.Subtle; uint8(r>>8) != s.R || uint8(g>>8) != s.G || uint8(b>>8) != s.B {
		t.Fatalf("title bar is not the theme's Subtle color: %v %v %v", r>>8, g>>8, b>>8)
	}

	// Gio's decorations put close last, at the right edge.
	w.click(f32.Pt(float32(img.Bounds().Dx())-12, titleBarHeight/2))
	if !w.Closed() {
		t.Fatal("the close button did not close the window")
	}

	// A platform title bar, or full screen, hands it back.
	w2 := openTest(t, Options{Title: "x", Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("y") }))})
	loop.Lock()
	w2.updateDecorations(false, false)
	if !w2.updateDecorations(true, false) || w2.drawsTitle {
		t.Fatal("server-side decorations did not hand the bar back")
	}
	w2.updateDecorations(false, true)
	if w2.drawsTitle {
		t.Fatal("a full-screen window drew a title bar")
	}
	loop.Unlock()

	// Frameless windows draw their own (kit.TitleBar) and never get Keel's.
	w3 := openTest(t, Options{Frameless: true, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("z") }))})
	loop.Lock()
	w3.updateDecorations(false, false)
	loop.Unlock()
	if w3.drawsTitle {
		t.Fatal("a frameless window got a second title bar")
	}
}

func TestFallbackTitleBarFitsScaledText(t *testing.T) {
	for _, density := range []float32{1, 1.25, 1.5, 2} {
		for _, textScale := range []float32{1, 1.6364, 2} {
			for _, title := range []string{"Keel Native Capabilities", "Keel 原生能力"} {
				var ops op.Ops
				gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: density, PxPerSp: density * textScale}, Constraints: layout.Exact(image.Pt(900, 500))}
				// Measure an unclipped single line independently of the decoration box.
				label := material.Label(theme.Material, theme.Material.TextSize*theme.TextMd/theme.TextBody, title)
				label.Font.Weight = 600
				label.MaxLines = 1
				textContext := gtx
				textContext.Constraints.Min = image.Point{}
				record := op.Record(&ops)
				textHeight := label.Layout(textContext).Size.Y
				record.Stop()
				w := &Window{opts: Options{Title: title}, focused: true, drawsTitle: true}
				content, pop := w.belowTitleBar(gtx)
				pop()
				height := gtx.Constraints.Max.Y - content.Constraints.Max.Y
				minimum := textHeight + 2*gtx.Dp(10) + gtx.Dp(1)
				if height < minimum {
					t.Fatalf("density=%g text=%g title=%q: bar=%d, needs at least %d", density, textScale, title, height, minimum)
				}
				if content.Constraints.Min.Y > content.Constraints.Max.Y {
					t.Fatal("invalid content constraints")
				}
			}
		}
	}
}
