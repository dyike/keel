//go:build darwin && !ios

package window

import (
	"sync"
	"sync/atomic"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/dyike/keel/ui/theme"
	"github.com/ebitengine/purego/objc"
)

var motionOnce sync.Once
var motionUpdates = make(chan bool, 1)
var scrollersAutoHide atomic.Bool

func watchSystemPreferences() {
	motionOnce.Do(func() {
		go func() {
			for reduce := range motionUpdates {
				core.Update(func() { theme.SetSystemReducedMotion(reduce) })
			}
		}()
		appkit.MainAsync(watchMotion)
		appkit.MainAsync(watchScroll)
		appkit.MainAsync(watchScrollers)
	})
}

// Observers live for the process; AppKit keeps the blocks it copied.
var preferenceObservers []appkit.ID

func observe(center appkit.ID, name string, object appkit.ID, fn func()) {
	block := appkit.NewBlock(func(_ objc.Block, _ appkit.ID) { fn() })
	defer block.Release()
	queue := appkit.Send(appkit.Class("NSOperationQueue"), "mainQueue")
	o := appkit.Send(center, "addObserverForName:object:queue:usingBlock:", uintptr(appkit.Constant(name)), uintptr(object), uintptr(queue), uintptr(block))
	preferenceObservers = append(preferenceObservers, appkit.Retain(o))
}

// Main thread.
func watchMotion() {
	workspace := appkit.Send(appkit.Class("NSWorkspace"), "sharedWorkspace")
	changed := func() { motionChanged(appkit.SendBool(workspace, "accessibilityDisplayShouldReduceMotion")) }
	observe(appkit.Send(workspace, "notificationCenter"), "NSWorkspaceAccessibilityDisplayOptionsDidChangeNotification", workspace, changed)
	changed()
}

// Main thread. Overlay scrollers ("Show scroll bars: automatically / when
// scrolling") hide at rest; legacy ones ("always") stay.
func watchScrollers() {
	changed := func() { scrollersChanged(overlayScrollers()) }
	observe(appkit.Send(appkit.Class("NSNotificationCenter"), "defaultCenter"), "NSPreferredScrollerStyleDidChangeNotification", 0, changed)
	changed()
}

const (
	nsEventMaskScrollWheel = 1 << 22
	phaseBegan             = 1
	phaseStationary        = 2
	phaseChanged           = 4
	phaseEnded             = 8
	phaseCancelled         = 16
	phaseMayBegin          = 32
)

// watchScroll observes scroll events before Gio's view receives them,
// reporting the device and gesture phase that Gio drops. Main thread.
func watchScroll() {
	handler := appkit.NewBlock(func(_ objc.Block, event appkit.ID) appkit.ID {
		phase, momentum := uintptr(appkit.Send(event, "phase")), uintptr(appkit.Send(event, "momentumPhase"))
		scrollEvent(appkit.SendBool(event, "hasPreciseScrollingDeltas"),
			phase&(phaseBegan|phaseStationary|phaseChanged|phaseMayBegin) != 0,
			momentum&(phaseBegan|phaseChanged) != 0,
			phase&(phaseEnded|phaseCancelled) != 0)
		return event
	})
	defer handler.Release()
	monitor := appkit.Send(appkit.Class("NSEvent"), "addLocalMonitorForEventsMatchingMask:handler:", nsEventMaskScrollWheel, uintptr(handler))
	preferenceObservers = append(preferenceObservers, appkit.Retain(monitor))
}

func motionChanged(reduce bool) {
	// AppKit callbacks must not block on the UI frame lock. Preserve the newest
	// value while a frame is busy, with one consumer applying updates in order.
	select {
	case motionUpdates <- reduce:
	default:
		select {
		case <-motionUpdates:
		default:
		}
		motionUpdates <- reduce
	}
}

func scrollEvent(precise, active, momentum, ended bool) {
	device := core.ScrollDeviceWheel
	if precise {
		device = core.ScrollDeviceTrackpad
	}
	core.ReportScrollGesture(device, active, momentum, ended)
}

func scrollersChanged(overlay bool) {
	// Gio's Invalidate can synchronously re-enter its event loop on AppKit's
	// main thread while holding invMu. Post from another goroutine instead,
	// and read the latest preference when applying possibly queued updates.
	scrollersAutoHide.Store(overlay)
	go core.Update(func() { theme.SetSystemScrollbarsAutoHide(scrollersAutoHide.Load()) })
}

const nsScrollerStyleOverlay = 1

func overlayScrollers() bool {
	return uintptr(appkit.Send(appkit.Class("NSScroller"), "preferredScrollerStyle")) == nsScrollerStyleOverlay
}
