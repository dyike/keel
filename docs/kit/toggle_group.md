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

All variants are borderless. `ToggleGhost` and `ToggleOutline` use a transparent background when not selected; `ToggleDefault` uses a `Subtle` fill. Segmented groups give every unselected segment a `Subtle` fill and separate segments with a 1dp gap. Selected items use `Highlight`, and keyboard/programmatic focus draws the inset focus ring.

`Size(...)` accepts `ToggleSizeXSmall`, `ToggleSizeSmall`, `ToggleSizeMedium`, `ToggleSizeLarge`, and adjusts the height, font size, icon and horizontal blank synchronously. The heights are 24/28/32/40dp respectively, and the default is Medium. Switching styles during operation retains the focus and selection status and does not trigger OnChange. Illegal enumeration values are ignored.

`Segmented(true)` groups adjacent buttons with a 1dp background gap and rounds only the outer corners. It draws no outer stroke or seam. Each item retains its keyboard focus ring. A single-item group keeps all four rounded corners; an empty group draws no buttons.

`Gap(dp)` sets finite non-negative spacing; a value other than the segmented default 1dp restores four rounded corners on every item. `ResetGap()` restores 1dp for segmented groups and SpaceXs for ordinary groups. `Segmented(false)` restores an ordinary group; explicit gaps survive mode switching. Changing mode or spacing does not change selection or focus and does not trigger callbacks.

Configure item by item using `Item(value, toggle)`:

```go
g := kit.ToggleGroup("star", "inbox").Multiple().Variant(kit.ToggleOutline).
    Item("star", kit.Toggle("", false).Icon(kit.IconStar)).
    Item("inbox", kit.Toggle("Inbox", false).Icon(kit.IconInbox))
g.SetValue("star")
```

value is still the option value during construction, separated from the display text. Agent names for icon-only items fall back to value. Item saves the configuration snapshot, and subsequent modifications to the source Toggle need to call the Item again to take effect; the Value and OnChange of the source Toggle do not participate in the group state. The group callback still returns the option values, arranged in the order of the original options.

Explicit itemized Size/Variant overrides group configuration, including explicit Medium/Default; unconfigured inherits group value. Item-by-item disabling and group disabling are superimposed, and existing selected values are not cleared and the program SetValue is not blocked. `Item(value, nil)` clears overwrites, unknown values are ignored. The replacement configuration retains the item's identity and focus. It is recommended that the connected segment groups have a unified height. Mixed sizes are still drawn according to each height and will not be automatically aligned.
