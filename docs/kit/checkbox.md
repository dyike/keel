# Checkbox

English | [简体中文](checkbox.zh-CN.md)

A labeled checkbox will toggle when clicked, Space, or Enter.

```go
agree := kit.Checkbox("同意条款", false).OnChange(func(on bool) { … })
all.SetMixed(true) // "Select All" displays half selection when partially selected
```

- `Value()` / `SetValue(bool)` reads or sets whether it is checked or not. `SetValue` does not trigger a callback and clears the half-selected state.
- `SetMixed(true)` displays half-selected and becomes checked when clicked.
- It cannot be operated after `SetDisabled(true)`, and the text can be modified after `SetLabel`.
- Checkbox is used for options that require the user to click "Submit" to take effect; Switch is used for switches that take effect immediately.

Agent: Role `checkbox`, `checked` indicates whether to check or not; when half-selected, `value` becomes `mixed`.

Verify: `go run ./examples/components -section checkbox`, add `-theme dark` to check the dark theme.

`Size(dp)` sets the box side length, positive values are limited to 12–64dp, 0 returns to the default 18dp; checkmarks and half-selected horizontal lines scale simultaneously. `TextSize(sp)` sets the label font size individually, positive values are limited to 8–128sp, 0 restores the parent element font size. Negative and non-finite values are ignored; modifying the appearance does not change the value, half-selected state, or existing focus.

```go
agree.Size(24).TextSize(18).TabIndex(2)
```

`TabStop(false)` Just skips tab traversal and still has mouse and program focus. `TabIndex(n)` traverses in ascending order, keeping the same value in tree order, and skipping negative values. Reuse el's single root rule, loop independently within the overlay, and do not sort across independent Embed or native Gio controls. The focus outline is still drawn along the entire line of labels and boxes.
