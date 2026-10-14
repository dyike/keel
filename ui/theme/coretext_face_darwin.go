//go:build darwin && !ios

package theme

import (
	"os"
	"sync"
	"unsafe"

	"github.com/dyike/keel/typesetting/font/opentype"
	"github.com/dyike/keel/typesetting/font/opentype/tables"
	"github.com/ebitengine/purego"
)

// CoreText expands variable faces into named instances. Its descriptor-array
// position is therefore NOT the sfnt/TTC face index used by the shaper.
var ctIdentity struct {
	sync.Once
	ready       bool
	url         func(uintptr, []byte, int64, bool) uintptr
	descriptors func(uintptr) uintptr
	count       func(uintptr) int64
	item        func(uintptr, int64) uintptr
	attribute   func(uintptr, uintptr) uintptr
	stringNew   func(uintptr, string, uint32) uintptr
	equal       func(uintptr, uintptr) bool
	release     func(uintptr)
	create      func(uintptr, float64, uintptr) uintptr
	name        func(uintptr) uintptr
	nameKey     uintptr
}

func initCTIdentity() {
	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return
	}
	ct, err := purego.Dlopen("/System/Library/Frameworks/CoreText.framework/CoreText", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return
	}
	purego.RegisterLibFunc(&ctIdentity.url, cf, "CFURLCreateFromFileSystemRepresentation")
	purego.RegisterLibFunc(&ctIdentity.count, cf, "CFArrayGetCount")
	purego.RegisterLibFunc(&ctIdentity.item, cf, "CFArrayGetValueAtIndex")
	purego.RegisterLibFunc(&ctIdentity.stringNew, cf, "CFStringCreateWithCString")
	purego.RegisterLibFunc(&ctIdentity.equal, cf, "CFEqual")
	purego.RegisterLibFunc(&ctIdentity.release, cf, "CFRelease")
	purego.RegisterLibFunc(&ctIdentity.descriptors, ct, "CTFontManagerCreateFontDescriptorsFromURL")
	purego.RegisterLibFunc(&ctIdentity.attribute, ct, "CTFontDescriptorCopyAttribute")
	purego.RegisterLibFunc(&ctIdentity.create, ct, "CTFontCreateWithFontDescriptor")
	purego.RegisterLibFunc(&ctIdentity.name, ct, "CTFontCopyPostScriptName")
	key, err := purego.Dlsym(ct, "kCTFontNameAttribute")
	if err != nil {
		return
	}
	var copyPointer func(*uintptr, uintptr, uintptr) uintptr
	purego.RegisterLibFunc(&copyPointer, cf, "memcpy")
	copyPointer(&ctIdentity.nameKey, key, unsafe.Sizeof(ctIdentity.nameKey))
	ctIdentity.ready = true
}

func fontPostScriptName(path string, index int, instance ...int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	loaders, err := opentype.NewLoaders(f)
	if err != nil || index < 0 || index >= len(loaders) {
		return ""
	}
	raw, err := loaders[index].RawTable(opentype.MustNewTag("name"))
	if err != nil {
		return ""
	}
	names, _, err := tables.ParseName(raw)
	if err != nil {
		return ""
	}
	if len(instance) > 0 && instance[0] > 0 {
		raw, err := loaders[index].RawTable(opentype.MustNewTag("fvar"))
		if err != nil {
			return ""
		}
		fv, _, err := tables.ParseFvar(raw)
		if err != nil || instance[0] > len(fv.Instances) {
			return ""
		}
		id := fv.Instances[instance[0]-1].PostScriptNameID
		if id == 0 || id == 0xffff {
			return ""
		}
		return names.Name(tables.NameID(id))
	}
	return names.Name(6)
}

// Caller owns the returned font. No matching identity means outline fallback,
// never substitution with an unrelated descriptor or another weight.
func openCTFontFace(path string, index int, size float64, instance ...int) uintptr {
	ctIdentity.Do(initCTIdentity)
	if !ctIdentity.ready {
		return 0
	}
	name := fontPostScriptName(path, index, instance...)
	if name == "" {
		return 0
	}
	wanted := ctIdentity.stringNew(0, name, 0x08000100)
	if wanted == 0 {
		return 0
	}
	defer ctIdentity.release(wanted)
	bytes := []byte(path)
	url := ctIdentity.url(0, bytes, int64(len(bytes)), false)
	if url == 0 {
		return 0
	}
	defer ctIdentity.release(url)
	descriptors := ctIdentity.descriptors(url)
	if descriptors == 0 {
		return 0
	}
	defer ctIdentity.release(descriptors)
	for i, n := int64(0), ctIdentity.count(descriptors); i < n; i++ {
		desc := ctIdentity.item(descriptors, i)
		candidate := ctIdentity.attribute(desc, ctIdentity.nameKey)
		if candidate == 0 {
			continue
		}
		match := ctIdentity.equal(candidate, wanted)
		ctIdentity.release(candidate)
		if !match {
			continue
		}
		f := ctIdentity.create(desc, size, 0)
		if f == 0 {
			return 0
		}
		actual := ctIdentity.name(f)
		valid := actual != 0 && ctIdentity.equal(actual, wanted)
		if actual != 0 {
			ctIdentity.release(actual)
		}
		if valid {
			return f
		}
		ctIdentity.release(f)
		return 0
	}
	return 0
}
