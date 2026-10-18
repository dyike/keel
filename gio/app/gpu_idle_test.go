// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"testing"
	"time"
)

// runIdle plays frames at the given interval for n frames, firing the
// debounce timer the way a context does, and counts the actions taken.
func runIdle(d *idleDebounce, interval time.Duration, n int, start time.Time) (releases, trims int, end time.Time) {
	now := start
	fire := func(until time.Time) {
		for !d.due.IsZero() && !d.due.After(until) {
			act, _ := d.next(d.due)
			switch act {
			case idleRelease:
				releases++
			case idleTrim:
				trims++
			}
		}
	}
	for i := 0; i < n; i++ {
		fire(now)
		d.resume()
		now = now.Add(2 * time.Millisecond) // render
		d.presented(now)
		now = now.Add(interval - 2*time.Millisecond)
	}
	return releases, trims, now
}

func TestIdleDebounceFrequentUpdates(t *testing.T) {
	d := &idleDebounce{trims: true}
	start := time.Unix(1000, 0)
	// 60 externally driven frames a second for 10 s: never idle.
	rel, trim, end := runIdle(d, time.Second/60, 600, start)
	if rel != 0 || trim != 0 {
		t.Fatalf("released %d, trimmed %d times while drawing at 60 Hz", rel, trim)
	}
	// Then quiet: one release after the release delay, one trim after the
	// trim delay, and nothing more.
	act, wait := d.next(end.Add(gpuReleaseDelay))
	if act != idleRelease || wait <= 0 {
		t.Fatalf("after release delay: %v, %v", act, wait)
	}
	if act, _ := d.next(end.Add(gpuReleaseDelay + wait - time.Millisecond)); act != idleNone {
		t.Fatalf("trimmed early: %v", act)
	}
	if act, wait := d.next(end.Add(gpuTrimDelay)); act != idleTrim || wait != 0 {
		t.Fatalf("after trim delay: %v, %v", act, wait)
	}
	if act, wait := d.next(end.Add(10 * gpuTrimDelay)); act != idleNone || wait != 0 {
		t.Fatalf("after trim: %v, %v", act, wait)
	}
	if !d.resume() {
		t.Fatal("resume did not report the trim")
	}
	if d.resume() {
		t.Fatal("second resume reported a trim")
	}
}

func TestIdleDebounceTimerChurn(t *testing.T) {
	d := &idleDebounce{trims: true}
	now := time.Unix(1000, 0)
	arms, fires := 0, 0
	// 60 frames a second for 10 s: the timer is set once and fires about
	// every release delay, not once per frame.
	for i := 0; i < 600; i++ {
		for !d.due.IsZero() && !d.due.After(now) {
			d.next(d.due)
			fires++
		}
		d.resume()
		if _, ok := d.presented(now); ok {
			arms++
		}
		now = now.Add(time.Second / 60)
	}
	if arms != 1 || fires > 45 {
		t.Fatalf("timer set %d times, fired %d times in 600 frames", arms, fires)
	}
}

func TestIdleDebounceSlowUpdates(t *testing.T) {
	d := &idleDebounce{trims: true}
	// A frame every 400 ms: each gap releases, none trims.
	rel, trim, _ := runIdle(d, 400*time.Millisecond, 10, time.Unix(1000, 0))
	if rel != 9 || trim != 0 {
		t.Fatalf("released %d, trimmed %d; want 9, 0", rel, trim)
	}
}

func TestIdleDebounceAnimation(t *testing.T) {
	d := &idleDebounce{trims: true}
	now := time.Unix(1000, 0)
	d.resume()
	d.presented(now)
	// The next frame asks for another frame soon: presented is not called.
	d.resume()
	if act, wait := d.next(now.Add(time.Hour)); act != idleNone || wait != 0 {
		t.Fatalf("animating window: %v, %v", act, wait)
	}
}

func TestIdleDebounceNoTrim(t *testing.T) {
	d := &idleDebounce{}
	now := time.Unix(1000, 0)
	d.presented(now)
	if act, wait := d.next(now.Add(gpuReleaseDelay)); act != idleRelease || wait != 0 {
		t.Fatalf("release: %v, %v", act, wait)
	}
	if act, _ := d.next(now.Add(time.Hour)); act != idleNone {
		t.Fatalf("trimmed without trims: %v", act)
	}
}
