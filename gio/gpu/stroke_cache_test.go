package gpu

import (
	"bytes"
	"encoding/binary"
	"hash/maphash"
	"math"
	"testing"

	"github.com/dyike/keel/gio/internal/f32"
	"github.com/dyike/keel/gio/internal/ops"
	"github.com/dyike/keel/gio/internal/scene"
	"github.com/dyike/keel/gio/internal/stroke"
)

func testStrokePath() []byte {
	data := make([]byte, scene.CommandSize+4)
	binary.LittleEndian.PutUint32(data, 1)
	ops.EncodeCommand(data[4:], scene.Cubic(f32.Pt(10, 20), f32.Pt(10, -10), f32.Pt(90, -10), f32.Pt(90, 20)))
	return data
}

// The pre-change buildVerts stroke path is the byte-for-byte oracle. In
// particular, stroking must precede transform even for shear and reflection.
func originalStrokeVerts(d *drawOps, path []byte, tr f32.Affine2D, width float32) ([]byte, f32.Rectangle) {
	inf := float32(math.Inf(1))
	d.qs.bounds = f32.Rectangle{Min: f32.Pt(inf, inf), Max: f32.Pt(-inf, -inf)}
	d.qs.d = d
	start := len(d.vertCache)
	for _, q := range stroke.StrokePathCommands(stroke.StrokeStyle{Width: width}, path) {
		d.qs.contour = q.Contour
		d.qs.splitAndEncode(q.Quad.Transform(tr))
	}
	fillMaxY(d.vertCache[start:])
	return d.vertCache[start:], d.qs.bounds
}

func TestStrokeCachePreservesAffineVertices(t *testing.T) {
	path := testStrokePath()
	var cached, original drawOps
	transforms := []f32.Affine2D{
		f32.AffineId(),
		f32.NewAffine2D(2, .7, .5, -.3, .6, .25),
		f32.NewAffine2D(-1, 0, 40, 0, 2, 30),
	}
	for i := range 120 {
		transforms = append(transforms, f32.AffineId().Rotate(f32.Pt(50, 50), float32(i)*.031))
	}
	for _, width := range []float32{.5, 3, 7} {
		for _, tr := range transforms {
			cached.vertCache, original.vertCache = cached.vertCache[:0], original.vertCache[:0]
			a, ab := cached.buildVerts(path, tr, false, width)
			b, bb := originalStrokeVerts(&original, path, tr, width)
			if ab != bb || !bytes.Equal(a, b) {
				t.Fatalf("vertices changed for width %v transform %v", width, tr)
			}
		}
	}
	if len(cached.strokeCache.index) != 3 {
		t.Fatalf("transforms created %d stroke entries, want 3 widths", len(cached.strokeCache.index))
	}
	i := 0
	if n := testing.AllocsPerRun(120, func() {
		cached.vertCache = cached.vertCache[:0]
		cached.buildVerts(path, transforms[i%len(transforms)], false, 3)
		i++
	}); n != 0 {
		t.Fatalf("warmed rotating vertices allocated %g times", n)
	}
}

func TestStrokeCacheOwnsStorageAndExpires(t *testing.T) {
	var cache strokeCache
	path := testStrokePath()
	quads := stroke.StrokePathCommands(stroke.StrokeStyle{Width: 3}, path)
	want := quads[0]
	cache.put(path, 3, quads)
	path[4], quads[0] = 0, stroke.StrokeQuad{}
	if got, ok := cache.get(testStrokePath(), 3); !ok || got[0] != want {
		t.Fatal("cache borrowed mutable input")
	}
	if _, ok := cache.get(testStrokePath(), 4); ok {
		t.Fatal("different widths shared geometry")
	}
	cache.frame()
	cache.frame()
	if cache.bytes != 0 || len(cache.index) != 0 || cache.entries[0].quads != nil || cache.entries[0].path != nil {
		t.Fatal("unused geometry retained")
	}
	// An artificial hash collision must not return the wrong path.
	cache.put(testStrokePath(), 3, quads)
	k := strokeKey{maphash.Bytes(pathSeed, testStrokePath()), 3}
	cache.entries[cache.index[k]].path[4] ^= 1
	if _, ok := cache.get(testStrokePath(), 3); ok {
		t.Fatal("hash collision reused geometry")
	}
}

func TestStrokeCacheBudgetAndWarmAllocs(t *testing.T) {
	var cache strokeCache
	path := testStrokePath()
	quads := stroke.StrokePathCommands(stroke.StrokeStyle{Width: 3}, path)
	for i := range maxStrokeEntries * 3 {
		cache.put(path, float32(i+1), quads)
	}
	if len(cache.index) != maxStrokeEntries || len(cache.entries) != maxStrokeEntries || cache.bytes > maxStrokeBytes {
		t.Fatal("entry bound exceeded")
	}
	large := make(stroke.StrokeQuads, maxStrokeBytes/28+1)
	cache.put(path, 900, large)
	if _, ok := cache.get(path, 900); ok {
		t.Fatal("oversized geometry cached")
	}
	// Exercise the byte budget independently of the entry-count budget.
	medium := make(stroke.StrokeQuads, 8192)
	for i := range 12 {
		cache.put(path, float32(1000+i), medium)
		if cache.bytes > maxStrokeBytes {
			t.Fatal("byte bound exceeded")
		}
	}
	cache.put(path, 3, quads)
	if n := testing.AllocsPerRun(100, func() { cache.get(path, 3) }); n != 0 {
		t.Fatalf("warm cache allocated %g times", n)
	}
}

func BenchmarkRotatingStrokeVertices(b *testing.B) {
	path := testStrokePath()
	for _, cached := range []bool{false, true} {
		name := "original"
		if cached {
			name = "cached"
		}
		b.Run(name, func(b *testing.B) {
			var d drawOps
			for i := range 120 {
				d.vertCache = d.vertCache[:0]
				d.buildVerts(path, f32.AffineId().Rotate(f32.Point{}, float32(i)*.01), false, 3)
			}
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				d.vertCache = d.vertCache[:0]
				tr := f32.AffineId().Rotate(f32.Point{}, float32(i%120)*.01)
				if cached {
					d.buildVerts(path, tr, false, 3)
				} else {
					originalStrokeVerts(&d, path, tr, 3)
				}
				i++
			}
		})
	}
}

func BenchmarkChangingStrokeVertices(b *testing.B) {
	path := testStrokePath()
	for _, cached := range []bool{false, true} {
		name := "original"
		if cached {
			name = "scratch"
		}
		b.Run(name, func(b *testing.B) {
			var d drawOps
			b.ReportAllocs()
			i := uint32(0)
			for b.Loop() {
				// The final cubic control coordinate changes every frame. This
				// measures cache misses rather than rotation-only reuse.
				binary.LittleEndian.PutUint32(path[4+5*4:], math.Float32bits(90+float32(i%10000)*.001))
				d.vertCache = d.vertCache[:0]
				if cached {
					d.buildVerts(path, f32.AffineId(), false, 3)
					d.strokeCache.frame()
				} else {
					originalStrokeVerts(&d, path, f32.AffineId(), 3)
				}
				i++
			}
		})
	}
}
