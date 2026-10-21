//go:build darwin && !ios

package window

import "github.com/dyike/keel/ui/internal/appkit"

const (
	nsViewLayerContentsRedrawDuringViewResize = 2
	nsViewLayerContentsPlacementTopLeft       = 11
)

// Let AppKit resize and redraw the view at each step, including animated
// resizing. If a transaction reaches the compositor before the next Metal
// frame, preserve the old pixels' size and top-left placement. Keep NSView's
// policy and CALayer's gravity consistent; changing the layer alone leaves
// AppKit's content placement set to scaleAxesIndependently.
// Retain both event handles across the async dispatch: neither the window
// goroutine nor the frame lock may wait for AppKit here.
func configureNativeResize(viewHandle, layerHandle uintptr) {
	view := appkit.Retain(appkit.ID(viewHandle))
	layer := appkit.Retain(appkit.ID(layerHandle))
	appkit.MainAsync(func() {
		defer appkit.Release(view)
		defer appkit.Release(layer)
		appkit.Send(view, "setLayerContentsRedrawPolicy:", nsViewLayerContentsRedrawDuringViewResize)
		appkit.Send(view, "setLayerContentsPlacement:", nsViewLayerContentsPlacementTopLeft)
		gravity := "kCAGravityTopLeft"
		// AppKit may flip layer contents independently of view coordinates.
		// Apple's gravity constants follow the contents' vertical axis.
		if appkit.SendBool(layer, "contentsAreFlipped") {
			gravity = "kCAGravityBottomLeft"
		}
		appkit.Send(layer, "setContentsGravity:", uintptr(appkit.Constant(gravity)))
	})
}
