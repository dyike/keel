package window

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"slices"
	"testing"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestShimmerTextAnimationAndRestoration(t *testing.T) {
	old := theme.ReducedMotion
	defer core.Update(func() { theme.SetReducedMotion(old) })
	core.Update(func() { theme.SetReducedMotion(false) })
	now := time.Unix(1000, 0)
	v := kit.ShimmerText("MMMM MMMM").Size(32).Color(color.NRGBA{R: 80, A: 255}).Highlight(color.NRGBA{R: 255, A: 255}).Duration(time.Second)
	root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Bg(color.NRGBA{A: 255}).Child(v.Render(cx)) }))
	w := openTest(t, Options{Width: 300, Height: 100, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
	capture := func() image.Image {
		t.Helper()
		b, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return im
	}
	first := capture()
	now = now.Add(400 * time.Millisecond)
	mid := capture()
	changes := 0
	for y := 0; y < first.Bounds().Max.Y; y++ {
		for x := 0; x < first.Bounds().Max.X; x++ {
			a := color.NRGBAModel.Convert(first.At(x, y)).(color.NRGBA)
			b := color.NRGBAModel.Convert(mid.At(x, y)).(color.NRGBA)
			if a != b {
				changes++
				if a.R == 0 && a.G == 0 && a.B == 0 && b.R > 15 {
					t.Fatal("shimmer painted outside glyph", x, y)
				}
			}
		}
	}
	if changes < 20 {
		t.Fatal("highlight did not move", changes)
	}
	count := 0
	for _, e := range w.snapshot() {
		if e.Name == "MMMM MMMM" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("text semantics duplicated", count)
	}
	same := func(a, b image.Image) bool {
		for y := 0; y < a.Bounds().Max.Y; y++ {
			for x := 0; x < a.Bounds().Max.X; x++ {
				if a.At(x, y) != b.At(x, y) {
					return false
				}
			}
		}
		return true
	}
	v.Reverse(true)
	if same(mid, capture()) {
		t.Fatal("reverse has no effect")
	}
	v.Reverse(false).Once(true)
	now = now.Add(time.Second)
	if !same(first, capture()) {
		t.Fatal("once did not restore base text")
	}
	v.Restart()
	capture()
	now = now.Add(400 * time.Millisecond)
	if same(first, capture()) {
		t.Fatal("restart did not animate")
	}
	core.Update(func() { theme.SetReducedMotion(true) })
	if !same(first, capture()) {
		t.Fatal("reduced motion not static")
	}
	core.Update(func() { theme.SetReducedMotion(false) })
	v.Enabled(false)
	if !same(first, capture()) {
		t.Fatal("disabled shimmer not static")
	}
	v.Enabled(true)
	if !same(first, capture()) {
		t.Fatal("re-enable did not start at beginning")
	}
}

func TestShimmerTextMatchesOrdinaryTypography(t *testing.T) {
	for _, value := range []string{"中英文 gyp العربية مرحبا", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 多行文字\n第二行 gyp"} {
		v := kit.ShimmerText(value).MaxLines(2)
		now := time.Unix(1000, 0)
		root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(170)).TextSize(24).Bold().LineHeight(1.4).Child(v.Render(cx))
		}))
		w := openTest(t, Options{Width: 220, Height: 200, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
		first, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		a := element(t, w, value)
		v.Enabled(false)
		second, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		b := element(t, w, value)
		if a.Width != b.Width || a.Height != b.Height {
			t.Fatal("typography changed size", a, b)
		}
		if !bytes.Equal(first, second) {
			t.Fatal("inactive sweep differs from ordinary text", value)
		}
	}
}

func TestAttachmentTitleShimmerLifecycle(t *testing.T) {
	old := theme.ReducedMotion
	defer core.Update(func() { theme.SetReducedMotion(old) })
	core.Update(func() { theme.SetReducedMotion(false) })
	now := time.Unix(1000, 0)
	a := kit.Attachment("MMMMMMMMMMMMM", 0)
	a.SetProgress(.5)
	root := el.Embed(a)
	w := openTest(t, Options{Width: 320, Height: 160, Content: core.Func(func(gtx core.C) core.D { gtx.Now = now; return root.Layout(gtx) })})
	sample := func() []color.NRGBA {
		t.Helper()
		var title Element
		for _, e := range w.snapshot() {
			if e.Role == "text" && e.Name == "MMMMMMMMMMMMM" {
				title = e
				break
			}
		}
		if title.Width == 0 {
			t.Fatal("missing title text")
		}
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		pixels := []color.NRGBA{}
		for y := title.Y; y < title.Y+title.Height; y++ {
			for x := title.X; x < title.X+title.Width; x++ {
				pixels = append(pixels, color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA))
			}
		}
		return pixels
	}
	before := sample()
	now = now.Add(time.Second)
	if slices.Equal(before, sample()) {
		t.Fatal("upload title does not shimmer")
	}
	a.SetStatus(kit.AttachmentStatusComplete)
	complete := sample()
	now = now.Add(time.Second)
	if !slices.Equal(complete, sample()) {
		t.Fatal("completed title still animates")
	}
}
