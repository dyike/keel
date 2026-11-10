package motion

import (
	"testing"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

func testTimeline() *Timeline {
	return new(Timeline).
		Track("x", TweenSpec{From: 0, To: 100, Duration: time.Second, Ease: Linear}).
		Label("title", WithPrevious(500*time.Millisecond)).
		Track("opacity", TweenSpec{From: 0, To: 1, Duration: time.Second, Ease: Linear}, AtLabel("title", 0))
}

func TestTimelineComposition(t *testing.T) {
	tl := testTimeline()
	start := time.Unix(50, 0)
	if tl.Duration() != 1500*time.Millisecond {
		t.Fatal(tl.Duration())
	}
	f := tl.sample(start, TimelineOptions{}, false)
	if f.Value("x") != 0 || f.Value("opacity") != 0 || !f.Active {
		t.Fatalf("initial: %+v", f)
	}
	f = tl.sample(start.Add(750*time.Millisecond), TimelineOptions{}, false)
	if f.Value("x") != 75 || f.Value("opacity") != .25 {
		t.Fatalf("overlap: x=%g opacity=%g", f.Value("x"), f.Value("opacity"))
	}
	f = tl.sample(start.Add(2*time.Second), TimelineOptions{}, false)
	if !f.Settled || !f.Arrived || f.Active || f.Value("opacity") != 1 {
		t.Fatalf("settled: %+v", f)
	}
	if !tl.sample(start.Add(2*time.Second), TimelineOptions{}, false).Arrived {
		t.Fatal("same-time resample lost arrival")
	}
	if tl.sample(start.Add(3*time.Second), TimelineOptions{}, false).Arrived {
		t.Fatal("arrival repeated")
	}
	if got := tl.sample(start.Add(3*time.Second), TimelineOptions{Run: 1}, false); got.Time != 0 || !got.Active {
		t.Fatalf("restart: %+v", got)
	}
}

func TestTimelinePauseReverseAndSeek(t *testing.T) {
	tl := testTimeline()
	start := time.Unix(60, 0)
	tl.sample(start, TimelineOptions{}, false)
	f := tl.sample(start.Add(500*time.Millisecond), TimelineOptions{Paused: true}, false)
	if f.Time != 500*time.Millisecond || f.Active {
		t.Fatalf("pause: %+v", f)
	}
	f = tl.sample(start.Add(time.Hour), TimelineOptions{}, false)
	if f.Time != 500*time.Millisecond || !f.Active {
		t.Fatalf("resume: %+v", f)
	}
	f = tl.sample(start.Add(time.Hour+100*time.Millisecond), TimelineOptions{Reverse: true}, false)
	if f.Time != 600*time.Millisecond {
		t.Fatalf("reverse discontinuity: %+v", f)
	}
	f = tl.sample(start.Add(time.Hour+300*time.Millisecond), TimelineOptions{Reverse: true}, false)
	if f.Time != 400*time.Millisecond {
		t.Fatalf("reverse: %+v", f)
	}
	progress := float32(.5)
	f = tl.sample(start.Add(2*time.Hour), TimelineOptions{Progress: &progress}, false)
	if f.Time != 750*time.Millisecond || f.Active || f.Settled {
		t.Fatalf("seek: %+v", f)
	}
	f = tl.sample(start.Add(3*time.Hour), TimelineOptions{}, false)
	if f.Time != 750*time.Millisecond {
		t.Fatalf("seek release jumped: %+v", f)
	}
	f = tl.sample(start.Add(3*time.Hour), TimelineOptions{Run: 1, Reverse: true}, false)
	if f.Time != tl.Duration() || !f.Active {
		t.Fatalf("reverse restart: %+v", f)
	}
	f = tl.sample(start.Add(4*time.Hour), TimelineOptions{Run: 1, Reverse: true}, false)
	if f.Time != 0 || !f.Arrived {
		t.Fatalf("reverse finish: %+v", f)
	}
}

func TestTimelineDelayRepeatAndReducedMotion(t *testing.T) {
	start := time.Unix(70, 0)
	for _, alternate := range []bool{false, true} {
		tl := new(Timeline).Track("x", TweenSpec{To: 1, Duration: time.Second, Ease: Linear})
		o := TimelineOptions{Delay: 200 * time.Millisecond, Repeat: 1, RepeatDelay: 300 * time.Millisecond, Alternate: alternate}
		tl.sample(start, o, false)
		for _, tt := range []struct {
			at    time.Duration
			value float32
			done  bool
		}{
			{100 * time.Millisecond, 0, false},
			{700 * time.Millisecond, .5, false},
			{1400 * time.Millisecond, 1, false},
			{2000 * time.Millisecond, .5, false},
			{2500 * time.Millisecond, 1, true},
		} {
			want := tt.value
			if tt.done && alternate {
				want = 0
			}
			f := tl.sample(start.Add(tt.at), o, false)
			if abs(f.Value("x")-want) > 1e-6 || f.Settled != tt.done {
				t.Fatalf("alternate=%v at=%s: value=%g frame=%+v", alternate, tt.at, f.Value("x"), f)
			}
		}
		o.Run++
		f := tl.sample(start.Add(time.Hour), o, true)
		want := float32(1)
		if alternate {
			want = 0
		}
		if f.Value("x") != want || f.Active || !f.Settled {
			t.Fatalf("reduced: %+v", f)
		}
		p := float32(1)
		o.Progress = &p
		f = tl.sample(start.Add(2*time.Hour), o, true)
		if f.Value("x") != 1 || f.Active {
			t.Fatalf("explicit seek endpoint: %+v", f)
		}
	}
	tl := new(Timeline).Track("x", TweenSpec{To: 1, Duration: time.Second})
	f := tl.sample(start, TimelineOptions{Repeat: -1, Alternate: true}, true)
	if f.Value("x") != 1 || f.Active {
		t.Fatalf("infinite reduced: %+v", f)
	}
}

func TestTimelineZeroDurationAndValidation(t *testing.T) {
	start := time.Unix(80, 0)
	tl := new(Timeline).Track("switch", TweenSpec{From: 2, To: 3})
	f := tl.sample(start, TimelineOptions{Delay: time.Second}, false)
	if !f.Active || f.Value("switch") != 2 {
		t.Fatalf("delay ended early: %+v", f)
	}
	f = tl.sample(start.Add(time.Second), TimelineOptions{Delay: time.Second}, false)
	if f.Active || !f.Settled || f.Value("switch") != 3 {
		t.Fatalf("instant track: %+v", f)
	}
	for name, fn := range map[string]func(){
		"unknown label":        func() { new(Timeline).Label("a", AtLabel("b", 0)) },
		"duplicate":            func() { new(Timeline).Track("a", TweenSpec{}).Track("a", TweenSpec{}) },
		"editing after sample": func() { tl.Track("late", TweenSpec{}) },
		"invalid repeat":       func() { new(Timeline).sample(start, TimelineOptions{Repeat: -2}, false) },
		"overflow":             func() { new(Timeline).Track("long", TweenSpec{Duration: time.Duration(1<<63 - 1)}, AtTime(1)) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			fn()
		})
	}
}

type timelineView struct {
	timeline *Timeline
	options  TimelineOptions
	frame    TimelineFrame
}

func (v *timelineView) Render(cx *el.Context) el.Element {
	v.frame = v.timeline.Sample(cx, v.options)
	return el.Div().Translate(v.frame.Value("x"), 0)
}

func TestTimelineFrameScheduling(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(false)
	v := &timelineView{timeline: testTimeline()}
	f := newFakeFrames(v)
	start := time.Unix(90, 0)
	if !f.at(start) {
		t.Fatal("active timeline did not request a frame")
	}
	v.options.Paused = true
	if f.at(start.Add(200 * time.Millisecond)) {
		t.Fatal("paused timeline requested a frame")
	}
	v.options.Paused = false
	f.at(start.Add(time.Second))
	if f.at(start.Add(3 * time.Second)) {
		t.Fatal("settled timeline requested a frame")
	}
	v.options.Run++
	theme.SetReducedMotion(true)
	if f.at(start.Add(4*time.Second)) || v.frame.Value("x") != 100 {
		t.Fatal("reduced timeline did not settle")
	}
}

func TestTimelineSampleAllocations(t *testing.T) {
	tl := testTimeline()
	start := time.Unix(100, 0)
	o := TimelineOptions{Repeat: -1}
	tl.sample(start, o, false)
	allocs := testing.AllocsPerRun(100, func() { start = start.Add(time.Millisecond); tl.sample(start, o, false).Value("x") })
	if allocs != 0 {
		t.Fatalf("sampling allocated %g times", allocs)
	}
}
