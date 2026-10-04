//go:build darwin && !ios && cgo

package sys

/*
#include <stdint.h>
void keel_clipboard_read(uintptr_t token);
*/
import "C"
import "runtime/cgo"

func ClipboardRead(done func([]byte, error)) {
	C.keel_clipboard_read(C.uintptr_t(cgo.NewHandle(done)))
}

//export keel_clipboard_read_done
func keel_clipboard_read_done(token C.uintptr_t, json *C.char, code C.int) {
	h := cgo.Handle(token)
	done := h.Value().(func([]byte, error))
	h.Delete()
	data := []byte(C.GoString(json))
	err := status(code)
	go done(data, err)
}
