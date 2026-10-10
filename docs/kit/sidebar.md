# Sidebar

English | [简体中文](sidebar.zh-CN.md)

The application's navigation sidebar: groups, selected items, and icons can be collapsed to only display icons.

```go
nav := kit.Sidebar().
    Section("Workspace", kit.SidebarItem{ID: "inbox", Label: "Inbox", Icon: kit.IconInbox, Badge: 12}).
    Section("", kit.SidebarItem{ID: "settings", Label: "Settings", Icon: kit.IconSettings}).
    OnChange(navigate)
```

- Mouse click only displays the selected background; Tab and arrow key navigation display the focus box. The scroll bar occupies an independent space on the right side and does not cover the row background, corner mark or click area.
- 36dp navigation line, 16dp icon, 14sp copy; selected items use neutral background color and accent icon, count is low contrast number, and folded copy is displayed in the expanded state at the bottom.
- Titled groups display the title; untitled groups are separated by separators.
- Each available item can be focused using Tab, ↑ ↓ / Home / End / PageUp / PageDown to move and scroll to the target, and Enter or Space to activate. `Disabled` or `SetItemDisabled(id, bool)` disables a single item; `SetDisabled` disables the entire sidebar and slot operations.
- `SidebarItem.Children` defines nested navigation. Click to expand/collapse the entire row of a branch without triggering the navigation callback; → expand, ← collapse or move to the parent. `SetExpanded` / `Expanded` manage the expansion state, `SetValue` expands the target ancestor and scrolls to the selection.
- `Header(el.View)` / `Footer(el.View)` are located outside the navigation scroll area and are suitable for workspace switching and account information. The slot can read `Collapsed()` in Render and switch the icon version content by itself.
- `Height(dp)` explicitly sets the height of the sidebar. When the navigation content overflows, only the middle area is scrolled. The height is set according to the content by default, and the upper limit is the window height. The height returned by `cx.ViewportSize()` can be passed when embedded in the application shell.
- The bottom button can collapse the sidebar, and the width after collapse is 56dp: only the icon is displayed, the name is changed to Tooltip, and the corner icon becomes a dot.
- `Icon` can be omitted (`IconNone`): only the text will be displayed when expanded, and the first letter of the name will be displayed when collapsed.
- `SidebarItem.IconView` accepts a display-only custom icon, overriding `Icon` in the same 16dp slot. It remains visible when collapsed; the application controls its colors. Branch disclosure arrows sit immediately after the label, before badges and independent suffixes.
- `Filter(query)` Displays only items whose names contain query (case-insensitive), and the parents leading to them, which are temporarily expanded; groups without matching items are hidden along with their titles. Pass an empty string to restore all. Selected items remain selected while being filtered out. The component library application `go run ./examples/components` uses it for searching.
- `Side(el.Right)` Move the divider to the left, and adjust the folding arrow and the prompt direction of the collapsed state; the default is `el.Left`. The application still needs to place the sidebar to the right of the main content, and the component does not change the parent layout order. `BorderWidth(dp)` adjusts the divider, 0 hides it.
- `SidebarItem.Tag` places display-only content, usually a `Tag`, right after the label and before the disclosure arrow. It keeps its size while the label truncates, keeps regular weight on a selected row, and is hidden when collapsed.
- `SectionAction(title, view, items...)` adds trailing content to a section heading, such as an add button; such a section keeps its heading while empty (hidden while filtering). `SectionTitleWeight(font.Weight)` sets the heading weight, regular by default.
- `Collapsible(false)` hides the built-in collapse button; programs can still call `SetCollapsed`.
- `SidebarItem.Suffix` / `SetSuffix(id, view)` Add independent tail content, you can put buttons or switches; click the tail to not select or expand the navigation items. When collapsed, the tail content is hidden. When expanded, the maximum width is half the configured width of the sidebar, and the height should fit within the 36dp row. Badge can still be displayed at the same time.
- `SidebarItem.ContextMenu` / `SetContextMenu(id, menu)` adds a right-click menu to the item, supporting Menu's submenus, shortcut key tips, links and custom content. A Menu instance belongs to only one item, and its Trigger is not used; the menu appearance and position are still configured by Menu. Hiding, disabling or replacing an entry menu closes the old menu. Right-clicking does not change the selected item.
- `Value()` / `SetValue(id)`, `SetBadge(id, n)`, `Collapsed()` / `SetCollapsed`, `Width(dp)` (default 220).

Options structures are copied recursively; IconView, Suffix and ContextMenu retain the passed-in instance. The ID of the entire Sidebar must be non-empty and unique; when the new Section contains empty values or duplicate IDs, the entire group will not be added. Parent disabling does not implicitly disable children.

Agent: container role `navigation`; the leaf item is `link`, the branch is `button` with an expanded Boolean value, `selected` represents the current page; the collapse button is named "Collapse Sidebar" and "Expand Sidebar".

Verify: `go run ./examples/components -section sidebar`, add `-theme dark` to check the dark theme.
