// SPDX-License-Identifier: Unlicense OR MIT

package metal

/*
#cgo CFLAGS: -Werror -xobjective-c -fobjc-arc

#include <CoreFoundation/CoreFoundation.h>
#include <Metal/Metal.h>
#include <stdlib.h>

static CFTypeRef quadPipeline(CFTypeRef devRef, CFTypeRef libRef, const char *frag, MTLPixelFormat format) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		id<MTLLibrary> lib = (__bridge id<MTLLibrary>)libRef;
		MTLRenderPipelineDescriptor *desc = [MTLRenderPipelineDescriptor new];
		desc.vertexFunction = [lib newFunctionWithName:@"quad_vert"];
		desc.fragmentFunction = [lib newFunctionWithName:[NSString stringWithUTF8String:frag]];
		desc.colorAttachments[0].pixelFormat = format;
		desc.colorAttachments[0].blendingEnabled = YES;
		desc.colorAttachments[0].sourceRGBBlendFactor = MTLBlendFactorOne;
		desc.colorAttachments[0].sourceAlphaBlendFactor = MTLBlendFactorOne;
		desc.colorAttachments[0].destinationRGBBlendFactor = MTLBlendFactorOneMinusSourceAlpha;
		desc.colorAttachments[0].destinationAlphaBlendFactor = MTLBlendFactorOneMinusSourceAlpha;
		NSError *err = nil;
		id<MTLRenderPipelineState> s = [dev newRenderPipelineStateWithDescriptor:desc error:&err];
		return s ? CFBridgingRetain(s) : NULL;
	}
}

static CFTypeRef quadNewBuffer(CFTypeRef devRef, NSUInteger size) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		return CFBridgingRetain([dev newBufferWithLength:size options:MTLResourceStorageModeShared|MTLResourceCPUCacheModeWriteCombined]);
	}
}

static CFTypeRef quadLibrary(CFTypeRef devRef, const char *src) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		NSError *err = nil;
		id<MTLLibrary> lib = [dev newLibraryWithSource:[NSString stringWithUTF8String:src] options:nil error:&err];
		return lib ? CFBridgingRetain(lib) : NULL;
	}
}

static void quadDraw(CFTypeRef encRef, CFTypeRef pipeRef, CFTypeRef bufRef, NSUInteger off, NSUInteger n, CFTypeRef *texs, CFTypeRef *samplers, NSUInteger ntex) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)encRef;
		[enc setRenderPipelineState:(__bridge id<MTLRenderPipelineState>)pipeRef];
		[enc setVertexBuffer:(__bridge id<MTLBuffer>)bufRef offset:off atIndex:0];
		for (NSUInteger i = 0; i < ntex; i++) {
			[enc setFragmentTexture:(__bridge id<MTLTexture>)texs[i] atIndex:i];
			[enc setFragmentSamplerState:(__bridge id<MTLSamplerState>)samplers[i] atIndex:i];
		}
		[enc drawPrimitives:MTLPrimitiveTypeTriangleStrip vertexStart:0 vertexCount:4 instanceCount:n];
	}
}
*/
import "C"

import (
	"unsafe"

	"gioui.org/gpu/internal/driver"
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
};

struct Varyings {
	float4 pos [[position]];
	float2 uv;
	float4 color;
	int slot [[flat]];
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
	return out;
}

fragment float4 quad_frag(Varyings in [[stage_in]], array<texture2d<float>, 8> texs [[texture(0)]], array<sampler, 8> smps [[sampler(0)]]) {
	if (in.slot < 0) {
		return in.color;
	}
	return in.color * texs[in.slot].sample(smps[in.slot], in.uv);
}
`

const quadSize = int(unsafe.Sizeof(driver.Quad{}))

type quadBatcher struct {
	failed   bool
	lib      C.CFTypeRef
	pipeline C.CFTypeRef
	format   C.MTLPixelFormat
	buf      C.CFTypeRef
	off      int
	texs     [driver.MaxQuadTextures]C.CFTypeRef
	samplers [driver.MaxQuadTextures]C.CFTypeRef
}

// frameStarted runs after BeginFrame waited for the previous frame: the
// instance buffer can be written from the start again.
func (q *quadBatcher) frameStarted() { q.off = 0 }

func (q *quadBatcher) release() {
	for _, r := range []C.CFTypeRef{q.lib, q.pipeline, q.buf} {
		if r != 0 {
			C.CFRelease(r)
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
		src := C.CString(quadSource)
		q.lib = C.quadLibrary(b.dev, src)
		C.free(unsafe.Pointer(src))
		if q.lib == 0 {
			q.failed = true
			return false
		}
	}
	if q.pipeline != 0 {
		C.CFRelease(q.pipeline)
	}
	frag := C.CString("quad_frag")
	q.pipeline = C.quadPipeline(b.dev, q.lib, frag, b.pixelFmt)
	C.free(unsafe.Pointer(frag))
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
			C.CFRelease(q.buf)
		}
		size := max(64<<10, 2*(q.off+n))
		q.buf = C.quadNewBuffer(b.dev, C.NSUInteger(size))
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
	C.quadDraw(b.renderEnc, q.pipeline, q.buf, C.NSUInteger(q.off), C.NSUInteger(len(quads)), &q.texs[0], &q.samplers[0], C.NSUInteger(len(textures)))
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
	buf := C.quadNewBuffer(b.dev, C.NSUInteger(max(64<<10, 2*size)))
	if buf == 0 {
		return false
	}
	if q.buf != 0 {
		C.CFRelease(q.buf)
	}
	q.buf, q.off = buf, 0
	return true
}
