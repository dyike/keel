# Form

English | [简体中文](form.zh-CN.md)

Two-column form: labels on the left and controls on the right, which are uniformly verified when submitted.

```go
name := kit.Input("")
amount := kit.NumberInput("").Range(0, 1e6)
f := kit.Form().
    Field("客户", name, func() string { return kit.Required(name.Value(), "请填写客户") }).
    Field("金额", amount, func() string {
        if amount.Value() <= 0 {
            return "金额必须大于 0"
        }
        return ""
    })
kit.Button("创建", func() {
    if f.Validate(cx) {
        create()
    }
})
```

- The verification function returns error information, and returns an empty string when legal; you can pass nil to indicate not to verify this item.
- `Validate(cx)` runs all verifications of visible fields, displays errors under each field, moves the focus to the first illegal field, and returns whether all pass. Must be called in callback.
- Controls with built-in error display clear errors according to their respective editing rules; errors in other controls are displayed by Form and retained until the next verification or asynchronous result update. `Errors()` returns a copy of the current error for each field.
- The control does not need to pass a label: Form will use the row label as the accessibility name of the control.
- Controls with built-in error display implement `kit.Validatable` (`SetError`, `FocusID`), including Input, TextArea, Select, NumberInput, OtpInput, TimeField, Combobox, and DatePicker.
- `kit.Required(s, msg)` returns msg if s is empty or entirely blank. `LabelWidth(dp)` sets the label column width, default is 72.
- `Actions(views...)` The row of buttons placed below the field is left aligned with the control column; it can still be clicked during submission for easy cancellation.

Agent: container role `form`, the row label is `text`, and the control is named after the row label.

Verify: `go run ./examples/components -section form`, add `-theme dark` to check the dark theme.

All non-nil validation functions for visible fields are executed. Errors in non-text fields such as Checkbox, Switch, Radio, Slider, Rating, etc. are placed under the control by Form; when `FocusID` is present, the control is focused, otherwise the error field container is focused. Disabled fields will not become focus targets. Available edit drafts for NumberInput, TimeField and Combobox are submitted before validation to avoid validation reading old values. Required treats Unicode whitespace (including full-width spaces) as null.

Asynchronous verification and submission share one request:

```go
token := f.BeginSubmit(cx)
if token == 0 { return } // Synchronization verification failed, disabled or already requested
payload := customer.Value() // Captured in UI callback, only snapshots are used in the background
go func() {
    errors := validateAndSave(payload) // []string, in Field order; nil indicates success
    core.Update(func() { f.FinishSubmit(token, errors) })
}()
```

When `Submitting()` is true, the field cannot be edited. Repeated `BeginSubmit` returns 0; the external submit button is available. `.Loading(f.Submitting())` displays busy status. `FinishSubmit` Resumes editing, displays individual field errors, and focuses on the first error after the field becomes available again. Error slices can be shorter than the number of fields, and the remaining fields will be processed as no errors; too long slices, expired or canceled tokens will return false and the status will not change.

`CancelSubmit()` invalidates the unfinished results and resumes editing; it does not cancel the network request of the business goroutine, and the application needs to cancel the I/O by itself. `SetDisabled(true)`, ancestor disabling, or resynchronizing `Validate` will also invalidate the request. When a program replaces a field value during submission, or removes a form, it should first call `CancelSubmit`. The background can only call the completion interface through `core.Update`.

## Multiple column and field configuration

`Columns(n)` arranges the fields into n equal-width grid columns, at least one column; `VerticalLabels(true)` places the label above the control. The two are independent, and the default is still a single field column with the label on the left. Responsive breakpoints are determined by the application and can be called in Render based on the available width of Columns. `Gap(dp)` sets the field spacing, `LabelTextSize(sp)` sets the label font size (0 restores inheritance). The size of the control is configured by each control itself.

```go
f := kit.Form().Columns(2).VerticalLabels(true).
    FieldWithOptions("姓名", name, validateName, kit.FormFieldOptions{
        Required: true, Description: "公开显示的姓名",
    }).
    FieldWithOptions("邮箱", email, validateEmail, kit.FormFieldOptions{
        Required: true,
    }).
    FieldWithOptions("介绍", bio, nil, kit.FormFieldOptions{ColSpan: 2}).
    Footer(kit.Button("保存", save))
```

`ColSpan` defaults to 1, limited to the current number of columns; `ColStart` counts from 1, and 0 indicates sequential arrangement. Specifies starting from the next row when the starting column is already occupied by the current row; shrinks to available columns when the starting column and span exceed the grid boundaries. Hidden fields do not occupy grid space. Actions continue to be arranged in the original way; Footer occupies the full width of the bottom and is aligned to the right, can be shared with Actions, and remains available in submissions.

`Required` only displays asterisks and does not replace the validator. `Description` Displays auxiliary text under the control; `DescriptionContent` can provide a rich content or dynamic view, taking precedence over plain text. `Hidden` defaults to false; hidden fields will not be rendered, drafts will not be submitted, validation will not be performed, and they will not become error focus targets. The insertion index and input status of other fields are not affected.

`SetFieldOptions(index, options)` replaces the configuration in the order of addition, and returns false for illegal indexes. Changing Hidden cancels an ongoing commit and clears this field error; other display configuration changes do not cancel commits. The asynchronous error array is still passed in in complete field order, and errors corresponding to hidden fields are ignored. Do not directly modify the business data in the submitted snapshot.

Automatic testing covers 1×/2× grid positioning, cross-column, starting column, hiding, shrinking to a single column, input status, required prompts/instructions/footers, and hidden field verification and asynchronous invalidation. The example shows two responsive columns, and the visual acceptance of the real device has not yet been completed.
