// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"encoding/binary"
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/gpu/internal/driver"
)

// The binding tests run under both bindings: go test, and go test -tags
// purego. They cover what the pixel tests do not reach: errors coming back
// from Metal, and compute dispatch, whose sizes are structs.

func TestLibraryError(t *testing.T) {
	b := testBackend(t)
	lib, err := newLibrary(b.dev, "broken", "not a metallib")
	if err == nil || lib != 0 {
		t.Fatalf("newLibrary of garbage = %#x, %v; want an error", lib, err)
	}
	if strings.TrimSpace(err.Error()) == "" {
		t.Fatal("empty error description")
	}
	if lib := quadLibrary(b.dev, "this is not Metal"); lib != 0 {
		cfRelease(lib)
		t.Fatal("quadLibrary compiled invalid source")
	}
}

const fillKernel = `
#include <metal_stdlib>
using namespace metal;

kernel void fill(device uint *out [[buffer(0)]], uint id [[thread_position_in_grid]], uint lid [[thread_position_in_threadgroup]]) {
	out[id] = id*3 + 1 + (lid << 16);
}
`

func TestComputeDispatch(t *testing.T) {
	b := testBackend(t)
	lib := quadLibrary(b.dev, fillKernel)
	if lib == 0 {
		t.Fatal("kernel did not compile")
	}
	defer cfRelease(lib)
	fn := libraryNewFunction(lib, "fill")
	if fn == 0 {
		t.Fatal("kernel function missing")
	}
	defer cfRelease(fn)
	pipe, err := newComputePipeline(b.dev, fn)
	if err != nil {
		t.Fatal(err)
	}
	const groups, size = 4, 16
	prog := &Program{pipeline: pipe, groupSize: [3]int{size, 1, 1}}
	defer prog.Release()
	buf, err := b.NewBuffer(driver.BufferBindingShaderStorageWrite, groups*size*4)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Release()
	b.BeginFrame(nil, false, image.Point{})
	b.BeginCompute()
	b.BindProgram(prog)
	b.BindStorageBuffer(0, buf)
	b.DispatchCompute(groups, 1, 1)
	b.EndCompute()
	got := make([]byte, groups*size*4)
	if err := buf.Download(got); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < groups*size; i++ {
		want := uint32(i*3 + 1 + (i%size)<<16)
		if v := binary.LittleEndian.Uint32(got[i*4:]); v != want {
			t.Fatalf("out[%d] = %#x, want %#x", i, v, want)
		}
	}
}
