package metal

// Keel patch: Gio allocated and freed GPU buffers every frame (the path
// data of anything that changes shape, such as a spinner), and newBuffer
// plus CFRelease took a fifth of an animating window's CPU. Released
// buffers are kept by size class and reused once the GPU is done with them.

const (
	minBufferClass  = 4 << 10
	maxPooledBytes  = 32 << 20
	maxPooledPerCls = 64
)

// bufferClass rounds size up to a power of two of at least 4 KiB.
func bufferClass(size int) int {
	c := minBufferClass
	for c < size {
		c <<= 1
	}
	return c
}

type pooledBuffer struct {
	key    bufferPoolKey
	buffer uintptr
}

type bufferStorage uint8

const (
	bufferStoragePrivate bufferStorage = iota
	bufferStorageShared
)

type bufferPoolKey struct {
	class   int
	storage bufferStorage
}

// bufferPool keeps released buffers. A buffer released during a frame may
// still be read by that frame's command buffer, so it stays pending until
// the next frame starts, after Backend.BeginFrame waited for the GPU.
type bufferPool struct {
	ready   map[bufferPoolKey][]uintptr
	pending []pooledBuffer
	bytes   int
}

func (p *bufferPool) get(key bufferPoolKey) uintptr {
	list := p.ready[key]
	if len(list) == 0 {
		return 0
	}
	buf := list[len(list)-1]
	p.ready[key] = list[:len(list)-1]
	p.bytes -= key.class
	return buf
}

// put keeps buf for reuse, or reports false when the pool is full.
func (p *bufferPool) put(key bufferPoolKey, buf uintptr) bool {
	if key.class == 0 || p.bytes+key.class > maxPooledBytes {
		return false
	}
	p.pending = append(p.pending, pooledBuffer{key, buf})
	p.bytes += key.class
	return true
}

func (p *bufferPool) frameStarted() {
	if len(p.pending) == 0 {
		return
	}
	if p.ready == nil {
		p.ready = map[bufferPoolKey][]uintptr{}
	}
	for _, b := range p.pending {
		if len(p.ready[b.key]) >= maxPooledPerCls {
			cfRelease(b.buffer)
			p.bytes -= b.key.class
			continue
		}
		p.ready[b.key] = append(p.ready[b.key], b.buffer)
	}
	p.pending = p.pending[:0]
}

// trim releases every kept buffer. Command buffers retain the buffers they
// use, so pending ones can go too.
func (p *bufferPool) trim() {
	for _, list := range p.ready {
		for _, buf := range list {
			cfRelease(buf)
		}
	}
	for _, b := range p.pending {
		cfRelease(b.buffer)
	}
	p.ready, p.pending, p.bytes = nil, nil, 0
}
