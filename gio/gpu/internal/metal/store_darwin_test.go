// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
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
