# Radio

English | [简体中文](radio.zh-CN.md)

The options are scattered in different cards and table rows. When RadioGroup is not used, use a single `kit.Radio(label)`:

```go
express, pickup := kit.Radio("Delivery"), kit.Radio("Store pickup")
express.SetValue(true)
express.OnChange(func(bool) { pickup.SetValue(false) })
pickup.OnChange(func(bool) { express.SetValue(false) })
```

- Click or space to select; clicking again while already selected will not cancel it, so the mutual exclusion is handled by the application in OnChange. The program call `SetValue` does not trigger the callback.
- Appearance is consistent with RadioGroup's options: `Size`, `TextSize`, `Content` Replace text labels with any view (the construction parameter is still the accessibility name).
- Clicking is invalid after disabling; the Agent role is `radio`, and `selected` is true when selected.
- A common set of options still recommends RadioGroup: it comes with arrow key switching and Tab entry rules.

Verification: `go run ./examples/components -section radio`.
