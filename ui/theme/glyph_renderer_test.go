package theme

import (
	"image"
	"image/color"
	"runtime"
	"strings"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/font"
	"github.com/dyike/keel/third_party/gio/font/gofont"
	"github.com/dyike/keel/third_party/gio/gpu/headless"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/text"
	"golang.org/x/image/math/fixed"
)

func TestGlyphRendererOriginsAndFallbacks(t *testing.T) {
	w, err := headless.NewWindow(500, 100)
	if err != nil {
		t.Skipf("GPU unavailable: %v", err)
	}
	defer w.Release()
	sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	var renderer GlyphRenderer
	defer renderer.Release()
	var ops op.Ops
	for _, value := range []string{"abcdef0123456789", "AV fi á", "مرحبا12", "12🙂34"} {
		for _, weight := range []font.Weight{font.Normal, font.Bold} {
			params := text.Parameters{Font: font.Font{Typeface: "Go Mono", Weight: weight}, PxPerEm: fixed.I(25), MaxWidth: 10000}
			sh.LayoutString(params, value)
			var glyphs []text.Glyph
			for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
				glyphs = append(glyphs, g)
			}
			for _, position := range []f32.Point{{X: 10, Y: 60}, {X: 10.25, Y: 60}, {X: 10.75, Y: 60}, {X: -0.25, Y: 60}, {X: 10.25, Y: 60.5}} {
				for _, alpha := range []uint8{255, 128} {
					run := GlyphRun{Params: params, Glyphs: glyphs, Color: color.NRGBA{R: 230, G: 190, B: 160, A: alpha}, Position: position}
					var images [2]*image.RGBA
					for mode := range images {
						ops.Reset()
						paint.Fill(&ops, color.NRGBA{R: 20, G: 30, B: 40, A: 255})
						if mode == 0 {
							tr := op.Affine(f32.AffineId().Offset(position)).Push(&ops)
							paintGlyphRun(&ops, sh, glyphs, run.Color)
							tr.Pop()
						} else {
							renderer.BeginFrame(sh)
							renderer.Prepare(run)
							renderer.Commit()
							renderer.Paint(&ops, run)
						}
						if err := w.Frame(&ops); err != nil {
							t.Fatal(err)
						}
						images[mode] = image.NewRGBA(image.Rect(0, 0, 500, 100))
						if err := w.Screenshot(images[mode]); err != nil {
							t.Fatal(err)
						}
					}
					maxDelta, sumDelta, ink := 0, 0, 0
					for i, want := range images[0].Pix {
						if want != []byte{20, 30, 40, 255}[i%4] {
							ink++
						}
						delta := int(images[1].Pix[i]) - int(want)
						if delta < 0 {
							delta = -delta
						}
						maxDelta = max(maxDelta, delta)
						sumDelta += delta
					}
					// Baking a fractional origin into a mask removes the extra
					// texture interpolation of the old vector path. Its edges
					// may be sharper, but local coverage and placement must match.
					fractionalMask := position.X != float32(int(position.X)) && position.Y == float32(int(position.Y))
					if fractionalMask {
						for y := 0; y < 100; y += 8 {
							for x := 0; x < 500; x += 8 {
								for channel := 0; channel < 3; channel++ {
									sum, count := 0, 0
									for yy := y; yy < min(y+8, 100); yy++ {
										for xx := x; xx < min(x+8, 500); xx++ {
											i := images[0].PixOffset(xx, yy) + channel
											sum += int(images[1].Pix[i]) - int(images[0].Pix[i])
											count++
										}
									}
									if sum < -12*count || sum > 12*count {
										t.Fatalf("%q weight=%v origin=%v alpha=%d tile=(%d,%d) coverage delta=%g", value, weight, position, alpha, x, y, float64(sum)/float64(count))
									}
								}
							}
						}
					} else if maxDelta > 24 || sumDelta > max(1, ink)*2 {
						t.Fatalf("%q weight=%v origin=%v alpha=%d max=%d mean=%g", value, weight, position, alpha, maxDelta, float64(sumDelta)/float64(max(1, ink)))
					}
				}
			}
		}
	}
	if renderer.Stats().RasterDraws == 0 || renderer.Stats().VectorDraws == 0 {
		t.Fatal("did not exercise both raster and complex-run fallback")
	}
	renderer.ReleaseScratch()
	if renderer.atlas.window != nil || renderer.atlas.readback != nil {
		t.Fatal("idle renderer retained scratch GPU resources")
	}
	other := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	renderer.BeginFrame(other)
	if renderer.Stats().Masks != 0 || len(renderer.blanks) != 0 || len(renderer.vector.fragments) != 0 {
		t.Fatal("font change retained old glyph resources")
	}
	renderer.Release()
	if renderer.shaper != nil || renderer.Stats().Pages != 0 || renderer.scratch != nil {
		t.Fatal("release retained view resources")
	}
}

// A changing grid catches whole-string outline cache churn in custom views.
// Keep this benchmark in Keel so downstream apps share the same regression.
func BenchmarkGlyphRendererScrolling(b *testing.B) {
	for _, mode := range []string{"WholeRun", "Cached"} {
		b.Run(mode, func(b *testing.B) {
			sh := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
			w, err := headless.NewWindow(1920, 992)
			if err != nil {
				b.Fatal(err)
			}
			defer w.Release()
			var renderer GlyphRenderer
			defer renderer.Release()
			var ops op.Ops
			var glyphs [31][]text.Glyph
			var runs [31]GlyphRun
			params := text.Parameters{Font: font.Font{Typeface: "Go Mono"}, PxPerEm: fixed.I(25), MaxWidth: 10000}
			b.ReportAllocs()
			b.ResetTimer()
			for frame := 0; frame < b.N; frame++ {
				ops.Reset()
				renderer.BeginFrame(sh)
				for y := range runs {
					var value strings.Builder
					value.Grow(127)
					n := uint32(frame*31+y+1) * 2654435761
					for x := 0; x < 127; x++ {
						n ^= n << 13
						n ^= n >> 17
						n ^= n << 5
						value.WriteByte(byte('a' + n%26))
					}
					sh.LayoutString(params, value.String())
					glyphs[y] = glyphs[y][:0]
					for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
						glyphs[y] = append(glyphs[y], g)
					}
					runs[y] = GlyphRun{Params: params, Glyphs: glyphs[y], Color: color.NRGBA{R: 220, G: 220, B: 220, A: 255}, Position: f32.Pt(.25, float32(25+y*32))}
					if mode == "Cached" {
						renderer.Prepare(runs[y])
					}
				}
				if mode == "Cached" {
					renderer.Commit()
				}
				for _, run := range runs {
					if mode == "Cached" {
						renderer.Paint(&ops, run)
					} else {
						tr := op.Affine(f32.AffineId().Offset(run.Position)).Push(&ops)
						paintGlyphRun(&ops, sh, run.Glyphs, run.Color)
						tr.Pop()
					}
				}
				if err := w.Frame(&ops); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			runtime.GC()
			var memory runtime.MemStats
			runtime.ReadMemStats(&memory)
			b.ReportMetric(float64(memory.HeapAlloc)/(1<<20), "live-heap-MiB")
			runtime.KeepAlive(sh)
		})
	}
}
