# RisingGlow

English | [简体中文](rising_glow.zh-CN.md)

A decorative particle field inspired by the current [Ruixen Rising Glow preview](https://ruixen.com/docs/components/rising-glow). This is a native Go/Gio implementation, with no web runtime or new cgo.

The reference page's prose describes a text glow wrapper, but its current preview and installation code render a separate particle field **below** a heading. Keel follows that implementation: dots of 4–10dp diameter, random horizontal positions, initial delays of 0–2 seconds, and stable individual cycles of 3–5 seconds. Each cycle moves upward and then back down, fading in and out, with the CSS ease-in-out curve on each half. Dots are clipped to the field. No blur or text shadow is added.

    glow := kit.RisingGlow().
        ParticleCount(80).
        Color(color.NRGBA{R: 0x13, G: 0x8d, B: 0xff, A: 255}).
        Height(100)

    // Keep glow in the view's state, not inside Render.
    return el.Div().Items(el.Center).Child(
        el.Text("keel").TextSize(60).Bold(),
        el.Div().WFull().MaxW(el.Dp(384)).Child(glow.Render(cx)),
    )

Import image/color, github.com/dyike/keel/ui/el and github.com/dyike/keel/ui/kit.

| Method | Behavior |
| --- | --- |
| Width(dp) | Zero (default) fills available width; positive values fix width in dp. |
| Height(dp) | Positive dp height; default 120. |
| ParticleCount(n) | 0–512, default 20. Zero hides particles and stops ticks. Changing count regenerates the field. |
| Color(c) | Default is the current theme's PrimaryText; the example explicitly uses the reference blue. |
| Enabled(on) | Default true. False shows a still representative pose; re-enabling restarts. |

Invalid dimensions or counts are ignored. Reapplying the same count or enabled state does not reset the effect. Particle parameters are generated once per count change, not per frame. Reduced motion shows the same still pose. Fully transparent, empty and clipped-out fields request no animation frames. Visible animation uses Keel's ambient paint cadence and retained layout; it is a continuous decorative animation, so it is not idle while playing.

The field adds no accessible role, label, focus target or input handler. Headings and controls remain separate ordinary elements. Place it in a constrained container when embedding it in rows; display density changes raster resolution, not the field's dp dimensions.

Preview:

    go run ./examples/components -section rising_glow
