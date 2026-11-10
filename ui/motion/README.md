# motion

Experimental: the API may change.

`motion` provides frame-time driven animation primitives inspired by Motion:
physics springs, tweens and cubic-bezier easing, keyframes and repeat modes,
inertia, stagger, presence, transforms, drag velocity tracking, color mixing,
range mapping, cursor-origin glow/ripples, and confetti. It depends on `ui/el`,
`ui/core`, `ui/theme`, and Gio. It is independent of `ui/kit` and `ui/window`.

Keep `Value` and transition slices in view state, call `Value.To` from every
render where the animation is visible, and do not use an infinite repeat for a
static status. Sampling uses `el.Context.Now`; only unsettled values call
`Context.Animating`, so a settled application remains event-driven. Reduced
motion jumps directly to the declared target.

`Transform` is paint-only: layout and automation bounds remain unchanged. It
replaces an element's existing `Decorate` callback. `LayoutGroup` performs
position-only FLIP for `Div` elements; it also owns `Decorate`, must stay in one
scroll coordinate space, and does not correct size changes or transformed
automation geometry.

## Pointer effects

Keep `PointerGlow` in the view and apply it to a `Div` every render:

```go
card := el.Div().Size(el.Dp(160))
v.glow.Apply(cx, "glow-card", card, motion.PointerGlowStyle{Color: color.NRGBA{R: 255, G: 255, B: 255, A: 120}})
```

The stable ID drives `Context.Hovered`; a pass-through pointer handler in the
element's paint coordinate space supplies the origin. `Origin` is element-local
dp under HiDPI scaling and inside a scroll area. Pointer input drives movement;
only fade and ripple lifetime call `Animating`. Reduced motion suppresses both
effects. `Apply` owns `Decorate`.

`Confetti` keeps an uncapped, reusable particle slice. Emission may grow the
slice; particle updates and painting allocate nothing after that capacity is
available. A seed makes the burst deterministic:

```go
v.confetti.Emit(cx.Now(), 180, 60, motion.ConfettiOptions{Count: 96, Seed: 7})
v.confetti.Apply(cx, stage)
```

Coordinates are local dp. Motion is sampled from elapsed time, not frame count.
When the last particle expires, the active slice becomes empty and no frame is
requested. Reduced motion clears the burst. `Apply` owns `Decorate`; wrap an
element if it must also use another decoration.


The numerical sampler is shared with `ui/el` through `ui/internal/animation` (standard library only). For element transforms whose frames can repaint retained layout, use `element.AnimatePose(el.Pose{...}, transition)`. Existing `Value.To` remains a Render-time value API, and the experimental `Transform`/`LayoutGroup` limitations above still apply.


## Timelines

Build a `Timeline` once in view state. Tracks share one playhead; labels and
placements allow sequences, overlap and stagger without independent timers:

```go
v.timeline.Track("x", motion.TweenSpec{To: 200, Duration: time.Second}).
    Label("arrival", motion.WithPrevious(600*time.Millisecond)).
    Track("fade", motion.TweenSpec{From: 1, To: 0, Duration: 400*time.Millisecond}, motion.AtLabel("arrival", 0))
// In Render:
f := v.timeline.Sample(cx, motion.TimelineOptions{Run: v.opened})
card.AnimatePose(el.Pose{X: f.Value("x"), Fade: f.Value("fade")}, el.Instant)
```

`AfterPrevious` is relative to the previous track's end (negative offsets
create overlap); `WithPrevious` uses its start; `AtTime` uses an absolute time.
A track name is unique. Composition is immutable after the first sample.
`TimelineOptions` supports pause, reverse, initial/repeat delays, extra repeat
count (`-1` for forever), alternating direction, and held `Seek` or `Progress`.
Bind a container directly with `TimelineOptions{Scroll: &motion.ScrollBinding{
ID: "list"}}`. The full scrollable extent maps to one cycle; `Horizontal` selects
ScrollX and `Start`/`End` choose a dp range. Missing/non-scrollable sources hold
zero. Scroll, Seek and Progress are mutually exclusive. A changed observed
content/viewport extent requests one follow-up frame; unchanged reads do not poll.
You can still supply Progress for custom mapping. Explicit seeks remain available with reduced
motion. Otherwise reduced motion finishes playback immediately. Infinite
playback then finishes its first play. A paused or settled timeline requests
no frames; sampling and track lookup allocate nothing after construction.

`Sample` runs in Render and therefore rebuilds that view for active playback.
Use `el.Instant` for its sampled poses to avoid a second interpolation. This
API does not make callbacks or report end-to-end display frame rate. These
additions are in the working branch and are not part of v0.4.2.
