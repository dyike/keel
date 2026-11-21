//go:build ios

package window

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

const (
	uiImpactFeedbackStyleLight  = 0
	uiImpactFeedbackStyleMedium = 1
	uiImpactFeedbackStyleHeavy  = 2
)

var (
	iosDispatchOnce sync.Once
	iosDispatchFn   uintptr
	iosMainQueue    uintptr
	iosCallback     uintptr
	iosPendingMu    sync.Mutex
	iosNextToken    uintptr
	iosPending      = map[uintptr]func(){}
)

func iosMainAsync(fn func()) {
	iosDispatchOnce.Do(func() {
		iosDispatchFn, _ = purego.Dlsym(purego.RTLD_DEFAULT, "dispatch_async_f")
		iosMainQueue, _ = purego.Dlsym(purego.RTLD_DEFAULT, "_dispatch_main_q")
		iosCallback = purego.NewCallback(func(token uintptr) {
			iosPendingMu.Lock()
			cb := iosPending[token]
			delete(iosPending, token)
			iosPendingMu.Unlock()
			if cb != nil {
				cb()
			}
		})
	})
	if iosDispatchFn == 0 || iosMainQueue == 0 {
		return
	}
	iosPendingMu.Lock()
	iosNextToken++
	token := iosNextToken
	iosPending[token] = fn
	iosPendingMu.Unlock()
	purego.SyscallN(iosDispatchFn, iosMainQueue, token, iosCallback)
}

func platformHaptic(style HapticStyle) {
	iosMainAsync(func() {
		pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(sel("new"))
		defer pool.Send(sel("drain"))
		if style == HapticSelection {
			cls := objc.GetClass("UISelectionFeedbackGenerator")
			if cls == 0 {
				return
			}
			gen := objc.ID(cls).Send(sel("alloc")).Send(sel("init"))
			if gen != 0 {
				gen.Send(sel("selectionChanged"))
				gen.Send(sel("release"))
			}
			return
		}
		cls := objc.GetClass("UIImpactFeedbackGenerator")
		if cls == 0 {
			return
		}
		impact := uintptr(uiImpactFeedbackStyleMedium)
		switch style {
		case HapticLight:
			impact = uiImpactFeedbackStyleLight
		case HapticHeavy:
			impact = uiImpactFeedbackStyleHeavy
		}
		gen := objc.ID(cls).Send(sel("alloc")).Send(sel("initWithStyle:"), impact)
		if gen != 0 {
			gen.Send(sel("impactOccurred"))
			gen.Send(sel("release"))
		}
	})
}
