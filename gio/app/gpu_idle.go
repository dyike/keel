// SPDX-License-Identifier: Unlicense OR MIT

package app

import "time"

// Keel patch: a window redrawn by frequent external updates (network data
// arriving through Invalidate) asks for no next frame after each of them,
// yet draws again a few milliseconds later. Releasing GPU memory right after
// such a frame paid kernel calls to mark textures volatile and freed the
// staging buffers only to allocate them again the next frame. idleDebounce
// releases only after the window has drawn nothing for a while.

const (
	// gpuReleaseDelay is how long a window draws nothing before its
	// scratch textures become volatile and its spare buffers are freed.
	gpuReleaseDelay = 250 * time.Millisecond
	// gpuTrimDelay is how long a window draws nothing before rebuildable
	// frame data and spare drawables are released.
	gpuTrimDelay = time.Second
)

// gpuIdler is the renderer's side of idle reclamation (package gpu).
type gpuIdler interface {
	Idle()
	Trim()
}

type idleAction uint8

const (
	idleNone    idleAction = iota
	idleRelease            // call gpu.Idle
	idleTrim               // call gpu.Trim and shrink the drawables
)

type idleStage uint8

const (
	idleActive idleStage = iota
	idleReleased
	idleTrimmed
)

// idleDebounce decides when a quiet window releases GPU memory. It has no
// timer of its own: its owner asks next when its timer fires.
type idleDebounce struct {
	// last is when the most recent frame was presented.
	last time.Time
	// requested reports whether that frame asked for no frame soon.
	requested bool
	stage     idleStage
	// due is when the owner's timer fires, zero when stopped.
	due time.Time
	// trims enables the second stage.
	trims bool
}

// resume records that a frame is being drawn. It reports whether the
// previous idle period went as far as trimming.
func (d *idleDebounce) resume() (trimmed bool) {
	trimmed = d.stage == idleTrimmed
	d.requested = false
	d.stage = idleActive
	return trimmed
}

// presented records a frame that asked for no frame soon. When the timer
// must be started or moved earlier, it returns the delay to set it to.
func (d *idleDebounce) presented(now time.Time) (time.Duration, bool) {
	d.last = now
	d.requested = true
	if due := now.Add(gpuReleaseDelay); d.due.IsZero() || d.due.After(due) {
		d.due = due
		return gpuReleaseDelay, true
	}
	return 0, false
}

// next returns the action due at now, and how long to wait before asking
// again; 0 means the timer stops until the next idle frame.
func (d *idleDebounce) next(now time.Time) (idleAction, time.Duration) {
	act, wait := d.step(now)
	d.due = time.Time{}
	if wait > 0 {
		d.due = now.Add(wait)
	}
	return act, wait
}

func (d *idleDebounce) step(now time.Time) (idleAction, time.Duration) {
	if !d.requested {
		// Drawing again, and the latest frame wants more: an animation.
		return idleNone, 0
	}
	quiet := now.Sub(d.last)
	switch d.stage {
	case idleActive:
		if quiet < gpuReleaseDelay {
			return idleNone, gpuReleaseDelay - quiet
		}
		d.stage = idleReleased
		if !d.trims {
			return idleRelease, 0
		}
		return idleRelease, max(gpuTrimDelay-quiet, time.Millisecond)
	case idleReleased:
		if !d.trims {
			return idleNone, 0
		}
		if quiet < gpuTrimDelay {
			return idleNone, gpuTrimDelay - quiet
		}
		d.stage = idleTrimmed
		return idleTrim, 0
	}
	return idleNone, 0
}
