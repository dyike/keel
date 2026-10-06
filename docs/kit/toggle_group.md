# ToggleGroup

English | [简体中文](toggle_group.zh-CN.md)

A row of switch buttons. Single selection by default, multiple selections available after `Multiple()`.

```go
align := kit.ToggleGroup("Align left", "Center", "Align right")
style := kit.ToggleGroup("B", "I", "U").Multiple().OnChange(func(on []string) { … })
```

- In radio selection mode, clicking the pressed option again will cancel the selection.
- `Value()` returns the pressed option in option order (returns a copy), `SetValue(values...)` does not trigger a callback; radio mode only retains the first one. `SetDisabled`.

Agent: container role `group`, each button is `toggle`.

Verify: `go run ./examples/components -section toggle_group`, add `-theme dark` to check the dark theme.

`Variant(kit.ToggleGhost)` sets the transparent borderless appearance when not selected; `ToggleOutline` sets the transparent background with borders. `ToggleDefault` retains the original Surface background and borders. The Highlight background color is displayed when selected; the focus border appears only when the Ghost is focused.

`Size(...)` accepts `ToggleSizeXSmall`, `ToggleSizeSmall`, `ToggleSizeMedium`, `ToggleSizeLarge`, and adjusts the height, font size, icon and horizontal blank synchronously. The heights are 24/28/32/40dp respectively, and the default is Medium. Switching styles during operation retains the focus and selection status and does not trigger OnChange. Illegal enumeration values are ignored.

`Segmented(true)` connects adjacent buttons, with the default spacing of 0, rounded corners at the beginning and end, and a single stroke at the middle seam; Ghost retains the strokeless appearance and focus outline of each item. The single-item group retains four rounded corners, and the empty group does not draw buttons.

`Gap(dp)` sets a limited non-negative spacing; after the segmented group is set to a positive value, each item restores the four-corner rounding. `ResetGap()` Restores the default spacing for segmented group 0, normal group SpaceXs. `Segmented(false)` Restores normal groups; explicit gaps are not lost due to mode switching. Switching mode/spacing does not change the selected value or focus, and does not trigger callbacks. The default is still single selection, you can call Multiple to enable multiple selection.

Configure item by item using `Item(value, toggle)`:

```go
g := kit.ToggleGroup("star", "inbox").Multiple().Variant(kit.ToggleOutline).
    Item("star", kit.Toggle("", false).Icon(kit.IconStar)).
    Item("inbox", kit.Toggle("Inbox", false).Icon(kit.IconInbox))
g.SetValue("star")
```

value is still the option value during construction, separated from the display text. Agent names for icon-only items fall back to value. Item saves the configuration snapshot, and subsequent modifications to the source Toggle need to call the Item again to take effect; the Value and OnChange of the source Toggle do not participate in the group state. The group callback still returns the option values, arranged in the order of the original options.

Explicit itemized Size/Variant overrides group configuration, including explicit Medium/Default; unconfigured inherits group value. Item-by-item disabling and group disabling are superimposed, and existing selected values are not cleared and the program SetValue is not blocked. `Item(value, nil)` clears overwrites, unknown values are ignored. The replacement configuration retains the item's identity and focus. It is recommended that the connected segment groups have a unified height. Mixed sizes are still drawn according to each height and will not be automatically aligned.
