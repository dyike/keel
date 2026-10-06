# Calendar

English | [简体中文](calendar.zh-CN.md)

Calendar displayed by month, optionally a day or a date range.

```go
cal := kit.Calendar().Bounds(time.Now(), time.Time{}).DisableDates(isWeekend)
span := kit.Calendar().Range().OnChange(func(start, end time.Time) { … })
```

- Keyboard: Tab falls on the current focus day when entered; arrow keys move by day or week, PageUp / PageDown moves by month (when the end of the month exceeds the target month, it falls on the last day of the month), Home / End jumps to the beginning and end of the week, Enter / Space selects.
- Range mode: Click for the first time to determine one end, and click for the second time to determine the other end. The order is not limited.
- `Bounds(min, max)` limits the optional range, a value of zero means there is no limit on this side; `DisableDates(fn)` disables certain dates.
- `Value()` returns `(start, end)`, which are the same in single-day mode; `SetValue` does not trigger a callback and jumps to the month where start is located. `SetMonth`, `SetDisabled`.
- The week name, the day the week starts, the month title, and the date format all come from `ui/locale`: Chinese starts from Monday, English starts from Sunday.

Agent: The calendar is `grid`, each day is `gridcell`, the name is the date (such as 2026-10-08), `selected` means selected or within the range; unselectable dates report `disabled`.

Verify: `go run ./examples/components -section calendar`, add `-theme dark` to check the dark theme.

`Months(n)` displays 1–12 consecutive months, wrapping when the width is insufficient; the cross-month panel does not generate date cell IDs repeatedly. Clicking on a displayed subsequent month will not jump to the first month. Click the top month title to enter the year and month selector. The year supports direct input. Click the month to complete the jump; the month name comes from `locale.MonthNames`.

The keyboard skips disabled dates; when it exceeds Bounds, it searches for optional dates inward from the boundary to avoid losing focus after it falls on a disabled cell. In order to avoid the "all disabled" predicate causing the UI to search endlessly, each navigation will be checked for up to 366 days, and if it is not found, the original focus will be maintained. Bounds reverse endpoints are swapped; out-of-bounds month page buttons are disabled.

Range selection uses draft: the first click only sets the pending endpoint, `Value()` still returns the originally submitted range; the second click is called back after completion. If any day within the range is not selectable, the submission will be rejected and the selection will be restarted from the date of this click, and a prompt will be displayed. `RangePending()` queries the draft, `CancelRange()`, the Cancel button, or Esc on the focused date abandons the draft without changing the value or triggering a callback. Disabling, modifying Bounds/DisableDates, and program SetValue will cancel the draft.

The standard width of a single-month grid is 252dp; the seven columns in the narrower container shrink to the same width, and the week title and date share the same column width. Each day and last column click are returned under 224dp, 1× / 2×; multi-month view still wraps by month.

`FirstWeekday(time.Sunday)` sets the first day of the week for the current calendar, supporting Sunday to Saturday; illegal values are ignored. The week header, date column and Home/End weekday/weekend navigation change simultaneously. `ResetFirstWeekday()` resumes following the current locale. Modifying the configuration does not change the selected date, draft range, or trigger OnChange; other calendars are not affected.

`Size(CalendarSizeXSmall/Small/Medium/Large)` Adjust the size of the date grid, week column, navigation and year and month selector. The default Medium retains the original layout and date text inheritance. The width and height of a single cell are 28×24, 32×28, 36×32, and 44×40dp respectively; for many months, the lines are still wrapped according to the available width. Switching the size does not change the selected value, range draft or trigger callback, and illegal gears are ignored.
