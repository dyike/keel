// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"unsafe"

	"github.com/dyike/keel/gio/gpu/internal/driver"
)

// Keel patch: driver.QuadBatcher for Metal. Gio drew every glyph and every
// rectangle with its own draw call and inline uniforms: 3,300 draws a frame
// for a terminal-sized text grid, which cost CPU (several cgo calls per
// draw) and command memory in the driver. Runs of quads go into one
// instanced draw here, computing exactly what Gio's blit shaders compute.
const quadSource = `
#include <metal_stdlib>
using namespace metal;

struct Quad {
	float4 transform; // scale.xy, offset.zw in clip space
	float4 uv0;       // first row of the UV affine transform (xyz), texture slot (w, -1: color)
	float4 uv1;       // second row
	float4 color;     // premultiplied, opacity applied; for textures the opacity
	float4 cover;
};

struct Varyings {
	float4 pos [[position]];
	float2 uv;
	float4 color;
	int slot [[flat]];
	float2 coverUV;
	int coverSlot [[flat]];
};

constant float2 corners[4] = {float2(-1, -1), float2(1, -1), float2(-1, 1), float2(1, 1)};
constant float2 uvs[4] = {float2(0, 0), float2(1, 0), float2(0, 1), float2(1, 1)};

vertex Varyings quad_vert(uint vid [[vertex_id]], uint iid [[instance_id]], const device Quad *quads [[buffer(0)]]) {
	Quad q = quads[iid];
	float2 p = corners[vid]*q.transform.xy + q.transform.zw;
	Varyings out;
	// Gio's windowTransform and fboTransform for Metal: flip y.
	out.pos = float4(p.x, -p.y, 0, 1);
	float3 v = float3(uvs[vid], 1);
	out.uv = float2(dot(q.uv0.xyz, v), dot(q.uv1.xyz, v));
	out.color = q.color;
	out.slot = int(q.uv0.w);
	out.coverUV = uvs[vid] * q.cover.xy + q.cover.zw;
	out.coverSlot = int(q.uv1.w) - 1;
	return out;
}

fragment float4 quad_frag(Varyings in [[stage_in]], array<texture2d<float>, 8> texs [[texture(0)]], array<sampler, 8> smps [[sampler(0)]]) {
	float4 color = in.color;
	if (in.slot >= 0) color *= texs[in.slot].sample(smps[in.slot], in.uv);
	if (in.coverSlot >= 0) color *= min(abs(texs[in.coverSlot].sample(smps[in.coverSlot], in.coverUV).r), 1.0);
	return color;
}
`

const quadSize = int(unsafe.Sizeof(driver.Quad{}))

type quadBatcher struct {
	failed   bool
	lib      uintptr
	pipeline uintptr
	format   uintptr
	buf      uintptr
	off      int
	texs     [driver.MaxQuadTextures]uintptr
	samplers [driver.MaxQuadTextures]uintptr
}

// frameStarted runs after BeginFrame waited for the previous frame: the
// instance buffer can be written from the start again.
func (q *quadBatcher) frameStarted() { q.off = 0 }

func (q *quadBatcher) release() {
	for _, r := range []uintptr{q.lib, q.pipeline, q.buf} {
		if r != 0 {
			cfRelease(r)
		}
	}
	*q = quadBatcher{}
}

func (q *quadBatcher) ready(b *Backend) bool {
	if q.failed {
		return false
	}
	if q.pipeline != 0 && q.format == b.pixelFmt {
		return true
	}
	if q.lib == 0 {
		q.lib = quadLibrary(b.dev, quadSource)
		if q.lib == 0 {
			q.failed = true
			return false
		}
	}
	if q.pipeline != 0 {
		cfRelease(q.pipeline)
	}
	q.pipeline = quadPipeline(b.dev, q.lib, "quad_frag", b.pixelFmt)
	q.format = b.pixelFmt
	if q.pipeline == 0 {
		q.failed = true
		return false
	}
	return true
}

// DrawQuads implements driver.QuadBatcher for the current render pass, which
// must target the window's pixel format.
func (b *Backend) DrawQuads(textures []driver.Texture, quads []driver.Quad) bool {
	q := &b.quads
	if len(quads) == 0 || len(textures) > driver.MaxQuadTextures || b.renderEnc == 0 || !q.ready(b) {
		return false
	}
	n := len(quads) * quadSize
	if q.buf == 0 || q.off+n > len(bufferStore(q.buf)) {
		// A command buffer retains the buffers it uses: the old one stays
		// alive until the GPU is done with it.
		if q.buf != 0 {
			cfRelease(q.buf)
		}
		size := max(64<<10, 2*(q.off+n))
		q.buf = quadNewBuffer(b.dev, uintptr(size))
		q.off = 0
		if q.buf == 0 {
			q.failed = true
			return false
		}
	}
	copy(bufferStore(q.buf)[q.off:], unsafe.Slice((*byte)(unsafe.Pointer(&quads[0])), n))
	for i, tex := range textures {
		t := tex.(*Texture)
		q.texs[i], q.samplers[i] = t.texture, t.sampler
	}
	quadDraw(b.renderEnc, q.pipeline, q.buf, uintptr(q.off), uintptr(len(quads)), q.texs[:len(textures)], q.samplers[:len(textures)])
	q.off += n
	return true
}

// PrepareQuads guarantees the tinted path has no allocation/compilation
// fallback inside a render pass. BeginFrame has waited for the previous use.
func (b *Backend) PrepareQuads(count int) bool {
	q := &b.quads
	if !q.ready(b) {
		return false
	}
	size := count * quadSize
	if q.buf != 0 && len(bufferStore(q.buf)) >= size {
		return true
	}
	buf := quadNewBuffer(b.dev, uintptr(max(64<<10, 2*size)))
	if buf == 0 {
		return false
	}
	if q.buf != 0 {
		cfRelease(q.buf)
	}
	q.buf, q.off = buf, 0
	return true
}

func (b *Backend) SupportsCoverageQuads() bool { return true }
