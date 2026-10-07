//go:build darwin && !ios && cgo && keelnativeqa

package window

/*
#include <stdlib.h>
void keel_qa_menu(const char *command);
void keel_qa_ime(int commit);
int keel_qa_icon(const char *path, const char *source, const char *reference);
*/
import "C"
import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"unsafe"
)

// NativeQAActivateMenu exercises an installed native menu action in development tests.
func NativeQAActivateMenu(id string) {
	value := C.CString(id)
	defer C.free(unsafe.Pointer(value))
	C.keel_qa_menu(value)
}

// NativeQACompose exercises the native input method bridge in development tests.
func NativeQACompose(commit bool) {
	value := 0
	if commit {
		value = 1
	}
	C.keel_qa_ime(C.int(value))
}

// NativeQACaptureIcon saves the actual and configured AppKit icons for comparison.
// Call from a background goroutine, outside core.Update.
func NativeQACaptureIcon(path, source, reference string) error {
	p, s, r := C.CString(path), C.CString(source), C.CString(reference)
	defer C.free(unsafe.Pointer(p))
	defer C.free(unsafe.Pointer(s))
	defer C.free(unsafe.Pointer(r))
	if C.keel_qa_icon(p, s, r) == 0 {
		return fmt.Errorf("window: native icon capture failed")
	}
	return nil
}

//export keel_native_qa_wake
func keel_native_qa_wake() { go core.Update(func() {}) }
