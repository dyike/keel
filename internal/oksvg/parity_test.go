package oksvg

import (
	"bytes"
	"image"
	"strings"
	"testing"

	upstream "github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func TestRasterParity(t *testing.T) {
	for name, body := range map[string]string{
		"paths":    `<path d="M2 2H28V28H2Z" fill="currentColor"/><circle cx="16" cy="16" r="10" fill="#00ff00" fill-opacity=".5"/>`,
		"gradient": `<defs><linearGradient id="g"><stop offset="0" stop-color="red"/><stop offset="1" stop-color="blue"/></linearGradient></defs><g transform="translate(2 3)"><path d="M0 0H24V24H0Z" fill="url(#g)"/></g>`,
		"use":      `<defs><path id="p" d="M0 0L12 0L6 12Z"/></defs><use href="#p" x="2" y="2" fill="blue"/><use href="#p" x="16" y="16" fill="red"/>`,
		"curves":   `<path d="M2 16C2 2 30 2 30 16Q16 30 2 16Z" fill="none" stroke="blue" stroke-width="2"/>`,
	} {
		t.Run(name, func(t *testing.T) {
			source := `<?xml version="1.0" encoding="UTF-8"?><svg viewBox="0 0 32 32"><title>中文 ✓</title>` + body + `</svg>`
			got, err := ReadReplacingCurrentColor(strings.NewReader(source), "black", StrictErrorMode)
			if err != nil {
				t.Fatal(err)
			}
			want, err := upstream.ReadReplacingCurrentColor(strings.NewReader(source), "black", upstream.StrictErrorMode)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SVGPaths) == 0 || got.Titles[0] != "中文 ✓" {
				t.Fatal("missing paths or UTF-8 title")
			}
			for _, scale := range []int{1, 2} {
				w := 32 * scale
				a, b := image.NewRGBA(image.Rect(0, 0, w, w)), image.NewRGBA(image.Rect(0, 0, w, w))
				got.SetTarget(0, 0, float64(w), float64(w))
				want.SetTarget(0, 0, float64(w), float64(w))
				got.Draw(rasterx.NewDasher(w, w, rasterx.NewScannerGV(w, w, a, a.Bounds())), 1)
				want.Draw(rasterx.NewDasher(w, w, rasterx.NewScannerGV(w, w, b, b.Bounds())), 1)
				if !bytes.Equal(a.Pix, b.Pix) {
					t.Fatalf("pixels differ at %dx", scale)
				}
				if bytes.Equal(a.Pix, make([]byte, len(a.Pix))) {
					t.Fatal("empty image")
				}
			}
		})
	}
}
