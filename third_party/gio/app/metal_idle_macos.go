// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !ios && !nometal

package app

import (
	"sync"
	"time"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// All state is accessed on the window's main thread. The timer only posts
// work there; it never touches the layer or the rendering state itself.
type metalIdleState struct {
	timer      *time.Timer
	generation uint64
	trimmed    bool
}

func (c *mtlContext) Idle(release func()) {
	if c.idle.timer != nil || c.idle.trimmed {
		return
	}
	c.idle.generation++
	generation := c.idle.generation
	c.idle.timer = time.AfterFunc(time.Second, func() {
		runOnMain(func() {
			if c.layer == 0 || c.idle.generation != generation {
				return
			}
			c.idle.timer = nil
			release()
			trimMetalLayer(uintptr(c.layer))
			c.idle.trimmed = true
		})
	})
}

func (c *mtlContext) stopIdle() {
	c.idle.generation++
	if c.idle.timer != nil {
		c.idle.timer.Stop()
		c.idle.timer = nil
	}
}

func (c *mtlContext) resumeIdle() {
	c.stopIdle()
	if c.idle.trimmed {
		resizeDrawable(c.view, c.layer)
		c.idle.trimmed = false
	}
}

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
