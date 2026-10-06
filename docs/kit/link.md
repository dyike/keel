# Link

English | [简体中文](link.zh-CN.md)

Clickable text links.

```go
kit.Link("View details", openDetail)
```

- Main color text, darkens on hover. It can be triggered with Tab focus, Enter or Space.
- `SetText`, `SetDisabled`; when disabled, it becomes the secondary text color and does not respond to operations.

Agent: role `link`.

Verify: `go run ./examples/components -section link`, add `-theme dark` to check the dark theme.
