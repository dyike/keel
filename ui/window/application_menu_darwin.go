//go:build darwin && !ios && cgo

package window

/*
#cgo LDFLAGS: -framework AppKit
#include <stdint.h>
#include <stddef.h>
#include <stdlib.h>
void keel_install_application_menu(const void *json, size_t len);
void keel_application_menu_edit(const char *action);
*/
import "C"
import "unsafe"

func platformNativeApplicationMenu() bool { return true }
func platformInstallApplicationMenu(data []byte) {
	C.keel_install_application_menu(unsafe.Pointer(&data[0]), C.size_t(len(data)))
}
func platformMenuEdit(action MenuAction) bool {
	text := C.CString(string(action))
	defer C.free(unsafe.Pointer(text))
	C.keel_application_menu_edit(text)
	return true
}

//export keel_application_menu_command
func keel_application_menu_command(generation C.uint64_t, id *C.char) {
	dispatchNativeMenu(uint64(generation), C.GoString(id))
}

func platformDrawApplicationMenu() bool       { return false }
func applicationMenuWindowEvent(*Window, any) {}

func platformNativeMenuShortcuts() bool { return true }
