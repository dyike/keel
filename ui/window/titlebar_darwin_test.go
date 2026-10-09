//go:build darwin && !ios && cgo

package window

import (
	"testing"
	"time"

	"github.com/dyike/keel/ui/internal/loop"
)

func TestNativeTitleBarCallbackDoesNotWaitForInvalidation(t *testing.T) {
	const view = uintptr(0x1234)
	w := &Window{}
	nativeWindows.Store(view, w)
	defer nativeWindows.Delete(view)
	key := new(int)
	entered, release := make(chan struct{}, 1), make(chan struct{})
	loop.Register(key, func() { entered <- struct{}{}; <-release })
	defer loop.Unregister(key)
	returned := make(chan struct{})
	go func() { postTitleBarAction(view, 2); close(returned) }()
	select {
	case <-returned:
	case <-time.After(time.Second):
		close(release)
		<-returned
		t.Fatal("native callback waited for Gio invalidation")
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("double-click did not request a frame")
	}
	close(release)
	loop.Lock()
	loop.Drain()
	if !w.Maximized() {
		loop.Unlock()
		t.Fatal("queued double-click did not maximize")
	}
	loop.Unlock()
}
