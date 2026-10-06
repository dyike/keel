# DropdownButton

English | [简体中文](dropdown_button.zh-CN.md)

Button with drop-down menu.

```go
kit.DropdownButton("Export", formats)              // The entire button opens the menu
kit.DropdownButton("Save", more).Split(save)     // "Save" executes save, and the arrow next to it opens the menu.
```

- The menu is a normal `kit.Menu`, and the keyboard, submenus, and disabled items all behave the same as Menu.
- `Variant(kit.ButtonSecondary)` etc. set the button appearance. In the split style, the two parts use the same appearance.
- `SetDisabled(true)` Also disables buttons and arrows, and closes open menus.
- `el.Root` REQUIRED.

Agent: In the overall style, the button name is the title; in the split style, the main operation button is named the title, and the arrow button is named "Title More Options".

Verification: `go run ./examples/components -section dropdown_button`.

When the menu argument is nil an empty menu is used and the main Split action is still available. `SetDisabled(true)` simultaneously disables the main button and the area where the menu is located; even if the Menu is held externally and SetValue(true) is called, the disabled anchor point will not display the pop-up layer. Ancestor disabling follows the same rules.

`Button(button)` uses an existing Button to configure the main action and enables split mode; supports its text, accessible name, icon, rich content, color matching, compact/stroke, loading and click callbacks. Rendering copies the configuration without modifying the original Button; internally uses the component's own stable ID and ignores the ID of the incoming Button. If Split is set at the same time, the action of Split takes priority. Button(nil) reverts to the original normal/Split usage.

`Size(dp)` sets the height of the two parts, 0 restores the inner button or default height; negative numbers and non-finite values are ignored. When Variant/Size is not explicitly set, the two parts inherit the variant and height of the inner button; the icon, content, color callback, etc. only belong to the main operation.

`Loading(bool)` Overrides the loading state of the main button. In normal mode, the menu cannot be opened through the loading button; in split mode, the arrow can still open the menu. Menus that have been opened are not closed by Loading. `SetDisabled(true)` always disables both parts and closes the menu. Disabling the inner Button itself only affects the main action.

```go
kit.DropdownButton("Save", more).
    Button(kit.Button("Save", save).Icon(kit.IconDone).Loading(saving)).
    Size(40)
```

`Placement(side, align)` Sets the top menu direction (Top/Bottom/Left/Right) and alignment (Start/Center/End), defaults to Bottom/Start; illegal combinations are ignored. `Offset(dp)` sets the spacing, the default is 4dp, supports 0 and negative value overlap, and non-finite values are ignored. It can be updated while it is open. When there is insufficient space, overlay flipping and in-window restrictions will be used. The submenu is still expanded by Right/Start, 2dp.

These two methods of DropdownButton directly configure the incoming Menu; normal mode anchors the whole button, and split mode anchors the arrow. Callers that share the same Menu will also see configuration changes.
