# Command

English | [简体中文](command.zh-CN.md)

Command panel: a search box plus a list of commands, which are filtered step by step as you type.

```go
palette := kit.Command(
    kit.CommandItem{Title: "New order", Group: "Orders", Shortcut: "mod+n", Action: newOrder},
    kit.CommandItem{Title: "Open settings", Action: openSettings},
)
// In Render:
cx.Shortcut("mod+k", palette.Toggle)
root.Child(palette.Render(cx))
```

- Matching rules: Prefix matching ranks first, followed by substring matching, and then fuzzy matching of "characters appearing in order". For example, "Settings" can find "Open Settings", and "nwo" can find "New window".
- ↑ ↓ Move the highlight and press Enter to execute. Esc first clears the non-empty query, then press it again to close; click outside to close it directly. By default, the modal panel is closed first and then the Action is called.
- When opened, focus is on the search box. The panel is modal and appears near the top of the window.
- `Shortcut` is the fixed display text. After setting `ActionName`, the default line reads the shortcut key from `core.Bind` and binds the key to the entry's Action while the panel holds focus; changing the binding will update the prompt and handler synchronously. The application needs to register the same Action outside the panel.
- `Toggle`, `Value()` / `SetValue(bool)`, `SetItems`. `el.Root` REQUIRED.

Agent: The panel is `dialog` named "Command Panel", the search box is `textbox`, the command is `option`, and `selected` indicates the current highlight.

Verify: `go run ./examples/components -section command`, add `-theme dark` to check the dark theme.

The group title is displayed before the continuous group command; the title cannot be selected. `CommandItem.Disabled` disables clicks and execution, and keyboard navigation skips disabled items. Construct and `SetItems` copy the slice, applying modifications to the original slice will not change the panel. Results build visible rows with a virtual list, search results are cached when items or queries don't change; ↑ ↓ and PageUp/PageDown navigate and scroll to reveal highlights, and Home/End retains the text editing behavior of the search box. After closing, the focus returns to the trigger control, and reopening resets the query, highlight and scroll position. `SetDisabled(true)` Closes and prevents opening.

`OnSearch(func(query string, token uint64))` switches to asynchronous search mode. Opening, changing words and retrying will generate new tokens. After the worker thread obtains the result, it calls `SetResults(token, items...)` or `SetSearchError(token, message)` through `core.Update`; returning false indicates that the result has expired or the panel has been closed. The old query will not overwrite the new query, and you can try again after failure. Asynchronous results are displayed in the order returned, no local fuzzy filtering is performed, and are suitable for semantic search. Old commands cannot be executed in loading or failed state. Example `-section command_async` input error simulation failed; static `-section command` contains thousands of results and disabled items.

`Inline(true)` puts the panel into a normal layout and opens it; it remains displayed after execution, does not block external controls, and does not automatically grab focus. `palette.Focus(cx)` is called when active entry is required. `SetValue(false)` or press Esc during empty query to hide the inline content; `Inline(false)` closes the current content and restores the default modal mode, and the pop-up layer will not be displayed until the next time it is opened. Switching will invalidate old asynchronous requests.

`Searchable(false)` hides the search box, displays all candidates, and stops calling OnSearch; the keyboard focus moves to the panel box, the arrow keys navigate, and Enter executes. Clear queries, loading state, and errors when switching, and old results will no longer be received; re-request when search is resumed and the panel is open. The default search behavior is retained.

`Header(view)` and `Footer(view)` are placed above the search box and below the results respectively. They are still displayed when loading, failure and empty results; nil is removed. Each area occupies at most one-fifth of the window height and does not exceed 80dp, beyond independent scrolling; this part of the height is conservatively reserved for the result area. `Empty(view)` replaces no matching content, nil returns to default. Interactive controls should reuse instances to preserve focus and internal state.

`RenderItem(func(CommandItem, bool) el.View)` Customize the visible candidates. The second parameter is the current highlighting status. It replaces text and shortcut key display, and the outer layer retains accessible names, disabled and selected semantics; nil callbacks or nil content use the default line. Click the displayed content to execute the command, and the nested button handles the operation independently. `RowHeight(dp)` specifies the unified virtual slot height, including a total of 4dp of white space at the top and bottom. The group titles share this height; 0 returns to 36dp, the minimum positive number is 5dp, and negative and non-finite values are ignored. The default is still to use uniform row height; for complex content, AutoRowHeight(true) can be turned on and measured according to the actual content.

```go
quick := kit.Command(items...).Searchable(false).Inline(true).
    Footer(kit.Button("Refresh", refresh))
// Put in a normal layout; the application can use quick.Focus(cx) in the event to give focus to the panel.
root.Child(quick.Render(cx))
```

Different from the upstream implementation agreement: Keel defaults to uniform row height, and the automatic measurement mode first estimates unvisited rows and then measures visible rows; the upstream measures all rows when it fails. Keel preserves fuzzy ordering rather than pure substring filtering, and events use flattened raw indexing rather than IndexPath. Inline mode can put the application's own elastic layer, and the closing of the outer elastic layer is managed by the application.


`CommandItem.Keywords` provides search aliases; the title and each keyword are fuzzy matched separately, sorted by the best score. Constructs, SetItems, and asynchronous SetResults all copy keyword slices; filtering does not change the original index of the item, and group headers do not occupy the index; explicitly delimited items occupy the source data position, but cannot be selected. After updating the entire candidate model, the index is based on the parameter order of the latest SetItems.

- `OnSelect(func(int))`: Notifies the original index when the highlight changes due to keyboard movement, clicks and filtering, −1 when there is no option. Notification will only occur if selection changes, no action will be executed. Automatic selections resulting from opening and model updates are sent after rendering; components can be updated in callbacks. The pointer entering an available row will also change the highlight and notify, but no action will be performed; a stationary pointer will not repeatedly overwrite the keyboard selection.
- `OnQuery(func(string))`: Notified when user input or Esc clears words, retains local filtering; filtering caused by OnSelect is sent before it. Open and retry without notification. OnSearch is still responsible for remote requests, triggering when opening, changing words, and retrying, and its purpose is different from OnQuery.
- `OnConfirm(func(int))`: Notify the original index after executing the Action, and also notify the entries without Action. Callbacks and indexes take snapshots before execution. Action reset candidates, open panels, or replace callbacks will not overwrite this confirmation.
- `OnCancel(func())`: Notified after the user closes it; no notification will be given when the program assigns a value or disables it. When searchable and the query is not empty, Esc will only clear the word for the first time; when there is no search mode or empty query, Esc will close and notify, and external clicks will close it directly.

When an entry is clicked, the selection change is notified first; if OnSelect replaces the candidate, closes the panel, or initiates a new query in the callback, the old entry will not continue to be executed this time. Custom rows maintain their identity before and after filtering with the original index; when replacing/rearranging the entire model, the application still needs to manage the child views it holds. All the above events can be removed by passing nil.


`AutoRowHeight(true)` Enables variable height virtualization, RowHeight becomes the minimum slot height and initial estimate. Default candidates, group titles, long text or complex custom content can be mixed. Window width, scaling, themes, and SetItems/RenderItem updates invalidate measurements; `InvalidateRows()` is called when the app changes off-screen content on its own. Only rows near the viewport are built, distant jumps correct positioning after measurement. `AutoRowHeight(false)` restores uniform row height and re-exposes the current highlight.

`CommandItem{Separator: true}` Inserts a non-interactive delimiter and ignores other fields. After filtering, the first and last and consecutive separated items are removed, and the empty group titles disappear at the same time. Sorting is performed between separated items to avoid cross-section mixing. `Icon` specifies the default row front icon, `Checked` displays the trailing check; valid shortcut key tips take precedence over the check. Custom rows draw these themselves.

`ActionName` Gets the current first set of bindings via `core.Bind`; does not fall back to Shortcut when name exists. Hide the prompt when not bound, and can fall back to displaying Checked. Shortcut keys only execute entries when the panel is available and focused within the panel, also respecting load/disable and confirm callbacks; closing or out-of-focus are not registered. If the same action is required outside the application, use the same Action function to register. Keel uses a single root registration sequence to handle conflicts, without upstream Command/application two-level action resolution.

Other configurations and status:

- `Placeholder(string)`: Search placeholder text, empty value restores localization default.
- `MaxHeight(dp)`: Result viewport maximum height, 0 restores 360dp; window space and headers and footers may shrink it further.
- `Bordered(false)`: Remove the default borders, rounded corners and shadows; `PanelStyle(func(*el.DivEl))` can adjust the width, background and other styles of the new panel in each frame.
- `Query()` / `SetQuery(string)`: read/set query; process and notify according to user input when it is opened and searchable, and the same text will not be notified repeatedly. Only saves when closed or not searchable; re-opening still resets the query.
- `SelectedIndex()`: Index of the currently highlighted original entry, −1 means none; data updates are reconciled on the next render.
- `MatchedCount()`: The current number of matches, including disabled items, excluding headers/separators.
- `SetLoading(bool)` / `IsLoading()`: The application manages the loading state; usually automatically managed by the token result interface when using OnSearch.

```go
core.Bind("orders.new", "mod+n")
commands := kit.Command(
    kit.CommandItem{Title: "New order", ActionName: "orders.new", Icon: kit.IconPlus, Action: newOrder},
    kit.CommandItem{Separator: true},
    kit.CommandItem{Title: "Current mode", Checked: true},
).AutoRowHeight(true).Placeholder("Search actions").MaxHeight(280)
```
