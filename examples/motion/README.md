# Motion example

Run the requested cursor-origin and order effects directly:

```sh
go run ./examples/motion -section effects
```

Move the pointer across the blue card: the spotlight follows it and a ripple
starts at the entry or press position. Click **Place order** to emit deterministic
Gio-drawn confetti from the button centre. Turn on **Reduced motion** to suppress
both effects.


All animation demonstrations live here. Run `go run ./examples/motion` for the complete lab, or choose a section:

- `-section presets`: Smooth / Snappy / Bouncy, including mid-animation reversal.
- `-section pages`: page fades with retained input.
- `-section effects`: pointer glow, ripples and confetti.

Use `-lang zh-CN` for Chinese. `examples/app-shell` demonstrates application layout.
