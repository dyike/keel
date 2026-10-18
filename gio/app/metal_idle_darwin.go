// SPDX-License-Identifier: Unlicense OR MIT

//go:build darwin && !nometal

package app

import "time"

// All state is accessed on the window's main thread. The timer only posts
// work there; it never touches the layer or the rendering state itself.
type metalIdleState struct {
	idleDebounce
	timer *time.Timer
	gpu   gpuIdler
}

// Idle records a presented frame that asked for no frame soon. The window
// releases GPU memory only once it has drawn nothing for a while; see
// idleDebounce.
func (c *mtlContext) Idle(g gpuIdler) {
	c.idle.gpu = g
	c.idle.trims = metalIdleTrims
	if wait, ok := c.idle.presented(time.Now()); ok {
		c.armIdle(wait)
	}
}

func (c *mtlContext) armIdle(wait time.Duration) {
	if c.idle.timer == nil {
		c.idle.timer = time.AfterFunc(wait, func() { runOnMain(c.fireIdle) })
		return
	}
	c.idle.timer.Reset(wait)
}

func (c *mtlContext) fireIdle() {
	if c.layer == 0 || c.idle.gpu == nil {
		return // released
	}
	if c.idle.due.IsZero() {
		return // a stale firing; the timer was rearmed
	}
	act, wait := c.idle.next(time.Now())
	switch act {
	case idleRelease:
		c.idle.gpu.Idle()
	case idleTrim:
		c.idle.gpu.Trim()
		trimMetalLayer(uintptr(c.layer))
	}
	if wait > 0 {
		c.armIdle(wait)
	}
}

func (c *mtlContext) stopIdle() {
	if c.idle.timer != nil {
		c.idle.timer.Stop()
	}
	c.idle = metalIdleState{}
}

func (c *mtlContext) resumeIdle() {
	if c.idle.resume() {
		resizeDrawable(c.view, c.layer)
	}
}
