//go:build windows

package uia

import (
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// COM objects without cgo. An object is a block of native heap memory
// holding one vtable pointer per interface it implements; a pointer to the
// i-th slot is the object as interface i. Vtables also live on the native
// heap, and their entries are callbacks created once per interface method
// and shared by every object (Windows callbacks are scarce and never
// freed). Go state is found from the block's address, so no Go pointer
// reaches native code.

const (
	ifSimple = iota
	ifFragment
	ifRoot
	ifInvoke
	ifToggle
	ifScrollItem
	ifValue
	numIfaces
)

const ptrSize = unsafe.Sizeof(uintptr(0))

var (
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	procGetHeap      = kernel32.NewProc("GetProcessHeap")
	procHeapAlloc    = kernel32.NewProc("HeapAlloc")
	procHeapFree     = kernel32.NewProc("HeapFree")
	oleaut32         = windows.NewLazySystemDLL("oleaut32.dll")
	procSysAlloc     = oleaut32.NewProc("SysAllocStringLen")
	procSafeArray    = oleaut32.NewProc("SafeArrayCreateVector")
	procSafeArrayPut = oleaut32.NewProc("SafeArrayPutElement")
)

func heapAlloc(size uintptr) uintptr {
	heap, _, _ := procGetHeap.Call()
	p, _, _ := procHeapAlloc.Call(heap, 8 /* HEAP_ZERO_MEMORY */, size)
	if p == 0 {
		panic("uia: out of memory")
	}
	return p
}

func heapFree(p uintptr) {
	heap, _, _ := procGetHeap.Call()
	procHeapFree.Call(heap, 0, p)
}

// object is the Go side of a COM object.
type object struct {
	base uintptr
	refs int32
	win  *Window
	key  string // the element; "" for the window's root
}

// mu guards every object and window.
var mu sync.Mutex

var objects = map[uintptr]*object{}

var vtables struct {
	once sync.Once
	ptr  [numIfaces]uintptr
}

// newObject allocates an object with no references; the caller holds mu.
func newObject(w *Window, key string) *object {
	vtables.once.Do(buildVtables)
	base := heapAlloc(numIfaces * ptrSize)
	for i := range numIfaces {
		*(*uintptr)(unsafe.Pointer(base + uintptr(i)*ptrSize)) = vtables.ptr[i]
	}
	o := &object{base: base, win: w, key: key}
	objects[base] = o
	return o
}

// as returns o as interface i with a new reference; the caller holds mu.
func (o *object) as(i int) uintptr {
	o.refs++
	return o.base + uintptr(i)*ptrSize
}

// release drops a reference; the caller holds mu.
func (o *object) release() int32 {
	o.refs--
	if o.refs == 0 {
		delete(objects, o.base)
		if o.win != nil && o.win.elements[o.key] == o {
			delete(o.win.elements, o.key)
		}
		heapFree(o.base)
		o.base = 0
	}
	return o.refs
}

// lookup finds the object behind interface pointer this of interface i;
// the caller holds mu.
func lookup(this uintptr, i int) *object {
	return objects[this-uintptr(i)*ptrSize]
}

func guidAt(p uintptr) windows.GUID { return *(*windows.GUID)(unsafe.Pointer(p)) }

func put(p, v uintptr) {
	if p != 0 {
		*(*uintptr)(unsafe.Pointer(p)) = v
	}
}

// interfaceIIDs lists what each interface slot answers to QueryInterface.
var interfaceIIDs = [numIfaces]windows.GUID{
	ifSimple: iidSimple, ifFragment: iidFragment, ifRoot: iidFragmentRoot,
	ifInvoke: iidInvokeProvider, ifToggle: iidToggleProvider,
	ifScrollItem: iidScrollItemProvider, ifValue: iidValueProvider,
}

func buildVtables() {
	methods := [numIfaces][]any{
		ifSimple:     {simpleProviderOptions, simplePatternProvider, simplePropertyValue, simpleHostProvider},
		ifFragment:   {fragmentNavigate, fragmentRuntimeID, fragmentBoundingRectangle, fragmentEmbeddedRoots, fragmentSetFocus, fragmentRoot},
		ifRoot:       {rootElementFromPoint, rootFocus},
		ifInvoke:     {invokeInvoke},
		ifToggle:     {toggleToggle, toggleState},
		ifScrollItem: {scrollItemScrollIntoView},
		ifValue:      {valueSetValue, valueValue, valueIsReadOnly},
	}
	for i := range numIfaces {
		entries := []uintptr{
			windows.NewCallback(func(this, riid, ppv uintptr) uintptr { return queryInterface(this, i, riid, ppv) }),
			windows.NewCallback(func(this uintptr) uintptr { return addRef(this, i) }),
			windows.NewCallback(func(this uintptr) uintptr { return release(this, i) }),
		}
		for _, m := range methods[i] {
			entries = append(entries, windows.NewCallback(bind(m, i)))
		}
		table := heapAlloc(uintptr(len(entries)) * ptrSize)
		for j, e := range entries {
			*(*uintptr)(unsafe.Pointer(table + uintptr(j)*ptrSize)) = e
		}
		vtables.ptr[i] = table
	}
}

// bind adapts a method, written against its object, to a callback taking
// the interface pointer.
func bind(m any, i int) any {
	switch m := m.(type) {
	case func(*object) uintptr:
		return func(this uintptr) uintptr { return call(this, i, func(o *object) uintptr { return m(o) }) }
	case func(*object, uintptr) uintptr:
		return func(this, a uintptr) uintptr { return call(this, i, func(o *object) uintptr { return m(o, a) }) }
	case func(*object, uintptr, uintptr) uintptr:
		return func(this, a, b uintptr) uintptr { return call(this, i, func(o *object) uintptr { return m(o, a, b) }) }
	case func(*object, [5]uintptr) uintptr:
		// ElementProviderFromPoint(double x, double y, out): the doubles
		// travel in floating-point registers, which callbacks do not
		// receive; the 386 stack holds them as two words each.
		return func(this, a, b, c, d, e uintptr) uintptr {
			return call(this, i, func(o *object) uintptr { return m(o, [5]uintptr{a, b, c, d, e}) })
		}
	}
	panic("uia: method signature")
}

// call runs a method on the object behind this, holding mu. A fault fails
// the call instead of the application.
func call(this uintptr, i int, fn func(*object) uintptr) (hr uintptr) {
	mu.Lock()
	defer mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			hr = 0x80004005 // E_FAIL
		}
	}()
	o := lookup(this, i)
	if o == nil {
		return uiaElementNotAvailable
	}
	return fn(o)
}

func queryInterface(this uintptr, i int, riid, ppv uintptr) uintptr {
	return call(this, i, func(o *object) uintptr {
		put(ppv, 0)
		if riid == 0 {
			return eInvalidArg
		}
		iid := guidAt(riid)
		if iid == iidUnknown {
			put(ppv, o.as(ifSimple))
			return sOK
		}
		for j, known := range interfaceIIDs {
			if iid == known && o.implements(j) {
				put(ppv, o.as(j))
				return sOK
			}
		}
		return eNoInterface
	})
}

func addRef(this uintptr, i int) uintptr {
	return call(this, i, func(o *object) uintptr { o.refs++; return uintptr(o.refs) })
}

func release(this uintptr, i int) uintptr {
	return call(this, i, func(o *object) uintptr { return uintptr(o.release()) })
}

// bstr returns a BSTR copy of s, which the receiver frees.
func bstr(s string) uintptr {
	u := windows.StringToUTF16(s)
	p, _, _ := procSysAlloc.Call(uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1))
	runtime.KeepAlive(u)
	return p
}

// variant writes a VARIANT: type at offset 0, value at offset 8.
func variant(p uintptr, vt uint16, value uint64) {
	if p == 0 {
		return
	}
	*(*[2]uint64)(unsafe.Pointer(p)) = [2]uint64{uint64(vt), value}
}

func variantBool(p uintptr, b bool) {
	v := uint64(0)
	if b {
		v = 0xffff // VARIANT_TRUE
	}
	variant(p, vtBool, v)
}

// intArray returns a SAFEARRAY of int32, which the receiver frees.
func intArray(values ...int32) uintptr {
	psa, _, _ := procSafeArray.Call(vtI4, 0, uintptr(len(values)))
	for i, v := range values {
		index := int32(i)
		procSafeArrayPut.Call(psa, uintptr(unsafe.Pointer(&index)), uintptr(unsafe.Pointer(&v)))
	}
	return psa
}
