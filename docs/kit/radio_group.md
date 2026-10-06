# RadioGroup

English | [简体中文](radio_group.zh-CN.md)

Choose one of several options.

```go
pay := kit.RadioGroup("付款方式", "转账", "支票", "现金").OnChange(func(s string) { … })
size := kit.RadioGroup("尺寸", "S", "M", "L").Horizontal()
```

- The keyboard behavior is consistent with the native radio group: Tab falls on the currently selected item when entering the group (it falls on the first available item when there is no selection or the selected item is disabled), the arrow keys move and select, and the head and tail cycle. By default, the entire group occupies only one Tab stop; an explicit ItemTab can override this.
- `SetOptionDisabled(value, bool)` disables a single item and retains the existing selection; the arrow keys skip disabled items and Home / End go to the first and last available items. When fully disabled, tab stops are not occupied.
- `Item(value)` Returns an independently renderable option, suitable for use in card layouts; all items in the same group share selection, keyboard order, and disabled state. Each value is rendered only once per frame; the outer layer declares the `radiogroup` role and name itself. Items that are not rendered or have disabled ancestors do not become arrow key targets.
- `Size(dp)` sets the diameter of the dots in the group, the default is 18dp, and the positive value is limited to 12–64dp; the `TextSize(sp)` setting option inherits the font size, and the positive value is limited to 8–128sp. Passing 0 to both returns to default, negative/non-finite values are ignored; independent `Item` uses the same configuration.
- `Content(value, view)` replaces the visible label of the corresponding option, and can combine title, description, icon and other display content. The original option string is still the value and accessible name; the content should not contain buttons or input boxes. Pass nil to restore text, unknown options are ignored; rearrange the retained content, and remove options to clean up the content. Content is rendered every frame, with explicit font sizes taking precedence over the group's inherited font sizes.
- `Options()` returns a copy; passed options are also copied, and null and duplicate values are removed. Reordering preserves option identity.
- `Value()` returns the selected item, and is empty when not selected; `SetValue` does not trigger a callback; `SetOptions` replaces the option, and clears it when the original selection is not in the new option; `SetDisabled`.

Agent: The role of the group is `radiogroup`, each option is `radio`, and `checked` indicates whether it is selected.

Verify: `go run ./examples/components -section radio_group`, add `-theme dark` to check the dark theme.

Rich tag example:

```go
plans := kit.RadioGroup("套餐", "基础", "专业").Size(28).TextSize(20)
plans.ItemSize("基础", 18, 14)
plans.Content("专业", el.ViewFunc(func(cx *el.Context) el.Element {
    return el.Div().Child(el.Text("专业版").Bold(),
        el.Text("支持团队协作").TextSize(theme.TextSm).TextColor(theme.Muted))
}))
```

The default is to use a single tab stop within the group. `TabStop(false)` or `TabIndex(-1)` can skip the entire group, still allowing mouse and arrow key selection; non-negative TabIndex is arranged in ascending order, the same value is in tree order, and the range is limited to el root. `ItemTab(value, stop, index)` can overwrite the single-item configuration, which explicitly participates in docking judgment and is no longer restricted by the default single docking point; `ClearItemTab(value)` restores the group configuration. Item-by-item coverage, rearrangement and retention, deletion and cleaning, unknown options are ignored. Disabled items can never be docked; Tab movement itself does not change the selected value.

`ItemSize(value, dp, sp)` can overwrite dots and font sizes item by item. When each parameter is 0, the group configuration is inherited; `ItemSize(value, 0, 0)` clears the override. Positive values inherit the upper and lower bounds of the group, unknown options or any negative/non-finite values ignore the entire call. Rearranging retains coverage, and deleting options cleans up. Independent Items also work.

See [Radio](radio.md) for a single, freely positioned radio button.
