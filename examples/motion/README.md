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
