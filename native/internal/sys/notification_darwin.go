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
	NotificationPostInteractive(id, title, body, nil, done)
}
func NotificationPostInteractive(id, title, body string, onClick func(), done func(error)) {
	finish := systemNotificationClicks.register(id, onClick)
	completed := func(err error) { finish(err); done(err) }
	a, b, c := C.CString(id), C.CString(title), C.CString(body)
	defer C.free(unsafe.Pointer(a))
	defer C.free(unsafe.Pointer(b))
	defer C.free(unsafe.Pointer(c))
	C.keel_notification_post(a, b, c, C.uintptr_t(cgo.NewHandle(completed)))
}
func NotificationRemove(id string, done func(error)) {
	a := C.CString(id)
	defer C.free(unsafe.Pointer(a))
	C.keel_notification_remove(a, C.uintptr_t(cgo.NewHandle(func(err error) {
		if err == nil {
			systemNotificationClicks.take(id)
		}
		done(err)
	})))
}

//export keel_notification_done
func keel_notification_done(token C.uintptr_t, code C.int) {
	h := cgo.Handle(token)
	done := h.Value().(func(error))
	h.Delete()
	done(status(code))
}

//export keel_notification_clicked
func keel_notification_clicked(id *C.char) {
	if fn := systemNotificationClicks.take(C.GoString(id)); fn != nil {
		go fn()
	}
}
