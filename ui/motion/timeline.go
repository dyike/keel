package motion

import (
	"math"
	"time"

	"github.com/dyike/keel/ui/el"
)

// TweenSpec describes one track of a Timeline. A nil Ease uses EaseInOut.
type TweenSpec struct {
	From, To float32
	Duration time.Duration
	Ease     Easing
}

// Placement locates a track or label on a timeline.
type Placement struct {
	kind   uint8
	offset time.Duration
	label  string
}

// AfterPrevious overlaps the preceding track when offset is negative.
func AfterPrevious(offset time.Duration) Placement { return Placement{offset: offset} }

// WithPrevious starts relative to the preceding track's start.
func WithPrevious(offset time.Duration) Placement { return Placement{kind: 1, offset: offset} }

// AtTime uses an absolute position in the timeline.
func AtTime(at time.Duration) Placement { return Placement{kind: 2, offset: at} }

// AtLabel starts relative to a previously declared label.
func AtLabel(label string, offset time.Duration) Placement {
	return Placement{kind: 3, offset: offset, label: label}
}

type timelineTrack struct {
	name  string
	start time.Duration
	TweenSpec
}

// Timeline coordinates named tracks on one frame-time playhead. Build it once
// in view state, then call Sample each visible Render. Tracks and labels cannot
// change after playback starts; create a new Timeline for a new composition.
// Its zero value is ready to use. It has no timers or background goroutines.
type Timeline struct {
	tracks                               []timelineTrack
	labels                               map[string]time.Duration
	duration, previousStart, previousEnd time.Duration
	started, moving, reverse, settled    bool
	run                                  uint64
	last, arrived                        time.Time
	raw                                  time.Duration
}

// Track adds a named tween. With no placement it follows the previous track.
// Names must be nonempty and unique. A zero duration switches at its start.
func (t *Timeline) Track(name string, spec TweenSpec, at ...Placement) *Timeline {
	if t.started {
		panic("motion: cannot edit a playing timeline")
	}
	if name == "" || spec.Duration < 0 || !finiteTimeline(spec.From) || !finiteTimeline(spec.To) {
		panic("motion: invalid timeline track")
	}
	for _, track := range t.tracks {
		if track.name == name {
			panic("motion: duplicate timeline track " + name)
		}
	}
	start := t.place(at)
	end := addTimelineTime(start, spec.Duration)
	t.tracks = append(t.tracks, timelineTrack{name, start, spec})
	t.previousStart, t.previousEnd = start, end
	t.duration = max(t.duration, end)
	return t
}

// Label names a position without changing the preceding track or duration.
func (t *Timeline) Label(name string, at ...Placement) *Timeline {
	if t.started {
		panic("motion: cannot edit a playing timeline")
	}
	if name == "" {
		panic("motion: empty timeline label")
	}
	if t.labels == nil {
		t.labels = make(map[string]time.Duration)
	}
	if _, exists := t.labels[name]; exists {
		panic("motion: duplicate timeline label " + name)
	}
	t.labels[name] = t.place(at)
	return t
}

func (t *Timeline) place(at []Placement) time.Duration {
	if len(at) > 1 {
		panic("motion: use one timeline placement")
	}
	p := Placement{}
	if len(at) == 1 {
		p = at[0]
	}
	base := t.previousEnd
	switch p.kind {
	case 1:
		base = t.previousStart
	case 2:
		base = 0
	case 3:
		var ok bool
		base, ok = t.labels[p.label]
		if !ok {
			panic("motion: unknown timeline label " + p.label)
		}
	}
	return max(0, addTimelineTime(base, p.offset))
}

func finiteTimeline(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
func addTimelineTime(a, b time.Duration) time.Duration {
	if b > 0 && a > time.Duration(math.MaxInt64)-b || b < 0 && a < time.Duration(math.MinInt64)-b {
		panic("motion: timeline duration overflow")
	}
	return a + b
}

// Duration is the length of one play, without delays or repeats.
func (t *Timeline) Duration() time.Duration { return t.duration }

// TimelineOptions controls playback. Change Run to restart. Repeat counts extra
// plays; -1 repeats forever. Reverse retraces the playhead, including repeats.
// Pause, seek and scroll binding request no frames. Seek, Progress and Scroll
// are mutually exclusive,
// hold the first cycle at that position, and resume there when removed.
type TimelineOptions struct {
	Run                        uint64
	Repeat                     int
	Alternate, Reverse, Paused bool
	Delay, RepeatDelay         time.Duration
	Seek                       *time.Duration
	Progress                   *float32
	Scroll                     *ScrollBinding
}

// TimelineFrame is a sample of a composition. Value returns a named track's
// value at Time. Paused/seeked frames may be neither Active nor Settled.
type TimelineFrame struct {
	timeline                 *Timeline
	beforeStart              bool
	Time                     time.Duration
	Progress                 float32
	Active, Settled, Arrived bool
}

// Value samples a named track, clamping to From/To outside its interval.
func (f TimelineFrame) Value(name string) float32 {
	if f.timeline != nil {
		for _, track := range f.timeline.tracks {
			if track.name != name {
				continue
			}
			if f.beforeStart || f.Time < track.start {
				return track.From
			}
			if f.Time >= track.start+track.Duration {
				return track.To
			}
			p := float32(f.Time-track.start) / float32(track.Duration)
			ease := track.Ease
			if ease == nil {
				ease = EaseInOut
			}
			return track.From + (track.To-track.From)*ease(p)
		}
	}
	panic("motion: unknown timeline track " + name)
}

// Sample advances all tracks with one clock. Reduced motion finishes playback
// immediately, while explicit pause and seek still hold the requested position.
// Feed returned values to styles or to AnimatePose with el.Instant to
// apply the sampled pose directly, without adding a second interpolation.
func (t *Timeline) Sample(cx *el.Context, options TimelineOptions) TimelineFrame {
	if options.Scroll != nil {
		if options.Seek != nil || options.Progress != nil {
			panic("motion: Scroll, Seek and Progress are mutually exclusive")
		}
		progress := options.Scroll.progress(cx)
		options.Progress = &progress
	}
	f := t.sample(cx.Now(), options, el.ReducedMotion())
	if f.Active {
		cx.Animating()
	}
	return f
}

func (t *Timeline) sample(now time.Time, o TimelineOptions, reduced bool) TimelineFrame {
	if o.Repeat < -1 || o.Delay < 0 || o.RepeatDelay < 0 || o.Seek != nil && o.Progress != nil || o.Progress != nil && !finiteTimeline(*o.Progress) {
		panic("motion: invalid timeline playback options")
	}
	cycle := addTimelineTime(t.duration, o.RepeatDelay)
	end := t.duration
	if o.Repeat >= 0 && cycle > 0 {
		if int64(o.Repeat) > (math.MaxInt64-int64(t.duration))/int64(cycle) {
			panic("motion: timeline repeat overflow")
		}
		end += time.Duration(o.Repeat) * cycle
	}
	limit := addTimelineTime(o.Delay, end)
	restart := !t.started || o.Run != t.run
	if restart {
		t.raw, t.run, t.arrived, t.settled = 0, o.Run, time.Time{}, false
		if o.Reverse {
			t.raw = limit
		}
	} else if t.moving && now.After(t.last) {
		dt := now.Sub(t.last)
		if t.reverse {
			t.raw = max(o.Delay, t.raw-dt)
		} else {
			t.raw += min(dt, time.Duration(math.MaxInt64)-t.raw)
		}
	}
	t.started, t.last, t.reverse = true, now, o.Reverse
	held := o.Seek != nil || o.Progress != nil
	if o.Seek != nil {
		t.raw = addTimelineTime(o.Delay, min(max(*o.Seek, 0), t.duration))
	}
	if o.Progress != nil {
		position := t.duration
		if *o.Progress < 1 {
			position = time.Duration(float64(max(*o.Progress, 0)) * float64(t.duration))
		}
		t.raw = addTimelineTime(o.Delay, position)
	}
	if reduced && !held && !o.Paused {
		t.raw = limit
		if o.Reverse {
			t.raw = o.Delay
		}
	}
	if o.Repeat >= 0 {
		t.raw = min(t.raw, limit)
	}
	raw := t.raw - o.Delay
	done := o.Repeat >= 0 && raw >= end
	if o.Reverse {
		done = raw <= 0
	}
	if reduced && !held && !o.Paused || cycle == 0 && raw >= 0 {
		done = true
	}
	position := time.Duration(0)
	if raw >= 0 {
		index := int64(0)
		if cycle > 0 {
			index, position = int64(raw/cycle), min(raw%cycle, t.duration)
		}
		if o.Repeat >= 0 && raw >= end {
			index, position = int64(o.Repeat), t.duration
		}
		if o.Alternate && index%2 != 0 {
			position = t.duration - position
		}
	}
	if o.Reverse && done {
		position = 0
	}
	if held {
		position = min(max(raw, 0), t.duration)
	}
	// Infinite motion has no last cycle; reduced motion ends the first play.
	if reduced && !held && !o.Paused && o.Repeat == -1 && !o.Reverse {
		position = t.duration
	}
	if done && !t.settled {
		t.arrived = now
	}
	if !done {
		t.arrived = time.Time{}
	}
	t.settled = done
	t.moving = !done && !held && !o.Paused
	progress := float32(1)
	if t.duration > 0 {
		progress = float32(position) / float32(t.duration)
	}
	return TimelineFrame{timeline: t, beforeStart: raw < 0, Time: position, Progress: progress, Active: t.moving, Settled: done, Arrived: done && t.arrived.Equal(now)}
}
