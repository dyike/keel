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

