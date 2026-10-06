# CopyButton

English | [简体中文](copy_button.zh-CN.md)

Copies text to the clipboard and displays "Copied" as feedback.

```go
kit.CopyButton(func() string { return order.ID })
```

- Text is only read when clicked, so the function always gets the latest content.
- After copying, the button displays "Copied" and a check icon, and returns to "Copy" after `kit.CopiedFeedback` (1.5 seconds).
- The button is Ghost style, 28dp high, supports Tab focus and Space/Enter triggering.

Agent: The role is `button` and the name switches between "Copied" and "Copied".

Verification: `go run ./examples/components -section copy_button`.

Continuous clicks will retain the full 1.5 seconds of feedback after each copy, regardless of the last cutoff time. `SetDisabled(true)` disables copying and clears feedback; ancestor disabling also prevents reading text and writing to the clipboard. When nil is passed to the value function, "Copied" will not be displayed when clicked.

`OnCopied(func(string))` is called after submitting the clipboard write request. The parameter is the original text submitted this time, and the value function will not be read again. The system clipboard did not successfully acknowledge the signal, so this callback does not represent the operating system acknowledging the write. Not called when the disabled or value function is nil; passing nil removes the callback.

`Content(el.View)` Replaces the default icon and text, preserving the button's keyboard action and "copy"/"copied" semantic name. Content should be text, icons and other display elements, and avoid nested input boxes or buttons. The minimum height of custom content is 28dp, which can be increased with the content; pass nil to restore the default button.

```go
copy := kit.CopyButton(func() string { return order.ID }).
    OnCopied(func(value string) { lastCopied = value })
copy.Content(el.ViewFunc(func(cx *el.Context) el.Element {
    label := "Copy order number"
    if copy.Copied() {
        label = "Order number copied"
    }
    return el.Text(label)
}))
```

`Copied()` provides the current feedback status. Custom content shares a 1.5-second timer with the default button; the timer is reset each time it is copied, and `SetDisabled(true)` clears feedback.
