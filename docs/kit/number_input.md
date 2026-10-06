# NumberInput

English | [简体中文](number_input.zh-CN.md)

Numeric input box with − / + buttons.

```go
qty := kit.NumberInput("数量").Range(1, 99).Step(1)
price := kit.NumberInput("单价").Range(0, 1e6).Step(0.5).Decimals(2)
```

- Intermediate states outside the range are allowed during the input process. For example, if you want to enter 15, the 1 typed first may be less than the lower limit; when you press Enter or leave the input box, it is limited to the range and formatted, and text that cannot be parsed, NaN, infinity, and overflow are restored to the original value.
- − / + buttons are adjusted in steps and disabled when reaching the boundary. The step size is only used for increase and decrease, and the handle input value is not absorbed to the step size multiple; the draft adds one step and only issues a final value callback. Decimal steps such as 0.1 do not accumulate binary addition errors.
- ↑ ↓ Add and subtract according to the step size, PageUp / PageDown 10 steps at a time.
- `Value()` / `SetValue`, `SetDisabled`, `SetError`; `Decimals(n)` will round the actual value and the displayed value to 0-15 decimal places. The default `-1` will retain the precision; illegal digits will not take effect. Range endpoints take precedence: For example, if the range is 0.001–0.009, even if two decimal places are set, the exact boundary will be displayed to avoid inconsistency between the text and the value.

`SetValue` rejects non-finite values; `Range` containing NaN or having only one endpoint value of infinity has no effect. Programmatic assignment does not trigger callbacks. Disabling the component or ancestor will abandon the uncommitted draft and restore the committed value; ordinary defocusing will still be submitted.

Agent: input box role `textbox`, `value` is the displayed text; the two buttons are named "Reduce Label" and "Increase Label".

Verify: `go run ./examples/components -section number_input`, add `-theme dark` to check the dark theme.

`StepBy(func(value float64, action kit.NumberStepAction) float64)` calculates the positive step size based on the current valid draft and direction. The direction is `NumberStepActionIncrement` or `NumberStepActionDecrement`. Called once per button/keyboard action, PageUp/PageDown multiplies the step size by ten and does not re-evaluate step by step. Returning zero, negative, or non-finite value will cancel this action and keep the draft; the callback should not modify the same NumberInput. `StepBy(nil)` restores the last fixed step size, and legal `Step` calls will replace the dynamic strategy. Rendering, procedural assignment and pure input do not call the strategy.

`Prefix(view)` / `Suffix(view)` Place currency symbols, units or action buttons before and after the text, inside the − / + buttons; pass nil to remove. Dynamic additions and deletions will not change the identity or content of the editor, and child controls inherit the overall disabled state. Slots and step buttons occupy a fixed content width, and narrow windows should avoid placing custom content that is too wide.

```go
price.StepBy(func(value float64, action kit.NumberStepAction) float64 {
    if value < 1 || value == 1 && action == kit.NumberStepActionDecrement {
        return 0.1
    }
    return 0.5
})
price.Suffix(kit.Button("帮助", showHelp))
```

When the application needs to take over increases and decreases, set `OnStep(func(kit.NumberStepEvent))`. The event contains the valid draft `Value`, the direction `Action` and the step number `Count` (1 for normal buttons/arrow keys, 10 for PageUp/PageDown). This mode does not automatically submit drafts, does not call dynamic strategies, and does not trigger `OnChange`; the callback can be used to update the display with `SetValue`, or not update it temporarily. Invalid drafts use the most recently submitted value, and the text remains unchanged; normal carriage return/out-of-focus submission behavior remains unchanged.

`OnStep(nil)` restores the previous fixed/dynamic strategy; legal `Step` or any `StepBy` call exits event mode. Stepping outward is still prohibited at the range boundary, and events will not be dispatched in the disabled state. `SetValue` in the callback follows the programmatic assignment rules, so no additional `OnChange` will be generated.

`Size(dp)` synchronously adjusts the minimum height, text size, step button and spacing, 28/36/48dp is recommended; `Size(0)` restores the theme height and inherited font size, negative values and non-finite values are ignored. The explicit size is scaled according to the default height of the current theme, and slot custom content can still specify its own font size or heighten the container.

`Appearance(false)` removes the default background, border, rounded corners and padding, retaining the minimum height, label/error prompt, button interaction and input focus; `Appearance(true)` restores it. No-decoration mode also does not display the error/focus color on the original border, and the application can draw it on the outer layer by itself.

Full-width digits `０–９`, symbols `＋/－`, and decimal points `．/。` appear as half-width immediately after typing or pasting, and the mapped selection still uses the rune position. Symbols in wrong positions and repeated decimal points will be rejected; intermediate drafts such as empty strings, individual symbols, `.5`, etc. are retained, and the range and precision are still applied when entering, defocusing, form submission, or increasing or decreasing. Pure input does not trigger numeric OnChange.

`ThousandsSeparator(',')` Adds thousandths digit to edit and commit display; also supports spaces, single quotes, non-breaking space U+00A0 and narrow non-breaking space U+202F, 0 is off, other characters are ignored. Decimal separators are fixed to points. Combined with `Decimals(2)`, Prefix, the amount can be displayed; the range endpoints are still fully displayed when more precision is required. Parsing and incrementing and decrementing will remove the configured grouping sign, and the Value will always be float64. Setting the delimiter restores the display of recently submitted values and discards uncommitted drafts.

```go
price := kit.NumberInput("金额").Decimals(2).ThousandsSeparator(',')
price.SetValue(12345.6) // Showing 12,345.60
```

Normalize use of `el.Input().Transform`: process text and map selection before writing Bind / OnChange. This mode saves undo status once per user modification, up to 100 times; format changes after programmatic assignment or submission reset this history. When a group symbol is deleted it is regenerated numerically, and numbers can be continued to be deleted after the cursor is mapped to the previous number. Input boxes that do not have a Transform set continue to use the original editor history.

Automatic testing covers instant full-width conversion, grouping, intermediate replacement, undo/redo, malformed input rejection, different delimiters, stepping and range endpoint accuracy; real device input method combination and visual acceptance have not yet been completed.
