# Table

English | [简体中文](table.zh-CN.md)

Data tables: sort, select, adjust column widths, customize cells, build only visible rows.

```go
t := kit.Table(kit.Col("单号").Width(120), kit.Col("客户").Flex(2), kit.Col("金额").Numeric()).
    Height(360).OnActivate(open)
t.SetRows(rows)
```

- List:
  - `Col(title)` defaults to equal width; `Flex(w)` distributes proportionally, `Width(dp)` has fixed width;
  - `Numeric()` right-justifies and sorts numerically; `NoSort()` disables sorting by this column;
  - `Cell(fn)` uses custom elements to render cells (such as labels, buttons), and the Agent still sees the row text.
- Click the header to sort, and click again to reverse; `SortBy(col, desc)` is used to sort, and when `col` is -1, the original order is restored.
- Drag the right edge of the table header to adjust the column width. After adjustment, the column becomes a fixed width. Drag the main body of the header horizontally to change the order. Point to the front half of the target column and insert it in front of it, and insert the second half behind it. The insertion position displays a theme-colored thin line. Release the application, move out of the header area or cancel the gesture to give up. Dragging does not trigger sorting. Stopping within 32dp of the left and right edges of the non-frozen header area will continue to scroll horizontally, and the closer to the edge, the faster the speed (up to 360dp/s); the frozen column area does not trigger scrolling. Stops when you release or cancel, and the insertion position is updated during scrolling.
- Click or use ↑ ↓ Home End PageUp PageDown to select, double-click or press Enter to activate.
- **The row numbers in callbacks, `Value`, and `SetValue` are the positions in the `SetRows` data and are not affected by sorting.**
- `SetValue(i)` will scroll the selected row to a visible position. When the table is not on the screen, it will scroll again after being displayed.
- `SetLoading(true)` displays a loading animation on the line and is suitable for asynchronous data retrieval: use `core.Update` to call `SetRows` and `SetLoading(false)` after the data arrives. `Empty(text)` sets the text when there is no data. By default, the locale's "no data" is used.
- When the width of the viewport is exceeded, use the trackpad or horizontal scroll wheel to scroll horizontally; the table header and data move synchronously, and the vertical direction is still constructed according to visible rows. Fixed-width columns retain the set width, and flexible columns fill the remaining space proportionally, with at least 40dp per column; scroll when the window is insufficient.
- The scroll range is recalculated after resizing the column width or window size. Freeze columns are configured via `FrozenColumns`.

Agent: role `table`, `value` is the number of rows (such as "36 rows"); the header is `columnheader`; each row is `row`, and the name is the column connected with "|", `selected` means selected.

Verify: `go run ./examples/components -section table`, add `-theme dark` to check the dark theme.

Copy column configuration when constructing the table, `SetRows` copies two-dimensional data, `Rows` / `Row` returns a copy. Reuse the same column configuration to create multiple tables, and dragging column widths will not affect each other. Subsequently use `SetRows` to update data, and use `SetColumnWidth(index, dp)` to modify the column width of a certain table; modifying `ColumnSpec` during original slicing or construction will not affect the created table.

When sorting, the virtual row uses the original data index as the element key, and the input/focus state in the custom cell that is still within the construction range moves with the data row. Identities are still interpreted according to the new data index after replacing the entire dataset; business state across datasets should be preserved in the application model.

`FrozenColumns(left, right)` fixes the starting left column and the last right column, and the middle column scrolls horizontally; the table header, data cell, drag column width and click area remain consistent. The count will be limited to the number of columns, with left priority, and repeated calls can adjust or cancel the freeze. The frozen column needs to have a fixed width, and the original flexible column will be converted to 120dp, which can be modified later with `SetColumnWidth` or dragging. When the narrow viewport cannot fit on both sides, the left side is given priority and the right side is cropped; when the middle column has no available width, it does not draw or respond to clicks. Vertical still uses the same virtual list and does not copy data rows.

Column management uses the source column index (the position when `Table` is constructed), and moving or hiding does not change this index:

- `MoveColumn(column, position)` Moves the source column to position in the display order, and the position is included in the hidden column; out-of-bounds ignore.
- `OnColumnMove(func(column, from, to int))` is called back after the user drags successfully; column is the source column index, and from/to is the display position of hidden columns. The procedure MoveColumn/SetLayoutState does not fire.
- `SetColumnVisible(column, visible)` Shows or hides the source column, preserving its position and width. Allows all columns to be hidden. Hiding sort columns will not cancel sorting.
- `LayoutState()` returns an independent `TableLayout` snapshot, including column order, width, elastic scale, hidden state and number of frozen columns on both sides, which can be saved to the application configuration with `encoding/json`.
- `SetLayoutState(state)` first checks the number of columns, unique index, limited width and proportion, and then restores it as a whole; invalid configuration returns an error and retains the original layout. Layouts only apply to the same source column structure, and applications should manage configuration versions along with the business data structure.

The frozen quantity acts on both ends of the currently visible order; the two sides do not overlap after the column is hidden, and the left side takes precedence. When moving columns and hiding other columns, custom cells that are still being built retain element identity; the hidden cells themselves are unloaded and persistent drafts still need to be saved by the app model. Column operations do not change the source data, sorting, or row selection, nor do they trigger row selection callbacks. The sample provides move, hide, save, and restore buttons.

`MultiSelect()` Enable row multi-selection: normal click to replace the selection, Ctrl/Cmd click to increase or decrease a single row, Shift click or Shift+arrow key to expand/reduce the range according to the current sorting, Ctrl/Cmd+Shift click to merge the range. `SelectedRows()` returns a copy of the source row index in the current display order, `Value()` represents the active row (which may have been deselected). `SetValue` replaces a single-line selection, `SetSelectedRows` replaces a multi-line selection, and no program operation triggers a callback; `OnSelectionChange` receives an independent copy. When data is shortened, out-of-bounds selections are removed, and sorting does not change the selected source rows.

When the table is focused, Ctrl/Cmd+A selects all rows of the multi-select table, and Ctrl/Cmd+C copies the selection. `SelectionText()` returns the same TSV content: only the currently visible columns are included, columns and rows are output in display order, and tabs, newlines, and double quotes within cells are escaped according to CSV quoting rules. The input box handles its text selection and copying on its own.

`CellSelect()` switches to cell mode and clears the selection, `MultiSelect()` switches back to row mode. Normally click to select a single cell, Ctrl/Cmd to click to increase or decrease a single cell, Shift to click or the arrow keys to expand the rectangular range; arrow keys to move, Home/End to jump to the beginning/end of the line, Ctrl/Cmd+Home/End to jump to the beginning/end of the entire table, PageUp/PageDown to move eight lines and scroll to reveal the target cell. Frozen columns remain fixed.

In this mode, click the table header to select the entire column, Shift-click to select multiple consecutive columns, Ctrl/Cmd click to increase or decrease the entire column, and double-click the table header to sort. `TableCell{Row, Column}` uses the source data index, `SelectedCells()` returns a copy in display order, and `SetSelectedCells` programmatic assignment does not trigger `OnCellSelectionChange`. `SetSelectedRows` selects all visible cells in the specified row; `SetValue` selects the first visible cell in the row. Hidden columns retain the selection, and only the visible columns participating in the selection will be output when copying; unselected intersecting cells in the sparse selection will output null values. Clean out-of-bounds selections after modifying the source data length.

`SetFilter(func(row []string) bool)` filters source data before sorting, incoming rows are copies; `nil` clears filtering. This method needs to be called again after the filtering conditions change. `Len()` is the number of source rows, `VisibleLen()` is the number of rows after filtering; the source data index and selection are retained, and the filtered rows will not participate in copying or selecting all.

Use `OnLoadMore(fn)` and `SetHasMore(true)` for paging loading: automatic request within two lines from the bottom, the component is set to loading before callback, and the same data can be automatically requested at most once. Asynchronous results are delivered via `core.Update`: `SetRows` updates all loaded data, `SetLoading(false)` ends the request, and then `SetHasMore(false)` for the last page. If the call to `SetLoadError(message)` fails, automatic retry will be stopped. The user will click Retry and call `OnLoadMore` again. Does not load automatically when disabled or hidden. After filtering, the content will continue to be paginated if it is less than one screen, and the application must mark the last page correctly. Example `go run ./examples/components -section table_data` simulates a first load failure, retries, and three pages of data; the status tip is always in the viewport.

`RowMenu(func(row int) *MenuView)` / `CellMenu(func(row, column int) *MenuView)` Constructs the menu on right-click request, the parameter is always the source data index. The cell menu takes precedence. If nil is returned, the row menu is used; the menu pops up against the target cell. Right-clicking the selected row/cell will retain the existing multi-selection, and the unselected targets will become the selection first. Shift+F10 opens the menu for the active row/cell; Esc, external click or command execution closes, and the original focus is restored after the keyboard is opened. Close menu when target is filtered, hidden or disabled. Menu reuse `Menu`'s submenus and keyboard navigation.

Performance: It takes less than a second to create, sort, and jump to the end of a table with 300,000 rows in the test, and scrolling takes about 1 millisecond per frame (`TestTableThreeHundredThousandRows`, `BenchmarkTableFrame300k`).

## Static combination table

Use `StaticTable` for small amounts of data or custom layouts. It returns a `el.DivEl` that can be styled directly, without creating the selection, sorting, or virtual list state of the data table. Each component is constructed in Render; stateful Input, Button and other views are held by the application, and the Render results are put into cells.

```go
kit.StaticTable().Name("订单").Child(
    kit.TableHeader().Child(kit.TableRow().Child(
        kit.TableHead().Child(el.Text("单号")),
        kit.TableHead().Items(el.End).Child(el.Text("金额")),
    )),
    kit.TableBody().Child(kit.TableRow().Child(
        kit.TableDataCell().Child(el.Text("SO-001")),
        kit.TableDataCell().Items(el.End).Child(el.Text("¥250.00")),
    )),
    kit.TableFooter().Child(kit.TableRow().Decorate(nil).Child(
        kit.TableDataCell().Child(el.Text("合计")),
        kit.TableDataCell().Items(el.End).Child(el.Text("¥250.00")),
    )),
    kit.TableCaption().Child(el.Text("最近一笔订单")),
)
```

Header, Body, and Footer can all contain any number of Rows; each Row can contain any number of cells, and fully customized elements can also be inserted. Header/Footer uses a weak background, Row draws the bottom divider by default, and Footer draws the top line by default; `Decorate(nil)` removes the corresponding default divider. The root container has borders, rounded corners, and surface background by default, which can be overridden via Border/Rounded/Bg. Caption accepts text or rich content and is displayed within the container by default. The position is determined by the Child order.

By default, the cell divides the row width equally; `Flex(0).W(el.Dp(120)).NoShrink()` fixes the column width, and `Flex(2)` adjusts the flexible ratio. Each row is laid out independently, and header, data, and summary rows should use a consistent column configuration. A single cell can fill an entire row; there is no cross-row merging or automatic measurement of the entire table column width. `Items(el.Center)` / `Items(el.End)` sets the horizontal alignment of the content, and P/Px/Py adjusts the white space; the text thickness, etc. can be set on the incoming Text. When long content scrolling is required, combine ScrollX/ScrollY in the outer layer.

Semantic roles include table, rowgroup, row, columnheader, cell, and caption. Static rows have no default selection, activation or keyboard navigation, and child controls receive events independently; OnClick/Focusable/OnKey can be explicitly configured when the entire row is required. The original `TableCell` is the data table selection coordinate type, so the static cell entry is named `TableDataCell`.

Run `go run ./examples/components -section table_static` to see examples of interactions, banner comments, and summaries. 1×/2× Automatic testing covers the three-segment alignment of some fixed columns, page footer/description position, sub-button events and input status; real device visual acceptance has not yet been completed.

## Whole column selection, column limits and density

`ColumnSelect()` Enter independent column mode and clear the original row/cell selection. Click the table header or data cell to select the column, Ctrl/Cmd click to increase or decrease, and Shift click to expand the range of consecutive columns; the left and right keys jump to adjacent optional columns, Home/End jumps to the beginning and end, Shift expands, and Ctrl/Cmd+A selects all visible and optional columns. Double-click the header to continue sorting. Whole column selection does not trigger row selection or row activation callbacks.

`SelectedColumns()` returns a copy of the source column index in display order, including hidden selected columns; `SetSelectedColumns([]int)` assigns silently, and `OnColumnSelectionChange` accepts user changes. The selection is saved in columns. The new rows added by SetRows naturally belong to the selected columns. Columns can also be selected in an empty table. Filtering does not lose selections; copying only outputs visible selected columns and filtered rows. `SelectedCells()` can be expanded to a snapshot of cells in the current filtered results, and the cost when called increases with the number of rows. `SetValue` / `SetSelectedRows` / `SetSelectedCells` Do not change the selection in column mode, use column interface settings. CellSelect/MultiSelect exits this mode.

Column configuration supports:

- `Selectable(false)`: This column is prohibited from participating in cell/whole column selection. Range selection, keyboard, all selection and program selection assignment are skipped; whole row selection or column sorting is not affected.
- `Resizable(false)`: Removes the handle for user dragging column width. Apps can still set the width using SetColumnWidth or layout restoration.
- `Movable(false)`: MoveColumn does not move the column, nor does it allow other columns to change their position across it. Explicit SetLayoutState still restores the full layout specified by the app. Header dragging follows the same locking rules.

Column configuration is copied when constructing the table. In the example, the width of the single number column is locked by dragging and moving, and the header and buttons can move other columns.

`Stripe(true)` alternately fills weak background in visible row order after filtering/sorting; selection highlighting takes priority. `RowHeight(dp)` Sets the actual row height and virtual list size simultaneously, ranging from 24–256dp, including 1dp separators, 0 returns to 40dp. Setting the row height does not change the cell control font size; customized content needs to adapt to the row height. Existing active guilds will be resurfaced.

The automatic test covers entire column selection, range skipping restricted columns, empty tables, data appending, filtering/hiding column copying, mode switching, disabling the keyboard, mobile lock and 1×/2× row height changes; real device dragging and visual acceptance have not yet been accepted.

Dragging the header hits the actual drawing position, supporting horizontally scrolled areas and frozen left and right columns; dragging to the left and right edges of the non-frozen header area automatically scrolls horizontally, bringing in columns outside the screen (see above). Hidden columns retain the source index and count towards the callback position. Disabling, hiding columns, or program-restoring layout cancels ongoing reordering. 1×/2× event testing covers reordering, locking, hiding, freezing, scrolling, canceling and sorting/column width operation isolation, and the real phone feel is still to be accepted.

Selecting the entire column will not force the focus of the inline input box to the table. The input box can continue to be edited, and the direction keys are handled by the input box; after clicking on the ordinary table header, the direction keys are still used for column navigation. Automatic testing covers both cell/column modes, 1×/2×, same-frame and split-frame clicks, and text retention after moving columns.
