# OtpInput

English | [简体中文](otp_input.zh-CN.md)

Enter the fixed-digit verification code in separate boxes.

```go
code := kit.OtpInput("Verification code", 6).OnComplete(func(s string) { verify(s) })
```

- Only numbers are accepted. After inputting, it will automatically jump to the next box. Backspace to return to the previous box. Pasting the entire verification code will fill all the boxes at once.
- `OnComplete` is called when the last digit is entered; `OnChange` is called after each edit.
- Implementation method: An invisible text box is covered on the grid, and focus, editing, and pasting are all handled by it. The grid is only responsible for display. Therefore the screen reader and Agent only see a text box.
- `Value()` / `SetValue` (discard non-digits and extra digits), `SetDisabled`, `SetError`.

`Masked(true)` masks the visible number and semantic value, `Masked(false)` restores the display; the original value is still handed to the application through `Value` and the callback. `Groups(n)` sets the number of groups, limited to 1-digit number. The default is two groups (single digits are one group). When it is not divisible, the previous group has one more digit. For example, 7 digits divided into 3 groups are 3–2–2. The spacing between groups is twice the spacing within the grid.

`Size(dp)` sets the grid height, the default is 48dp; the grid width, spacing and font size change proportionally, ignoring non-positive numbers and non-finite numbers. Modifying the layout or mask does not change the value and does not trigger the edit callback.

Agent: role `textbox`, the name is the label; `value` in normal mode is the entered number, and the mask mode is dots of equal length.

Verify: `go run ./examples/components -section otp_input`, add `-theme dark` to check the dark theme.

Each bit in the narrow container shrinks to the same width, and the actual editing area is consistent with the width of the visible grid. The default six digits are still fully displayed under the 224dp content width; there are regression tests for the 1× / 2× number boundary and the entire paragraph pasting.
