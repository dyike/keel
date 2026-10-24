// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((darwin && !ios) || (linux && !android)) && (!cgo || purego)

package gl

import (
	"runtime"
	"testing"
	"unsafe"
)

// Every required GL symbol resolves, and the float wrappers are registered.
// Rendering through these functions is covered by the headless tests with
// -tags nometal on macOS.
func TestPuregoFunctionsLoad(t *testing.T) {
	f, err := NewFunctions(nil, false)
	if err != nil {
		if runtime.GOOS == "darwin" {
			t.Fatal(err)
		}
		t.Skip(err) // no GL ES library on this machine
	}
	if f.glClear == 0 || f.glDrawElements == 0 || f.clearColor == nil || f.uniform4f == nil {
		t.Fatal("functions missing after load")
	}
}

func TestPuregoStrings(t *testing.T) {
	c := cString("glClear")
	if len(c) != len("glClear")+1 || c[len(c)-1] != 0 {
		t.Fatalf("cString %q", c)
	}
	if got := goString(uintptr(unsafe.Pointer(&c[0]))); got != "glClear" {
		t.Fatalf("goString %q", got)
	}
	runtime.KeepAlive(c)
	if goString(0) != "" {
		t.Fatal("goString(0)")
	}
}
