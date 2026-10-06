# Spinner

English | [简体中文](spinner.zh-CN.md)

`kit.Spinner().Size(20).Label("Loading")` shows indeterminate progress. When Label is empty, only graphics are drawn, and the default name is "Loading". Agent role progressbar, value is indeterminate. No keyboard operation.

The rotation phase comes from cx.Now, cx.Animating requests the next frame, no goroutine is started. Remain static after theme.SetReducedMotion(true). Fixed graphic size, text wraps according to parent container constraints.

Verification: `go run ./examples/components -section spinner -theme dark`, omit theme to see the light theme; example button switching reduces animation. Pixel testing injects different frame times, checking rotation and stillness.

`Icon(kit.IconSettings)` replaces the ring with a rotating icon, `Icon(kit.IconNone)` restores the ring; `VectorIcon` accepts a custom Gio icon, and passes nil to restore the ring. `Color(color.NRGBA{...})` sets the graphic color, and the label still uses the theme's Muted color. PrimaryText is used with the theme when no color is set. Custom icons and rings default to one circle per second and remain stationary when animation is reduced; the size is consistent with Agent semantics.

`Period(2 * time.Second)` sets the time for one rotation. The larger the value, the slower the rotation; 0 returns to the default one second, and negative values are ignored. Period changes immediately recalculate the phase according to the shared frame clock, possibly changing the current angle. Reduced motion takes precedence over period settings. Loading controls such as buttons still use the default cycle.
