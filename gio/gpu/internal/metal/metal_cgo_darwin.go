// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && (ios || (cgo && !purego))

package metal

// Gio's Objective-C Metal calls, behind the handle-based functions that
// metal_darwin.go uses. metal_purego_darwin.go sends the same messages
// without cgo. Keel patch: the helpers are Gio's; the Go wrappers are new.

/*
#cgo CFLAGS: -Werror -xobjective-c -fobjc-arc
#cgo LDFLAGS: -framework CoreGraphics -framework Metal -framework Foundation

#include <CoreFoundation/CoreFoundation.h>
#include <Metal/Metal.h>
#include <stdlib.h>

typedef struct {
	void *addr;
	NSUInteger size;
} slice;

typedef struct {
	CFTypeRef obj;
	char *err;
} CFTypeAndError;

static CFTypeAndError error(NSError *err) {
	char *errStr = strdup(err.localizedDescription.UTF8String);
	if (errStr == NULL) {
		abort();
	}
	CFTypeAndError res = {
		.err = errStr,
	};
	return res;
}

static CFTypeAndError ok(id obj) {
	CFTypeAndError res = {
		.obj = CFBridgingRetain(obj),
	};
	return res;
}

static void textureSetVolatile(CFTypeRef texRef, int volatile_) {
	@autoreleasepool {
		id<MTLTexture> tex = (__bridge id<MTLTexture>)texRef;
		[tex setPurgeableState:volatile_ ? MTLPurgeableStateVolatile : MTLPurgeableStateNonVolatile];
	}
}

static CFTypeRef queueNewBuffer(CFTypeRef queueRef) {
	@autoreleasepool {
		id<MTLCommandQueue> queue = (__bridge id<MTLCommandQueue>)queueRef;
		return CFBridgingRetain([queue commandBuffer]);
	}
}

static void cmdBufferCommit(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		[cmdBuf commit];
	}
}

static void cmdBufferWaitUntilCompleted(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		[cmdBuf waitUntilCompleted];
	}
}

static CFTypeRef cmdBufferRenderEncoder(CFTypeRef cmdBufRef, CFTypeRef textureRef, MTLLoadAction act, float r, float g, float b, float a) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		MTLRenderPassDescriptor *desc = [MTLRenderPassDescriptor new];
		desc.colorAttachments[0].texture = (__bridge id<MTLTexture>)textureRef;
		desc.colorAttachments[0].loadAction = act;
		desc.colorAttachments[0].clearColor = MTLClearColorMake(r, g, b, a);
		return CFBridgingRetain([cmdBuf renderCommandEncoderWithDescriptor:desc]);
	}
}

static CFTypeRef cmdBufferComputeEncoder(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		return CFBridgingRetain([cmdBuf computeCommandEncoder]);
	}
}

static CFTypeRef cmdBufferBlitEncoder(CFTypeRef cmdBufRef) {
	@autoreleasepool {
		id<MTLCommandBuffer> cmdBuf = (__bridge id<MTLCommandBuffer>)cmdBufRef;
		return CFBridgingRetain([cmdBuf blitCommandEncoder]);
	}
}

static void renderEncEnd(CFTypeRef renderEncRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc endEncoding];
	}
}

static void renderEncViewport(CFTypeRef renderEncRef, MTLViewport viewport) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc setViewport:viewport];
	}
}

static void renderEncSetFragmentTexture(CFTypeRef renderEncRef, NSUInteger index, CFTypeRef texRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLTexture> tex = (__bridge id<MTLTexture>)texRef;
		[enc setFragmentTexture:tex atIndex:index];
	}
}

static void renderEncSetFragmentSamplerState(CFTypeRef renderEncRef, NSUInteger index, CFTypeRef samplerRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLSamplerState> sampler = (__bridge id<MTLSamplerState>)samplerRef;
		[enc setFragmentSamplerState:sampler atIndex:index];
	}
}

static void renderEncSetVertexBuffer(CFTypeRef renderEncRef, CFTypeRef bufRef, NSUInteger idx, NSUInteger offset) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc setVertexBuffer:buf offset:offset atIndex:idx];
	}
}

static void renderEncSetFragmentBuffer(CFTypeRef renderEncRef, CFTypeRef bufRef, NSUInteger idx, NSUInteger offset) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc setFragmentBuffer:buf offset:offset atIndex:idx];
	}
}

static void renderEncSetFragmentBytes(CFTypeRef renderEncRef, const void *bytes, NSUInteger length, NSUInteger idx) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc setFragmentBytes:bytes length:length atIndex:idx];
	}
}

static void renderEncSetVertexBytes(CFTypeRef renderEncRef, const void *bytes, NSUInteger length, NSUInteger idx) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc setVertexBytes:bytes length:length atIndex:idx];
	}
}

static void renderEncSetRenderPipelineState(CFTypeRef renderEncRef, CFTypeRef pipeRef) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLRenderPipelineState> pipe = (__bridge id<MTLRenderPipelineState>)pipeRef;
		[enc setRenderPipelineState:pipe];
	}
}

static void renderEncDrawPrimitives(CFTypeRef renderEncRef, MTLPrimitiveType type, NSUInteger start, NSUInteger count) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		[enc drawPrimitives:type vertexStart:start vertexCount:count];
	}
}

static void renderEncDrawIndexedPrimitives(CFTypeRef renderEncRef, MTLPrimitiveType type, CFTypeRef bufRef, NSUInteger offset, NSUInteger count) {
	@autoreleasepool {
		id<MTLRenderCommandEncoder> enc = (__bridge id<MTLRenderCommandEncoder>)renderEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc drawIndexedPrimitives:type indexCount:count indexType:MTLIndexTypeUInt16 indexBuffer:buf indexBufferOffset:offset];
	}
}

static void computeEncSetPipeline(CFTypeRef computeEncRef, CFTypeRef pipeRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		id<MTLComputePipelineState> pipe = (__bridge id<MTLComputePipelineState>)pipeRef;
		[enc setComputePipelineState:pipe];
	}
}

static void computeEncSetTexture(CFTypeRef computeEncRef, NSUInteger index, CFTypeRef texRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		id<MTLTexture> tex = (__bridge id<MTLTexture>)texRef;
		[enc setTexture:tex atIndex:index];
	}
}

static void computeEncEnd(CFTypeRef computeEncRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		[enc endEncoding];
	}
}

static void computeEncSetBuffer(CFTypeRef computeEncRef, NSUInteger index, CFTypeRef bufRef) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		[enc setBuffer:buf offset:0 atIndex:index];
	}
}

static void computeEncDispatch(CFTypeRef computeEncRef, MTLSize threadgroupsPerGrid, MTLSize threadsPerThreadgroup) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		[enc dispatchThreadgroups:threadgroupsPerGrid threadsPerThreadgroup:threadsPerThreadgroup];
	}
}

static void computeEncSetBytes(CFTypeRef computeEncRef, const void *bytes, NSUInteger length, NSUInteger index) {
	@autoreleasepool {
		id<MTLComputeCommandEncoder> enc = (__bridge id<MTLComputeCommandEncoder>)computeEncRef;
		[enc setBytes:bytes length:length atIndex:index];
	}
}

static void blitEncEnd(CFTypeRef blitEncRef) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		[enc endEncoding];
	}
}

static void blitEncCopyFromTexture(CFTypeRef blitEncRef, CFTypeRef srcRef, MTLOrigin srcOrig, MTLSize srcSize, CFTypeRef dstRef, MTLOrigin dstOrig) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLTexture> src = (__bridge id<MTLTexture>)srcRef;
		id<MTLTexture> dst = (__bridge id<MTLTexture>)dstRef;
		[enc copyFromTexture:src
				 sourceSlice:0
				 sourceLevel:0
			    sourceOrigin:srcOrig
				  sourceSize:srcSize
				   toTexture:dst
			destinationSlice:0
			destinationLevel:0
		   destinationOrigin:dstOrig];
	}
}

static void blitEncCopyBufferToTexture(CFTypeRef blitEncRef, CFTypeRef bufRef, CFTypeRef texRef, NSUInteger offset, NSUInteger stride, NSUInteger length, MTLSize dims, MTLOrigin orig) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLBuffer> src = (__bridge id<MTLBuffer>)bufRef;
		id<MTLTexture> dst = (__bridge id<MTLTexture>)texRef;
		[enc copyFromBuffer:src
			   sourceOffset:offset
		  sourceBytesPerRow:stride
		sourceBytesPerImage:length
				 sourceSize:dims
				  toTexture:dst
		   destinationSlice:0
		   destinationLevel:0
		  destinationOrigin:orig];
	}
}

static void blitEncGenerateMipmapsForTexture(CFTypeRef blitEncRef, CFTypeRef texRef) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLTexture> tex = (__bridge id<MTLTexture>)texRef;
		[enc generateMipmapsForTexture: tex];
	}
}

static void blitEncCopyTextureToBuffer(CFTypeRef blitEncRef, CFTypeRef texRef, CFTypeRef bufRef, NSUInteger offset, NSUInteger stride, NSUInteger length, MTLSize dims, MTLOrigin orig) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLTexture> src = (__bridge id<MTLTexture>)texRef;
		id<MTLBuffer> dst = (__bridge id<MTLBuffer>)bufRef;
		[enc		 copyFromTexture:src
						 sourceSlice:0
						 sourceLevel:0
						sourceOrigin:orig
						  sourceSize:dims
							toBuffer:dst
				   destinationOffset:offset
			  destinationBytesPerRow:stride
			destinationBytesPerImage:length];
	}
}

static void blitEncCopyBufferToBuffer(CFTypeRef blitEncRef, CFTypeRef srcRef, CFTypeRef dstRef, NSUInteger srcOff, NSUInteger dstOff, NSUInteger size) {
	@autoreleasepool {
		id<MTLBlitCommandEncoder> enc = (__bridge id<MTLBlitCommandEncoder>)blitEncRef;
		id<MTLBuffer> src = (__bridge id<MTLBuffer>)srcRef;
		id<MTLBuffer> dst = (__bridge id<MTLBuffer>)dstRef;
		[enc   copyFromBuffer:src
				 sourceOffset:srcOff
					 toBuffer:dst
			destinationOffset:dstOff
						 size:size];
	}
}

static CFTypeRef newTexture(CFTypeRef devRef, NSUInteger width, NSUInteger height, MTLPixelFormat format, MTLTextureUsage usage, int mipmapped) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		MTLTextureDescriptor *mtlDesc = [MTLTextureDescriptor texture2DDescriptorWithPixelFormat: format
																						   width: width
																						  height: height
																				   	   mipmapped: mipmapped ? YES : NO];
		mtlDesc.usage = usage;
		mtlDesc.storageMode =  MTLStorageModePrivate;
		return CFBridgingRetain([dev newTextureWithDescriptor:mtlDesc]);
	}
}

static CFTypeRef newSampler(CFTypeRef devRef, MTLSamplerMinMagFilter minFilter, MTLSamplerMinMagFilter magFilter, MTLSamplerMipFilter mipFilter) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		MTLSamplerDescriptor *desc = [MTLSamplerDescriptor new];
		desc.minFilter = minFilter;
		desc.magFilter = magFilter;
		desc.mipFilter = mipFilter;
		return CFBridgingRetain([dev newSamplerStateWithDescriptor:desc]);
	}
}

static CFTypeRef newBuffer(CFTypeRef devRef, NSUInteger size, MTLResourceOptions opts) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		id<MTLBuffer> buf = [dev newBufferWithLength:size
											 options:opts];
		return CFBridgingRetain(buf);
	}
}

static slice bufferContents(CFTypeRef bufRef) {
	@autoreleasepool {
		id<MTLBuffer> buf = (__bridge id<MTLBuffer>)bufRef;
		slice s = {.addr = [buf contents], .size = [buf length]};
		return s;
	}
}

static CFTypeAndError newLibrary(CFTypeRef devRef, char *name, void *mtllib, size_t size) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		dispatch_data_t data = dispatch_data_create(mtllib, size, DISPATCH_TARGET_QUEUE_DEFAULT, DISPATCH_DATA_DESTRUCTOR_DEFAULT);
		NSError *err = NULL;
		id<MTLLibrary> lib = [dev newLibraryWithData:data error:&err];
		if (lib == nil) {
			return error(err);
		}
		lib.label = [NSString stringWithUTF8String:name];
		return ok(lib);
	}
}

static CFTypeRef libraryNewFunction(CFTypeRef libRef, char *funcName) {
	@autoreleasepool {
		id<MTLLibrary> lib = (__bridge id<MTLLibrary>)libRef;
		NSString *name = [NSString stringWithUTF8String:funcName];
		return CFBridgingRetain([lib newFunctionWithName:name]);
	}
}

static CFTypeAndError newComputePipeline(CFTypeRef devRef, CFTypeRef funcRef) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		id<MTLFunction> func = (__bridge id<MTLFunction>)funcRef;
		NSError *err = NULL;
		id<MTLComputePipelineState> s = [dev newComputePipelineStateWithFunction:func error:&err];
		if (s == nil) {
			return error(err);
		}
		return ok(s);
	}
}

static CFTypeAndError newRenderPipeline(CFTypeRef devRef, CFTypeRef vertFunc, CFTypeRef fragFunc, MTLPixelFormat pixelFormat, NSUInteger bufIdx, NSUInteger nverts, MTLVertexFormat *fmts, NSUInteger *offsets, NSUInteger stride, int blend, MTLBlendFactor srcFactor, MTLBlendFactor dstFactor, NSUInteger nvertBufs, NSUInteger nfragBufs) {
	@autoreleasepool {
		id<MTLDevice> dev = (__bridge id<MTLDevice>)devRef;
		id<MTLFunction> vfunc = (__bridge id<MTLFunction>)vertFunc;
		id<MTLFunction> ffunc = (__bridge id<MTLFunction>)fragFunc;
		MTLVertexDescriptor *vdesc = [MTLVertexDescriptor vertexDescriptor];
		vdesc.layouts[bufIdx].stride = stride;
		for (NSUInteger i = 0; i < nverts; i++) {
			vdesc.attributes[i].format = fmts[i];
			vdesc.attributes[i].offset = offsets[i];
			vdesc.attributes[i].bufferIndex = bufIdx;
		}
		MTLRenderPipelineDescriptor *desc = [MTLRenderPipelineDescriptor new];
		desc.vertexFunction = vfunc;
		desc.fragmentFunction = ffunc;
		desc.vertexDescriptor = vdesc;
		for (NSUInteger i = 0; i < nvertBufs; i++) {
			if (@available(iOS 11.0, *)) {
				desc.vertexBuffers[i].mutability = MTLMutabilityImmutable;
			}
		}
		for (NSUInteger i = 0; i < nfragBufs; i++) {
			if (@available(iOS 11.0, *)) {
				desc.fragmentBuffers[i].mutability = MTLMutabilityImmutable;
			}
		}
		desc.colorAttachments[0].pixelFormat = pixelFormat;
		desc.colorAttachments[0].blendingEnabled = blend ? YES : NO;
		desc.colorAttachments[0].sourceAlphaBlendFactor = srcFactor;
		desc.colorAttachments[0].sourceRGBBlendFactor = srcFactor;
		desc.colorAttachments[0].destinationAlphaBlendFactor = dstFactor;
		desc.colorAttachments[0].destinationRGBBlendFactor = dstFactor;
		NSError *err = NULL;
		id<MTLRenderPipelineState> s = [dev newRenderPipelineStateWithDescriptor:desc
											error:&err];
		if (s == nil) {
			return error(err);
		}
		return ok(s);
	}
}
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
	"errors"
	"unsafe"

	"github.com/dyike/keel/gio/gpu/internal/driver"
)

// The enumeration values metal_darwin.go spells out must match the SDK.
func init() {
	for _, c := range [...][2]uint64{
		{C.MTLPixelFormatR8Unorm, mtlPixelFormatR8Unorm},
		{C.MTLPixelFormatR16Float, mtlPixelFormatR16Float},
		{C.MTLPixelFormatRGBA8Unorm, mtlPixelFormatRGBA8Unorm},
		{C.MTLPixelFormatRGBA8Unorm_sRGB, mtlPixelFormatRGBA8Unorm_sRGB},
		{C.MTLTextureUsageShaderRead, mtlTextureUsageShaderRead},
		{C.MTLTextureUsageShaderWrite, mtlTextureUsageShaderWrite},
		{C.MTLTextureUsageRenderTarget, mtlTextureUsageRenderTarget},
		{C.MTLSamplerMinMagFilterNearest, mtlSamplerMinMagFilterNearest},
		{C.MTLSamplerMinMagFilterLinear, mtlSamplerMinMagFilterLinear},
		{C.MTLSamplerMipFilterNotMipmapped, mtlSamplerMipFilterNotMipmapped},
		{C.MTLSamplerMipFilterLinear, mtlSamplerMipFilterLinear},
		{C.MTLBlendFactorZero, mtlBlendFactorZero},
		{C.MTLBlendFactorOne, mtlBlendFactorOne},
		{C.MTLBlendFactorOneMinusSourceAlpha, mtlBlendFactorOneMinusSourceAlpha},
		{C.MTLBlendFactorDestinationColor, mtlBlendFactorDestinationColor},
		{C.MTLVertexFormatFloat, mtlVertexFormatFloat},
		{C.MTLVertexFormatFloat2, mtlVertexFormatFloat2},
		{C.MTLVertexFormatFloat3, mtlVertexFormatFloat3},
		{C.MTLVertexFormatFloat4, mtlVertexFormatFloat4},
		{C.MTLLoadActionDontCare, mtlLoadActionDontCare},
		{C.MTLLoadActionLoad, mtlLoadActionLoad},
		{C.MTLLoadActionClear, mtlLoadActionClear},
		{C.MTLPrimitiveTypeTriangle, mtlPrimitiveTypeTriangle},
		{C.MTLPrimitiveTypeTriangleStrip, mtlPrimitiveTypeTriangleStrip},
		{C.MTLResourceCPUCacheModeWriteCombined, mtlResourceCPUCacheModeWriteCombined},
		{C.MTLResourceStorageModeShared, mtlResourceStorageModeShared},
		{C.MTLResourceStorageModePrivate, mtlResourceStorageModePrivate},
	} {
		if c[0] != c[1] {
			panic("metal: enumeration value differs from the SDK")
		}
	}
}

func loadMetal() error { return nil }

func ref(r uintptr) C.CFTypeRef { return C.CFTypeRef(r) }

func cfRetain(r uintptr)  { C.CFRetain(ref(r)) }
func cfRelease(r uintptr) { C.CFRelease(ref(r)) }

func splitErr(o C.CFTypeAndError) (uintptr, error) {
	if o.err != nil {
		defer C.free(unsafe.Pointer(o.err))
		return 0, errors.New(C.GoString(o.err))
	}
	return uintptr(o.obj), nil
}

func textureSetVolatile(tex uintptr, volatile bool) {
	v := C.int(0)
	if volatile {
		v = 1
	}
	C.textureSetVolatile(ref(tex), v)
}

func queueNewBuffer(queue uintptr) uintptr { return uintptr(C.queueNewBuffer(ref(queue))) }
func cmdBufferCommit(cb uintptr)           { C.cmdBufferCommit(ref(cb)) }
func cmdBufferWaitUntilCompleted(cb uintptr) {
	C.cmdBufferWaitUntilCompleted(ref(cb))
}

func cmdBufferRenderEncoder(cb, tex, act uintptr, r, g, b, a float32) uintptr {
	return uintptr(C.cmdBufferRenderEncoder(ref(cb), ref(tex), C.MTLLoadAction(act), C.float(r), C.float(g), C.float(b), C.float(a)))
}

func cmdBufferComputeEncoder(cb uintptr) uintptr { return uintptr(C.cmdBufferComputeEncoder(ref(cb))) }
func cmdBufferBlitEncoder(cb uintptr) uintptr    { return uintptr(C.cmdBufferBlitEncoder(ref(cb))) }

func renderEncEnd(enc uintptr) { C.renderEncEnd(ref(enc)) }

func renderEncViewport(enc uintptr, v mtlViewport) {
	C.renderEncViewport(ref(enc), C.MTLViewport{
		originX: C.double(v.OriginX),
		originY: C.double(v.OriginY),
		width:   C.double(v.Width),
		height:  C.double(v.Height),
		znear:   C.double(v.ZNear),
		zfar:    C.double(v.ZFar),
	})
}

func renderEncSetFragmentTexture(enc, index, tex uintptr) {
	C.renderEncSetFragmentTexture(ref(enc), C.NSUInteger(index), ref(tex))
}

func renderEncSetFragmentSamplerState(enc, index, sampler uintptr) {
	C.renderEncSetFragmentSamplerState(ref(enc), C.NSUInteger(index), ref(sampler))
}

func renderEncSetVertexBuffer(enc, buf, idx, offset uintptr) {
	C.renderEncSetVertexBuffer(ref(enc), ref(buf), C.NSUInteger(idx), C.NSUInteger(offset))
}

func renderEncSetFragmentBuffer(enc, buf, idx, offset uintptr) {
	C.renderEncSetFragmentBuffer(ref(enc), ref(buf), C.NSUInteger(idx), C.NSUInteger(offset))
}

func renderEncSetFragmentBytes(enc uintptr, bytes unsafe.Pointer, length, idx uintptr) {
	C.renderEncSetFragmentBytes(ref(enc), bytes, C.NSUInteger(length), C.NSUInteger(idx))
}

func renderEncSetVertexBytes(enc uintptr, bytes unsafe.Pointer, length, idx uintptr) {
	C.renderEncSetVertexBytes(ref(enc), bytes, C.NSUInteger(length), C.NSUInteger(idx))
}

func renderEncSetRenderPipelineState(enc, pipe uintptr) {
	C.renderEncSetRenderPipelineState(ref(enc), ref(pipe))
}

func renderEncDrawPrimitives(enc, typ, start, count uintptr) {
	C.renderEncDrawPrimitives(ref(enc), C.MTLPrimitiveType(typ), C.NSUInteger(start), C.NSUInteger(count))
}

func renderEncDrawIndexedPrimitives(enc, typ, buf, offset, count uintptr) {
	C.renderEncDrawIndexedPrimitives(ref(enc), C.MTLPrimitiveType(typ), ref(buf), C.NSUInteger(offset), C.NSUInteger(count))
}

func computeEncSetPipeline(enc, pipe uintptr) { C.computeEncSetPipeline(ref(enc), ref(pipe)) }
func computeEncSetTexture(enc, index, tex uintptr) {
	C.computeEncSetTexture(ref(enc), C.NSUInteger(index), ref(tex))
}
func computeEncEnd(enc uintptr) { C.computeEncEnd(ref(enc)) }
func computeEncSetBuffer(enc, index, buf uintptr) {
	C.computeEncSetBuffer(ref(enc), C.NSUInteger(index), ref(buf))
}

func computeEncDispatch(enc uintptr, groups, threads mtlSize) {
	C.computeEncDispatch(ref(enc), cSize(groups), cSize(threads))
}

func computeEncSetBytes(enc uintptr, bytes unsafe.Pointer, length, index uintptr) {
	C.computeEncSetBytes(ref(enc), bytes, C.NSUInteger(length), C.NSUInteger(index))
}

func cSize(s mtlSize) C.MTLSize {
	return C.MTLSize{width: C.NSUInteger(s.Width), height: C.NSUInteger(s.Height), depth: C.NSUInteger(s.Depth)}
}

func cOrigin(o mtlOrigin) C.MTLOrigin {
	return C.MTLOrigin{x: C.NSUInteger(o.X), y: C.NSUInteger(o.Y), z: C.NSUInteger(o.Z)}
}

func blitEncEnd(enc uintptr) { C.blitEncEnd(ref(enc)) }

func blitEncCopyFromTexture(enc, src uintptr, srcOrig mtlOrigin, srcSize mtlSize, dst uintptr, dstOrig mtlOrigin) {
	C.blitEncCopyFromTexture(ref(enc), ref(src), cOrigin(srcOrig), cSize(srcSize), ref(dst), cOrigin(dstOrig))
}

func blitEncCopyBufferToTexture(enc, buf, tex, offset, stride, length uintptr, dims mtlSize, orig mtlOrigin) {
	C.blitEncCopyBufferToTexture(ref(enc), ref(buf), ref(tex), C.NSUInteger(offset), C.NSUInteger(stride), C.NSUInteger(length), cSize(dims), cOrigin(orig))
}

func blitEncGenerateMipmapsForTexture(enc, tex uintptr) {
	C.blitEncGenerateMipmapsForTexture(ref(enc), ref(tex))
}

func blitEncCopyTextureToBuffer(enc, tex, buf, offset, stride, length uintptr, dims mtlSize, orig mtlOrigin) {
	C.blitEncCopyTextureToBuffer(ref(enc), ref(tex), ref(buf), C.NSUInteger(offset), C.NSUInteger(stride), C.NSUInteger(length), cSize(dims), cOrigin(orig))
}

func blitEncCopyBufferToBuffer(enc, src, dst, srcOff, dstOff, size uintptr) {
	C.blitEncCopyBufferToBuffer(ref(enc), ref(src), ref(dst), C.NSUInteger(srcOff), C.NSUInteger(dstOff), C.NSUInteger(size))
}

func newTexture(dev, width, height, format, usage uintptr, mipmapped bool) uintptr {
	m := C.int(0)
	if mipmapped {
		m = 1
	}
	return uintptr(C.newTexture(ref(dev), C.NSUInteger(width), C.NSUInteger(height), C.MTLPixelFormat(format), C.MTLTextureUsage(usage), m))
}

func newSampler(dev, minFilter, magFilter, mipFilter uintptr) uintptr {
	return uintptr(C.newSampler(ref(dev), C.MTLSamplerMinMagFilter(minFilter), C.MTLSamplerMinMagFilter(magFilter), C.MTLSamplerMipFilter(mipFilter)))
}

func newBuffer(dev, size, opts uintptr) uintptr {
	return uintptr(C.newBuffer(ref(dev), C.NSUInteger(size), C.MTLResourceOptions(opts)))
}

func bufferContents(buf uintptr) (unsafe.Pointer, int) {
	s := C.bufferContents(ref(buf))
	return s.addr, int(s.size)
}

func newLibrary(dev uintptr, name, data string) (uintptr, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return splitErr(C.newLibrary(ref(dev), cname, unsafe.Pointer(unsafe.StringData(data)), C.size_t(len(data))))
}

func libraryNewFunction(lib uintptr, name string) uintptr {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return uintptr(C.libraryNewFunction(ref(lib), cname))
}

func newComputePipeline(dev, fn uintptr) (uintptr, error) {
	return splitErr(C.newComputePipeline(ref(dev), ref(fn)))
}

func newRenderPipeline(dev, vertFunc, fragFunc, pixelFormat, bufIdx uintptr, fmts, offsets []uintptr, stride uintptr, blend bool, srcFactor, dstFactor, nvertBufs, nfragBufs uintptr) (uintptr, error) {
	b := C.int(0)
	if blend {
		b = 1
	}
	return splitErr(C.newRenderPipeline(
		ref(dev), ref(vertFunc), ref(fragFunc), C.MTLPixelFormat(pixelFormat), C.NSUInteger(bufIdx),
		C.NSUInteger(len(fmts)), (*C.MTLVertexFormat)(unsafe.Pointer(unsafe.SliceData(fmts))), (*C.NSUInteger)(unsafe.Pointer(unsafe.SliceData(offsets))),
		C.NSUInteger(stride), b, C.MTLBlendFactor(srcFactor), C.MTLBlendFactor(dstFactor),
		C.NSUInteger(nvertBufs), C.NSUInteger(nfragBufs),
	))
}

func quadLibrary(dev uintptr, src string) uintptr {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))
	return uintptr(C.quadLibrary(ref(dev), csrc))
}

func quadPipeline(dev, lib uintptr, frag string, format uintptr) uintptr {
	cfrag := C.CString(frag)
	defer C.free(unsafe.Pointer(cfrag))
	return uintptr(C.quadPipeline(ref(dev), ref(lib), cfrag, C.MTLPixelFormat(format)))
}

func quadNewBuffer(dev, size uintptr) uintptr {
	return uintptr(C.quadNewBuffer(ref(dev), C.NSUInteger(size)))
}

// quadDraw draws n instances; texs and samplers have the same length.
// The handles are copied: cgo checks the whole allocation a slice points
// into, and the batcher's lives inside the Backend.
func quadDraw(enc, pipe, buf, off, n uintptr, texs, samplers []uintptr) {
	var t, s [driver.MaxQuadTextures]C.CFTypeRef
	for i := range texs {
		t[i], s[i] = ref(texs[i]), ref(samplers[i])
	}
	C.quadDraw(ref(enc), ref(pipe), ref(buf), C.NSUInteger(off), C.NSUInteger(n), &t[0], &s[0], C.NSUInteger(len(texs)))
}
