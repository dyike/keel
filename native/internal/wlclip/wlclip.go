//go:build linux && !android && cgo && !nowayland

// Package wlclip reads the Wayland clipboard through an application's own
// wl_display. Wayland offers the selection only to the client with keyboard
// focus, so a separate connection would see nothing; the caller passes the
// display of its focused window. Build with -tags nowayland to leave out
// libwayland-client.
package wlclip

/*
#cgo pkg-config: wayland-client
#include <stdlib.h>
#include "wlclip.h"
*/
import "C"

import (
	"errors"
	"io"
	"os"
	"time"
	"unsafe"

	"github.com/dyike/keel/native"
)

// Selection is the clipboard's offer at the time of Open.
type Selection struct{ c *C.struct_keel_wlclip }

// Open binds the seat's data device on display and waits for the current
// selection. Close it when done.
func Open(display unsafe.Pointer) (*Selection, error) {
	if display == nil {
		return nil, native.ErrUnsupported
	}
	var code C.int
	c := C.keel_wlclip_open(display, &code)
	switch {
	case c != nil:
		return &Selection{c}, nil
	case code == C.KEEL_WLCLIP_UNSUPPORTED:
		return nil, native.ErrUnsupported
	}
	return nil, native.ErrFailed
}

// MIMEs lists the offered types; nil when the clipboard is empty.
func (s *Selection) MIMEs() []string {
	n := int(C.keel_wlclip_count(s.c))
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, C.GoString(C.keel_wlclip_mime(s.c, C.int(i))))
	}
	return out
}

// Read receives one type, at most limit bytes, by the deadline. A larger
// payload is an error rather than a truncated value.
func (s *Selection) Read(mime string, limit int, deadline time.Time) ([]byte, error) {
	cs := C.CString(mime)
	defer C.free(unsafe.Pointer(cs))
	fd := C.keel_wlclip_receive(s.c, cs)
	if fd < 0 {
		return nil, native.ErrFailed
	}
	f := os.NewFile(uintptr(fd), "wayland-selection")
	defer f.Close()
	f.SetReadDeadline(deadline)
	data, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return nil, native.ErrTimeout
	}
	if err != nil {
		return nil, native.ErrFailed
	}
	if len(data) > limit {
		return nil, errors.New("wayland selection exceeds the limit")
	}
	return data, nil
}

// Close releases the data device and offers.
func (s *Selection) Close() { C.keel_wlclip_close(s.c) }
