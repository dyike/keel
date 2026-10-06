# Stepper

English | [简体中文](stepper.zh-CN.md)

Shows the progress of a multi-step process.

```go
steps := kit.Stepper("Order details", "Confirm payment", "Shipping").Navigable()
steps.SetValue(1) // Go to the second step
```

- The steps before the current step are completed (showing check marks), the steps before the current step are bolded, and the steps after the step are not started.
- `Navigable()` allows clicking on a completed step to return to that step, in which case `OnChange` is called.
- `Value()` returns the sequence number of the current step. When it is equal to the number of steps, it means all steps are completed; `SetValue` does not trigger a callback.

Agent: container role `list`, each step is `step`, `value` is `done` / `current` / `upcoming`.

Verify: `go run ./examples/components -section stepper`, add `-theme dark` to check the dark theme.

Provides horizontal scrolling when there are more steps than visible width, preserving step order and connecting lines. Navigable completed steps support keyboard focus; `SetDisabled` disables scrolling and step modification. Label slices are copied during construction. There are regression tests for click after last item scrolling, disabling and 1× / 2×.

`Vertical()` changes to vertical arrangement and vertical scrolling, and the connecting lines are arranged along the center of the marked circle. `Size(dp)` sets the mark diameter, commonly used are 20, 24 (default), and 32dp; the length of the icon and connecting line scales accordingly, and the text uses the theme font size file. Zero, negative, NaN, and infinity values do not change the current size.

Use `SetEntries` when you need an icon, a single item disabled, or a multi-line description:

```go
steps := kit.Stepper().Vertical().Size(32).Navigable()
steps.SetEntries(
    kit.StepperItem{Label: "Orders", Icon: kit.IconReceipt},
    kit.StepperItem{Label: "Payment", Icon: kit.IconLock, Disabled: true},
    kit.StepperItem{Label: "Shipping", Icon: kit.IconInbox},
)
steps.SetValue(2)
```

- `StepperItem.Icon` Replaces the number/complete check mark; retains the original mark if no icon is provided.
- `StepperItem.Disabled` or `SetItemDisabled(index, on)` prohibits the click and keyboard focus of this step, does not change its completion status, and does not prevent the `SetValue` program from jumping. Out-of-bounds indexes are ignored.
- `StepperItem.Content` can provide display content such as multi-line description; `Label` is still the Agent name. Content should not nest buttons or input boxes.
- `SetEntries` and `Entries()` both copy entry slices, and the content View does not make a deep copy. When replacing an entry, keep the current index and converge to the new length, without triggering `OnChange`.
- Custom items respect the configured navigation scope; single item, entire component, and ancestor disabling all prevent navigation.

`TextCenter(true)` Centers the horizontal step's text/rich content below the marker, with connecting lines centered around adjacent marker circles. Each column equally divides the available width. The minimum width is three times the diameter of the mark. If it is insufficient, it scrolls horizontally; custom content can set the internal layout by itself. Portrait mode keeps the markup to the left of the content and only sets the text to be centered. false restores the original layout. `Horizontal()` switches from portrait to landscape orientation.

`Navigation(kit.StepperNavigationNone / StepperNavigationCompleted / StepperNavigationAll)` is read-only, can only return completed steps, and can select any non-disabled steps. Default is None, `Navigable()` is equivalent to Completed. All Contains steps that have not yet been started; clicking the current step does not trigger OnChange. Neither configuration changes nor SetValue trigger callbacks.

Three navigation strategies have been verified, including repeated clicks on the current item, Tab skipping disabled items, dynamic direction switching, double-ratio equal-width columns and narrow window scrolling; the return of light and dark window pixels confirms that the connection line is colored with the completion status.
