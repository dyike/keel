//go:build darwin && !ios && keelnativeqa

package window

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// This development probe saves only the requested application's window area.
// It includes this process's windows and system desktop layers in the filter:
// removing the desktop layers makes ScreenCaptureKit flatten behind-window
// materials. Other applications' content windows are excluded. Current-process
// shareable content does not request a screen-recording grant.
var materialCapture struct {
	once           sync.Once
	mu             sync.Mutex
	path, title    string
	done           chan error
	content, image objc.Block
}

// NativeQACaptureMaterial saves the requested window's final composition with
// other windows from this process behind it, including system desktop layers. Call once per QA process, off the
// UI thread and frame lock. It does not request screen-recording permission.
func NativeQACaptureMaterial(title, path string) error {
	q := &materialCapture
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.done != nil {
		return fmt.Errorf("material capture already used")
	}
	q.path, q.title, q.done = path, title, make(chan error, 1)
	q.once.Do(func() {
		_, _ = purego.Dlopen("/System/Library/Frameworks/ScreenCaptureKit.framework/ScreenCaptureKit", purego.RTLD_GLOBAL|purego.RTLD_NOW)
		q.image = appkit.NewBlock(func(_ objc.Block, image, nativeError appkit.ID) {
			defer q.image.Release()
			if image == 0 || nativeError != 0 {
				q.done <- fmt.Errorf("material compositor capture failed: %v", nativeError)
				return
			}
			appkit.Pool(func() {
				rep := appkit.Send(appkit.Send(appkit.Class("NSBitmapImageRep"), "alloc"), "initWithCGImage:", uintptr(image))
				defer appkit.Release(rep)
				png := appkit.Send(rep, "representationUsingType:properties:", 4, uintptr(appkit.Send(appkit.Class("NSDictionary"), "dictionary")))
				if !appkit.SendBool(png, "writeToFile:atomically:", uintptr(appkit.String(q.path)), 1) {
					q.done <- fmt.Errorf("material capture write failed")
					return
				}
				q.done <- nil
			})
		})
		q.content = appkit.NewBlock(func(_ objc.Block, content, nativeError appkit.ID) {
			defer q.content.Release()
			requested := false
			defer func() {
				if !requested {
					q.image.Release()
				}
			}()
			if content == 0 || nativeError != 0 {
				q.done <- fmt.Errorf("current-process capture unavailable: %v", nativeError)
				return
			}
			appkit.Pool(func() {
				available := appkit.Send(content, "windows")
				windows := appkit.Send(appkit.Class("NSMutableArray"), "array")
				for i := uintptr(0); i < uintptr(appkit.Send(available, "count")); i++ {
					w := appkit.Send(available, "objectAtIndex:", i)
					owner := appkit.Send(w, "owningApplication")
					if int(appkit.Send(owner, "processID")) == os.Getpid() || int64(appkit.Send(w, "windowLayer")) < 0 {
						appkit.Send(windows, "addObject:", uintptr(w))
					}
				}
				var target appkit.ID
				for i := uintptr(0); i < uintptr(appkit.Send(windows, "count")); i++ {
					w := appkit.Send(windows, "objectAtIndex:", i)
					if appkit.Equal(appkit.Send(w, "title"), q.title) && appkit.SendBool(w, "isOnScreen") {
						target = w
						break
					}
				}
				if target == 0 {
					q.done <- fmt.Errorf("material window %q unavailable", q.title)
					return
				}
				rect := appkit.MsgRect(target, appkit.Sel("frame"))
				displays := appkit.Send(content, "displays")
				var display appkit.ID
				for i := uintptr(0); i < uintptr(appkit.Send(displays, "count")); i++ {
					d := appkit.Send(displays, "objectAtIndex:", i)
					frame := appkit.MsgRect(d, appkit.Sel("frame"))
					if frame.Contains(rect.Origin) {
						display = d
						rect.Origin.X -= frame.Origin.X
						rect.Origin.Y -= frame.Origin.Y
						break
					}
				}
				if display == 0 {
					q.done <- fmt.Errorf("material display unavailable")
					return
				}
				filter := appkit.Send(appkit.Send(appkit.Class("SCContentFilter"), "alloc"), "initWithDisplay:includingWindows:", uintptr(display), uintptr(windows))
				defer appkit.Release(filter)
				config := appkit.Send(appkit.Class("SCStreamConfiguration"), "new")
				defer appkit.Release(config)
				appkit.MsgSetRect(config, appkit.Sel("setSourceRect:"), rect)
				appkit.Send(config, "setWidth:", uintptr(rect.Size.Width*2))
				appkit.Send(config, "setHeight:", uintptr(rect.Size.Height*2))
				appkit.Send(config, "setShowsCursor:", 0)
				requested = true
				appkit.Send(appkit.Class("SCScreenshotManager"), "captureImageWithFilter:configuration:completionHandler:", uintptr(filter), uintptr(config), uintptr(q.image))
			})
		})
	})
	if !appkit.SendBool(appkit.Class("SCShareableContent"), "respondsToSelector:", uintptr(appkit.Sel("getCurrentProcessShareableContentWithCompletionHandler:"))) {
		q.content.Release()
		q.image.Release()
		return fmt.Errorf("current-process capture requires macOS 14.4")
	}
	appkit.Send(appkit.Class("SCShareableContent"), "getCurrentProcessShareableContentWithCompletionHandler:", uintptr(q.content))
	select {
	case err := <-q.done:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("material capture timeout")
	}
}
