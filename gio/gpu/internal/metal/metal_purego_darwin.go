// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && (!cgo || purego)

package metal

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// The Metal calls of metal_cgo_darwin.go without cgo: the same messages,
// sent through objc_msgSend at run time. Keel patch.
//
// Encoder calls run for every draw. purego.SyscallN allocates its argument
// slice, so they go through send, which reuses one argument array under a
// mutex. Pointers passed through send must point into memory the garbage
// collector cannot move (heap or package data) and stay alive until the
// call returns. Calls with pointers to locals, and calls that block, use
// purego.SyscallN directly.
//
// Structs: MTLOrigin, MTLSize and MTLViewport are larger than 16 bytes and
// not homogeneous float aggregates of at most four members, so arm64 passes
// a pointer to a copy; amd64 copies them onto the stack, which the
// reflective wrappers do. MTLClearColor is set once per cleared pass.

type mtlClearColor struct{ R, G, B, A float64 }

const (
	mtlStorageModePrivate         = 2
	mtlIndexTypeUInt16            = 0
	mtlMutabilityImmutable        = 2
	mtlPurgeableStateNonVolatile  = 2
	mtlPurgeableStateVolatile     = 3
	maxSendArgs                   = 12
	structArgsByPointer           = runtime.GOARCH == "arm64"
	dispatchDataDestructorDefault = 0
)

var mtl struct {
	once sync.Once
	err  error

	msg, poolPush, poolPop  uintptr
	cfRetain, cfRelease     uintptr
	dataCreate, dataRelease uintptr

	// Classes.
	renderPassDesc, textureDesc, samplerDesc uintptr
	vertexDesc, pipelineDesc, nsString       uintptr
	sel                                      selectors
	setClearColor                            func(id, sel uintptr, c mtlClearColor)
	setViewport                              func(id, sel uintptr, v mtlViewport)
	dispatch                                 func(id, sel uintptr, groups, threads mtlSize)
	copyTexToTex                             func(id, sel, src, srcSlice, srcLevel uintptr, srcOrig mtlOrigin, srcSize mtlSize, dst, dstSlice, dstLevel uintptr, dstOrig mtlOrigin)
	copyBufToTex                             func(id, sel, buf, off, stride, length uintptr, size mtlSize, tex, slice, level uintptr, orig mtlOrigin)
	copyTexToBuf                             func(id, sel, tex, slice, level uintptr, orig mtlOrigin, size mtlSize, buf, off, stride, length uintptr)
}

// call holds the reused argument array of send, and the struct copies arm64
// passes by pointer.
var call struct {
	sync.Mutex
	args    [maxSendArgs]uintptr
	origins [2]mtlOrigin
	sizes   [2]mtlSize
	view    mtlViewport
}

type selectors struct {
	commandBuffer, commit, waitUntilCompleted                        uintptr
	renderEncoder, computeEncoder, blitEncoder, endEncoding          uintptr
	setViewport, setFragmentTexture, setFragmentSampler              uintptr
	setVertexBuffer, setFragmentBuffer                               uintptr
	setVertexBytes, setFragmentBytes, setRenderPipelineState         uintptr
	drawPrimitives, drawIndexedPrimitives, drawInstanced             uintptr
	setComputePipelineState, setTexture, setBuffer, setBytes         uintptr
	dispatch, copyTexToTex, copyBufToTex, copyTexToBuf, copyBufToBuf uintptr
	generateMipmaps                                                  uintptr
	texture2D, setUsage, setStorageMode, newTexture                  uintptr
	new, release, setMinFilter, setMagFilter, setMipFilter           uintptr
	newSampler, newBuffer, contents, length, setPurgeableState       uintptr
	newLibraryWithData, newLibraryWithSource, setLabel               uintptr
	stringWithUTF8String, localizedDescription, utf8String           uintptr
	newFunction, newComputePipeline, newRenderPipeline               uintptr
	vertexDescriptor, layouts, attributes, objectAtIndex             uintptr
	setStride, setFormat, setOffset, setBufferIndex                  uintptr
	setVertexFunction, setFragmentFunction, setVertexDescriptor      uintptr
	vertexBuffers, fragmentBuffers, setMutability                    uintptr
	colorAttachments, setPixelFormat, setBlendingEnabled             uintptr
	setSourceAlpha, setSourceRGB, setDestinationAlpha, setDestRGB    uintptr
	setTextureAttachment, setLoadAction, setClearColor               uintptr
}

func loadMetal() error {
	mtl.once.Do(func() { mtl.err = initMetal() })
	return mtl.err
}

func initMetal() error {
	if _, err := purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_NOW|purego.RTLD_GLOBAL); err != nil {
		return err
	}
	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return err
	}
	for _, sym := range []struct {
		lib  uintptr
		name string
		addr *uintptr
	}{
		{purego.RTLD_DEFAULT, "objc_msgSend", &mtl.msg},
		{purego.RTLD_DEFAULT, "objc_autoreleasePoolPush", &mtl.poolPush},
		{purego.RTLD_DEFAULT, "objc_autoreleasePoolPop", &mtl.poolPop},
		{purego.RTLD_DEFAULT, "dispatch_data_create", &mtl.dataCreate},
		{purego.RTLD_DEFAULT, "dispatch_release", &mtl.dataRelease},
		{cf, "CFRetain", &mtl.cfRetain},
		{cf, "CFRelease", &mtl.cfRelease},
	} {
		if *sym.addr, err = purego.Dlsym(sym.lib, sym.name); err != nil {
			return err
		}
	}
	for _, c := range []struct {
		name string
		addr *uintptr
	}{
		{"MTLRenderPassDescriptor", &mtl.renderPassDesc},
		{"MTLTextureDescriptor", &mtl.textureDesc},
		{"MTLSamplerDescriptor", &mtl.samplerDesc},
		{"MTLVertexDescriptor", &mtl.vertexDesc},
		{"MTLRenderPipelineDescriptor", &mtl.pipelineDesc},
		{"NSString", &mtl.nsString},
	} {
		if *c.addr = uintptr(objc.GetClass(c.name)); *c.addr == 0 {
			return errors.New("metal: class " + c.name + " not found")
		}
	}
	s := &mtl.sel
	for _, n := range []struct {
		addr *uintptr
		name string
	}{
		{&s.commandBuffer, "commandBuffer"},
		{&s.commit, "commit"},
		{&s.waitUntilCompleted, "waitUntilCompleted"},
		{&s.renderEncoder, "renderCommandEncoderWithDescriptor:"},
		{&s.computeEncoder, "computeCommandEncoder"},
		{&s.blitEncoder, "blitCommandEncoder"},
		{&s.endEncoding, "endEncoding"},
		{&s.setViewport, "setViewport:"},
		{&s.setFragmentTexture, "setFragmentTexture:atIndex:"},
		{&s.setFragmentSampler, "setFragmentSamplerState:atIndex:"},
		{&s.setVertexBuffer, "setVertexBuffer:offset:atIndex:"},
		{&s.setFragmentBuffer, "setFragmentBuffer:offset:atIndex:"},
		{&s.setVertexBytes, "setVertexBytes:length:atIndex:"},
		{&s.setFragmentBytes, "setFragmentBytes:length:atIndex:"},
		{&s.setRenderPipelineState, "setRenderPipelineState:"},
		{&s.drawPrimitives, "drawPrimitives:vertexStart:vertexCount:"},
		{&s.drawIndexedPrimitives, "drawIndexedPrimitives:indexCount:indexType:indexBuffer:indexBufferOffset:"},
		{&s.drawInstanced, "drawPrimitives:vertexStart:vertexCount:instanceCount:"},
		{&s.setComputePipelineState, "setComputePipelineState:"},
		{&s.setTexture, "setTexture:atIndex:"},
		{&s.setBuffer, "setBuffer:offset:atIndex:"},
		{&s.setBytes, "setBytes:length:atIndex:"},
		{&s.dispatch, "dispatchThreadgroups:threadsPerThreadgroup:"},
		{&s.copyTexToTex, "copyFromTexture:sourceSlice:sourceLevel:sourceOrigin:sourceSize:toTexture:destinationSlice:destinationLevel:destinationOrigin:"},
		{&s.copyBufToTex, "copyFromBuffer:sourceOffset:sourceBytesPerRow:sourceBytesPerImage:sourceSize:toTexture:destinationSlice:destinationLevel:destinationOrigin:"},
		{&s.copyTexToBuf, "copyFromTexture:sourceSlice:sourceLevel:sourceOrigin:sourceSize:toBuffer:destinationOffset:destinationBytesPerRow:destinationBytesPerImage:"},
		{&s.copyBufToBuf, "copyFromBuffer:sourceOffset:toBuffer:destinationOffset:size:"},
		{&s.generateMipmaps, "generateMipmapsForTexture:"},
		{&s.texture2D, "texture2DDescriptorWithPixelFormat:width:height:mipmapped:"},
		{&s.setUsage, "setUsage:"},
		{&s.setStorageMode, "setStorageMode:"},
		{&s.newTexture, "newTextureWithDescriptor:"},
		{&s.new, "new"},
		{&s.release, "release"},
		{&s.setMinFilter, "setMinFilter:"},
		{&s.setMagFilter, "setMagFilter:"},
		{&s.setMipFilter, "setMipFilter:"},
		{&s.newSampler, "newSamplerStateWithDescriptor:"},
		{&s.newBuffer, "newBufferWithLength:options:"},
		{&s.contents, "contents"},
		{&s.length, "length"},
		{&s.setPurgeableState, "setPurgeableState:"},
		{&s.newLibraryWithData, "newLibraryWithData:error:"},
		{&s.newLibraryWithSource, "newLibraryWithSource:options:error:"},
		{&s.setLabel, "setLabel:"},
		{&s.stringWithUTF8String, "stringWithUTF8String:"},
		{&s.localizedDescription, "localizedDescription"},
		{&s.utf8String, "UTF8String"},
		{&s.newFunction, "newFunctionWithName:"},
		{&s.newComputePipeline, "newComputePipelineStateWithFunction:error:"},
		{&s.newRenderPipeline, "newRenderPipelineStateWithDescriptor:error:"},
		{&s.vertexDescriptor, "vertexDescriptor"},
		{&s.layouts, "layouts"},
		{&s.attributes, "attributes"},
		{&s.objectAtIndex, "objectAtIndexedSubscript:"},
		{&s.setStride, "setStride:"},
		{&s.setFormat, "setFormat:"},
		{&s.setOffset, "setOffset:"},
		{&s.setBufferIndex, "setBufferIndex:"},
		{&s.setVertexFunction, "setVertexFunction:"},
		{&s.setFragmentFunction, "setFragmentFunction:"},
		{&s.setVertexDescriptor, "setVertexDescriptor:"},
		{&s.vertexBuffers, "vertexBuffers"},
		{&s.fragmentBuffers, "fragmentBuffers"},
		{&s.setMutability, "setMutability:"},
		{&s.colorAttachments, "colorAttachments"},
		{&s.setPixelFormat, "setPixelFormat:"},
		{&s.setBlendingEnabled, "setBlendingEnabled:"},
		{&s.setSourceAlpha, "setSourceAlphaBlendFactor:"},
		{&s.setSourceRGB, "setSourceRGBBlendFactor:"},
		{&s.setDestinationAlpha, "setDestinationAlphaBlendFactor:"},
		{&s.setDestRGB, "setDestinationRGBBlendFactor:"},
		{&s.setTextureAttachment, "setTexture:"},
		{&s.setLoadAction, "setLoadAction:"},
		{&s.setClearColor, "setClearColor:"},
	} {
		*n.addr = uintptr(objc.RegisterName(n.name))
	}
	purego.RegisterFunc(&mtl.setClearColor, mtl.msg)
	purego.RegisterFunc(&mtl.setViewport, mtl.msg)
	purego.RegisterFunc(&mtl.dispatch, mtl.msg)
	purego.RegisterFunc(&mtl.copyTexToTex, mtl.msg)
	purego.RegisterFunc(&mtl.copyBufToTex, mtl.msg)
	purego.RegisterFunc(&mtl.copyTexToBuf, mtl.msg)
	return nil
}

// send sends sel to id with integer arguments without allocating.
func send(id, sel uintptr, args ...uintptr) uintptr {
	call.Lock()
	defer call.Unlock()
	return sendLocked(id, sel, args...)
}

func sendLocked(id, sel uintptr, args ...uintptr) uintptr {
	call.args[0], call.args[1] = id, sel
	n := 2 + copy(call.args[2:], args)
	r, _, _ := purego.SyscallN(mtl.msg, call.args[:n]...)
	return r
}

// ccall calls a C function with integer arguments without allocating. It
// must not block.
func ccall(fn uintptr, args ...uintptr) uintptr {
	call.Lock()
	defer call.Unlock()
	n := copy(call.args[:], args)
	r, _, _ := purego.SyscallN(fn, call.args[:n]...)
	return r
}

// pushPool locks the goroutine to its thread, which the pool belongs to,
// and pushes an autorelease pool; popPool undoes both.
func pushPool() uintptr {
	runtime.LockOSThread()
	return ccall(mtl.poolPush)
}

func popPool(p uintptr) {
	ccall(mtl.poolPop, p)
	runtime.UnlockOSThread()
}

// retained returns obj with a reference the caller owns, as
// CFBridgingRetain does for an autoreleased result.
func retained(obj uintptr) uintptr {
	if obj != 0 {
		ccall(mtl.cfRetain, obj)
	}
	return obj
}

func boolArg(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

// cString returns s NUL-terminated in Go memory.
func cString(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

// goString copies a NUL-terminated C string.
func goString(p uintptr) string {
	if p == 0 {
		return ""
	}
	ptr := *(*unsafe.Pointer)(unsafe.Pointer(&p))
	n := 0
	for *(*byte)(unsafe.Add(ptr, n)) != 0 {
		n++
	}
	return string(unsafe.Slice((*byte)(ptr), n))
}

// nsString returns an autoreleased NSString; call inside a pool.
func nsString(s string) uintptr {
	b := cString(s)
	r := send(mtl.nsString, mtl.sel.stringWithUTF8String, uintptr(unsafe.Pointer(&b[0])))
	runtime.KeepAlive(b)
	return r
}

// objErr returns obj, or the NSError's description; call inside a pool.
func objErr(obj, nserr uintptr) (uintptr, error) {
	if obj != 0 {
		return obj, nil
	}
	if nserr == 0 {
		return 0, errors.New("metal: unknown error")
	}
	desc := send(nserr, mtl.sel.localizedDescription)
	return 0, errors.New(goString(send(desc, mtl.sel.utf8String)))
}

func cfRetain(r uintptr)  { ccall(mtl.cfRetain, r) }
func cfRelease(r uintptr) { ccall(mtl.cfRelease, r) }

func textureSetVolatile(tex uintptr, volatile bool) {
	state := uintptr(mtlPurgeableStateNonVolatile)
	if volatile {
		state = mtlPurgeableStateVolatile
	}
	send(tex, mtl.sel.setPurgeableState, state)
}

func queueNewBuffer(queue uintptr) uintptr {
	p := pushPool()
	defer popPool(p)
	return retained(send(queue, mtl.sel.commandBuffer))
}

func cmdBufferCommit(cb uintptr) {
	p := pushPool()
	defer popPool(p)
	send(cb, mtl.sel.commit)
}

// cmdBufferWaitUntilCompleted blocks, so it does not hold send's lock; its
// argument slice is the one allocation a frame the cgo binding avoids.
func cmdBufferWaitUntilCompleted(cb uintptr) {
	p := pushPool()
	defer popPool(p)
	purego.SyscallN(mtl.msg, cb, mtl.sel.waitUntilCompleted)
}

// passes are render pass descriptors, reused: an encoder copies its
// descriptor. Setting a clear color takes a reflective call, and a frame
// alternates between a few colors (the window's, the stencil's, the
// intersection's), so each descriptor keeps one.
var passes struct {
	sync.Mutex
	descs [4]passDesc
	next  int
}

type passDesc struct {
	desc, att uintptr
	color     mtlClearColor
}

// passFor returns a descriptor clearing to c; call with passes locked.
func passFor(c mtlClearColor) *passDesc {
	for i := range passes.descs {
		if d := &passes.descs[i]; d.desc != 0 && d.color == c {
			return d
		}
	}
	d := &passes.descs[passes.next]
	passes.next = (passes.next + 1) % len(passes.descs)
	if d.desc == 0 {
		d.desc = send(mtl.renderPassDesc, mtl.sel.new)
		d.att = send(send(d.desc, mtl.sel.colorAttachments), mtl.sel.objectAtIndex, 0)
	}
	mtl.setClearColor(d.att, mtl.sel.setClearColor, c)
	d.color = c
	return d
}

func cmdBufferRenderEncoder(cb, tex, act uintptr, r, g, b, a float32) uintptr {
	p := pushPool()
	defer popPool(p)
	passes.Lock()
	defer passes.Unlock()
	// MTLClearColorMake takes doubles; the cgo helper widens floats. A pass
	// that does not clear ignores the color.
	d := &passes.descs[0]
	if c := (mtlClearColor{float64(r), float64(g), float64(b), float64(a)}); act == mtlLoadActionClear || d.desc == 0 {
		d = passFor(c)
	}
	send(d.att, mtl.sel.setTextureAttachment, tex)
	send(d.att, mtl.sel.setLoadAction, act)
	enc := retained(send(cb, mtl.sel.renderEncoder, d.desc))
	// Keep no target alive, such as a window's drawable.
	send(d.att, mtl.sel.setTextureAttachment, 0)
	return enc
}

func cmdBufferComputeEncoder(cb uintptr) uintptr {
	p := pushPool()
	defer popPool(p)
	return retained(send(cb, mtl.sel.computeEncoder))
}

func cmdBufferBlitEncoder(cb uintptr) uintptr {
	p := pushPool()
	defer popPool(p)
	return retained(send(cb, mtl.sel.blitEncoder))
}

func renderEncEnd(enc uintptr)  { send(enc, mtl.sel.endEncoding) }
func computeEncEnd(enc uintptr) { send(enc, mtl.sel.endEncoding) }
func blitEncEnd(enc uintptr)    { send(enc, mtl.sel.endEncoding) }

func renderEncViewport(enc uintptr, v mtlViewport) {
	if !structArgsByPointer {
		mtl.setViewport(enc, mtl.sel.setViewport, v)
		return
	}
	call.Lock()
	defer call.Unlock()
	call.view = v
	sendLocked(enc, mtl.sel.setViewport, uintptr(unsafe.Pointer(&call.view)))
}

func renderEncSetFragmentTexture(enc, index, tex uintptr) {
	send(enc, mtl.sel.setFragmentTexture, tex, index)
}

func renderEncSetFragmentSamplerState(enc, index, sampler uintptr) {
	send(enc, mtl.sel.setFragmentSampler, sampler, index)
}

func renderEncSetVertexBuffer(enc, buf, idx, offset uintptr) {
	send(enc, mtl.sel.setVertexBuffer, buf, offset, idx)
}

func renderEncSetFragmentBuffer(enc, buf, idx, offset uintptr) {
	send(enc, mtl.sel.setFragmentBuffer, buf, offset, idx)
}

// The bytes passed to the set*Bytes calls are heap memory of the caller,
// which keeps it alive; Metal copies it before returning.
func renderEncSetFragmentBytes(enc uintptr, bytes unsafe.Pointer, length, idx uintptr) {
	send(enc, mtl.sel.setFragmentBytes, uintptr(bytes), length, idx)
}

func renderEncSetVertexBytes(enc uintptr, bytes unsafe.Pointer, length, idx uintptr) {
	send(enc, mtl.sel.setVertexBytes, uintptr(bytes), length, idx)
}

func renderEncSetRenderPipelineState(enc, pipe uintptr) {
	send(enc, mtl.sel.setRenderPipelineState, pipe)
}

func renderEncDrawPrimitives(enc, typ, start, count uintptr) {
	send(enc, mtl.sel.drawPrimitives, typ, start, count)
}

func renderEncDrawIndexedPrimitives(enc, typ, buf, offset, count uintptr) {
	send(enc, mtl.sel.drawIndexedPrimitives, typ, count, mtlIndexTypeUInt16, buf, offset)
}

func computeEncSetPipeline(enc, pipe uintptr) {
	send(enc, mtl.sel.setComputePipelineState, pipe)
}

func computeEncSetTexture(enc, index, tex uintptr) {
	send(enc, mtl.sel.setTexture, tex, index)
}

func computeEncSetBuffer(enc, index, buf uintptr) {
	send(enc, mtl.sel.setBuffer, buf, 0, index)
}

func computeEncDispatch(enc uintptr, groups, threads mtlSize) {
	if !structArgsByPointer {
		mtl.dispatch(enc, mtl.sel.dispatch, groups, threads)
		return
	}
	call.Lock()
	defer call.Unlock()
	call.sizes[0], call.sizes[1] = groups, threads
	sendLocked(enc, mtl.sel.dispatch, uintptr(unsafe.Pointer(&call.sizes[0])), uintptr(unsafe.Pointer(&call.sizes[1])))
}

func computeEncSetBytes(enc uintptr, bytes unsafe.Pointer, length, index uintptr) {
	send(enc, mtl.sel.setBytes, uintptr(bytes), length, index)
}

func blitEncCopyFromTexture(enc, src uintptr, srcOrig mtlOrigin, srcSize mtlSize, dst uintptr, dstOrig mtlOrigin) {
	if !structArgsByPointer {
		mtl.copyTexToTex(enc, mtl.sel.copyTexToTex, src, 0, 0, srcOrig, srcSize, dst, 0, 0, dstOrig)
		return
	}
	call.Lock()
	defer call.Unlock()
	call.origins[0], call.sizes[0], call.origins[1] = srcOrig, srcSize, dstOrig
	sendLocked(enc, mtl.sel.copyTexToTex, src, 0, 0, uintptr(unsafe.Pointer(&call.origins[0])), uintptr(unsafe.Pointer(&call.sizes[0])), dst, 0, 0, uintptr(unsafe.Pointer(&call.origins[1])))
}

func blitEncCopyBufferToTexture(enc, buf, tex, offset, stride, length uintptr, dims mtlSize, orig mtlOrigin) {
	if !structArgsByPointer {
		mtl.copyBufToTex(enc, mtl.sel.copyBufToTex, buf, offset, stride, length, dims, tex, 0, 0, orig)
		return
	}
	call.Lock()
	defer call.Unlock()
	call.sizes[0], call.origins[0] = dims, orig
	sendLocked(enc, mtl.sel.copyBufToTex, buf, offset, stride, length, uintptr(unsafe.Pointer(&call.sizes[0])), tex, 0, 0, uintptr(unsafe.Pointer(&call.origins[0])))
}

func blitEncGenerateMipmapsForTexture(enc, tex uintptr) {
	send(enc, mtl.sel.generateMipmaps, tex)
}

func blitEncCopyTextureToBuffer(enc, tex, buf, offset, stride, length uintptr, dims mtlSize, orig mtlOrigin) {
	if !structArgsByPointer {
		mtl.copyTexToBuf(enc, mtl.sel.copyTexToBuf, tex, 0, 0, orig, dims, buf, offset, stride, length)
		return
	}
	call.Lock()
	defer call.Unlock()
	call.origins[0], call.sizes[0] = orig, dims
	sendLocked(enc, mtl.sel.copyTexToBuf, tex, 0, 0, uintptr(unsafe.Pointer(&call.origins[0])), uintptr(unsafe.Pointer(&call.sizes[0])), buf, offset, stride, length)
}

func blitEncCopyBufferToBuffer(enc, src, dst, srcOff, dstOff, size uintptr) {
	send(enc, mtl.sel.copyBufToBuf, src, srcOff, dst, dstOff, size)
}

func newTexture(dev, width, height, format, usage uintptr, mipmapped bool) uintptr {
	p := pushPool()
	defer popPool(p)
	desc := send(mtl.textureDesc, mtl.sel.texture2D, format, width, height, boolArg(mipmapped))
	send(desc, mtl.sel.setUsage, usage)
	send(desc, mtl.sel.setStorageMode, mtlStorageModePrivate)
	return send(dev, mtl.sel.newTexture, desc)
}

func newSampler(dev, minFilter, magFilter, mipFilter uintptr) uintptr {
	p := pushPool()
	defer popPool(p)
	desc := send(mtl.samplerDesc, mtl.sel.new)
	defer send(desc, mtl.sel.release)
	send(desc, mtl.sel.setMinFilter, minFilter)
	send(desc, mtl.sel.setMagFilter, magFilter)
	send(desc, mtl.sel.setMipFilter, mipFilter)
	return send(dev, mtl.sel.newSampler, desc)
}

func newBuffer(dev, size, opts uintptr) uintptr {
	p := pushPool()
	defer popPool(p)
	return send(dev, mtl.sel.newBuffer, size, opts)
}

func bufferContents(buf uintptr) (unsafe.Pointer, int) {
	addr := send(buf, mtl.sel.contents)
	return *(*unsafe.Pointer)(unsafe.Pointer(&addr)), int(send(buf, mtl.sel.length))
}

func newLibrary(dev uintptr, name, lib string) (uintptr, error) {
	p := pushPool()
	defer popPool(p)
	// The default destructor copies the bytes.
	data, _, _ := purego.SyscallN(mtl.dataCreate, uintptr(unsafe.Pointer(unsafe.StringData(lib))), uintptr(len(lib)), 0, dispatchDataDestructorDefault)
	if data == 0 {
		return 0, errors.New("dispatch_data_create failed")
	}
	defer purego.SyscallN(mtl.dataRelease, data)
	var nserr uintptr
	l, _, _ := purego.SyscallN(mtl.msg, dev, mtl.sel.newLibraryWithData, data, uintptr(unsafe.Pointer(&nserr)))
	l, err := objErr(l, nserr)
	if err != nil {
		return 0, err
	}
	send(l, mtl.sel.setLabel, nsString(name))
	return l, nil
}

func libraryNewFunction(lib uintptr, name string) uintptr {
	p := pushPool()
	defer popPool(p)
	return send(lib, mtl.sel.newFunction, nsString(name))
}

func newComputePipeline(dev, fn uintptr) (uintptr, error) {
	p := pushPool()
	defer popPool(p)
	var nserr uintptr
	s, _, _ := purego.SyscallN(mtl.msg, dev, mtl.sel.newComputePipeline, fn, uintptr(unsafe.Pointer(&nserr)))
	return objErr(s, nserr)
}

func newRenderPipeline(dev, vertFunc, fragFunc, pixelFormat, bufIdx uintptr, fmts, offsets []uintptr, stride uintptr, blend bool, srcFactor, dstFactor, nvertBufs, nfragBufs uintptr) (uintptr, error) {
	p := pushPool()
	defer popPool(p)
	s := &mtl.sel
	vdesc := send(mtl.vertexDesc, s.vertexDescriptor)
	send(send(send(vdesc, s.layouts), s.objectAtIndex, bufIdx), s.setStride, stride)
	attrs := send(vdesc, s.attributes)
	for i := range fmts {
		a := send(attrs, s.objectAtIndex, uintptr(i))
		send(a, s.setFormat, fmts[i])
		send(a, s.setOffset, offsets[i])
		send(a, s.setBufferIndex, bufIdx)
	}
	desc := send(mtl.pipelineDesc, s.new)
	defer send(desc, s.release)
	send(desc, s.setVertexFunction, vertFunc)
	send(desc, s.setFragmentFunction, fragFunc)
	send(desc, s.setVertexDescriptor, vdesc)
	vbufs := send(desc, s.vertexBuffers)
	for i := uintptr(0); i < nvertBufs; i++ {
		send(send(vbufs, s.objectAtIndex, i), s.setMutability, mtlMutabilityImmutable)
	}
	fbufs := send(desc, s.fragmentBuffers)
	for i := uintptr(0); i < nfragBufs; i++ {
		send(send(fbufs, s.objectAtIndex, i), s.setMutability, mtlMutabilityImmutable)
	}
	setBlend(desc, pixelFormat, blend, srcFactor, dstFactor)
	return newPipelineState(dev, desc)
}

// setBlend configures the descriptor's first color attachment.
func setBlend(desc, pixelFormat uintptr, blend bool, srcFactor, dstFactor uintptr) {
	s := &mtl.sel
	att := send(send(desc, s.colorAttachments), s.objectAtIndex, 0)
	send(att, s.setPixelFormat, pixelFormat)
	send(att, s.setBlendingEnabled, boolArg(blend))
	send(att, s.setSourceAlpha, srcFactor)
	send(att, s.setSourceRGB, srcFactor)
	send(att, s.setDestinationAlpha, dstFactor)
	send(att, s.setDestRGB, dstFactor)
}

func newPipelineState(dev, desc uintptr) (uintptr, error) {
	var nserr uintptr
	r, _, _ := purego.SyscallN(mtl.msg, dev, mtl.sel.newRenderPipeline, desc, uintptr(unsafe.Pointer(&nserr)))
	return objErr(r, nserr)
}

func quadLibrary(dev uintptr, src string) uintptr {
	p := pushPool()
	defer popPool(p)
	var nserr uintptr
	lib, _, _ := purego.SyscallN(mtl.msg, dev, mtl.sel.newLibraryWithSource, nsString(src), 0, uintptr(unsafe.Pointer(&nserr)))
	return lib
}

func quadPipeline(dev, lib uintptr, frag string, format uintptr) uintptr {
	p := pushPool()
	defer popPool(p)
	s := &mtl.sel
	desc := send(mtl.pipelineDesc, s.new)
	defer send(desc, s.release)
	vfn := send(lib, s.newFunction, nsString("quad_vert"))
	ffn := send(lib, s.newFunction, nsString(frag))
	// The descriptor retains the functions.
	send(desc, s.setVertexFunction, vfn)
	send(desc, s.setFragmentFunction, ffn)
	for _, fn := range [...]uintptr{vfn, ffn} {
		if fn != 0 {
			send(fn, s.release)
		}
	}
	setBlend(desc, format, true, mtlBlendFactorOne, mtlBlendFactorOneMinusSourceAlpha)
	pipe, _ := newPipelineState(dev, desc)
	return pipe
}

func quadNewBuffer(dev, size uintptr) uintptr {
	return newBuffer(dev, size, mtlResourceStorageModeShared|mtlResourceCPUCacheModeWriteCombined)
}

// quadDraw draws n instances; texs and samplers have the same length. It
// holds send's lock once for the whole draw.
func quadDraw(enc, pipe, buf, off, n uintptr, texs, samplers []uintptr) {
	s := &mtl.sel
	call.Lock()
	defer call.Unlock()
	sendLocked(enc, s.setRenderPipelineState, pipe)
	sendLocked(enc, s.setVertexBuffer, buf, off, 0)
	for i := range texs {
		sendLocked(enc, s.setFragmentTexture, texs[i], uintptr(i))
		sendLocked(enc, s.setFragmentSampler, samplers[i], uintptr(i))
	}
	sendLocked(enc, s.drawInstanced, mtlPrimitiveTypeTriangleStrip, 0, 4, n)
}
