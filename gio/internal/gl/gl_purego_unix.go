// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((darwin && !ios) || (linux && !android)) && (!cgo || purego)

package gl

// OpenGL without cgo, for desktop platforms: the functions of gl_unix.go,
// loaded with purego. Keel patch, see docs/desktop-cgo-free.md. Integer and
// pointer calls go by address through glCall; the six calls with floating
// arguments use RegisterFunc wrappers, which are ABI-correct but allocate.

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
)

// glCall calls a GL function by address. Like purego.SyscallN, it keeps the
// objects behind uintptr(unsafe.Pointer(p)) arguments alive and in place
// for the call.
//
//go:uintptrescapes
func glCall(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// cString returns s with a terminating NUL.
func cString(s string) []byte { return append([]byte(s), 0) }

// goString copies the NUL-terminated string at p; "" for 0.
func goString(p uintptr) string {
	if p == 0 {
		return ""
	}
	b := *(**byte)(unsafe.Pointer(&p))
	n := 0
	for *(*byte)(unsafe.Add(unsafe.Pointer(b), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(b, n))
}

type Context any

type Functions struct {
	// Calls with floating-point arguments, through ABI-correct wrappers.
	clearColor  func(r, g, b, a float32)
	clearDepthf func(d float32)
	uniform1f   func(loc int32, v float32)
	uniform2f   func(loc int32, v0, v1 float32)
	uniform3f   func(loc int32, v0, v1, v2 float32)
	uniform4f   func(loc int32, v0, v1, v2, v3 float32)

	// Query caches.
	uints  [100]uint32
	ints   [100]int32
	floats [100]float32

	glActiveTexture                       uintptr
	glAttachShader                        uintptr
	glBindAttribLocation                  uintptr
	glBindBuffer                          uintptr
	glBindFramebuffer                     uintptr
	glBindRenderbuffer                    uintptr
	glBindTexture                         uintptr
	glBlendEquation                       uintptr
	glBlendFuncSeparate                   uintptr
	glBufferData                          uintptr
	glBufferSubData                       uintptr
	glCheckFramebufferStatus              uintptr
	glClear                               uintptr
	glClearColor                          uintptr
	glClearDepthf                         uintptr
	glCompileShader                       uintptr
	glCopyTexSubImage2D                   uintptr
	glCreateProgram                       uintptr
	glCreateShader                        uintptr
	glDeleteBuffers                       uintptr
	glDeleteFramebuffers                  uintptr
	glDeleteProgram                       uintptr
	glDeleteRenderbuffers                 uintptr
	glDeleteShader                        uintptr
	glDeleteTextures                      uintptr
	glDepthFunc                           uintptr
	glDepthMask                           uintptr
	glDisable                             uintptr
	glDisableVertexAttribArray            uintptr
	glDrawArrays                          uintptr
	glDrawElements                        uintptr
	glEnable                              uintptr
	glEnableVertexAttribArray             uintptr
	glFinish                              uintptr
	glFlush                               uintptr
	glFramebufferRenderbuffer             uintptr
	glFramebufferTexture2D                uintptr
	glGenBuffers                          uintptr
	glGenerateMipmap                      uintptr
	glGenFramebuffers                     uintptr
	glGenRenderbuffers                    uintptr
	glGenTextures                         uintptr
	glGetError                            uintptr
	glGetFramebufferAttachmentParameteriv uintptr
	glGetFloatv                           uintptr
	glGetIntegerv                         uintptr
	glGetIntegeri_v                       uintptr
	glGetProgramiv                        uintptr
	glGetProgramInfoLog                   uintptr
	glGetRenderbufferParameteriv          uintptr
	glGetShaderiv                         uintptr
	glGetShaderInfoLog                    uintptr
	glGetString                           uintptr
	glGetUniformLocation                  uintptr
	glGetVertexAttribiv                   uintptr
	glGetVertexAttribPointerv             uintptr
	glIsEnabled                           uintptr
	glLinkProgram                         uintptr
	glPixelStorei                         uintptr
	glReadPixels                          uintptr
	glRenderbufferStorage                 uintptr
	glScissor                             uintptr
	glShaderSource                        uintptr
	glTexImage2D                          uintptr
	glTexParameteri                       uintptr
	glTexSubImage2D                       uintptr
	glUniform1f                           uintptr
	glUniform1i                           uintptr
	glUniform2f                           uintptr
	glUniform3f                           uintptr
	glUniform4f                           uintptr
	glUseProgram                          uintptr
	glVertexAttribPointer                 uintptr
	glViewport                            uintptr
	glBindVertexArray                     uintptr
	glBindBufferBase                      uintptr
	glGetUniformBlockIndex                uintptr
	glUniformBlockBinding                 uintptr
	glInvalidateFramebuffer               uintptr
	glBeginQuery                          uintptr
	glDeleteQueries                       uintptr
	glDeleteVertexArrays                  uintptr
	glEndQuery                            uintptr
	glGenQueries                          uintptr
	glGenVertexArrays                     uintptr
	glGetProgramBinary                    uintptr
	glGetQueryObjectuiv                   uintptr
	glGetStringi                          uintptr
	glDispatchCompute                     uintptr
	glMemoryBarrier                       uintptr
	glMapBufferRange                      uintptr
	glUnmapBuffer                         uintptr
	glBindImageTexture                    uintptr
	glTexStorage2D                        uintptr
	glBlitFramebuffer                     uintptr
}

func NewFunctions(ctx Context, forceES bool) (*Functions, error) {
	if ctx != nil {
		panic("non-nil context")
	}
	f := new(Functions)
	err := f.load(forceES)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func dlsym(handle uintptr, s string) uintptr {
	p, err := purego.Dlsym(handle, s)
	if err != nil {
		return 0
	}
	return p
}

func dlopen(lib string) uintptr {
	h, err := purego.Dlopen(lib, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return 0
	}
	return h
}

func (f *Functions) load(forceES bool) error {
	var (
		loadErr  error
		libNames []string
		handles  []uintptr
	)
	switch {
	case runtime.GOOS == "darwin" && !forceES:
		libNames = []string{"/System/Library/Frameworks/OpenGL.framework/OpenGL"}
	case runtime.GOOS == "darwin" && forceES:
		libNames = []string{"libGLESv2.dylib"}
	case runtime.GOOS == "ios":
		libNames = []string{"/System/Library/Frameworks/OpenGLES.framework/OpenGLES"}
	case runtime.GOOS == "android":
		libNames = []string{"libGLESv3.so", "libGLESv2.so"}
	default:
		libNames = []string{"libGLESv2.so.3.0", "libGLESv2.so.2"}
	}
	for _, lib := range libNames {
		if h := dlopen(lib); h != 0 {
			handles = append(handles, h)
		}
	}
	if len(handles) == 0 {
		return fmt.Errorf("gl: no OpenGL implementation could be loaded (tried %q)", libNames)
	}
	load := func(s string) uintptr {
		for _, h := range handles {
			if f := dlsym(h, s); f != 0 {
				return f
			}
		}
		return 0
	}
	must := func(s string) uintptr {
		ptr := load(s)
		if ptr == 0 {
			loadErr = fmt.Errorf("gl: failed to load symbol %q", s)
		}
		return ptr
	}
	// GL ES 2.0 functions.
	f.glActiveTexture = must("glActiveTexture")
	f.glAttachShader = must("glAttachShader")
	f.glBindAttribLocation = must("glBindAttribLocation")
	f.glBindBuffer = must("glBindBuffer")
	f.glBindFramebuffer = must("glBindFramebuffer")
	f.glBindRenderbuffer = must("glBindRenderbuffer")
	f.glBindTexture = must("glBindTexture")
	f.glBlendEquation = must("glBlendEquation")
	f.glBlendFuncSeparate = must("glBlendFuncSeparate")
	f.glBufferData = must("glBufferData")
	f.glBufferSubData = must("glBufferSubData")
	f.glCheckFramebufferStatus = must("glCheckFramebufferStatus")
	f.glClear = must("glClear")
	f.glClearColor = must("glClearColor")
	f.glClearDepthf = must("glClearDepthf")
	f.glCompileShader = must("glCompileShader")
	f.glCopyTexSubImage2D = must("glCopyTexSubImage2D")
	f.glCreateProgram = must("glCreateProgram")
	f.glCreateShader = must("glCreateShader")
	f.glDeleteBuffers = must("glDeleteBuffers")
	f.glDeleteFramebuffers = must("glDeleteFramebuffers")
	f.glDeleteProgram = must("glDeleteProgram")
	f.glDeleteRenderbuffers = must("glDeleteRenderbuffers")
	f.glDeleteShader = must("glDeleteShader")
	f.glDeleteTextures = must("glDeleteTextures")
	f.glDepthFunc = must("glDepthFunc")
	f.glDepthMask = must("glDepthMask")
	f.glDisable = must("glDisable")
	f.glDisableVertexAttribArray = must("glDisableVertexAttribArray")
	f.glDrawArrays = must("glDrawArrays")
	f.glDrawElements = must("glDrawElements")
	f.glEnable = must("glEnable")
	f.glEnableVertexAttribArray = must("glEnableVertexAttribArray")
	f.glFinish = must("glFinish")
	f.glFlush = must("glFlush")
	f.glFramebufferRenderbuffer = must("glFramebufferRenderbuffer")
	f.glFramebufferTexture2D = must("glFramebufferTexture2D")
	f.glGenBuffers = must("glGenBuffers")
	f.glGenerateMipmap = must("glGenerateMipmap")
	f.glGenFramebuffers = must("glGenFramebuffers")
	f.glGenRenderbuffers = must("glGenRenderbuffers")
	f.glGenTextures = must("glGenTextures")
	f.glGetError = must("glGetError")
	f.glGetFramebufferAttachmentParameteriv = must("glGetFramebufferAttachmentParameteriv")
	f.glGetIntegerv = must("glGetIntegerv")
	f.glGetFloatv = must("glGetFloatv")
	f.glGetProgramiv = must("glGetProgramiv")
	f.glGetProgramInfoLog = must("glGetProgramInfoLog")
	f.glGetRenderbufferParameteriv = must("glGetRenderbufferParameteriv")
	f.glGetShaderiv = must("glGetShaderiv")
	f.glGetShaderInfoLog = must("glGetShaderInfoLog")
	f.glGetString = must("glGetString")
	f.glGetUniformLocation = must("glGetUniformLocation")
	f.glGetVertexAttribiv = must("glGetVertexAttribiv")
	f.glGetVertexAttribPointerv = must("glGetVertexAttribPointerv")
	f.glIsEnabled = must("glIsEnabled")
	f.glLinkProgram = must("glLinkProgram")
	f.glPixelStorei = must("glPixelStorei")
	f.glReadPixels = must("glReadPixels")
	f.glRenderbufferStorage = must("glRenderbufferStorage")
	f.glScissor = must("glScissor")
	f.glShaderSource = must("glShaderSource")
	f.glTexImage2D = must("glTexImage2D")
	f.glTexParameteri = must("glTexParameteri")
	f.glTexSubImage2D = must("glTexSubImage2D")
	f.glUniform1f = must("glUniform1f")
	f.glUniform1i = must("glUniform1i")
	f.glUniform2f = must("glUniform2f")
	f.glUniform3f = must("glUniform3f")
	f.glUniform4f = must("glUniform4f")
	f.glUseProgram = must("glUseProgram")
	f.glVertexAttribPointer = must("glVertexAttribPointer")
	f.glViewport = must("glViewport")

	// Extensions and GL ES 3 functions.
	f.glBindBufferBase = load("glBindBufferBase")
	f.glBindVertexArray = load("glBindVertexArray")
	f.glGetIntegeri_v = load("glGetIntegeri_v")
	f.glGetUniformBlockIndex = load("glGetUniformBlockIndex")
	f.glUniformBlockBinding = load("glUniformBlockBinding")
	f.glInvalidateFramebuffer = load("glInvalidateFramebuffer")
	f.glGetStringi = load("glGetStringi")
	// Fall back to EXT_invalidate_framebuffer if available.
	if f.glInvalidateFramebuffer == 0 {
		f.glInvalidateFramebuffer = load("glDiscardFramebufferEXT")
	}

	f.glBeginQuery = load("glBeginQuery")
	if f.glBeginQuery == 0 {
		f.glBeginQuery = load("glBeginQueryEXT")
	}
	f.glDeleteQueries = load("glDeleteQueries")
	if f.glDeleteQueries == 0 {
		f.glDeleteQueries = load("glDeleteQueriesEXT")
	}
	f.glEndQuery = load("glEndQuery")
	if f.glEndQuery == 0 {
		f.glEndQuery = load("glEndQueryEXT")
	}
	f.glGenQueries = load("glGenQueries")
	if f.glGenQueries == 0 {
		f.glGenQueries = load("glGenQueriesEXT")
	}
	f.glGetQueryObjectuiv = load("glGetQueryObjectuiv")
	if f.glGetQueryObjectuiv == 0 {
		f.glGetQueryObjectuiv = load("glGetQueryObjectuivEXT")
	}

	f.glDeleteVertexArrays = load("glDeleteVertexArrays")
	f.glGenVertexArrays = load("glGenVertexArrays")
	f.glMemoryBarrier = load("glMemoryBarrier")
	f.glDispatchCompute = load("glDispatchCompute")
	f.glMapBufferRange = load("glMapBufferRange")
	f.glUnmapBuffer = load("glUnmapBuffer")
	f.glBindImageTexture = load("glBindImageTexture")
	f.glTexStorage2D = load("glTexStorage2D")
	f.glBlitFramebuffer = load("glBlitFramebuffer")
	f.glGetProgramBinary = load("glGetProgramBinary")

	if loadErr != nil {
		return loadErr
	}
	purego.RegisterFunc(&f.clearColor, f.glClearColor)
	purego.RegisterFunc(&f.clearDepthf, f.glClearDepthf)
	purego.RegisterFunc(&f.uniform1f, f.glUniform1f)
	purego.RegisterFunc(&f.uniform2f, f.glUniform2f)
	purego.RegisterFunc(&f.uniform3f, f.glUniform3f)
	purego.RegisterFunc(&f.uniform4f, f.glUniform4f)
	return nil
}

func (f *Functions) ActiveTexture(texture Enum) {
	glCall(f.glActiveTexture, uintptr(texture))
}

func (f *Functions) AttachShader(p Program, s Shader) {
	glCall(f.glAttachShader, uintptr(p.V), uintptr(s.V))
}

func (f *Functions) BeginQuery(target Enum, query Query) {
	glCall(f.glBeginQuery, uintptr(target), uintptr(query.V))
}

func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	cname := cString(name)
	glCall(f.glBindAttribLocation, uintptr(p.V), uintptr(a), uintptr(unsafe.Pointer(unsafe.SliceData(cname))))
}

func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	glCall(f.glBindBufferBase, uintptr(target), uintptr(index), uintptr(b.V))
}

func (f *Functions) BindBuffer(target Enum, b Buffer) {
	glCall(f.glBindBuffer, uintptr(target), uintptr(b.V))
}

func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	glCall(f.glBindFramebuffer, uintptr(target), uintptr(fb.V))
}

func (f *Functions) BindRenderbuffer(target Enum, fb Renderbuffer) {
	glCall(f.glBindRenderbuffer, uintptr(target), uintptr(fb.V))
}

func (f *Functions) BindImageTexture(unit int, t Texture, level int, layered bool, layer int, access, format Enum) {
	l := uintptr(FALSE)
	if layered {
		l = TRUE
	}
	glCall(f.glBindImageTexture, uintptr(unit), uintptr(t.V), uintptr(level), l, uintptr(layer), uintptr(access), uintptr(format))
}

func (f *Functions) BindTexture(target Enum, t Texture) {
	glCall(f.glBindTexture, uintptr(target), uintptr(t.V))
}

func (f *Functions) BindVertexArray(a VertexArray) {
	glCall(f.glBindVertexArray, uintptr(a.V))
}

func (f *Functions) BlendEquation(mode Enum) {
	glCall(f.glBlendEquation, uintptr(mode))
}

func (f *Functions) BlendFuncSeparate(srcRGB, dstRGB, srcA, dstA Enum) {
	glCall(f.glBlendFuncSeparate, uintptr(srcRGB), uintptr(dstRGB), uintptr(srcA), uintptr(dstA))
}

func (f *Functions) BlitFramebuffer(sx0, sy0, sx1, sy1, dx0, dy0, dx1, dy1 int, mask Enum, filter Enum) {
	glCall(f.glBlitFramebuffer,
		uintptr(sx0), uintptr(sy0), uintptr(sx1), uintptr(sy1),
		uintptr(dx0), uintptr(dy0), uintptr(dx1), uintptr(dy1),
		uintptr(mask), uintptr(filter),
	)
}

func (f *Functions) BufferData(target Enum, size int, usage Enum, data []byte) {
	glCall(f.glBufferData, uintptr(target), uintptr(size), uintptr(unsafe.Pointer(unsafe.SliceData(data))), uintptr(usage))
}

func (f *Functions) BufferSubData(target Enum, offset int, src []byte) {
	glCall(f.glBufferSubData, uintptr(target), uintptr(offset), uintptr(len(src)),
		uintptr(unsafe.Pointer(unsafe.SliceData(src))))
}

func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(uint32(glCall(f.glCheckFramebufferStatus, uintptr(target))))
}

func (f *Functions) Clear(mask Enum) {
	glCall(f.glClear, uintptr(mask))
}

func (f *Functions) ClearColor(red float32, green float32, blue float32, alpha float32) {
	f.clearColor(red, green, blue, alpha)
}

func (f *Functions) ClearDepthf(d float32) {
	f.clearDepthf(d)
}

func (f *Functions) CompileShader(s Shader) {
	glCall(f.glCompileShader, uintptr(s.V))
}

func (f *Functions) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	glCall(f.glCopyTexSubImage2D, uintptr(target), uintptr(level), uintptr(xoffset), uintptr(yoffset), uintptr(x), uintptr(y), uintptr(width), uintptr(height))
}

func (f *Functions) CreateBuffer() Buffer {
	glCall(f.glGenBuffers, 1, uintptr(unsafe.Pointer(&f.uints[0])))
	return Buffer{uint(f.uints[0])}
}

func (f *Functions) CreateFramebuffer() Framebuffer {
	glCall(f.glGenFramebuffers, 1, uintptr(unsafe.Pointer(&f.uints[0])))
	return Framebuffer{uint(f.uints[0])}
}

func (f *Functions) CreateProgram() Program {
	return Program{uint(uint32(glCall(f.glCreateProgram)))}
}

func (f *Functions) CreateQuery() Query {
	glCall(f.glGenQueries, 1, uintptr(unsafe.Pointer(&f.uints[0])))
	return Query{uint(f.uints[0])}
}

func (f *Functions) CreateRenderbuffer() Renderbuffer {
	glCall(f.glGenRenderbuffers, 1, uintptr(unsafe.Pointer(&f.uints[0])))
	return Renderbuffer{uint(f.uints[0])}
}

func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader{uint(uint32(glCall(f.glCreateShader, uintptr(ty))))}
}

func (f *Functions) CreateTexture() Texture {
	glCall(f.glGenTextures, 1, uintptr(unsafe.Pointer(&f.uints[0])))
	return Texture{uint(f.uints[0])}
}

func (f *Functions) CreateVertexArray() VertexArray {
	glCall(f.glGenVertexArrays, 1, uintptr(unsafe.Pointer(&f.uints[0])))
	return VertexArray{uint(f.uints[0])}
}

func (f *Functions) DeleteBuffer(v Buffer) {
	f.uints[0] = uint32(v.V)
	glCall(f.glDeleteBuffers, 1, uintptr(unsafe.Pointer(&f.uints[0])))
}

func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	f.uints[0] = uint32(v.V)
	glCall(f.glDeleteFramebuffers, 1, uintptr(unsafe.Pointer(&f.uints[0])))
}

func (f *Functions) DeleteProgram(p Program) {
	glCall(f.glDeleteProgram, uintptr(p.V))
}

func (f *Functions) DeleteQuery(query Query) {
	f.uints[0] = uint32(query.V)
	glCall(f.glDeleteQueries, 1, uintptr(unsafe.Pointer(&f.uints[0])))
}

func (f *Functions) DeleteVertexArray(array VertexArray) {
	f.uints[0] = uint32(array.V)
	glCall(f.glDeleteVertexArrays, 1, uintptr(unsafe.Pointer(&f.uints[0])))
}

func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	f.uints[0] = uint32(v.V)
	glCall(f.glDeleteRenderbuffers, 1, uintptr(unsafe.Pointer(&f.uints[0])))
}

func (f *Functions) DeleteShader(s Shader) {
	glCall(f.glDeleteShader, uintptr(s.V))
}

func (f *Functions) DeleteTexture(v Texture) {
	f.uints[0] = uint32(v.V)
	glCall(f.glDeleteTextures, 1, uintptr(unsafe.Pointer(&f.uints[0])))
}

func (f *Functions) DepthFunc(v Enum) {
	glCall(f.glDepthFunc, uintptr(v))
}

func (f *Functions) DepthMask(mask bool) {
	m := uintptr(FALSE)
	if mask {
		m = uintptr(TRUE)
	}
	glCall(f.glDepthMask, m)
}

func (f *Functions) DisableVertexAttribArray(a Attrib) {
	glCall(f.glDisableVertexAttribArray, uintptr(a))
}

func (f *Functions) Disable(cap Enum) {
	glCall(f.glDisable, uintptr(cap))
}

func (f *Functions) DrawArrays(mode Enum, first int, count int) {
	glCall(f.glDrawArrays, uintptr(mode), uintptr(first), uintptr(count))
}

func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	glCall(f.glDrawElements, uintptr(mode), uintptr(count), uintptr(ty), uintptr(offset))
}

func (f *Functions) DispatchCompute(x, y, z int) {
	glCall(f.glDispatchCompute, uintptr(x), uintptr(y), uintptr(z))
}

func (f *Functions) Enable(cap Enum) {
	glCall(f.glEnable, uintptr(cap))
}

func (f *Functions) EndQuery(target Enum) {
	glCall(f.glEndQuery, uintptr(target))
}

func (f *Functions) EnableVertexAttribArray(a Attrib) {
	glCall(f.glEnableVertexAttribArray, uintptr(a))
}

func (f *Functions) Finish() {
	glCall(f.glFinish)
}

func (f *Functions) Flush() {
	// As the cgo binding does: its glFlush trampoline is given glFinish.
	glCall(f.glFinish)
}

func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	glCall(f.glFramebufferRenderbuffer, uintptr(target), uintptr(attachment), uintptr(renderbuffertarget), uintptr(renderbuffer.V))
}

func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	glCall(f.glFramebufferTexture2D, uintptr(target), uintptr(attachment), uintptr(texTarget), uintptr(t.V), uintptr(level))
}

func (f *Functions) GenerateMipmap(target Enum) {
	glCall(f.glGenerateMipmap, uintptr(target))
}

func (c *Functions) GetBinding(pname Enum) Object {
	return Object{uint(c.GetInteger(pname))}
}

func (c *Functions) GetBindingi(pname Enum, idx int) Object {
	return Object{uint(c.GetIntegeri(pname, idx))}
}

func (f *Functions) GetError() Enum {
	return Enum(uint32(glCall(f.glGetError)))
}

func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	glCall(f.glGetRenderbufferParameteriv, uintptr(target), uintptr(pname), uintptr(unsafe.Pointer(&f.ints[0])))
	return int(f.ints[0])
}

func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	glCall(f.glGetFramebufferAttachmentParameteriv, uintptr(target), uintptr(attachment), uintptr(pname), uintptr(unsafe.Pointer(&f.ints[0])))
	return int(f.ints[0])
}

func (f *Functions) GetFloat4(pname Enum) [4]float32 {
	glCall(f.glGetFloatv, uintptr(pname), uintptr(unsafe.Pointer(&f.floats[0])))
	var r [4]float32
	for i := range r {
		r[i] = float32(f.floats[i])
	}
	return r
}

func (f *Functions) GetFloat(pname Enum) float32 {
	glCall(f.glGetFloatv, uintptr(pname), uintptr(unsafe.Pointer(&f.floats[0])))
	return float32(f.floats[0])
}

func (f *Functions) GetInteger4(pname Enum) [4]int {
	glCall(f.glGetIntegerv, uintptr(pname), uintptr(unsafe.Pointer(&f.ints[0])))
	var r [4]int
	for i := range r {
		r[i] = int(f.ints[i])
	}
	return r
}

func (f *Functions) GetInteger(pname Enum) int {
	glCall(f.glGetIntegerv, uintptr(pname), uintptr(unsafe.Pointer(&f.ints[0])))
	return int(f.ints[0])
}

func (f *Functions) GetIntegeri(pname Enum, idx int) int {
	glCall(f.glGetIntegeri_v, uintptr(pname), uintptr(idx), uintptr(unsafe.Pointer(&f.ints[0])))
	return int(f.ints[0])
}

func (f *Functions) GetProgrami(p Program, pname Enum) int {
	glCall(f.glGetProgramiv, uintptr(p.V), uintptr(pname), uintptr(unsafe.Pointer(&f.ints[0])))
	return int(f.ints[0])
}

func (f *Functions) GetProgramBinary(p Program) []byte {
	sz := f.GetProgrami(p, PROGRAM_BINARY_LENGTH)
	buf := make([]byte, sz)
	var format uint32
	glCall(f.glGetProgramBinary, uintptr(p.V), uintptr(sz), 0, uintptr(unsafe.Pointer(&format)), uintptr(unsafe.Pointer(unsafe.SliceData(buf))))
	return buf
}

func (f *Functions) GetProgramInfoLog(p Program) string {
	n := f.GetProgrami(p, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	glCall(f.glGetProgramInfoLog, uintptr(p.V), uintptr(len(buf)), 0, uintptr(unsafe.Pointer(unsafe.SliceData(buf))))
	return string(buf)
}

func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	glCall(f.glGetQueryObjectuiv, uintptr(query.V), uintptr(pname), uintptr(unsafe.Pointer(&f.uints[0])))
	return uint(f.uints[0])
}

func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	glCall(f.glGetShaderiv, uintptr(s.V), uintptr(pname), uintptr(unsafe.Pointer(&f.ints[0])))
	return int(f.ints[0])
}

func (f *Functions) GetShaderInfoLog(s Shader) string {
	n := f.GetShaderi(s, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	glCall(f.glGetShaderInfoLog, uintptr(s.V), uintptr(len(buf)), 0,
		uintptr(unsafe.Pointer(unsafe.SliceData(buf))))
	return string(buf)
}

func (f *Functions) getStringi(pname Enum, index int) string {
	return goString(glCall(f.glGetStringi, uintptr(pname), uintptr(index)))
}

func (f *Functions) GetString(pname Enum) string {
	switch {
	case runtime.GOOS == "darwin" && pname == EXTENSIONS:
		// macOS OpenGL 3 core profile doesn't support glGetString(GL_EXTENSIONS).
		// Use glGetStringi(GL_EXTENSIONS, <index>).
		var exts []string
		nexts := f.GetInteger(NUM_EXTENSIONS)
		for i := 0; i < nexts; i++ {
			ext := f.getStringi(EXTENSIONS, i)
			exts = append(exts, ext)
		}
		return strings.Join(exts, " ")
	default:
		return goString(glCall(f.glGetString, uintptr(pname)))
	}
}

func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	cname := cString(name)
	return uint(uint32(glCall(f.glGetUniformBlockIndex, uintptr(p.V), uintptr(unsafe.Pointer(unsafe.SliceData(cname))))))
}

func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	cname := cString(name)
	return Uniform{int(int32(glCall(f.glGetUniformLocation, uintptr(p.V), uintptr(unsafe.Pointer(unsafe.SliceData(cname))))))}
}

func (f *Functions) GetVertexAttrib(index int, pname Enum) int {
	glCall(f.glGetVertexAttribiv, uintptr(index), uintptr(pname), uintptr(unsafe.Pointer(&f.ints[0])))
	return int(f.ints[0])
}

func (f *Functions) GetVertexAttribBinding(index int, pname Enum) Object {
	return Object{uint(f.GetVertexAttrib(index, pname))}
}

func (f *Functions) GetVertexAttribPointer(index int, pname Enum) uintptr {
	var ptr uintptr
	glCall(f.glGetVertexAttribPointerv, uintptr(index), uintptr(pname), uintptr(unsafe.Pointer(&ptr)))
	return ptr
}

func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	// Framebuffer invalidation is just a hint and can safely be ignored.
	if f.glInvalidateFramebuffer != 0 {
		a := uint32(attachment)
		glCall(f.glInvalidateFramebuffer, uintptr(target), 1, uintptr(unsafe.Pointer(&a)))
	}
}

func (f *Functions) IsEnabled(cap Enum) bool {
	return uint8(glCall(f.glIsEnabled, uintptr(cap))) == TRUE
}

func (f *Functions) LinkProgram(p Program) {
	glCall(f.glLinkProgram, uintptr(p.V))
}

func (f *Functions) PixelStorei(pname Enum, param int) {
	glCall(f.glPixelStorei, uintptr(pname), uintptr(param))
}

func (f *Functions) MemoryBarrier(barriers Enum) {
	glCall(f.glMemoryBarrier, uintptr(barriers))
}

func (f *Functions) MapBufferRange(target Enum, offset, length int, access Enum) []byte {
	p := glCall(f.glMapBufferRange, uintptr(target), uintptr(offset), uintptr(length), uintptr(access))
	if p == 0 {
		return nil
	}
	return unsafe.Slice(*(**byte)(unsafe.Pointer(&p)), length) // driver memory, valid until UnmapBuffer
}

func (f *Functions) Scissor(x, y, width, height int32) {
	glCall(f.glScissor, uintptr(x), uintptr(y), uintptr(width), uintptr(height))
}

func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	glCall(f.glReadPixels, uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(format), uintptr(ty),
		uintptr(unsafe.Pointer(unsafe.SliceData(data))))
}

func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	glCall(f.glRenderbufferStorage, uintptr(target), uintptr(internalformat), uintptr(width), uintptr(height))
}

func (f *Functions) ShaderSource(s Shader, src string) {
	strlen := int32(len(src))
	pstr := unsafe.StringData(src)
	glCall(f.glShaderSource, uintptr(s.V), 1, uintptr(unsafe.Pointer(&pstr)), uintptr(unsafe.Pointer(&strlen)))
}

func (f *Functions) TexImage2D(target Enum, level int, internalFormat Enum, width int, height int, format Enum, ty Enum) {
	glCall(f.glTexImage2D, uintptr(target), uintptr(level), uintptr(internalFormat), uintptr(width), uintptr(height), 0, uintptr(format), uintptr(ty), 0)
}

func (f *Functions) TexStorage2D(target Enum, levels int, internalFormat Enum, width, height int) {
	glCall(f.glTexStorage2D, uintptr(target), uintptr(levels), uintptr(internalFormat), uintptr(width), uintptr(height))
}

func (f *Functions) TexSubImage2D(target Enum, level int, x int, y int, width int, height int, format Enum, ty Enum, data []byte) {
	glCall(f.glTexSubImage2D, uintptr(target), uintptr(level), uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(format), uintptr(ty),
		uintptr(unsafe.Pointer(unsafe.SliceData(data))))
}

func (f *Functions) TexParameteri(target, pname Enum, param int) {
	glCall(f.glTexParameteri, uintptr(target), uintptr(pname), uintptr(param))
}

func (f *Functions) UniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint) {
	glCall(f.glUniformBlockBinding, uintptr(p.V), uintptr(uniformBlockIndex), uintptr(uniformBlockBinding))
}

func (f *Functions) Uniform1f(dst Uniform, v float32) {
	f.uniform1f(int32(dst.V), v)
}

func (f *Functions) Uniform1i(dst Uniform, v int) {
	glCall(f.glUniform1i, uintptr(dst.V), uintptr(v))
}

func (f *Functions) Uniform2f(dst Uniform, v0 float32, v1 float32) {
	f.uniform2f(int32(dst.V), v0, v1)
}

func (f *Functions) Uniform3f(dst Uniform, v0 float32, v1 float32, v2 float32) {
	f.uniform3f(int32(dst.V), v0, v1, v2)
}

func (f *Functions) Uniform4f(dst Uniform, v0 float32, v1 float32, v2 float32, v3 float32) {
	f.uniform4f(int32(dst.V), v0, v1, v2, v3)
}

func (f *Functions) UseProgram(p Program) {
	glCall(f.glUseProgram, uintptr(p.V))
}

func (f *Functions) UnmapBuffer(target Enum) bool {
	return uint8(glCall(f.glUnmapBuffer, uintptr(target))) == TRUE
}

func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride int, offset int) {
	var n uintptr = FALSE
	if normalized {
		n = TRUE
	}
	glCall(f.glVertexAttribPointer, uintptr(dst), uintptr(size), uintptr(ty), n, uintptr(stride), uintptr(offset))
}

func (f *Functions) Viewport(x int, y int, width int, height int) {
	glCall(f.glViewport, uintptr(x), uintptr(y), uintptr(width), uintptr(height))
}
