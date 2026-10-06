# List

English | [简体中文](list.zh-CN.md)

The default single-select virtual list supports multi-selection, single-item disabling, stable ID, and drag-and-drop rearrangement.

```go
contacts := kit.List(names...).Height(240).OnChange(show).OnActivate(open)
```

- Click or ↑ ↓ Home End PageUp PageDown to select, double-click or press Enter to activate.
- The focus is on the entire list, not on a certain row: the row will be recycled with scrolling, and the focus cannot stay on a certain row.
- `Value()` returns the serial number of the selected item (-1 if none), `SetValue` does not trigger the callback; `SetItems` replaces the content, and clears it when the original selection does not exist; `Items()`, `SetDisabled`.
- `Height(dp)` or `Fill()` sets the height. `Plain()` removes the border and background and is used in panels and sidebars that already have borders; the outline is still displayed when focused.

Agent: container role `listbox`, each item is `option`, `selected` means selected.

Verify: `go run ./examples/components -section list`, add `-theme dark` to check the dark theme.

Constructs and `SetItems` copies the option slice, `Items()` returns a copy. Update data via `SetItems`; modifying the slice passed in or returned does not affect the list. The selection is still represented by the index. After replacement, the index will be cleared when it goes out of bounds, and the user callback will not be triggered.

`SetEntries(...ListItem)` receives `{ID, Label, Disabled}`, copies the data and keeps the selection by ID; empty ID, duplicate ID panics before modification. `Entries()` returns a copy. `SetItems` still uses index identity, use `SetEntries` when you need to maintain selections across insertions, deletions, and rearrangements. `SetItemDisabled(index, on)` disables a single item. The mouse, arrow keys, range selection and select all will skip it, and pressing Enter will not activate it. Program selection allows disabled items to be retained.

`MultiSelect()` Enables Ctrl/Cmd plus selection, Shift range selection, and Ctrl/Cmd+A selection all. `SelectedValues()` returns a copy of the index in display order, `SetSelectedValues` is assigned programmatically, `OnSelectionChange` receives a copy of the selection; `Value()` is the active item, and `SetValue` is replaced with a single selection. Programmatic assignment does not trigger user callbacks.

`Reorderable(func(from, to int))` enables drag rearrangement. The internal order will be modified and notified after releasing the button. Cancel dragging without modifying the order. `Move(from, to)` provides program rearrangement without triggering callbacks. Both methods retain the selection by ID, and the callback index refers to the position before and after the move respectively; the current drag is oriented to the target within the viewport and does not automatically scroll to distant items.

## Group, search and customize content

`ListItem` can set Group, Keywords and Icon. Adjacent items with the same Group form a group. The group title occupies one line but cannot be selected; the group header is not displayed when there are no search results in the group. `RenderGroupHeader(func(cx, group) el.Element)` replaces the header content and `RenderGroupFooter` inserts a line at the end of each non-empty group. All virtual rows (including group headers/footers) use the same row height. `RowHeight(dp)` can be set to 20–512dp, and the default is 32dp.

`Searchable(true)` displays the search box, which is not displayed by default; `SetQuery` / `Query` controls the query. The search is case-insensitive and matches Label and Keywords, but not group titles. Keywords are copied during SetEntries, and the keywords returned by Entries are also copies. `OnSearch` receives query changes and is notified when the program calls SetQuery. The application can obtain remote data accordingly; local filtering is always in effect.

Searches are not renumbered: Value, select/activate callbacks, and Index in custom row contexts are all source data indexes for SetEntries. Hidden selections are retained. Keyboard navigation, range selection, and select all only work on visible available items. The hidden current item cannot be activated by Enter. Drag to position the target according to the visible layout including the group header/footer, and then update the source order; the items retain the original Group, and changes in adjacent relationships may cause groups with the same name to appear in segments.

```go
list.Searchable(true).RenderItem(func(cx *el.Context, row kit.ListItemContext) el.Element {
    return el.Div().Row().Grow().Child(
        el.Text(row.Item.Label).Grow(),
        kit.Button("详情", func() { open(row.Item.ID) }).Size(24).Render(cx),
    )
})
```

`ListItemContext` with entry copy, source Index, Selected and Disabled status. RenderItem returns nil to use the default icon/label. Sub-buttons with custom content handle events independently, without selecting/activating list rows; the rest of the row's background still supports original selection and dragging. Stateful instances of custom views should be saved by the application with a stable ID, and the content height should fit into the uniform row height.

## Load more

`OnLoadMore(fn)` works with `SetHasMore(true)` to request the next page within two lines from the bottom. The component is set to loading first and then fn is called. The same data will not be automatically requested repeatedly. The results are delivered through SetEntries in the UI thread or core.Update, then SetLoading(false); the last page is set by SetHasMore(false). Appending data will not automatically pull the viewport back to the old selection.

If failed, use `SetLoadError(message)` to stop the automatic request and display the retry button; click it and then call the loading function. When disabled, no requests are initiated and issued network tasks are canceled by the application. Expiration verification of remote search or loaded results is also managed by the application, and the query/request identifier should be checked before delivery; the component does not provide an asynchronous token API. SetQuery and SetEntries reset the current round of request flags, and the filtered content can continue to load if it is less than one screen.

## Status content

```go
files := kit.List().Searchable(true).
    InitialContent(kit.Label("输入文件名开始搜索")).
    EmptyContent(kit.Label("这个文件夹是空的")).
    NoMatchesContent(kit.Label("没有找到匹配的文件")).
    LoadingContent(kit.Skeleton()).
    ErrorContent(func(msg string, retry func()) el.View { /* 自定义错误和重试 */ })
```

- `EmptyContent`: Displayed when the list has no data and no search terms.
- `NoMatchesContent`: Displayed when there is a search term but no results.
- `InitialContent`: When the searchable list has not yet entered a search term, it will be displayed instead of the entire list, and the results will be displayed after input.
- `LoadingContent`, `ErrorContent`: replace the circle when loading more and the text when an error occurs and add "retry"; `retry` reinitiates loading.
- Pass nil to restore default. The status content is displayed below the list, in the same position as the default text.

Example `go run ./examples/components -section list` demonstrates search, grouping, icons, and inline operations. The automatic test covers the original selection/drag regression, as well as search source index, group footer, range skip hidden items, sub-button isolation, load deduplication/failure retry/disable; real device vision and drag acceptance are not done.
