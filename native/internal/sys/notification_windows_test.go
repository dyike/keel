//go:build windows

package sys

import (
	"testing"
	"unsafe"
)

// NOTIFYICONDATAW must match the Windows SDK layout, including hBalloonIcon.
func TestNotifyIconDataSize(t *testing.T) {
	want := uintptr(976)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		want = 956
	}
	if got := unsafe.Sizeof(notifyIconData{}); got != want {
		t.Fatalf("NOTIFYICONDATAW is %d bytes, want %d", got, want)
	}
}
