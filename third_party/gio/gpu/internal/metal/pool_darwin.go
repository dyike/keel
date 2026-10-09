package metal

/*
#include <CoreFoundation/CoreFoundation.h>
*/
import "C"

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
	class  int
	buffer C.CFTypeRef
}

// bufferPool keeps released buffers. A buffer released during a frame may
// still be read by that frame's command buffer, so it stays pending until
// the next frame starts, after Backend.BeginFrame waited for the GPU.
type bufferPool struct {
	ready   map[int][]C.CFTypeRef
	pending []pooledBuffer
	bytes   int
}

func (p *bufferPool) get(class int) C.CFTypeRef {
	list := p.ready[class]
	if len(list) == 0 {
		return 0
	}
	buf := list[len(list)-1]
	p.ready[class] = list[:len(list)-1]
	p.bytes -= class
	return buf
}

// put keeps buf for reuse, or reports false when the pool is full.
func (p *bufferPool) put(class int, buf C.CFTypeRef) bool {
	if class == 0 || p.bytes+class > maxPooledBytes {
		return false
	}
	p.pending = append(p.pending, pooledBuffer{class, buf})
	p.bytes += class
	return true
}

func (p *bufferPool) frameStarted() {
	if len(p.pending) == 0 {
		return
	}
	if p.ready == nil {
		p.ready = map[int][]C.CFTypeRef{}
	}
	for _, b := range p.pending {
		if len(p.ready[b.class]) >= maxPooledPerCls {
			C.CFRelease(b.buffer)
			p.bytes -= b.class
			continue
		}
		p.ready[b.class] = append(p.ready[b.class], b.buffer)
	}
	p.pending = p.pending[:0]
}

// trim releases every kept buffer. Command buffers retain the buffers they
// use, so pending ones can go too.
func (p *bufferPool) trim() {
	for _, list := range p.ready {
		for _, buf := range list {
			C.CFRelease(buf)
		}
	}
	for _, b := range p.pending {
		C.CFRelease(b.buffer)
	}
	p.ready, p.pending, p.bytes = nil, nil, 0
}
