// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"errors"
	"fmt"
	"image"
	"unsafe"

	"gioui.org/shader"
	"github.com/dyike/keel/gio/gpu/internal/driver"
)

// The Metal calls themselves live in metal_cgo_darwin.go (Gio's Objective-C)
// and metal_purego_darwin.go (the same messages sent through purego).
// Keel patch: this file holds the logic both share, over plain handles.

// Metal enumeration values; metal_cgo_darwin.go checks them against the
// SDK headers at compile time.
const (
	mtlPixelFormatR8Unorm         = 10
	mtlPixelFormatR16Float        = 25
	mtlPixelFormatRGBA8Unorm      = 70
	mtlPixelFormatRGBA8Unorm_sRGB = 71

	mtlTextureUsageShaderRead   = 1
	mtlTextureUsageShaderWrite  = 2
	mtlTextureUsageRenderTarget = 4

	mtlSamplerMinMagFilterNearest   = 0
	mtlSamplerMinMagFilterLinear    = 1
	mtlSamplerMipFilterNotMipmapped = 0
	mtlSamplerMipFilterLinear       = 2

	mtlBlendFactorZero                = 0
	mtlBlendFactorOne                 = 1
	mtlBlendFactorOneMinusSourceAlpha = 5
	mtlBlendFactorDestinationColor    = 6

	mtlVertexFormatFloat  = 28
	mtlVertexFormatFloat2 = 29
	mtlVertexFormatFloat3 = 30
	mtlVertexFormatFloat4 = 31

	mtlLoadActionDontCare = 0
	mtlLoadActionLoad     = 1
	mtlLoadActionClear    = 2

	mtlPrimitiveTypeTriangle      = 3
	mtlPrimitiveTypeTriangleStrip = 4

	mtlResourceCPUCacheModeWriteCombined = 1
	mtlResourceStorageModeShared         = 0 << 4
	mtlResourceStorageModePrivate        = 2 << 4
)

// MTLOrigin, MTLSize and MTLViewport.
type mtlOrigin struct{ X, Y, Z uintptr }
type mtlSize struct{ Width, Height, Depth uintptr }
type mtlViewport struct{ OriginX, OriginY, Width, Height, ZNear, ZFar float64 }

type Backend struct {
	dev      uintptr
	queue    uintptr
	pixelFmt uintptr

	cmdBuffer     uintptr
	lastCmdBuffer uintptr
	renderEnc     uintptr
	computeEnc    uintptr
	blitEnc       uintptr

	prog     *Program
	topology uintptr

	unified bool // CPU uploads avoid activating a second GPU engine on Apple silicon.
	pool    bufferPool
	stores  storePool
	quads   quadBatcher

	stagingBuf uintptr
	stagingOff int

	indexBuf *Buffer

	// bufSizes is scratch space for filling out the spvBufferSizeConstants
	// that spirv-cross generates for emulating buffer.length expressions in
	// shaders.
	bufSizes []uint32
}

type Texture struct {
	backend *Backend
	texture uintptr
	sampler uintptr
	width   int
	height  int
	mipmap  bool
	foreign bool
	direct  bool // Only the initial complete upload may bypass GPU synchronization.
	srgb    bool
	mask    bool
}

type Shader struct {
	function uintptr
	inputs   []shader.InputLocation
}

type Program struct {
	pipeline  uintptr
	groupSize [3]int
}

type Pipeline struct {
	pipeline uintptr
	topology uintptr
}

type Buffer struct {
	backend *Backend
	size    int
	buffer  uintptr
	poolKey bufferPoolKey

	// store is the buffer contents for buffers not allocated on the GPU.
	store []byte
}

const (
	uniformBufferIndex   = 0
	attributeBufferIndex = 1

	spvBufferSizeConstantsBinding = 25
)

const (
	texUnits    = 4
	bufferUnits = 4
)

func init() {
	driver.NewMetalDevice = newMetalDevice
}

func newMetalDevice(api driver.Metal) (driver.Device, error) {
	if err := loadMetal(); err != nil {
		return nil, fmt.Errorf("metal: %w", err)
	}
	dev := uintptr(api.Device)
	cfRetain(dev)
	queue := uintptr(api.Queue)
	cfRetain(queue)
	b := &Backend{
		dev:      dev,
		unified:  unifiedMemory(uintptr(dev)),
		queue:    queue,
		pixelFmt: uintptr(api.PixelFormat),
		bufSizes: make([]uint32, bufferUnits),
	}
	return b, nil
}

func (b *Backend) BeginFrame(target driver.RenderTarget, clear bool, viewport image.Point) driver.Texture {
	if b.lastCmdBuffer != 0 {
		cmdBufferWaitUntilCompleted(b.lastCmdBuffer)
		b.stagingOff = 0
	}
	// Every command buffer committed so far has completed: buffers released
	// before this frame are no longer read by the GPU.
	b.pool.frameStarted()
	b.quads.frameStarted()
	if target == nil {
		return nil
	}
	switch t := target.(type) {
	case driver.MetalRenderTarget:
		texture := uintptr(t.Texture)
		return &Texture{texture: texture, foreign: true}
	case *Texture:
		return t
	default:
		panic(fmt.Sprintf("metal: unsupported render target type: %T", t))
	}
}

func (b *Backend) startBlit() uintptr {
	if b.blitEnc != 0 {
		return b.blitEnc
	}
	b.endEncoder()
	b.ensureCmdBuffer()
	b.blitEnc = cmdBufferBlitEncoder(b.cmdBuffer)
	if b.blitEnc == 0 {
		panic("metal: [MTLCommandBuffer blitCommandEncoder:] failed")
	}
	return b.blitEnc
}

func (b *Backend) CopyTexture(dst driver.Texture, dorig image.Point, src driver.Texture, srect image.Rectangle) {
	enc := b.startBlit()
	dstTex := dst.(*Texture).texture
	srcTex := src.(*Texture).texture
	ssz := srect.Size()
	blitEncCopyFromTexture(
		enc,
		srcTex,
		mtlOrigin{
			X: uintptr(srect.Min.X),
			Y: uintptr(srect.Min.Y),
		},
		mtlSize{
			Width:  uintptr(ssz.X),
			Height: uintptr(ssz.Y),
			Depth:  1,
		},
		dstTex,
		mtlOrigin{
			X: uintptr(dorig.X),
			Y: uintptr(dorig.Y),
		},
	)
}

func (b *Backend) EndFrame() {
	b.endCmdBuffer(false)
}

func (b *Backend) endCmdBuffer(wait bool) {
	b.endEncoder()
	if b.cmdBuffer == 0 {
		return
	}
	cmdBufferCommit(b.cmdBuffer)
	if wait {
		cmdBufferWaitUntilCompleted(b.cmdBuffer)
	}
	if b.lastCmdBuffer != 0 {
		cfRelease(b.lastCmdBuffer)
	}
	b.lastCmdBuffer = b.cmdBuffer
	b.cmdBuffer = 0
}

func (b *Backend) Caps() driver.Caps {
	return driver.Caps{
		MaxTextureSize: 8192,
		Features:       driver.FeatureSRGB | driver.FeatureCompute | driver.FeatureFloatRenderTargets,
	}
}

func (b *Backend) NewTimer() driver.Timer {
	panic("timers not supported")
}

func (b *Backend) IsTimeContinuous() bool {
	panic("timers not supported")
}

func (b *Backend) Release() {
	b.pool.trim()
	b.quads.release()
	if b.cmdBuffer != 0 {
		cfRelease(b.cmdBuffer)
	}
	if b.lastCmdBuffer != 0 {
		cfRelease(b.lastCmdBuffer)
	}
	if b.stagingBuf != 0 {
		cfRelease(b.stagingBuf)
	}
	cfRelease(b.queue)
	cfRelease(b.dev)
	*b = Backend{}
}

// NewSurfaceTexture wraps a BGRA IOSurface in a texture without copying
// it. Keel patch.
func (b *Backend) NewSurfaceTexture(surface uintptr, width, height int, minFilter, magFilter driver.TextureFilter) (driver.Texture, error) {
	tex := uintptr(surfaceTexture(uintptr(b.dev), surface, uintptr(width), uintptr(height), b.unified))
	if tex == 0 {
		return nil, errors.New("metal: [MTLDevice newTextureWithDescriptor:iosurface:plane:] failed")
	}
	min, mip := samplerFilterFor(minFilter)
	max, _ := samplerFilterFor(magFilter)
	s := newSampler(b.dev, min, max, mip)
	if s == 0 {
		cfRelease(tex)
		return nil, errors.New("metal: [MTLDevice newSamplerStateWithDescriptor:] failed")
	}
	return &Texture{backend: b, texture: tex, sampler: s, width: width, height: height, srgb: true}, nil
}

func (b *Backend) NewTexture(format driver.TextureFormat, width, height int, minFilter, magFilter driver.TextureFilter, bindings driver.BufferBinding) (driver.Texture, error) {
	mformat := pixelFormatFor(format)
	var usage uintptr
	if bindings&(driver.BufferBindingTexture|driver.BufferBindingShaderStorageRead) != 0 {
		usage |= mtlTextureUsageShaderRead
	}
	if bindings&driver.BufferBindingFramebuffer != 0 {
		usage |= mtlTextureUsageRenderTarget
	}
	if bindings&driver.BufferBindingShaderStorageWrite != 0 {
		usage |= mtlTextureUsageShaderWrite
	}
	min, mip := samplerFilterFor(minFilter)
	max, _ := samplerFilterFor(magFilter)
	mipmap := mip != mtlSamplerMipFilterNotMipmapped
	direct := b.unified && bindings == driver.BufferBindingTexture && format != driver.TextureFormatFloat
	var tex uintptr
	if direct {
		tex = uintptr(uploadTexture(uintptr(b.dev), uintptr(mformat), uintptr(width), uintptr(height), uintptr(usage), mipmap))
	}
	if tex == 0 {
		if format == driver.TextureFormatAlpha8 {
			return nil, errors.New("metal: single-channel masks unavailable")
		}
		direct = false
		tex = newTexture(b.dev, uintptr(width), uintptr(height), mformat, usage, mipmap)
	}
	if tex == 0 {
		return nil, errors.New("metal: [MTLDevice newTextureWithDescriptor:] failed")
	}
	s := newSampler(b.dev, min, max, mip)
	if s == 0 {
		cfRelease(tex)
		return nil, errors.New("metal: [MTLDevice newSamplerStateWithDescriptor:] failed")
	}
	return &Texture{backend: b, texture: tex, sampler: s, width: width, height: height, mipmap: mipmap, direct: direct, srgb: format == driver.TextureFormatSRGBA, mask: format == driver.TextureFormatAlpha8}, nil
}

func (b *Backend) NewMaskTexture(width, height int, min, mag driver.TextureFilter) (driver.Texture, error) {
	if !b.unified {
		return nil, errors.New("metal: coverage upload requires unified memory")
	}
	return b.NewTexture(driver.TextureFormatAlpha8, width, height, min, mag, driver.BufferBindingTexture)
}

func samplerFilterFor(f driver.TextureFilter) (uintptr, uintptr) {
	switch f {
	case driver.FilterNearest:
		return mtlSamplerMinMagFilterNearest, mtlSamplerMipFilterNotMipmapped
	case driver.FilterLinear:
		return mtlSamplerMinMagFilterLinear, mtlSamplerMipFilterNotMipmapped
	case driver.FilterLinearMipmapLinear:
		return mtlSamplerMinMagFilterLinear, mtlSamplerMipFilterLinear
	default:
		panic("invalid texture filter")
	}
}

func (b *Backend) NewPipeline(desc driver.PipelineDesc) (driver.Pipeline, error) {
	vsh, fsh := desc.VertexShader.(*Shader), desc.FragmentShader.(*Shader)
	layout := desc.VertexLayout.Inputs
	if got, exp := len(layout), len(vsh.inputs); got != exp {
		return nil, fmt.Errorf("metal: number of input descriptors (%d) doesn't match number of inputs (%d)", got, exp)
	}
	formats := make([]uintptr, len(layout))
	offsets := make([]uintptr, len(layout))
	for i, inp := range layout {
		index := vsh.inputs[i].Location
		formats[index] = vertFormatFor(vsh.inputs[i])
		offsets[index] = uintptr(inp.Offset)
	}
	srcFactor := blendFactorFor(desc.BlendDesc.SrcFactor)
	dstFactor := blendFactorFor(desc.BlendDesc.DstFactor)
	pf := b.pixelFmt
	if f := desc.PixelFormat; f != driver.TextureFormatOutput {
		pf = pixelFormatFor(f)
	}
	pipe, err := newRenderPipeline(
		b.dev,
		vsh.function,
		fsh.function,
		pf,
		attributeBufferIndex,
		formats, offsets,
		uintptr(desc.VertexLayout.Stride),
		desc.BlendDesc.IsEnabled(), srcFactor, dstFactor,
		2, // Number of vertex buffers.
		1, // Number of fragment buffers.
	)
	if err != nil {
		return nil, fmt.Errorf("metal: pipeline construction failed: %w", err)
	}
	return &Pipeline{pipeline: pipe, topology: primitiveFor(desc.Topology)}, nil
}

func blendFactorFor(f driver.BlendFactor) uintptr {
	switch f {
	case driver.BlendFactorZero:
		return mtlBlendFactorZero
	case driver.BlendFactorOne:
		return mtlBlendFactorOne
	case driver.BlendFactorOneMinusSrcAlpha:
		return mtlBlendFactorOneMinusSourceAlpha
	case driver.BlendFactorDstColor:
		return mtlBlendFactorDestinationColor
	default:
		panic("unsupported blend factor")
	}
}

func vertFormatFor(f shader.InputLocation) uintptr {
	t := f.Type
	s := f.Size
	switch {
	case t == shader.DataTypeFloat && s == 1:
		return mtlVertexFormatFloat
	case t == shader.DataTypeFloat && s == 2:
		return mtlVertexFormatFloat2
	case t == shader.DataTypeFloat && s == 3:
		return mtlVertexFormatFloat3
	case t == shader.DataTypeFloat && s == 4:
		return mtlVertexFormatFloat4
	default:
		panic("unsupported data type")
	}
}

func pixelFormatFor(f driver.TextureFormat) uintptr {
	switch f {
	case driver.TextureFormatFloat:
		return mtlPixelFormatR16Float
	case driver.TextureFormatRGBA8:
		return mtlPixelFormatRGBA8Unorm
	case driver.TextureFormatAlpha8:
		return mtlPixelFormatR8Unorm
	case driver.TextureFormatSRGBA:
		return mtlPixelFormatRGBA8Unorm_sRGB
	default:
		panic("unsupported pixel format")
	}
}

func (b *Backend) NewBuffer(typ driver.BufferBinding, size int) (driver.Buffer, error) {
	// Transfer buffer contents in command encoders on every use for
	// smaller buffers. The advantage is that buffer re-use during a frame
	// won't occur a GPU wait.
	// We can't do this for buffers written to by the GPU and read by the client,
	// and Metal doesn't require a buffer for indexed draws.
	if size <= maxStoreSize && typ&(driver.BufferBindingShaderStorageWrite|driver.BufferBindingIndices) == 0 {
		return &Buffer{backend: b, size: size, store: b.stores.get(size)}, nil
	}
	key := bufferPoolKey{class: bufferClass(size), storage: bufferStoragePrivate}
	buf := b.pool.get(key)
	if buf == 0 {
		buf = newBuffer(b.dev, uintptr(key.class), mtlResourceStorageModePrivate)
	}
	return &Buffer{backend: b, size: size, buffer: buf, poolKey: key}, nil
}

func (b *Backend) NewImmutableBuffer(typ driver.BufferBinding, data []byte) (driver.Buffer, error) {
	// Immutable geometry is initialized before its first GPU use. Keep mutable
	// buffers on the ordered upload path, including later Upload calls on this buffer.
	if b.unified && (len(data) > 4096 || typ&driver.BufferBindingIndices != 0) {
		key := bufferPoolKey{class: bufferClass(len(data)), storage: bufferStorageShared}
		buf := b.pool.get(key)
		if buf == 0 {
			buf = newBuffer(b.dev, uintptr(key.class), mtlResourceStorageModeShared)
		}
		if buf != 0 {
			copy(bufferStore(buf)[:len(data)], data)
			return &Buffer{backend: b, size: len(data), buffer: buf, poolKey: key}, nil
		}
	}
	buf, err := b.NewBuffer(typ, len(data))
	if err != nil {
		return nil, err
	}
	buf.Upload(data)
	return buf, nil
}

func (b *Backend) NewComputeProgram(src shader.Sources) (driver.Program, error) {
	sh, err := b.newShader(src)
	if err != nil {
		return nil, err
	}
	defer sh.Release()
	pipe, err := newComputePipeline(b.dev, sh.function)
	if err != nil {
		return nil, fmt.Errorf("metal: compute program %q load failed: %w", src.Name, err)
	}
	return &Program{pipeline: pipe, groupSize: src.WorkgroupSize}, nil
}

func (b *Backend) NewVertexShader(src shader.Sources) (driver.VertexShader, error) {
	return b.newShader(src)
}

func (b *Backend) NewFragmentShader(src shader.Sources) (driver.FragmentShader, error) {
	return b.newShader(src)
}

func (b *Backend) newShader(src shader.Sources) (*Shader, error) {
	vsrc := src.MetalLib
	vlib, err := newLibrary(b.dev, src.Name, vsrc)
	if err != nil {
		return nil, fmt.Errorf("metal: shader %q load failed: %w", src.Name, err)
	}
	defer cfRelease(vlib)
	f := libraryNewFunction(vlib, "main0")
	if f == 0 {
		return nil, fmt.Errorf("metal: main function not found in %q", src.Name)
	}
	return &Shader{function: f, inputs: src.Inputs}, nil
}

func (b *Backend) Viewport(x, y, width, height int) {
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	renderEncViewport(enc, mtlViewport{
		OriginX: float64(x),
		OriginY: float64(y),
		Width:   float64(width),
		Height:  float64(height),
		ZNear:   0.0,
		ZFar:    1.0,
	})
}

func (b *Backend) DrawArrays(off, count int) {
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	renderEncDrawPrimitives(enc, b.topology, uintptr(off), uintptr(count))
}

func (b *Backend) DrawElements(off, count int) {
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	renderEncDrawIndexedPrimitives(enc, b.topology, b.indexBuf.buffer, uintptr(off), uintptr(count))
}

func primitiveFor(mode driver.Topology) uintptr {
	switch mode {
	case driver.TopologyTriangles:
		return mtlPrimitiveTypeTriangle
	case driver.TopologyTriangleStrip:
		return mtlPrimitiveTypeTriangleStrip
	default:
		panic("metal: unknown draw mode")
	}
}

func (b *Backend) BindImageTexture(unit int, tex driver.Texture) {
	b.BindTexture(unit, tex)
}

func (b *Backend) BeginCompute() {
	b.endEncoder()
	b.ensureCmdBuffer()
	for i := range b.bufSizes {
		b.bufSizes[i] = 0
	}
	b.computeEnc = cmdBufferComputeEncoder(b.cmdBuffer)
	if b.computeEnc == 0 {
		panic("metal: [MTLCommandBuffer computeCommandEncoder:] failed")
	}
}

func (b *Backend) EndCompute() {
	if b.computeEnc == 0 {
		panic("no active compute pass")
	}
	computeEncEnd(b.computeEnc)
	cfRelease(b.computeEnc)
	b.computeEnc = 0
}

func (b *Backend) DispatchCompute(x, y, z int) {
	enc := b.computeEnc
	if enc == 0 {
		panic("no active compute pass")
	}
	computeEncSetBytes(enc, unsafe.Pointer(unsafe.SliceData(b.bufSizes)), uintptr(len(b.bufSizes)*4), spvBufferSizeConstantsBinding)
	threadgroupsPerGrid := mtlSize{
		Width: uintptr(x), Height: uintptr(y), Depth: uintptr(z),
	}
	sz := b.prog.groupSize
	threadsPerThreadgroup := mtlSize{
		Width: uintptr(sz[0]), Height: uintptr(sz[1]), Depth: uintptr(sz[2]),
	}
	computeEncDispatch(enc, threadgroupsPerGrid, threadsPerThreadgroup)
}

func (b *Backend) stagingBuffer(size int) (uintptr, int) {
	if b.stagingBuf == 0 || b.stagingOff+size > len(bufferStore(b.stagingBuf)) {
		if b.stagingBuf != 0 {
			cfRelease(b.stagingBuf)
		}
		cap := 2 * (b.stagingOff + size)
		b.stagingBuf = newBuffer(b.dev, uintptr(cap), mtlResourceStorageModeShared|mtlResourceCPUCacheModeWriteCombined)
		if b.stagingBuf == 0 {
			panic(fmt.Errorf("metal: failed to allocate %d bytes of staging buffer", cap))
		}
		b.stagingOff = 0
	}
	off := b.stagingOff
	b.stagingOff += size
	return b.stagingBuf, off
}

func (t *Texture) Upload(offset, size image.Point, pixels []byte, stride int) {
	if len(pixels) == 0 {
		return
	}
	bpp := 4
	if t.mask {
		bpp = 1
	}
	if stride == 0 {
		stride = size.X * bpp
	}
	// A later or partial write must remain ordered against encoded GPU reads.
	direct := t.direct
	t.direct = false
	if direct && offset == (image.Point{}) && size == image.Pt(t.width, t.height) {
		directUpload(uintptr(t.texture), offset, size, pixels, stride, 0)
		if t.mipmap {
			if t.mask {
				uploadMaskMipmaps(uintptr(t.texture), size, pixels, stride)
			} else {
				uploadMipmaps(uintptr(t.texture), size, pixels, stride, t.srgb)
			}
		}
		return
	}
	dstStride := size.X * bpp
	n := size.Y * dstStride
	buf, off := t.backend.stagingBuffer(n)
	store := bufferSlice(buf, off, n)
	var srcOff, dstOff int
	for y := 0; y < size.Y; y++ {
		srcRow := pixels[srcOff : srcOff+dstStride]
		dstRow := store[dstOff : dstOff+dstStride]
		copy(dstRow, srcRow)
		dstOff += dstStride
		srcOff += stride
	}
	enc := t.backend.startBlit()
	orig := mtlOrigin{
		X: uintptr(offset.X),
		Y: uintptr(offset.Y),
	}
	msize := mtlSize{
		Width:  uintptr(size.X),
		Height: uintptr(size.Y),
		Depth:  1,
	}
	blitEncCopyBufferToTexture(enc, buf, t.texture, uintptr(off), uintptr(dstStride), uintptr(len(store)), msize, orig)
	if t.mipmap {
		blitEncGenerateMipmapsForTexture(enc, t.texture)
	}
}

// SetVolatile lets the system reclaim the texture's memory while it is
// volatile, which also takes it out of the process's footprint. Keel patch.
func (t *Texture) SetVolatile(volatile bool) {
	textureSetVolatile(t.texture, volatile)
}

func (t *Texture) Release() {
	if t.foreign {
		panic("metal: release of external texture")
	}
	cfRelease(t.texture)
	cfRelease(t.sampler)
	*t = Texture{}
}

func (p *Pipeline) Release() {
	cfRelease(p.pipeline)
	*p = Pipeline{}
}

func (b *Backend) PrepareTexture(tex driver.Texture) {}

func (b *Backend) BindTexture(unit int, tex driver.Texture) {
	t := tex.(*Texture)
	if enc := b.renderEnc; enc != 0 {
		renderEncSetFragmentTexture(enc, uintptr(unit), t.texture)
		renderEncSetFragmentSamplerState(enc, uintptr(unit), t.sampler)
	} else if enc := b.computeEnc; enc != 0 {
		computeEncSetTexture(enc, uintptr(unit), t.texture)
	} else {
		panic("no active render nor compute pass")
	}
}

func (b *Backend) ensureCmdBuffer() {
	if b.cmdBuffer != 0 {
		return
	}
	b.cmdBuffer = queueNewBuffer(b.queue)
	if b.cmdBuffer == 0 {
		panic("metal: [MTLCommandQueue cmdBuffer] failed")
	}
}

func (b *Backend) BindPipeline(pipe driver.Pipeline) {
	p := pipe.(*Pipeline)
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	renderEncSetRenderPipelineState(enc, p.pipeline)
	b.topology = p.topology
}

func (b *Backend) BindProgram(prog driver.Program) {
	enc := b.computeEnc
	if enc == 0 {
		panic("no active compute pass")
	}
	p := prog.(*Program)
	computeEncSetPipeline(enc, p.pipeline)
	b.prog = p
}

func (s *Shader) Release() {
	cfRelease(s.function)
	*s = Shader{}
}

func (p *Program) Release() {
	cfRelease(p.pipeline)
	*p = Program{}
}

func (b *Backend) BindStorageBuffer(binding int, buffer driver.Buffer) {
	buf := buffer.(*Buffer)
	b.bufSizes[binding] = uint32(buf.size)
	enc := b.computeEnc
	if enc == 0 {
		panic("no active compute pass")
	}
	if buf.buffer != 0 {
		computeEncSetBuffer(enc, uintptr(binding), buf.buffer)
	} else if buf.size > 0 {
		computeEncSetBytes(enc, unsafe.Pointer(unsafe.SliceData(buf.store)), uintptr(buf.size), uintptr(binding))
	}
}

func (b *Backend) BindUniforms(buf driver.Buffer) {
	bf := buf.(*Buffer)
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	if bf.buffer != 0 {
		renderEncSetVertexBuffer(enc, bf.buffer, uniformBufferIndex, 0)
		renderEncSetFragmentBuffer(enc, bf.buffer, uniformBufferIndex, 0)
	} else if bf.size > 0 {
		renderEncSetVertexBytes(enc, unsafe.Pointer(unsafe.SliceData(bf.store)), uintptr(bf.size), uniformBufferIndex)
		renderEncSetFragmentBytes(enc, unsafe.Pointer(unsafe.SliceData(bf.store)), uintptr(bf.size), uniformBufferIndex)
	}
}

func (b *Backend) BindVertexBuffer(buf driver.Buffer, offset int) {
	bf := buf.(*Buffer)
	enc := b.renderEnc
	if enc == 0 {
		panic("no active render pass")
	}
	if bf.buffer != 0 {
		renderEncSetVertexBuffer(enc, bf.buffer, attributeBufferIndex, uintptr(offset))
	} else if n := bf.size - offset; n > 0 {
		renderEncSetVertexBytes(enc, unsafe.Pointer(&bf.store[offset]), uintptr(n), attributeBufferIndex)
	}
}

func (b *Backend) BindIndexBuffer(buf driver.Buffer) {
	b.indexBuf = buf.(*Buffer)
}

func (b *Buffer) Download(data []byte) error {
	if len(data) > b.size {
		panic(fmt.Errorf("len(data) (%d) larger than len(content) (%d)", len(data), b.size))
	}
	buf, off := b.backend.stagingBuffer(len(data))
	enc := b.backend.startBlit()
	blitEncCopyBufferToBuffer(enc, b.buffer, buf, 0, uintptr(off), uintptr(len(data)))
	b.backend.endCmdBuffer(true)
	store := bufferSlice(buf, off, len(data))
	copy(data, store)
	return nil
}

func (b *Buffer) Upload(data []byte) {
	if len(data) > b.size {
		panic(fmt.Errorf("len(data) (%d) larger than len(content) (%d)", len(data), b.size))
	}
	if b.buffer == 0 {
		copy(b.store, data)
		return
	}
	buf, off := b.backend.stagingBuffer(len(data))
	store := bufferSlice(buf, off, len(data))
	copy(store, data)
	enc := b.backend.startBlit()
	blitEncCopyBufferToBuffer(enc, buf, b.buffer, uintptr(off), 0, uintptr(len(store)))
}

func bufferStore(buf uintptr) []byte {
	addr, size := bufferContents(buf)
	return unsafe.Slice((*byte)(addr), size)
}

func bufferSlice(buf uintptr, off, len int) []byte {
	store := bufferStore(buf)
	return store[off : off+len]
}

func (b *Buffer) Release() {
	if b.buffer != 0 && !b.backend.pool.put(b.poolKey, b.buffer) {
		cfRelease(b.buffer)
	}
	if b.store != nil && b.backend != nil {
		b.backend.stores.put(b.store)
	}
	*b = Buffer{}
}

// Trim releases the buffers kept for reuse. Keel patch; the gpu package
// calls it when a window goes idle.
func (b *Backend) Trim() {
	// The next frame must not reuse storage still read by the GPU. Finish
	// this frame before dropping the completion fence and temporary buffers.
	if b.lastCmdBuffer != 0 {
		cmdBufferWaitUntilCompleted(b.lastCmdBuffer)
		cfRelease(b.lastCmdBuffer)
		b.lastCmdBuffer = 0
	}
	b.pool.trim()
	b.stores.trim()
	if b.stagingBuf != 0 {
		cfRelease(b.stagingBuf)
		b.stagingBuf = 0
	}
	b.stagingOff = 0
	if b.quads.buf != 0 {
		cfRelease(b.quads.buf)
		b.quads.buf = 0
	}
	b.quads.off = 0
}

func (t *Texture) ReadPixels(src image.Rectangle, pixels []byte, stride int) error {
	if len(pixels) == 0 {
		return nil
	}
	sz := src.Size()
	orig := mtlOrigin{
		X: uintptr(src.Min.X),
		Y: uintptr(src.Min.Y),
	}
	msize := mtlSize{
		Width:  uintptr(sz.X),
		Height: uintptr(sz.Y),
		Depth:  1,
	}
	bpp := 4
	if t.mask {
		bpp = 1
	}
	stageStride := sz.X * bpp
	n := sz.Y * stageStride
	buf, off := t.backend.stagingBuffer(n)
	enc := t.backend.startBlit()
	blitEncCopyTextureToBuffer(enc, t.texture, buf, uintptr(off), uintptr(stageStride), uintptr(n), msize, orig)
	t.backend.endCmdBuffer(true)
	store := bufferSlice(buf, off, n)
	var srcOff, dstOff int
	for y := 0; y < sz.Y; y++ {
		dstRow := pixels[srcOff : srcOff+sz.X*4]
		srcRow := store[dstOff : dstOff+stageStride]
		if t.mask {
			for x, a := range srcRow {
				j := x * 4
				dstRow[j], dstRow[j+1], dstRow[j+2], dstRow[j+3] = a, a, a, a
			}
		} else {
			copy(dstRow, srcRow)
		}
		dstOff += stageStride
		srcOff += stride
	}
	return nil
}

func (b *Backend) BeginRenderPass(tex driver.Texture, d driver.LoadDesc) {
	b.endEncoder()
	b.ensureCmdBuffer()
	f := tex.(*Texture)
	col := d.ClearColor
	var act uintptr
	switch d.Action {
	case driver.LoadActionKeep:
		act = mtlLoadActionLoad
	case driver.LoadActionClear:
		act = mtlLoadActionClear
	case driver.LoadActionInvalidate:
		act = mtlLoadActionDontCare
	}
	b.renderEnc = cmdBufferRenderEncoder(b.cmdBuffer, f.texture, act, float32(col.R), float32(col.G), float32(col.B), float32(col.A))
	if b.renderEnc == 0 {
		panic("metal: [MTLCommandBuffer renderCommandEncoderWithDescriptor:] failed")
	}
}

func (b *Backend) EndRenderPass() {
	if b.renderEnc == 0 {
		panic("no active render pass")
	}
	renderEncEnd(b.renderEnc)
	cfRelease(b.renderEnc)
	b.renderEnc = 0
}

func (b *Backend) endEncoder() {
	if b.renderEnc != 0 {
		panic("active render pass")
	}
	if b.computeEnc != 0 {
		panic("active compute pass")
	}
	if b.blitEnc != 0 {
		blitEncEnd(b.blitEnc)
		cfRelease(b.blitEnc)
		b.blitEnc = 0
	}
}

func (f *Texture) ImplementsRenderTarget() {}
