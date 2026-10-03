//go:build darwin && !ios && cgo

package sys

/*
#cgo LDFLAGS: -framework UserNotifications
#include <stdint.h>
#include <stdlib.h>
int keel_notification_available(void);
void keel_notification_permission(uintptr_t token);
void keel_notification_post(const char *id, const char *title, const char *body, uintptr_t token);
void keel_notification_remove(const char *id, uintptr_t token);
*/
import "C"
import (
	"runtime/cgo"
	"unsafe"
)

func NotificationAvailable() bool { return C.keel_notification_available() != 0 }
func NotificationPermission(done func(error)) {
	C.keel_notification_permission(C.uintptr_t(cgo.NewHandle(done)))
}
func NotificationPost(id, title, body string, done func(error)) {
	a, b, c := C.CString(id), C.CString(title), C.CString(body)
	defer C.free(unsafe.Pointer(a))
	defer C.free(unsafe.Pointer(b))
	defer C.free(unsafe.Pointer(c))
	C.keel_notification_post(a, b, c, C.uintptr_t(cgo.NewHandle(done)))
}
func NotificationRemove(id string, done func(error)) {
	a := C.CString(id)
	defer C.free(unsafe.Pointer(a))
	C.keel_notification_remove(a, C.uintptr_t(cgo.NewHandle(done)))
}

//export keel_notification_done
func keel_notification_done(token C.uintptr_t, code C.int) {
	h := cgo.Handle(token)
	done := h.Value().(func(error))
	h.Delete()
	done(status(code))
}
