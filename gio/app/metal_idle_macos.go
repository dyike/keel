// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

import (
	"sync"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// metalIdleTrims enables the second idle stage: after a second without
// frames, frame data is released and the spare drawables shrink.
const metalIdleTrims = true

var metalTrimOnce sync.Once
var metalSetDrawableSize func(uintptr, uintptr, struct{ Width, Height float64 })
var metalDrawableSizeSelector uintptr

// Core Animation retains the displayed frame. Shrinking drawableSize lets
// Metal discard spare full-size surfaces; RenderTarget restores the size
// before requesting another drawable. No new frame is requested to trim.
func trimMetalLayer(layer uintptr) {
	metalTrimOnce.Do(func() {
		addr, err := purego.Dlsym(purego.RTLD_DEFAULT, "objc_msgSend")
		if err != nil {
			panic(err)
		}
		purego.RegisterFunc(&metalSetDrawableSize, addr)
		metalDrawableSizeSelector = uintptr(objc.RegisterName("setDrawableSize:"))
	})
	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(objc.RegisterName("new"))
	defer pool.Send(objc.RegisterName("drain"))
	tx := objc.ID(objc.GetClass("CATransaction"))
	tx.Send(objc.RegisterName("begin"))
	tx.Send(objc.RegisterName("setDisableActions:"), true)
	metalSetDrawableSize(layer, metalDrawableSizeSelector, struct{ Width, Height float64 }{1, 1})
	tx.Send(objc.RegisterName("commit"))
}
