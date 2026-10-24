// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/gpu/internal/driver"
)

func TestStorePoolClasses(t *testing.T) {
	var p storePool
	for _, size := range []int{0, 1, 64, 65, 1000, 4096} {
		s := p.get(size)
		if len(s) != size || cap(s) < size || cap(s)&(cap(s)-1) != 0 {
			t.Fatalf("get(%d): len %d cap %d", size, len(s), cap(s))
		}
		for i := range s {
			s[i] = 0xff
		}
		p.put(s)
		r := p.get(size)
		if &r[:1][0] != &s[:1][0] {
			t.Fatalf("get(%d) did not reuse the released store", size)
		}
		for i, v := range r {
			if v != 0 {
				t.Fatalf("get(%d): reused store not cleared at %d", size, i)
			}
		}
	}
	// Foreign slices are not kept.
	p.put(make([]byte, 100))
	p.put(make([]byte, 8192))
	for i, l := range p.free {
		if len(l) != 0 {
			t.Fatalf("class %d kept a foreign slice", i)
		}
	}
	for range maxStoresPerCls + 10 {
		p.put(make([]byte, 0, 128))
	}
	if n := len(p.free[storeClass(128)]); n != maxStoresPerCls {
		t.Fatalf("pool kept %d stores, want %d", n, maxStoresPerCls)
	}
	p.trim()
	if len(p.free[storeClass(128)]) != 0 {
		t.Fatal("trim kept stores")
	}
}

func TestBufferPoolCompletionAndStorageClasses(t *testing.T) {
	var p bufferPool
	shared := bufferPoolKey{class: 8192, storage: bufferStorageShared}
	private := bufferPoolKey{class: 8192, storage: bufferStoragePrivate}
	if !p.put(shared, 0) {
		t.Fatal("shared buffer was not admitted")
	}
	if len(p.pending) != 1 || len(p.ready[shared]) != 0 {
		t.Fatal("released buffer became ready before frame completion")
	}
	// A zero native reference is sufficient for testing pool bookkeeping; no
	// native release occurs while the per-class limit is not exceeded.
	p.frameStarted()
	if len(p.pending) != 0 || len(p.ready[shared]) != 1 {
		t.Fatal("completed shared buffer did not become ready")
	}
	p.get(private)
	if len(p.ready[shared]) != 1 {
		t.Fatal("private allocation consumed shared storage")
	}
	p.get(shared)
	if len(p.ready[shared]) != 0 || p.bytes != 0 {
		t.Fatal("shared allocation did not consume completed shared storage")
	}
}

// Paths rebuilt every frame go through small immutable buffers; their
// stores must not be allocated anew each time.
func TestSmallBufferStoresReused(t *testing.T) {
	b := testBackend(t)
	data := make([]byte, 1500)
	for i := range data {
		data[i] = byte(i)
	}
	// Warm the pool and check the contents round trip.
	buf, err := b.NewImmutableBuffer(driver.BufferBindingVertices, data)
	if err != nil {
		t.Fatal(err)
	}
	if got := buf.(*Buffer).store; string(got) != string(data) {
		t.Fatal("store does not hold the uploaded data")
	}
	buf.Release()
	allocs := testing.AllocsPerRun(100, func() {
		buf, err := b.NewImmutableBuffer(driver.BufferBindingVertices, data)
		if err != nil {
			t.Fatal(err)
		}
		buf.Release()
	})
	// One allocation remains: the *Buffer itself.
	if allocs > 1 {
		t.Fatalf("%v allocations per small buffer, want at most 1", allocs)
	}
	ubuf, err := b.NewBuffer(driver.BufferBindingUniforms, 64)
	if err != nil {
		t.Fatal(err)
	}
	ubuf.Upload([]byte{1, 2, 3})
	ubuf.Release()
	ubuf, err = b.NewBuffer(driver.BufferBindingUniforms, 64)
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range ubuf.(*Buffer).store {
		if v != 0 {
			t.Fatalf("new buffer has stale byte %d at %d", v, i)
		}
	}
	ubuf.Release()
}

func TestLargeImmutableBuffersReuseSharedStorage(t *testing.T) {
	b := testBackend(t)
	if !b.unified {
		t.Skip("shared immutable buffers require unified memory")
	}
	data := make([]byte, 5000)
	for i := range data {
		data[i] = byte(i * 17)
	}
	firstBuffer, err := b.NewImmutableBuffer(driver.BufferBindingVertices, data)
	if err != nil {
		t.Fatal(err)
	}
	first := firstBuffer.(*Buffer)
	if first.poolKey.storage != bufferStorageShared {
		t.Fatal("large immutable buffer did not use shared storage")
	}
	ref := first.buffer
	first.Release()
	if len(b.pool.pending) != 1 {
		t.Fatalf("released buffer count = %d, want 1 pending completion", len(b.pool.pending))
	}
	// BeginFrame waits for the previous command buffer before making released
	// buffers available. There is no command here, but the same lifetime gate
	// still advances the pool.
	b.BeginFrame(nil, false, image.Point{})
	for i := range data {
		data[i] ^= 0xff
	}
	secondBuffer, err := b.NewImmutableBuffer(driver.BufferBindingVertices, data)
	if err != nil {
		t.Fatal(err)
	}
	second := secondBuffer.(*Buffer)
	if second.buffer != ref {
		t.Fatal("large immutable buffer did not reuse completed shared storage")
	}
	if got := bufferStore(second.buffer)[:len(data)]; string(got) != string(data) {
		t.Fatal("reused shared storage does not hold the new path vertices")
	}
	second.Release()

	mutableBuffer, err := b.NewBuffer(driver.BufferBindingVertices, len(data))
	if err != nil {
		t.Fatal(err)
	}
	mutable := mutableBuffer.(*Buffer)
	if mutable.poolKey.storage != bufferStoragePrivate {
		t.Fatal("mutable vertex buffer did not use private storage")
	}
	mutable.Release()
}
