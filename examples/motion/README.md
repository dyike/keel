# Motion example

Run the requested cursor-origin and order effects directly:

```sh
go run ./examples/motion -section effects
```

Move the pointer across the blue card: the spotlight follows it and a ripple
starts at the entry or press position. Click **Place order** to emit deterministic
Gio-drawn confetti from the button centre. Turn on **Reduced motion** to suppress
both effects.

Run `go run ./examples/motion` for the complete nine-section Motion lab. Use
`-lang zh-CN` for Chinese, or capture the focused section with:

```sh
go run ./examples/motion -section effects -screenshot /tmp/keel-motion-effects.png
```

Run `go run ./examples/motion -section pose` for paint-only element transforms. Click Move/reverse or catch the moving card; its hit region and anchored overlay follow it. Reduced motion jumps to the target.

Run `go run ./examples/motion -section timeline` for overlapping tracks, replay, pause, reverse and seek.

The timeline page also has Scroll / clock: scroll its lower pane to drive the same tracks directly.
