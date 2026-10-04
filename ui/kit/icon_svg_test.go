package kit

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestSVGIconValidationAndFile(t *testing.T) {
	for _, s := range []string{"", "<svg>", `<svg viewBox="0 0 0 20"/>`, `<svg viewBox="0 0 -1 20"/>`, `<svg viewBox="NaN 0 20 20"/>`, `<svg viewBox="0 0 20 20"><script>bad</script></svg>`, `<svg viewBox="0 0 20 20"><text>unsupported</text></svg>`} {
		if _, err := SVGIcon([]byte(s)); err == nil {
			t.Errorf("accepted invalid SVG: %s", s)
		}
	}
	if _, err := SVGIcon(bytes.Repeat([]byte("x"), maxSVGBytes+1)); err == nil {
		t.Fatal("oversized bytes accepted")
	}
	p := filepath.Join(t.TempDir(), "icon.svg")
	valid := []byte(`<svg width="20" height="10"><rect width="20" height="10"/></svg>`)
	if err := os.WriteFile(p, valid, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := SVGIconFile(p); err != nil {
		t.Fatal(err)
	}
	if _, err := SVGIconFile(p + "missing"); err == nil {
		t.Fatal("missing file accepted")
	}
	if err := os.WriteFile(p, bytes.Repeat([]byte("x"), maxSVGBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := SVGIconFile(p); err == nil {
		t.Fatal("oversized file accepted")
	}
}

func TestSVGIconPixelsTintRotationAndScale(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v, err := SVGIcon([]byte(`<svg viewBox="10 20 40 20"><rect x="10" y="20" width="20" height="20" fill="#ff0000"/><rect x="30" y="20" width="20" height="20" fill="#00ff00"/></svg>`))
		if err != nil {
			t.Fatal(err)
		}
		v.Size(40).OriginalColors(true)
		gpu, err := headless.NewWindow(64*scale, 64*scale)
		if err != nil {
			t.Fatal(err)
		}
		var img *image.RGBA
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Bg(color.NRGBA{R: 255, G: 255, B: 255, A: 255}).Items(el.Start).Child(v.Render(cx))
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(64*scale, 64*scale)
			root.Layout(gtx)
			if err := gpu.Frame(gtx.Ops); err != nil {
				t.Fatal(err)
			}
			img = image.NewRGBA(image.Rect(0, 0, 64*scale, 64*scale))
			if err := gpu.Screenshot(img); err != nil {
				t.Fatal(err)
			}
		})
		pixel := func(x, y int) color.RGBA { return img.RGBAAt(x*scale, y*scale) }
		check := func(x, y int, want color.RGBA) {
			t.Helper()
			got := pixel(x, y)
			if got != want {
				t.Fatalf("scale%d point%d,%d got%v want%v", scale, x, y, got, want)
			}
		}
		check(5, 20, color.RGBA{255, 0, 0, 255})
		check(35, 20, color.RGBA{0, 255, 0, 255})
		check(20, 5, color.RGBA{255, 255, 255, 255})
		v.Rotate(90)
		h.Frame()
		check(20, 5, color.RGBA{255, 0, 0, 255})
		check(20, 35, color.RGBA{0, 255, 0, 255})
		check(5, 20, color.RGBA{255, 255, 255, 255})
		v.Rotate(0).OriginalColors(false).Color(color.NRGBA{B: 255, A: 255})
		h.Frame()
		check(5, 20, color.RGBA{0, 0, 255, 255})
		check(35, 20, color.RGBA{0, 0, 255, 255})
		if v.svg.size != image.Pt(40*scale, 40*scale) {
			t.Fatal("not physical resolution", v.svg.size)
		}
		v.Color(color.NRGBA{B: 255, A: 128})
		h.Frame()
		p := pixel(5, 20)
		if p.B < 250 || p.R < 120 || p.R > 195 || p.G != p.R {
			t.Fatal("alpha tint lost", p)
		}
		v.OriginalColors(true)
		h.Frame()
		check(5, 20, color.RGBA{255, 0, 0, 255})
		gradient, err := SVGIcon([]byte(`<svg viewBox="0 0 40 40"><defs><linearGradient id="g"><stop offset="0" stop-color="#ff0000"/><stop offset="1" stop-color="#0000ff"/></linearGradient></defs><g transform="translate(0 10)"><path d="M0 0H40V20H0Z" fill="url(#g)"/></g></svg>`))
		if err != nil {
			t.Fatal(err)
		}
		*v = *gradient.Size(40).OriginalColors(true)
		h.Frame()
		left, right := pixel(5, 20), pixel(35, 20)
		if left.R <= left.B || right.B <= right.R {
			t.Fatal("gradient direction", left, right)
		}
		check(20, 5, color.RGBA{255, 255, 255, 255})
		alpha, err := SVGIcon([]byte(`<svg viewBox="0 0 40 40"><rect width="40" height="40" fill="#ff0000" fill-opacity="0.5"/></svg>`))
		if err != nil {
			t.Fatal(err)
		}
		*v = *alpha.Size(40).OriginalColors(true)
		h.Frame()
		p = pixel(20, 20)
		if p.R < 250 || p.G < 180 || p.G > 195 || p.B != p.G {
			t.Fatal("source opacity darkened", p)
		}
		v.Size(20)
		h.Frame()
		if v.svg.size != image.Pt(20*scale, 20*scale) {
			t.Fatal("stale raster size", v.svg.size)
		}
		gpu.Release()
	}
}

func TestSVGIconThemeAndRasterLimit(t *testing.T) {
	old := theme.Current()
	defer theme.Apply(old)
	v, err := SVGIcon([]byte(`<svg viewBox="0 0 10 10"><path fill="currentColor" d="M0 0H10V10H0Z"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	h := renderView(v, 100, 1)
	theme.Apply(theme.Dark())
	h.Frame()
	if v.svg.color != theme.Text || !v.svg.valid {
		t.Fatal("theme tint did not update")
	}
	v.Rotate(450).Rotate(float32(math.NaN())).Size(float32(math.Inf(1)))
	if v.rotation != 90 || v.size != 18 {
		t.Fatal("invalid geometry accepted")
	}
	v.Size(4096)
	renderView(v, 4096, 1)
	if v.svg.size.X > 2048 || v.svg.size.Y > 2048 {
		t.Fatal("raster limit", v.svg.size)
	}
}
