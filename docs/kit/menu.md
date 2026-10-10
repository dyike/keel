# Menu

English | [简体中文](menu.zh-CN.md)

A menu of commands that pops up next to the triggering element.

```go
export := kit.Menu().Item("PDF", "", exportPDF).Item("CSV", "", exportCSV)
more := kit.Menu().
    Item("Copy", "mod+c", copy).
    Separator().
    Sub("Export", export).
    Item("Delete", "delete", remove)
more.SetItemDisabled("Copy", !hasSelection)
more.Trigger(kit.Button("More", more.Toggle).Variant(kit.ButtonGhost))
```

- `Item(label, shortcut, action)`: `shortcut` uses the writing method of `core.ParseShortcut`, only uses `kit.Kbd` to display, **does not register** the shortcut key; pass an empty string when not needed.
- `Sub(label, menu)` adds a submenu, `Separator()` adds a divider, and `SetItemDisabled(label, bool)` disables or enables menu items.
- The menu is modal: the rest of the page does not respond to clicks when opened, focus is limited to the menu, and when opened it focuses on the first available item. Click outside or press Esc to close. After closing, the focus returns to the triggering element.
- keyboard:
  - ↑ ↓ Move between available items, skip disabled items and separators, and loop from beginning to end;
  - Home / End Jump to the first or last item;
  - Enter / Space executes the current item;
  - → Open the submenu, ← or Esc only close the current submenu.
- After executing any of the items, the entire menu (including all submenus) is closed before the action is called.
- The submenu is displayed on the right side by default and flips to the left if it cannot fit.
- `Value()` / `SetValue(bool)` reads or sets whether to open, `Toggle` is used as a click callback to trigger the element, and `Width(dp)` sets the minimum width (default 220).

Agent: The role of the menu container is `menu` (the name of the submenu is its title in the parent menu), the role of the menu item is `menuitem`; items with submenus `value` are `submenu`; disabled items report `disabled`.

Verification: `go run ./examples/components -section menu`.

Long menus are constrained to the window and scroll vertically. The arrow keys and Home / End will scroll the target item into the visible range; if the previous items are disabled when opening, the first available item will also be displayed. When the width is too large, it is constrained by the window width.

`SetDisabled(true)` Closes and disables the entire menu; disabling the expanding parent or submenu closes the corresponding branch. The expanded state is cleared recursively when closing the top level, and reopening does not restore the old deep submenus. `Sub` ignores nil, circular references and repeatedly mounted submenu instances; please create separate instances for different branches.

For right-click triggering, you can pass a custom View to `Trigger` and call `Toggle` in `OnContextMenu` of the element. The keyboard entrance is defined by the `OnKey` of the triggering View. The example uses F10; the elastic layer is still anchored to the trigger area, and the closing behavior of external clicks and Esc is the same.

`Placement(side, align)` Sets the top menu direction (Top/Bottom/Left/Right) and alignment (Start/Center/End), defaults to Bottom/Start; illegal combinations are ignored. `Offset(dp)` sets the spacing, the default is 4dp, supports 0 and negative value overlap, and non-finite values are ignored. It can be updated while it is open. When there is insufficient space, overlay flipping and in-window restrictions will be used. The submenu is still expanded by Right/Start, 2dp.

These two methods of DropdownButton directly configure the incoming Menu; normal mode anchors the whole button, and split mode anchors the arrow. Callers that share the same Menu will also see configuration changes.

`IconItem(label, shortcut, icon, action)` adds the icon command, `SetItemIcon(label, icon)` can update the icons of ordinary items, check items or submenu entries; IconNone is removed. Menus with prefixed tags uniformly reserve the tag column, and the text remains aligned.

`CheckItem(label, shortcut, checked, onChange)` Add check option. Click /Enter/Space to switch the internal state, close the entire menu chain, and then call onChange(bool); disabled items are not switched. `SetItemChecked(label, checked)` does not trigger a callback; `ItemChecked(label)` returns status and whether it is found. The update method applies to all items with the same name in the current menu. The query returns the first matching item. It is recommended to use a unique label.

`CheckSide(el.Left/Right)` sets the check mark position of the current menu; by default, the left side replaces the item icon, and the right side can display the icon at the same time. Unknown directions are ignored, and submenus are configured separately. The role in the Agent is menuitemcheckbox and reports the checked status.

`Label(text)` Inserts a non-interactive group title. Empty text is ignored. The title uses a secondary color and a smaller bold font, has a fixed line height of 30dp, and long text is truncated in a single line; it does not respond to clicks, and will not be the target of arrow keys, Home/End or text search. Long menu positioning counts towards title height. The heading's Agent role is heading; it is just a visual section heading and does not create independent submenus or nested groups.

`ContentItem(label, shortcut, content, action)` Add custom content lines, which can combine multi-line text, descriptions, icons and other display elements; do not nest buttons/input boxes. label remains an accessible name and text search basis. Clicks and Enter/Space in the entire row execute the same action and close the menu chain.

`SetItemContent(label, content)` can replace the display content of the common item, check item or submenu entry of the current menu with the same name; nil restores the text. The row identity remains unchanged, and disabled, checked, shortcut keys, and submenu arrows retain the original configuration. The minimum height of the custom row is 30dp, which increases according to the content; the long menu is positioned according to the actual layout height, and the keyboard focus is transferred after scrolling to the visible position.

`Link(label, url)` adds a link item, retains the menuitem role, and the semantic value is the URL; after clicking or typing with the keyboard, first close the entire menu chain, and then open the link. By default the platform browser/mail handler is called via `core.OpenURL`, which only accepts absolute HTTP, HTTPS and mailto URLs. `SetItemIcon` can add the front icon, and `ExternalLinkIcon(false)` can hide the external link icon on the right.

`OnLink(func(string))` can take over the opening behavior and URL strategy. The submenu uses its own callback first, otherwise it looks along the parent menu. By default, `OnLinkError(func(error))` notification is used when opening fails, and the search is also performed along the parent menu; errors are ignored when not configured. The default entry only reports a checksum process startup error and does not indicate that the web page is loaded successfully. Web uses window.open, which is restricted by browser pop-up policy.


## Show shortcut keys by target area

`ActionItem` The key is resolved based on the trigger's `KeyContext` ancestor, and the submenu inherits the top-level target. You can also use `ActionContext("editor")` to specify an element ID to display the key position of the area where the element is located; no prompt will be displayed when the specified target does not exist, is hidden, or is disabled. An empty string restores the trigger target. The prompt is generated after the current frame is constructed and before measurement, so there is no need to wait for one more frame when opening for the first time or switching contexts in the same frame.

```go
core.Bind("save", "mod+s")
core.BindIn("editor", "save", "mod+shift+s")
menu := kit.Menu().ActionContext("document").ActionItem("Save", "save", save)
// In Render:
cx.ActionAt("document", "save", save)
return el.Div().ID("document").KeyContext("editor").Child(input.Render(cx), menu.Render(cx))
```

`core.BindIn` Sets area coverage for an action. The order of priority is inner layer, outer layer, and global. Not passing the key bit means disabling the binding of the action in this area; `ClearBindingIn` deletes the override and restores inheritance. Invalid key positions will not change the original binding, and changing the binding will require all windows to be refreshed. `Keymap/LoadKeymap` still only reads and writes global bindings, regional bindings are configured by the application.

### Context conditional expression

The first parameter of `BindIn` can be a conditional expression, which is consistent with the GPUI keymap writing method:

```go
core.BindIn("Editor && !ReadOnly", "format", "mod+shift+f") // editable editor
core.BindIn("Pane > Editor", "close", "mod+w")               // Editor in panel
core.BindIn("Terminal || Shell", "clear", "mod+k")          // one of the two
```

- Each level on the focus path is a `el.KeyContext`, and the name can contain multiple identifiers separated by spaces, such as `KeyContext("Editor ReadOnly")`.
- Identifies the layer that checks the current matching; `a > b` means that this layer satisfies b and an outer layer satisfies a. Priority from high to low: `!`, `>`, `&&`, `||`, grouped by brackets.
- When parsing, looking from the innermost layer to the outside, the first layer with established binding will win; when there are multiple establishments in the same layer, the later binding will take effect. If neither is true, use global binding.
- The incorrectly written expression `BindIn` returns an error and changes nothing. A single name is an expression with only one identifier, and the original writing method still takes effect.

`cx.ActionAt` Handles resolved keystrokes only when there is focus within the specified element. Among nested targets, deeper targets take precedence, and the same level is in the order of declaration; explicit null binding in the inner level will also prevent the same action in the outer level from being processed. Do not simultaneously register the same action with global `cx.Action`. When a function is passed to `ActionItem`, the click only calls this function; when nil is passed, the click is routed according to the action name: from the trigger (or the element specified by `ActionContext`), look out for the innermost `cx.ActionAt` processor. If the global processor of `cx.Action` cannot be found, the effect is the same as pressing the shortcut key there, and there is no need to bind any keys. In this way, menus, shortcut keys and command panels can be implemented using one command. Routing does not switch focus. The bottom layer is `cx.Perform(targetID, action)`, and the command panel can also be called directly. Explicit shortcut keys for normal `Item` and standalone `KbdFor` retain their original behavior. The underlying `el.KeyHint` can be used to display content with the same rules and is not responsible for registering shortcut keys.

## Scrollbar visibility

Use `kit.Menu().Scrollbars(el.ScrollbarAlways)` to keep the scrollbar visible
whenever the menu overflows. Other modes are `ScrollbarHover` (pointer inside
or dragging), `ScrollbarScrolling` (during scrolling and briefly afterward),
and `ScrollbarSystem` (platform preference). Without an explicit setting,
the global `el.SetScrollbarDefault` applies. Submenus inherit the nearest
parent's explicit mode and can override it. Invalid values are ignored.
Short menus have symmetric row margins; overflowing menus reserve a scrollbar
hit area even while the bar fades out, so rows do not jump during scrolling.
The long-menu component example uses `ScrollbarAlways`.
