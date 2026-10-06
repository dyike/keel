# StatusMarker

English | [简体中文](status_marker.zh-CN.md)

`kit.StatusMarker("Synced")` is used for message status, timeline boundaries, and system prompts. By default, the width of the parent container is filled, the content is moved to the left, and the text uses the theme Muted. Existing `kit.Marker(shape)` continues drawing chart geometry markers.

```go
kit.StatusMarker("Today").Variant(kit.StatusMarkerSeparator)
kit.StatusMarker("3 unread messages").Variant(kit.StatusMarkerBorder).
    Content(kit.Button("View", openMessages))
kit.StatusMarker("Generating…").Loading(true).
    LoadingStyle(kit.StatusMarkerLoadingStyleShimmer).ID("generation").Role("status")
```

Variant supports Plain, Separator, and Border; the separator is centered by default, Alignment(el.Start/Center/End) explicitly overrides it, and ResetAlignment restores the default. Separator only leaves the right line when moving to the left, and only the left line when moving to the right; Border draws lines at the bottom of the entire row.

Icon accepts any View, with a default 16dp slot; Content adds rich content after the text, and SetText("") can only use rich content. Children replaces children added directly to the end of the line, and the input slice is copied. Child controls such as buttons retain their own click, focus, and disable inheritance, and the status line itself does not become a clickable control, nor does it maintain an unread count or notification lifecycle.

Loading adds a spinner by default; retains the icon if it already exists. Shimmer mode sweeps text, and pure rich content changes transparency in a two-second cycle; icons and lines remain static. The ShimmerStyle callback configures the persistent ShimmerText (Duration, Spread, Highlight, Reverse, Once, etc.), and the text and Enabled are managed by the status line. Switch Loading/LoadingStyle to restart text sweeping; reduce static display when animation is reduced. Rich content transparency is fixed at 0.7–1, text sweep configuration does not change it.

PartStyle configures Root, Row, Icon, Content, and Separator respectively; the callback receives a new element every frame, and the ID is restored after the style to avoid losing the child control state. The Separator style also works on the Border bottom line. By default, status roles are not added, and ID/Role is used when needed; Agent can read text, and system screen reading is still limited by Keel/Gio platform support.

The layout uses Keel's row layout and theme scales and does not copy upstream dimensions. Alignment simultaneously sets the content group position and text alignment; child elements can be independently overridden with TextAlign.

Verification: `go run ./examples/components -section status_marker`.
