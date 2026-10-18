// SPDX-License-Identifier: Unlicense OR MIT

package metal

// Keel patch: buffers of at most 4 KiB live in Go memory and reach the GPU
// through set*Bytes, which copies them into the command encoder. Paths are
// rebuilt into such buffers every frame, so allocating each store made an
// animating window produce megabytes of garbage a second. Released stores
// are kept by size class; set*Bytes already copied their contents, so a
// store is free for reuse as soon as its buffer is released.

const (
	minStoreClass   = 64
	maxStoreSize    = 4096
	maxStoresPerCls = 256
	storeClasses    = 7 // 64, 128, ..., 4096
)

type storePool struct {
	free [storeClasses][][]byte
}

// storeClass returns the index of the smallest class holding size bytes.
func storeClass(size int) int {
	i, c := 0, minStoreClass
	for c < size {
		c <<= 1
		i++
	}
	return i
}

// get returns a zeroed slice of length size.
func (p *storePool) get(size int) []byte {
	i := storeClass(size)
	if list := p.free[i]; len(list) > 0 {
		s := list[len(list)-1]
		list[len(list)-1] = nil
		p.free[i] = list[:len(list)-1]
		s = s[:size]
		clear(s)
		return s
	}
	return make([]byte, size, minStoreClass<<i)
}

// put keeps s for reuse. s must come from get.
func (p *storePool) put(s []byte) {
	if cap(s) < minStoreClass || cap(s) > maxStoreSize {
		return
	}
	i := storeClass(cap(s))
	if minStoreClass<<i != cap(s) || len(p.free[i]) >= maxStoresPerCls {
		return
	}
	p.free[i] = append(p.free[i], s[:0])
}

func (p *storePool) trim() {
	*p = storePool{}
}
