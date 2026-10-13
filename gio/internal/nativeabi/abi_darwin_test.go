// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && (amd64 || arm64)

package nativeabi

import (
	"runtime"
	"structs"
	"sync"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

type point struct {
	_    structs.HostLayout
	X, Y float64
}

type rect struct {
	_            structs.HostLayout
	Origin, Size point
}

type nsrange struct {
	_                structs.HostLayout
	Location, Length uintptr
}

var (
	registerOnce sync.Once
	probeClass   objc.Class
	childClass   objc.Class
	registerErr  error
)

// Register exactly once: Objective-C classes and purego callbacks are permanent.
// No window, retained Go pointer, or mutable per-request callback state is needed.
func registerProbe() {
	registerOnce.Do(func() {
		_, registerErr = purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if registerErr != nil {
			return
		}
		probeClass, registerErr = objc.RegisterClass("GioNativeABIProbe", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
			{Cmd: objc.RegisterName("selectedRange"), Fn: func(objc.ID, objc.SEL) nsrange {
				return nsrange{Location: ^uintptr(0), Length: 1 << 40}
			}},
			{Cmd: objc.RegisterName("setMarkedText:selectedRange:replacementRange:"), Fn: func(_ objc.ID, _ objc.SEL, object objc.ID, a, b nsrange) nsrange {
				// self, _cmd, object and two ranges exhaust amd64 integer registers.
				return nsrange{Location: a.Location ^ b.Length ^ uintptr(object), Length: a.Length ^ b.Location}
			}},
			{Cmd: objc.RegisterName("characterIndexForPoint:"), Fn: func(_ objc.ID, _ objc.SEL, p point) uintptr {
				if p.X == 17.25 && p.Y == -39.5 {
					return 12345
				}
				return 0
			}},
			{Cmd: objc.RegisterName("firstRectForCharacterRange:actualRange:"), Fn: func(_ objc.ID, _ objc.SEL, r nsrange, actual *nsrange) rect {
				if actual != nil {
					*actual = r
				}
				return rect{Origin: point{X: float64(r.Location), Y: -2.5}, Size: point{X: 31.25, Y: float64(r.Length)}}
			}},
			{Cmd: objc.RegisterName("echoRect:"), Fn: func(_ objc.ID, _ objc.SEL, r rect) rect { return r }},
		})
		if registerErr != nil {
			return
		}
		childClass, registerErr = objc.RegisterClass("GioNativeABIChildProbe", probeClass, nil, nil, []objc.MethodDef{
			{Cmd: objc.RegisterName("echoRect:"), Fn: func(self objc.ID, sel objc.SEL, r rect) rect {
				return objc.SendSuper[rect](self, sel, r)
			}},
		})
	})
}

// objc.Send enters the system objc_msgSend implementation, which invokes our
// IMP through the native ABI. This is not a direct Go-to-Go callback test.
func TestTextInputAggregateCallbacks(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	registerProbe()
	if registerErr != nil {
		t.Fatal(registerErr)
	}
	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
	defer pool.Send(objc.RegisterName("drain"))
	obj := objc.ID(probeClass).Send(objc.RegisterName("new"))
	defer obj.Send(objc.RegisterName("release"))

	want := nsrange{Location: ^uintptr(0), Length: 1 << 40}
	if got := objc.Send[nsrange](obj, objc.RegisterName("selectedRange")); got != want {
		t.Fatalf("range return: got %+v, want %+v", got, want)
	}
	a, b := nsrange{Location: 7, Length: 1 << 42}, nsrange{Location: 1 << 43, Length: 11}
	want = nsrange{Location: a.Location ^ b.Length ^ uintptr(obj), Length: a.Length ^ b.Location}
	if got := objc.Send[nsrange](obj, objc.RegisterName("setMarkedText:selectedRange:replacementRange:"), obj, a, b); got != want {
		t.Fatalf("two-range argument spill: got %+v, want %+v", got, want)
	}
	if got := objc.Send[uintptr](obj, objc.RegisterName("characterIndexForPoint:"), point{X: 17.25, Y: -39.5}); got != 12345 {
		t.Fatalf("point argument: got %d", got)
	}

	// Native-allocated output storage avoids passing retained Go pointers and
	// checks canaries around the actualRange output buffer.
	var malloc func(uintptr) unsafe.Pointer
	var free func(unsafe.Pointer)
	purego.RegisterLibFunc(&malloc, purego.RTLD_DEFAULT, "malloc")
	purego.RegisterLibFunc(&free, purego.RTLD_DEFAULT, "free")
	mem := malloc(unsafe.Sizeof(nsrange{}) * 3)
	if mem == nil {
		t.Fatal("malloc failed")
	}
	defer free(mem)
	buffer := unsafe.Slice((*nsrange)(mem), 3)
	canary := nsrange{Location: 0xabc, Length: 0xdef}
	buffer[0], buffer[2] = canary, canary
	rng := nsrange{Location: 19, Length: 41}
	wantRect := rect{Origin: point{X: 19, Y: -2.5}, Size: point{X: 31.25, Y: 41}}
	for _, actual := range []*nsrange{&buffer[1], nil} {
		got := objc.Send[rect](obj, objc.RegisterName("firstRectForCharacterRange:actualRange:"), rng, actual)
		if got != wantRect {
			t.Fatalf("rect return/hidden pointer: got %+v, want %+v", got, wantRect)
		}
	}
	if buffer[1] != rng || buffer[0] != canary || buffer[2] != canary {
		t.Fatalf("out range/canaries: %+v", buffer)
	}
	if got := objc.Send[rect](obj, objc.RegisterName("echoRect:"), wantRect); got != wantRect {
		t.Fatalf("nested rect argument/return: got %+v", got)
	}
}

// A block header is native-owned after NewBlock/Copy. Read its invoke pointer
// and use a typed C-ABI call: objc.InvokeBlock would call Go directly instead.
type blockHeader struct {
	_               structs.HostLayout
	Isa             uintptr
	Flags, Reserved uint32
	Invoke          uintptr
}

func TestSuperAndCopiedAggregateBlock(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	registerProbe()
	if registerErr != nil {
		t.Fatal(registerErr)
	}
	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
	defer pool.Send(objc.RegisterName("drain"))
	obj := objc.ID(childClass).Send(objc.RegisterName("new"))
	defer obj.Send(objc.RegisterName("release"))
	want := rect{Origin: point{X: 13.25, Y: -17.5}, Size: point{X: 101.5, Y: 83.75}}
	if got := objc.Send[rect](obj, objc.RegisterName("echoRect:"), want); got != want {
		t.Fatalf("super rect arguments/return: got %+v, want %+v", got, want)
	}
	block := objc.NewBlock(func(_ objc.Block, token uintptr, p point, r nsrange) rect {
		return rect{Origin: p, Size: point{X: float64(token ^ r.Location), Y: float64(r.Length)}}
	})
	copied := block.Copy()
	block.Release()
	defer copied.Release()
	// Read the native header during a bounded call, without converting an
	// opaque handle to a Go pointer. memcpy retains neither destination nor data.
	var header blockHeader
	var memcpy func(*blockHeader, objc.Block, uintptr)
	purego.RegisterLibFunc(&memcpy, purego.RTLD_DEFAULT, "memcpy")
	memcpy(&header, copied, unsafe.Sizeof(header))
	var invoke func(objc.Block, uintptr, point, nsrange) rect
	purego.RegisterFunc(&invoke, header.Invoke)
	for i := 0; i < 3; i++ {
		got := invoke(copied, 71, want.Origin, nsrange{Location: 19, Length: 83})
		expected := rect{Origin: want.Origin, Size: point{X: 71 ^ 19, Y: 83}}
		if got != expected {
			t.Fatalf("copied block native aggregate call: got %+v, want %+v", got, expected)
		}
	}
}

func TestScalarFloatCallbackResultRejected(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("purego scalar floating callback result unexpectedly supported; recheck adapter requirements")
		}
	}()
	purego.NewCallback(func() float64 { return 1.25 })
}
