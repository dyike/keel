# Pagination

English | [简体中文](pagination.zh-CN.md)

Pagination strips.

```go
pages := kit.Pagination(total, 20).OnChange(func(p int) { load(p) })
start, end := pages.Bounds() // Data range of current page [start, end)
```

- Page numbers start from 1. When there are many pages, the page numbers near the first page, last page and current page are displayed, with ellipsis in between.
- "Previous Page" is disabled for the first page, and "Next Page" is disabled for the last page.
- `Value()` / `SetValue(p)` (will be limited to the legal range and will not trigger a callback), `SetTotal(n)`, `Pages()`.
- "Total N items", "previous page" and "next page" come from locale.

Agent: container roles `navigation`, `value` are "current page/total number of pages"; the page number is a button named number.

Verify: `go run ./examples/components -section pagination`, add `-theme dark` to check the dark theme.

In a narrow container, the total number, page number, and previous and next page buttons are automatically wrapped, and the last page button is still clickable. Calculation of page number and last page range to avoid integer addition overflow.

`Compact(true)` only displays the previous and next page icon buttons, hiding the total number and page number; Agent still reports the current page/total page number. false restores the complete layout and switches without changing the current page.

`VisiblePages(n)` sets the upper limit of number buttons, excluding ellipsis and previous and next pages. Positive numbers are limited to 3–101 so that the first, current, and last pages are always preserved; the middle window moves with the current page. 0 restores the original window strategy, negative numbers are ignored. The default layout follows Keel's original layout (all displayed when there are less than eight pages), which is different from the upstream default of five buttons.

`Size(dp)` sets the button height, positive numbers are limited to 16–128dp, 0 returns to the default 28dp; negative numbers and non-finite values are ignored. Use continuous size, optional 20/24/28/36dp corresponding to different densities. `SetDisabled(true)` disables entire paging; ancestor disabling also takes effect. Program calls to SetValue/SetTotal still update the state and do not trigger OnChange.

```go
pages.VisiblePages(9).Size(36)
compact := kit.Pagination(total, 20).Compact(true).Size(24)
```

Page button uses stable identity, and the remaining button retains keyboard focus when window is moved and compact mode is switched. When the total number of pages is reduced, the current page will converge to the last page.
