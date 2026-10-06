# Skeleton

English | [简体中文](skeleton.zh-CN.md)

`kit.Skeleton().W(el.Dp(240)).H(el.Dp(16))` creates a gray loading placeholder with a default width of 16dp and a height of 16dp. Circle draws a centered circle, often with the same width and height. Shimmer Enables sweeping, which uses light and dark pulses by default.

Animation is based on cx.Now and frames are requested via cx.Animating; remains static when reduced motion. Skeleton is a decoration, does not expose Agent semantics, and does not handle the keyboard. Shimmer is an option, not a standalone component.

Verification: `go run ./examples/components -section skeleton -theme dark`, omit theme to see the light theme. Example button switching reduces animation; pixel test injects frame time to verify both animations and still state.

`Secondary(true)` halve the transparency of the entire placeholder pattern (including pulse or sweep) and restore it with `Secondary(false)`. `Rounded(dp)` Customize the rounded corners of the rectangle, 0 is a right angle, and the default is RadiusSm; negative numbers, NaN, and infinite values are ignored, and the drawing is limited to half of the short side. Rounded and Circle take effect after the caller; Circle still draws a centered circle in a non-square area. Colors are read from the theme each frame, and rounded corners and secondary color levels are retained when animation is reduced.
