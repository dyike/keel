# Glass Studio

[简体中文](README.zh-CN.md)

Run the native macOS glass demo:

```sh
go run ./examples/glass -backdrop
```

The sidebar and header expose a native backdrop; the main content remains opaque. Click navigation items, switch appearance or open Clear and frosted comparison windows. `-style clear`, `-style frosted`, `-dark` and `-opaque` select startup options. `-backdrop` opens a colored reference window behind the demo.

Liquid Glass needs macOS 26+ and a macOS 26+ SDK. Older macOS versions/SDKs use vibrancy; other platforms retain an opaque background. Native effects are visible in desktop windows, not off-screen screenshots. See [the window API](../../docs/app.md#macos-glass-backdrop).
