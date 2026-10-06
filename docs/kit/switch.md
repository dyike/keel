# Switch

English | [简体中文](switch.zh-CN.md)

Switches that take effect immediately, such as "Receive notifications".

```go
notify := kit.Switch("Receive notifications", true).OnChange(func(on bool) { save(on) })
```

- Click, Space, Enter to switch; `Value()` / `SetValue(bool)`, `SetValue` does not trigger the callback; `SetDisabled`.

Agent: role `switch`, `checked` represents the switch status.

Verify: `go run ./examples/components -section switch`, add `-theme dark` to check the dark theme.

`Size(kit.SwitchSmall)` uses a 28×16dp track, and the default `SwitchMedium` is 36×20dp; the sliders are 12/16dp respectively. The label font size continues to inherit from the parent element. `LabelSide(el.Left/el.Right)` sets the label position, defaults to the right side; labels and tracks share a clickable, focusable control, and the focus is retained when switching styles. Illegal enumeration values are ignored.

`Color(color.NRGBA)` only covers the selected track, `ClearColor()` restores the theme color. When disabled and selected, the alpha of the custom color is halved; when no custom color is set, the original disabled color is retained. Ancestor disables inherit the disable behavior of el. The theme color should be passed in on Render to follow theme switches.

```go
notify.Size(kit.SwitchSmall).LabelSide(el.Left).Color(theme.Success)
```

Tooltips are available through external composition.

The slider position uses a 180ms smooth transition (`kit.SwitchDuration`) to quickly reverse the transition from the current display position. Show target position directly when first showing and reduced motion; track color, Value, callback and Agent checked status update immediately. The animation uses the frame clock and does not create a timer; disabling does not change the existing value, and the program SetValue can still update and trigger position transitions.

`FocusRing(false)` hides the keyboard/program focus outline, true returns to default. It does not remove focus or tab stops, nor does it affect Space/Enter; used in scenes where there is already an outer focus hint. The focus outline is drawn on the outer circle of the track, excluding labels; mouse clicks follow the original strategy of not displaying the focus ring. When a custom control wants to draw an outline on a certain component, you can set a transparent `FocusStyle` for the focusable element, and then use `cx.FocusVisible(id)` to determine whether to display it.

`TabStop(false)` Skip switches from Tab/Shift+Tab traversal while still allowing mouse or `cx.Focus(s.FocusID())` focus and keyboard use. `TabIndex(n)` sorts the stops in ascending order, keeping the tree order with the same value; defaults to 0, and skips with negative values. When explicitly set, el root takes over tab traversal; when not configured, Gio's native ordering is maintained. Disabled, hidden and currently undrawn controls are skipped; modal/TrapFocus loops independently within the overlay, and the first focus also follows the ordering.

```go
notify.TabIndex(2)
secondary.TabStop(false)
```

The sorting range is the same el root and cannot be sorted across multiple independent Embed or native Gio controls. Calling Gio Router.MoveFocus directly does not go through this rule; the app should send normal tab events. The editor's explicit handling of Tabs takes precedence over global traversal.
