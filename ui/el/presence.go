package el

import (
	"time"

	"github.com/dyike/keel/ui/internal/animation"
)

// Inert keeps a subtree's appearance while excluding it from pointer input,
// keyboard focus, automation and platform accessibility. It also dismisses
// overlays owned by its descendants. Use it for content animating out.
func (s *Styled[T]) Inert(on bool) *T { s.node().inert = on; return s.self }

// PresenceOptions describes the poses at the start of entry and end of exit.
// The resting pose is Pose{}. Animation must be a finite spring or tween.
type PresenceOptions struct {
	Enter, Exit Pose
	Animation   Animation
}

// PresenceFrame says whether to build content for a presence declaration.
// Visible includes the exit interval; Exiting content must use Inert(true).
// Pose is already sampled: apply it with AnimatePose(Pose, Instant).
type PresenceFrame struct {
	Visible, Exiting, Settled, Arrived bool
	Pose                               Pose
}

// Presence applies a sampled frame, including input and accessibility
// suppression while exiting. Build children only while p.Visible to avoid
// retaining their work after the exit. It owns AnimatePose and Inert.
func (s *Styled[T]) Presence(p PresenceFrame) *T {
	s.AnimatePose(p.Pose, Instant)
	s.node().inert = p.Exiting
	if !p.Visible {
		s.node().style.hidden = true
	}
	return s.self
}

type presenceState struct {
	values  [7]animation.Value
	epoch   uint64
	settled bool
	arrival time.Time
}

// Presence owns entry/exit state under a comparable key unique to this root.
// Declare it each Render, including while present is false, and build its
// content while Visible. Changing present in flight preserves spring velocity.
// Omitting the declaration cancels it and releases state after this build.
// Finished exits release their state immediately; they retain no element tree
// or callback. Declare outside Cache builders.
//
// p := cx.Presence("card", v.show, el.PresenceOptions{Enter: el.Pose{Y: 12, Fade: 1}, Exit: el.Pose{Y: -12, Fade: 1}, Animation: el.Snappy})
// if p.Visible { body.Child(card.AnimatePose(p.Pose, el.Instant).Inert(p.Exiting)) }
func (cx *Context) Presence(key any, present bool, options PresenceOptions) PresenceFrame {
	if options.Animation.Kind != animation.KindSpring && options.Animation.Kind != animation.KindTween || options.Animation.Repeat != 0 {
		panic("el: presence requires a non-repeating spring or tween")
	}
	validatePose(options.Enter)
	validatePose(options.Exit)
	r := cx.root
	now := r.e.gtx.Now
	state := r.presences[key]
	if state == nil && !present {
		return PresenceFrame{Settled: true}
	}
	live := r.e.gtx.Enabled()
	if live && state != nil && state.epoch == r.presenceEpoch {
		panic("el: presence key declared twice in one build")
	}
	if state == nil {
		state = new(presenceState)
		for i, value := range options.Enter.values() {
			state.values[i].Set(value)
		}
		if live {
			if r.presences == nil {
				r.presences = make(map[any]*presenceState)
			}
			r.presences[key] = state
		}
	} else if !live {
		copy := *state
		state = &copy
	}
	state.epoch = r.presenceEpoch
	target := Pose{}
	if !present {
		target = options.Exit
	}
	values := target.values()
	active := false
	tr := defaultAnimation(options.Animation)
	for i, target := range values {
		v := &state.values[i]
		if !v.Animating() && v.Get() == target {
			continue
		}
		var moving bool
		values[i], moving = v.Sample(now, target, tr, ReducedMotion() || tr.Duration < 0)
		active = active || moving
	}
	if active {
		cx.Animating()
		state.arrival = time.Time{}
	} else if !state.settled {
		state.arrival = now
	}
	state.settled = !active
	if !active && state.arrival.Equal(now) {
		cx.noteRenderTime()
	}
	if !present && !active && live {
		delete(r.presences, key)
	}
	return PresenceFrame{Visible: present || active, Exiting: !present, Settled: !active, Arrived: !active && state.arrival.Equal(now),
		Pose: Pose{X: values[0], Y: values[1], XPercent: values[2], YPercent: values[3], Scale: max(values[4], .0001), Rotate: values[5], Fade: min(max(values[6], 0), 1)}}
}

func (r *RootWidget) sweepPresences() {
	if !r.e.gtx.Enabled() {
		return
	}
	for key, state := range r.presences {
		if state.epoch != r.presenceEpoch {
			delete(r.presences, key)
		}
	}
	for key, state := range r.presenceLists {
		if state.epoch() != r.presenceEpoch {
			delete(r.presenceLists, key)
		}
	}
}
