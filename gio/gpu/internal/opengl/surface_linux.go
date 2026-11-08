// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android

package opengl

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"github.com/dyike/keel/gio/gpu/internal/driver"
	"github.com/dyike/keel/gio/internal/gl"
	"github.com/dyike/keel/gio/internal/surface"
	"github.com/ebitengine/purego"
)

const externalTarget gl.Enum = 0x8D65

type retiredSurface struct {
	fence uintptr
	free  func()
}

type surfaceState struct {
	display  uintptr
	enabled  bool
	external bool
	disable  func()
	retired  []retiredSurface
}

var dmaEGL struct {
	sync.Once
	current, query, createImage, destroyImage, targetImage uintptr
	queryString                                            func(uintptr, int32) unsafe.Pointer
	createSync, waitSync, destroySync, flush               uintptr
}

func dmaCall(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

func dmaString(p unsafe.Pointer) string {
	if p == nil {
		return ""
	}
	var b []byte
	for i := 0; i < 1<<16; i++ {
		c := *(*byte)(unsafe.Add(p, i))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

func loadDMAEGL() {
	h, err := purego.Dlopen("libEGL.so.1", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return
	}
	dmaEGL.current, _ = purego.Dlsym(h, "eglGetCurrentDisplay")
	dmaEGL.query, _ = purego.Dlsym(h, "eglQueryString")
	if dmaEGL.query != 0 {
		purego.RegisterFunc(&dmaEGL.queryString, dmaEGL.query)
	}
	var proc func(string) uintptr
	purego.RegisterLibFunc(&proc, h, "eglGetProcAddress")
	dmaEGL.createImage = proc("eglCreateImageKHR")
	dmaEGL.destroyImage = proc("eglDestroyImageKHR")
	dmaEGL.targetImage = proc("glEGLImageTargetTexture2DOES")
	dmaEGL.createSync = proc("eglCreateSyncKHR")
	dmaEGL.waitSync = proc("eglClientWaitSyncKHR")
	dmaEGL.destroySync = proc("eglDestroySyncKHR")
	dmaEGL.flush = proc("glFlush")
}

func (b *Backend) initSurfaces() {
	// Shared contexts need additional external-target state restoration.
	if b.sharedCtx || !b.gles || !hasExtension(strings.Fields(b.funcs.GetString(gl.EXTENSIONS)), "GL_OES_EGL_image_external") {
		return
	}
	dmaEGL.Do(loadDMAEGL)
	for _, fn := range []uintptr{dmaEGL.current, dmaEGL.query, dmaEGL.createImage, dmaEGL.destroyImage, dmaEGL.targetImage, dmaEGL.createSync, dmaEGL.waitSync, dmaEGL.destroySync, dmaEGL.flush} {
		if fn == 0 {
			return
		}
	}
	display := dmaCall(dmaEGL.current)
	if display == 0 {
		return
	}
	exts := strings.Fields(dmaString(dmaEGL.queryString(display, 0x3055)))
	if !hasExtension(exts, "EGL_EXT_image_dma_buf_import") || !hasExtension(exts, "EGL_EXT_image_dma_buf_import_modifiers") || !hasExtension(exts, "EGL_KHR_fence_sync") {
		return
	}
	b.surfaces = surfaceState{display: display, enabled: true, disable: surface.EnableDMABuf()}
}

func dmaAttributes(img surface.DMABuf) ([]int32, error) {
	if img.Width <= 0 || img.Height <= 0 || len(img.Planes) < 1 || len(img.Planes) > 4 {
		return nil, errors.New("opengl: invalid DMA-BUF")
	}
	attrs := []int32{0x3057, int32(img.Width), 0x3056, int32(img.Height), 0x3271, int32(img.FourCC)}
	fd := [...]int32{0x3272, 0x3275, 0x3278, 0x3440}
	lo := [...]int32{0x3443, 0x3445, 0x3447, 0x3449}
	for i, p := range img.Planes {
		attrs = append(attrs, fd[i], int32(p.FD), fd[i]+1, int32(p.Offset), fd[i]+2, int32(p.Pitch))
		// DRM_FORMAT_MOD_INVALID means the producer supplied no modifier.
		if p.Modifier != 0x00ffffffffffffff {
			attrs = append(attrs, lo[i], int32(p.Modifier), lo[i]+1, int32(p.Modifier>>32))
		}
	}
	if img.FourCC == 0x3231564e || img.FourCC == 0x30313050 { // NV12 / P010
		space := int32(0x327F)
		if img.ColorSpace == 709 {
			space = 0x3280
		} else if img.ColorSpace == 2020 {
			space = 0x3281
		}
		rangeHint := int32(0x3283)
		if img.FullRange {
			rangeHint = 0x3282
		}
		attrs = append(attrs, 0x327B, space, 0x327C, rangeHint)
	}
	return append(attrs, 0x3038), nil
}

func (b *Backend) NewSurfaceTexture(handle uintptr, width, height int, minFilter, magFilter driver.TextureFilter) (driver.Texture, error) {
	img, release, ok := surface.AcquireDMABuf(handle)
	if !ok {
		return nil, errors.New("opengl: expired DMA-BUF")
	}
	fail := func(err error) (driver.Texture, error) {
		if img.OnError != nil {
			img.OnError(err)
		}
		release()
		return nil, err
	}
	if !b.surfaces.enabled || width != img.Width || height != img.Height {
		return fail(errors.New("opengl: DMA-BUF import unavailable or size mismatch"))
	}
	attrs, err := dmaAttributes(img)
	if err != nil {
		return fail(err)
	}
	eglImage := dmaCall(dmaEGL.createImage, b.surfaces.display, 0, 0x3270, 0, uintptr(unsafe.Pointer(&attrs[0])))
	runtime.KeepAlive(attrs)
	if eglImage == 0 {
		return fail(fmt.Errorf("opengl: EGL rejected DMA-BUF fourcc=%#x modifier=%#x", img.FourCC, img.Planes[0].Modifier))
	}
	t := &texture{backend: b, obj: b.funcs.CreateTexture(), width: width, height: height, external: true, bindings: driver.BufferBindingTexture}
	// Binding goes through the same state transition used during drawing.
	b.BindTexture(0, t)
	min, _ := toTexFilter(minFilter)
	mag, _ := toTexFilter(magFilter)
	b.funcs.TexParameteri(externalTarget, gl.TEXTURE_MIN_FILTER, min)
	b.funcs.TexParameteri(externalTarget, gl.TEXTURE_MAG_FILTER, mag)
	b.funcs.TexParameteri(externalTarget, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.funcs.TexParameteri(externalTarget, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	dmaCall(dmaEGL.targetImage, uintptr(externalTarget), eglImage)
	if err := glErr(b.funcs); err != nil {
		b.funcs.DeleteTexture(t.obj)
		dmaCall(dmaEGL.destroyImage, b.surfaces.display, eglImage)
		return fail(err)
	}
	t.surfaceRelease = func() {
		b.funcs.DeleteTexture(t.obj)
		dmaCall(dmaEGL.destroyImage, b.surfaces.display, eglImage)
		release()
	}
	return t, nil
}

func (b *Backend) externalTexture() bool { return b.surfaces.external }

func (b *Backend) bindSurfaceTexture(unit int, tex *texture) bool {
	if !b.surfaces.enabled {
		return false
	}
	if unit == 0 {
		changed := b.surfaces.external != tex.external
		b.surfaces.external = tex.external
		if changed && b.state.pipeline != nil {
			b.BindPipeline(b.state.pipeline)
		}
	}
	if tex.external {
		b.glstate.bindTexture(b.funcs, unit, gl.Texture{})
		b.funcs.BindTexture(externalTarget, tex.obj)
		return true
	}
	return false
}

// The external sampler performs YUV conversion in the texture unit. Decode
// sRGB before Gio's blending, clipping and opacity passes, just as an ordinary
// sRGB image texture does. No intermediate RGBA image or upload is created.
func (b *Backend) newExternalProgram(desc driver.PipelineDesc) (*program, error) {
	if !b.surfaces.enabled {
		return nil, nil
	}
	fsh := desc.FragmentShader.(*glshader)
	src := b.glslFor(fsh.src)
	const decl = "uniform mediump sampler2D tex;"
	if !strings.Contains(src, decl) {
		return nil, nil
	}
	src = strings.Replace(src, "#version 100", "#version 100\n#extension GL_OES_EGL_image_external : require", 1)
	src = strings.Replace(src, decl, "uniform mediump samplerExternalOES tex;", 1)
	src = strings.ReplaceAll(src, "texture2D(tex,", "keelSampleExternal(tex,")
	const sample = `
vec4 keelSampleExternal(samplerExternalOES image, vec2 at) {
    vec4 c = texture2D(image, at);
    vec3 s = clamp(c.rgb, 0.0, 1.0);
    vec3 lo = s / 12.92;
    vec3 hi = pow((s + 0.055) / 1.055, vec3(2.4));
    c.rgb = mix(lo, hi, step(vec3(0.04045), s));
    return c;
}
`
	src = strings.Replace(src, "void main()", sample+"\nvoid main()", 1)
	sh, err := gl.CreateShader(b.funcs, gl.FRAGMENT_SHADER, src)
	if err != nil {
		return nil, fmt.Errorf("opengl: DMA-BUF sampler: %w", err)
	}
	defer b.funcs.DeleteShader(sh)
	alt := *fsh
	alt.obj = sh
	desc.FragmentShader = &alt
	return b.newProgram(desc)
}

func (b *Backend) retireSurface(t *texture) {
	free := t.surfaceRelease
	t.surfaceRelease = nil
	fence := dmaCall(dmaEGL.createSync, b.surfaces.display, 0x30F9, 0)
	if fence == 0 {
		// Correctness on resource exhaustion: don't return a live VA surface.
		b.funcs.Finish()
		free()
		return
	}
	dmaCall(dmaEGL.flush)
	b.surfaces.retired = append(b.surfaces.retired, retiredSurface{fence, free})
	b.collectSurfaces(len(b.surfaces.retired) >= 8)
}

func (b *Backend) collectSurfaces(wait bool) {
	for len(b.surfaces.retired) > 0 {
		r := b.surfaces.retired[0]
		timeout := uintptr(0)
		if wait {
			timeout = ^uintptr(0)
		}
		result := dmaCall(dmaEGL.waitSync, b.surfaces.display, r.fence, 0, timeout)
		if result == 0x30F5 {
			return
		} // EGL_TIMEOUT_EXPIRED_KHR
		if result != 0x30F6 {
			b.funcs.Finish()
		} // failed wait must not free in-flight data
		dmaCall(dmaEGL.destroySync, b.surfaces.display, r.fence)
		r.free()
		b.surfaces.retired[0] = retiredSurface{}
		b.surfaces.retired = b.surfaces.retired[1:]
		wait = false
	}
}

func (b *Backend) releaseSurfaces() {
	for len(b.surfaces.retired) > 0 {
		b.collectSurfaces(true)
	}
	if b.surfaces.disable != nil {
		b.surfaces.disable()
		b.surfaces.disable = nil
	}
}
