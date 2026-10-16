// SPDX-License-Identifier: Unlicense OR MIT
//go:build darwin && !ios

package text

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Libraries are opened lazily while installing the generated ABI table. Keep
// them open on success: native objects and function addresses outlive this loader.
type coreTextLoader struct {
	libraries map[string]uintptr
	err       error
}

func newCoreTextLoader() *coreTextLoader {
	return &coreTextLoader{libraries: make(map[string]uintptr)}
}

func (l *coreTextLoader) symbol(library, name string) uintptr {
	if l.err != nil {
		return 0
	}
	handle := l.libraries[library]
	if handle == 0 {
		var path string
		switch library {
		case "cf":
			path = "/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation"
		case "text":
			path = "/System/Library/Frameworks/CoreText.framework/CoreText"
		case "graphics":
			path = "/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics"
		case "appkit":
			path = "/System/Library/Frameworks/AppKit.framework/AppKit"
		case "objc":
			path = "/usr/lib/libobjc.A.dylib"
		default:
			l.err = fmt.Errorf("text: unknown native font framework %q", library)
			return 0
		}
		handle, l.err = purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
		if l.err != nil {
			l.err = fmt.Errorf("text: load %s: %w", library, l.err)
			return 0
		}
		l.libraries[library] = handle
	}
	address, err := purego.Dlsym(handle, name)
	if err != nil || address == 0 {
		l.err = fmt.Errorf("text: resolve %s/%s: %v", library, name, err)
		return 0
	}
	return address
}

func (l *coreTextLoader) function(library, name string, target any) {
	if address := l.symbol(library, name); address != 0 {
		purego.RegisterFunc(target, address)
	}
}

func (l *coreTextLoader) address(library, name string) uintptr {
	return l.symbol(library, name)
}

func (l *coreTextLoader) indirect(library, name string) uintptr {
	address := l.symbol(library, name)
	if address == 0 {
		return 0
	}
	return **(**uintptr)(unsafe.Pointer(&address))
}

func (l *coreTextLoader) scalar(library, name string) float64 {
	address := l.symbol(library, name)
	if address == 0 {
		return 0
	}
	return **(**float64)(unsafe.Pointer(&address))
}

func (l *coreTextLoader) closeOnError() {
	if l.err != nil {
		for _, handle := range l.libraries {
			_ = purego.Dlclose(handle)
		}
	}
}
