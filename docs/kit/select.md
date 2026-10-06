# Select

English | [简体中文](select.zh-CN.md)

Choose an item from the list.

```go
status := kit.Select("Status", "Awaiting payment", "Paid", "Shipped").OnChange(func(s string) { … })
city := kit.Select("City", cities...).Searchable()
```

- Click, Enter, Space or ↓ to open the list. ↑ ↓ moves in the list (looping from beginning to end), Home / End jumps to the beginning and end, Enter selects, Esc closes; after closing, the focus returns to the drop-down box.
- `Searchable()` Add a search box at the top of the list. When opening, focus on the search box and press Enter to select the first matching item. Press ↓ in the search box to enter the list.
- `Hint(s)` is the text displayed when no selection is made. By default, the locale's "Please select" is used.
- `Value()` / `SetValue`, `SetOptions` (cleared when the original selection is not in the new option), `SetDisabled`, `SetError`. `el.Root` REQUIRED.

Agent: drop-down box roles `select`, `value` are the current options; after opening, the list is `listbox`, each item is `option`, `selected` represents the currently selected item.

Verify: `go run ./examples/components -section select`, add `-theme dark` to check the dark theme.

`SetEntries(...SelectOption)` supports standalone `Value`, `Label`, `Group` and `Disabled`. Display the title before consecutive groups, and use Value for empty Label. Value must be non-empty and unique. Illegal data will panic before modification; `Entries()` returns a copy. The old `Select(label, options...)` / `SetOptions` still take literals as values. Copy data for all entries and clean up removed selections after dynamic replacement. `SetOptionDisabled(value, on)` controls a single option; clicks and keyboard skip over disabled items, programmatic assignment still allows it.

`Multiple()` Enables multi-selection, click, space or enter to switch the current item and keep the drop-down box open, Esc or external click to close. `Values()` returns an independent copy in option order, `SetValues` replaces the selection without triggering a callback, and `OnValuesChange` receives the user-modified copy. `Value()` is the primary value, multiple selections should use `Values()`; `SetValue` will be replaced with a single value. Use Label for display and Value for callback.

Search for matching tags or values. List virtualization only builds rows near the viewport, and scrolls to the current option when opened; the arrow keys bypass titles and disabled items, Home/End jumps to the beginning and end, and PageUp/PageDown turns pages. Keyboard focus remains on the option container, preventing long lists from losing focus after recycling rows; returning to fields after closing. Examples include grouping, disabled items, and 10,000 multiple-select options.

## Customized displays and menus

`RenderItem(func(*el.Context, SelectItemContext) el.Element)` Customizes visible row content; context includes Entries index, option copy, selection, and keyboard activity state. Return nil to use the default text, the outer layer is still responsible for option semantics, disabling, checking and selecting. Content should be presented as display elements, and interactions should be placed in independent controls. `RenderValue` receives a copy of the selected item and customizes the display when closed; the placeholder prompt remains the default and returns nil to restore the label. The value of Agent continues to use the stored value and does not follow the custom drawn text.

`TitlePrefix("City: ")` Adds prefix only when selected, up to half field width and truncated. `Empty(view)` replaces no matching content, nil returns to default; empty content area can be scrolled. `Match(func(SelectOption,string) bool)` Replaces the search match on the original query received; nil restores the tag/value containment match. Recalling Match after the data the closure depends on changes will invalidate the cache.

`Clearable(true)` displays an independent clear button, clears selections, queries and errors and closes the menu. The focus returns to the field; the multi-select callback receives nil, and OnChange receives an empty string when the main value changes. Empty selections do not send callbacks repeatedly, the procedure SetValue/SetValues remains silent, and disabling inheritance blocks the button.

`MenuWidth(dp)` sets the menu width, 0 follows the field; `MenuMaxHeight(dp)` sets the height budget including search/blank space, 0 defaults to 240dp, the minimum positive number is 64dp, and the window available space takes priority. `Size(dp)` scales fields, text, spacing and default line height, recommended 28/36/48, minimum positive number is 20; `RowHeight(dp)` independently sets unified options/group height, minimum positive number is 16, 0 follows Size. The custom content must fit within the row height, and the excess part will be clipped by the virtual list. The above interface ignores negative numbers and non-finite values.

`Appearance(false)` Removes field background and borders, retaining labels, error text, keyboard, options, and disabled behavior. The menu still uses the theme overlay appearance.
