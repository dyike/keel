# Tabs

English | [简体中文](tabs.zh-CN.md)

Tabs with support for underline, capsule, stroke and segmented appearances.

```go
tabs := kit.Tabs().Add("General", basicForm).Add("Notifications", notifySettings).OnChange(onTab)
```

- `Variant(TabsUnderline/TabsPill/TabsOutline/TabsSegmented)` Select the appearance, retain the original underline style by default. Colors switch with the theme.
- `AddItem(TabItem{Title, Page, Icon, Content, Disabled})` Add an icon or rich label. Content replaces the visible title and should be a display element; Title is still used for the Agent name and overflow menu. Rich tags are laid out within a limited width, and text truncation requires the content to set MaxLines by itself.
- `SetItem(i, item)` updates the entry and preserves stable identity; `SetItemDisabled(i, bool)` disables a single entry. Disabled items cannot be clicked, closed, or dragged, and arrow keys and overflow menus skip them. Disable the current item to automatically select the next available item; when all are disabled, the current page is retained, the tab does not enter Tab navigation, and the page content remains available. Restore selection after re-enabling an entry. Program modification does not trigger OnChange; SetValue does not select disabled items.
- Only render the current page. Each page is a View held by the application itself, and the state remains after switching.
- The Tab key focuses on the current tab, ← → switches, cycles through the beginning and end, and Home / End jumps to the beginning and end.
- When the labels cannot be placed, the extra ones will be put into the "More" menu at the end. The currently selected label will always remain in the column, and you can continue to use the arrow keys to operate it; long titles will be truncated. Like Toolbar, tabs should be placed in a position with width constraints.
- `Closable(fn)` adds a close button to each label, `fn(i)` determines the meaning of close, and usually passes `tabs.Remove` directly. The close button is next to the label but not inside the label, so clicking it will not switch to this label first.
- `Reorderable(fn)` allows dragging the visible labels to rearrange them and submitting after releasing the pointer; canceling the drag does not change the order. `Move(from, to)` can programmatically rearrange any label without triggering a callback; the page and editing status follow the stable identity, and the selected page remains unchanged. Dragging does not automatically open the overflow menu.
- Delete calls the close callback when the tab is focused. After closing the current page, the focus moves to the remaining adjacent tabs; closing other pages keeps the current page.
- `Leading(el.View)` / `Trailing(el.View)` places fixed left and right areas; `Size(dp)` sets the label height, default 40, minimum 24. `SetDisabled` Disables tabs, close buttons, dragging, and the current page.
- `MaxWidth(dp)` limits the width of a single label area (including icons and padding, excluding the close button next to it); 0 cancels the upper limit, and negative and non-finite values are ignored. Text is automatically omitted by default, and rich content is truncated by itself.
- `Scrollable(true)` Changed the tab bar to horizontal scrolling, all tabs are kept in the track, and are no longer included in the overflow menu; Fixed Leading/Trailing not to participate in scrolling. Automatically scroll into the viewport after the selected item changes, and the keyboard focus will be transferred after the scrolling is completed. Switching back to false restores the original overflow menu behavior.
- `ScrollTo(i)` requests to display an item without changing the selected item or triggering a callback; it can be called before the first layout and requests to follow the label identity. `ScrollState(cx)` returns the horizontal offset, viewport width, and content width of the most recent drawing, in dp. Manual scrolling will not be forced back by the currently selected item.
- `Value()` / `SetValue(i)` (no callback is triggered), `Remove(i)`, `Len()`.

Agent: The label bar is `tablist`, each label is `tab`, `selected` represents the current label; the current page is `tabpanel` named after the label title.

Verify: `go run ./examples/components -section tabs`, add `-theme dark` to check the dark theme.
