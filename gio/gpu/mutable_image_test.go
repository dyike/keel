// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"image/color"
	"sync"
	"testing"

	"github.com/dyike/keel/gio/gpu/internal/driver"
	"github.com/dyike/keel/gio/internal/dirty"
	"github.com/dyike/keel/gio/internal/ops"
	gioop "github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
)

type uploadRecord struct {
	offset, size image.Point
	pixels       []byte // tightly packed
}

type mutableTestTexture struct {
	driver.Texture
	uploads  []uploadRecord
	released bool
}

func (t *mutableTestTexture) Upload(offset, size image.Point, pixels []byte, stride int) {
	rec := uploadRecord{offset: offset, size: size}
	for y := 0; y < size.Y; y++ {
		rec.pixels = append(rec.pixels, pixels[y*stride:y*stride+size.X*4]...)
	}
	t.uploads = append(t.uploads, rec)
}
func (t *mutableTestTexture) Release() { t.released = true }

type mutableTestDevice struct {
	driver.Device
	textures []*mutableTestTexture
	filters  [][2]driver.TextureFilter
}

func (d *mutableTestDevice) NewTexture(_ driver.TextureFormat, _, _ int, minF, magF driver.TextureFilter, _ driver.BufferBinding) (driver.Texture, error) {
	t := new(mutableTestTexture)
	d.textures = append(d.textures, t)
	d.filters = append(d.filters, [2]driver.TextureFilter{minF, magF})
	return t, nil
}

func TestMutableImageUploadsOnlyChanges(t *testing.T) {
	// The source is a sub-image, so image coordinates are not texture
	// coordinates: Invalidate takes the former, uploads use the latter.
	backing := image.NewRGBA(image.Rect(0, 0, 20, 20))
	src := backing.SubImage(image.Rect(4, 4, 14, 12)).(*image.RGBA)
	op := paint.NewMutableImageOp(src)
	tracker := mutableTracker(t, op)

	dev := &mutableTestDevice{}
	r := renderer{ctx: dev}
	cache := newTextureCache()
	defer cache.release()
	data := imageOpData{src: src, handle: tracker, tint: whiteTint}

	tex := r.texHandle(cache, data)
	if len(dev.textures) != 1 || len(dev.textures[0].uploads) != 1 {
		t.Fatalf("first use: %d textures", len(dev.textures))
	}
	if u := dev.textures[0].uploads[0]; u.offset != (image.Point{}) || u.size != image.Pt(10, 8) {
		t.Fatalf("initial upload %v %v", u.offset, u.size)
	}

	// Unchanged: no upload.
	cache.frame()
	if r.texHandle(cache, data) != tex || len(dev.textures[0].uploads) != 1 {
		t.Fatal("unchanged image uploaded again")
	}

	// Two changes in one frame upload their union into the same texture.
	set := func(x, y int, c color.RGBA) { backing.SetRGBA(x, y, c) }
	set(5, 5, color.RGBA{R: 1, A: 255})
	op.Invalidate(image.Rect(5, 5, 6, 6))
	set(7, 6, color.RGBA{G: 2, A: 255})
	op.Invalidate(image.Rect(7, 6, 8, 7))
	cache.frame()
	if r.texHandle(cache, data) != tex || len(dev.textures) != 1 {
		t.Fatal("change created a new texture")
	}
	u := dev.textures[0].uploads[1]
	if u.offset != image.Pt(1, 1) || u.size != image.Pt(3, 2) {
		t.Fatalf("partial upload at %v size %v, want (1,1) 3x2", u.offset, u.size)
	}
	if u.pixels[0] != 1 || u.pixels[(1*3+2)*4+1] != 2 {
		t.Fatalf("partial upload pixels %v", u.pixels)
	}

	// Changes outside the image are clipped away.
	op.Invalidate(image.Rect(0, 0, 3, 3))
	cache.frame()
	r.texHandle(cache, data)
	if len(dev.textures[0].uploads) != 2 {
		t.Fatal("uploaded an area outside the image")
	}

	// A renderer that fell behind the change log uploads everything.
	for i := range 40 {
		op.Invalidate(image.Rect(4+i%10, 4, 5+i%10, 5))
	}
	cache.frame()
	r.texHandle(cache, data)
	if u := dev.textures[0].uploads[2]; u.offset != (image.Point{}) || u.size != image.Pt(10, 8) {
		t.Fatalf("catch-up upload %v %v, want the whole image", u.offset, u.size)
	}

	// Unused textures are still released.
	cache.frame()
	cache.frame()
	if !dev.textures[0].released {
		t.Fatal("texture leaked")
	}
}

func TestMutableImageTint(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	op := paint.NewMutableImageOp(src)
	dev := &mutableTestDevice{}
	r := renderer{ctx: dev}
	cache := newTextureCache()
	defer cache.release()
	red := color.NRGBA{R: 255, A: 255}
	data := imageOpData{src: src, handle: mutableTracker(t, op), tint: red}
	r.texHandle(cache, data)
	src.SetRGBA(2, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	op.Invalidate(image.Rect(2, 3, 3, 4))
	cache.frame()
	r.texHandle(cache, data)
	u := dev.textures[0].uploads[1]
	if u.offset != image.Pt(2, 3) || u.size != image.Pt(1, 1) {
		t.Fatalf("tinted upload %v %v", u.offset, u.size)
	}
	if u.pixels[0] != 255 || u.pixels[1] != 0 || u.pixels[2] != 0 {
		t.Fatalf("tint not applied to the update: %v", u.pixels)
	}
}

func TestLinearNoMipmapFilter(t *testing.T) {
	if paint.FilterLinearNoMipmap != filterLinearNoMipmap {
		t.Fatal("paint and gpu filter numbers differ")
	}
	dev := &mutableTestDevice{}
	r := renderer{ctx: dev}
	cache := newTextureCache()
	defer cache.release()
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for _, f := range []byte{filterLinear, filterLinearNoMipmap, filterNearest} {
		r.texHandle(cache, imageOpData{src: src, handle: new(int), tint: whiteTint, filter: f})
	}
	want := [][2]driver.TextureFilter{
		{driver.FilterLinearMipmapLinear, driver.FilterLinear},
		{driver.FilterLinear, driver.FilterLinear},
		{driver.FilterNearest, driver.FilterNearest},
	}
	for i, w := range want {
		if dev.filters[i] != w {
			t.Errorf("filter %d: got %v want %v", i, dev.filters[i], w)
		}
	}
}

func TestInvalidateIgnoresImmutableImages(t *testing.T) {
	op := paint.NewImageOp(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	op.Invalidate(image.Rect(0, 0, 1, 1)) // must not panic
}

// mutableTracker extracts the tracker the way the GPU sees it: through the
// op encoding.
func mutableTracker(t *testing.T, op paint.ImageOp) *dirty.Tracker {
	t.Helper()
	tr, ok := imageHandle(op).(*dirty.Tracker)
	if !ok {
		t.Fatal("mutable image has no tracker")
	}
	return tr
}

// imageHandle encodes op and decodes it as the renderer does.
func imageHandle(op paint.ImageOp) any {
	var o gioop.Ops
	op.Add(&o)
	var r ops.Reader
	r.Reset(&o.Internal)
	for enc, ok := r.Decode(); ok; enc, ok = r.Decode() {
		if ops.OpType(enc.Data[0]) == ops.TypeImage {
			return decodeImageOp(enc.Data, enc.Refs).handle
		}
	}
	return nil
}

// heldLock records whether it is held and how often it was taken.
type heldLock struct {
	sync.Mutex
	held  bool
	locks int
}

func (l *heldLock) Lock()   { l.Mutex.Lock(); l.held = true; l.locks++ }
func (l *heldLock) Unlock() { l.held = false; l.Mutex.Unlock() }

type lockCheckTexture struct {
	mutableTestTexture
	lock     *heldLock
	unlocked int // uploads made without the lock
}

func (t *lockCheckTexture) Upload(offset, size image.Point, pixels []byte, stride int) {
	if !t.lock.held {
		t.unlocked++
	}
	t.mutableTestTexture.Upload(offset, size, pixels, stride)
}

type lockCheckDevice struct {
	driver.Device
	lock     *heldLock
	textures []*lockCheckTexture
}

func (d *lockCheckDevice) NewTexture(driver.TextureFormat, int, int, driver.TextureFilter, driver.TextureFilter, driver.BufferBinding) (driver.Texture, error) {
	t := &lockCheckTexture{lock: d.lock}
	d.textures = append(d.textures, t)
	return t, nil
}

// A shared image is read only under its owner's lock, and only when it
// changed, so the owner can write its pixels outside the frame.
func TestSharedImageReadsUnderLock(t *testing.T) {
	mu := new(heldLock)
	fb := image.NewRGBA(image.Rect(0, 0, 16, 8))
	// Two tiles of one framebuffer share its lock and its pixels.
	left := paint.NewSharedImageOp(fb.SubImage(image.Rect(0, 0, 8, 8)).(*image.RGBA), mu)
	right := paint.NewSharedImageOp(fb.SubImage(image.Rect(8, 0, 16, 8)).(*image.RGBA), mu)
	dev := &lockCheckDevice{lock: mu}
	r := renderer{ctx: dev}
	cache := newTextureCache()
	defer cache.release()
	tiles := []imageOpData{
		{src: fb.SubImage(image.Rect(0, 0, 8, 8)).(*image.RGBA), handle: mutableTracker(t, left), tint: whiteTint},
		{src: fb.SubImage(image.Rect(8, 0, 16, 8)).(*image.RGBA), handle: mutableTracker(t, right), tint: whiteTint},
	}
	frame := func() {
		cache.frame()
		for _, d := range tiles {
			r.texHandle(cache, d)
		}
		if mu.held {
			t.Fatal("lock still held after the frame's uploads")
		}
	}
	frame()
	if mu.locks != 2 {
		t.Fatalf("initial uploads took the lock %d times, want 2", mu.locks)
	}
	frame()
	if mu.locks != 2 {
		t.Fatal("unchanged images took the lock")
	}
	mu.Lock()
	fb.SetRGBA(9, 2, color.RGBA{B: 7, A: 255})
	right.Invalidate(image.Rect(9, 2, 10, 3))
	mu.Unlock()
	frame()
	if mu.locks != 4 { // the test's own and one upload
		t.Fatalf("lock taken %d times, want 4", mu.locks)
	}
	up := dev.textures[1].uploads
	if len(up) != 2 || up[1].offset != image.Pt(1, 2) || up[1].pixels[2] != 7 {
		t.Fatalf("right tile uploads %+v", up)
	}
	if len(dev.textures[0].uploads) != 1 {
		t.Fatal("unchanged left tile uploaded again")
	}
	for i, tex := range dev.textures {
		if tex.unlocked != 0 {
			t.Fatalf("texture %d: %d uploads without the lock", i, tex.unlocked)
		}
	}

	// A writer on another goroutine; go test -race reports any read of the
	// pixels outside the lock.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := range 2000 {
			mu.Lock()
			x := i % 16
			fb.SetRGBA(x, i%8, color.RGBA{R: uint8(i), A: 255})
			if x < 8 {
				left.Invalidate(image.Rect(x, i%8, x+1, i%8+1))
			} else {
				right.Invalidate(image.Rect(x, i%8, x+1, i%8+1))
			}
			mu.Unlock()
		}
	}()
	for running := true; running; {
		select {
		case <-done:
			running = false
		default:
		}
		cache.frame()
		for _, d := range tiles {
			r.texHandle(cache, d)
		}
	}
}
