# Marker

English | [简体中文](marker.zh-CN.md)

`kit.Marker(kit.MarkerDiamond).Color(theme.Info).Size(12)` creates pure graphic markers and supports MarkerDot, MarkerSquare, and MarkerDiamond. Default 8dp, color is read every time Render reads theme.Text; explicit Color remains fixed.

Markers do not expose semantics to Agents and have no keyboard operations. Used for legends, list items, and charts. When a readable label is required, the caller puts Text next to it. Parent container constraints limit the drawing size.

Verification: `go run ./examples/components -section marker -theme dark`, omit theme to see the light theme.

Message status, timeline boundaries, unread reminders, and loading lines use [StatusMarker](status_marker.md).
