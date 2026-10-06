# DatePicker

English | [简体中文](date_picker.zh-CN.md)

When clicked, the date field of the calendar pops up.

```go
due := kit.DatePicker("交货日期").Placeholder("选择日期")
trip := kit.DatePicker("出差日期").Range().Months(2)
```

- Click, Enter or ↓ to open the calendar. The focus falls on the selected date. If there is no selection, it falls on today; if the target date is disabled, the disabled date will be skipped. It will automatically close after selecting the date (or selecting both ends of the range); Esc or click outside to close without changing the selection.
- The first end of the range acts as a draft only; Esc, click outside, disabled field, or ancestor container discards the draft, retains the committed value, and does not trigger `OnChange`. Reopening starts from the submitted date, and the year and month selection panel will also be reset.
- `Months(n)` Request display 1–12 months. The narrow window automatically reduces the number of months side by side, and the month arrows can still be used to access the remaining months; when the height is insufficient, the pop-up layer can be scrolled, and the keyboard will scroll to the target grid when switching dates.
- `Value()` / `SetValue(start, end)`, `OnChange(func(start, end time.Time))`; `Bounds`, `DisableDates` Same as Calendar; `SetDisabled`, `SetError`. `el.Root` REQUIRED.

Agent: The field is `button`, the name is the label, and `value` is the displayed date; when opened, it is `dialog` named label, and inside it is `grid` of Calendar.

Verify: `go run ./examples/components -section date_picker`, add `-theme dark` to check the dark theme.

`Format(layout)` uses Go's time layout to format a single date and both ends of a range at the same time, such as `Format("2006-01-02")`; passing an empty string restores the date format of the current locale. It only changes the display, does not modify the date, and does not trigger a callback. The format syntax is different from the GPUI chrono format.

`Clearable(true)` displays an independent clear button when there is a date; clearing it with the mouse or keyboard will close the pop-up layer, discard the range draft, remove errors, and call back `OnChange(time.Time{}, time.Time{})` once, and the focus will return to the date trigger. Clearing will not open the calendar by mistake, and the disabled state is inherited to the clear button; `Clearable(false)` hides the button, and the program SetValue does not trigger a callback.

`Presets(...DatePickerPreset)` displays shortcut buttons below the calendar, and the narrow layout automatically wraps and scrolls with the pop-up layer. Each item contains stable `ID`, `Label`, `Start`, `End`; the single date mode ignores End, the empty End in the range mode means the same day, and the reverse range is automatically sorted. Empty IDs/tags are ignored, duplicate IDs retain the first item; components copy input slices, `Presets()` clears all presets.

The default respects Bounds and DisableDates, and the range cannot cross the disabled date; unavailable items remain displayed but disabled, and the latest configuration is rechecked when clicked. Selecting a preset cancels the range draft, clears errors, closes the popup, and calls back once, returning focus to the field. The custom disable function checks day by day and should remain fast and side-effect free. Dates are snapshots, and mobile presets like "Today" and "Last Seven Days" are updated by the app.

```go
trip.Presets(kit.DatePickerPreset{
    ID: "week", Label: "最近七天",
    Start: today.AddDate(0, 0, -6), End: today,
})
```

`Size(dp)` synchronizes the minimum height, font size, calendar icon, clear button and spacing of the zoom field, 28/36/48dp is recommended; `Size(0)` restores the theme height and inherited font size. Negative and non-finite values are ignored. Calendar grid and preset buttons do not scale with fields.

`Appearance(false)` removes the field background, border, rounded corners and padding, retains the minimum height, label, error copy and keyboard interaction, and the pop-up layer still retains the original decoration; `Appearance(true)` restores it. Plain fields do not show the focus/error color on the default border and can be drawn themselves on the outer layer. Switching these configurations while the popup is open will not clear the date or close the popup.

`WithTime()` adds segmented time input to a single date, and the default is accurate to minutes; `TimeSeconds()` turns on seconds, and `TimeHour12(bool)` covers the 12/24-hour format of the locale. `DefaultTime(9*time.Hour)` sets the current clock, if it is negative or exceeds one day, it wraps back by one day; these configurations do not trigger callbacks. Configure precision first, then SetValue; values received by Calendar before enabling time only retain dates.

The time patterns `Value`, `SetValue` and `OnChange` contain hours, minutes and seconds, the same on both ends of a single date, preserving the date's time zone. Select a new date to keep the current clock and call back immediately, keeping the popup open; select the same day to close without repeating the callback. Time is submitted via Enter, Focus, or Arrow keys; Esc/External Close retains the submitted value and discards the time entry draft. Adjusting the clock when there is no date yet does not generate a date or trigger a callback. Clear the reserved clock for next selection.

The date and time are displayed by default; the custom Format is a complete layout and needs to include the time itself, such as `2006-01-02 15:04:05`. Range calendars still only edit dates, date boundary/disable rules do not limit moments. Date presets retain the current clock by default; setting `DatePickerPreset.IncludeTime: true` uses the hours, minutes and seconds of Start (truncated to component precision, midnight can also be specified explicitly), still checking date boundaries and disabling rules first. Pure date and range modes also store these moments, read via DateTimeValue. Combine two single date components when start and end times are required.

Default example: `kit.DatePickerPreset{ID: "meeting", Label: "下午会议", Start: meetingTime, IncludeTime: true}`. After adding the IncludeTime field, the code that uses positional parameters to construct DatePickerPreset needs to be changed to a named field. The range calendar still only edits dates, and saves the start and end times independently.


`DateValue()` always returns the date part; `SetDateValue(start,end)` replaces the date and preserves the clock at both ends. `DateTimeValue()` / `SetDateTimeValue(start,end)` accesses the complete date and time, supports single date and range, does not automatically open time input, and does not trigger callbacks. Use Start for the empty End of the range, and sort the reverse range together with the time; retain nanoseconds when the time precision is not configured, and truncate both ends according to the precision after configuring the minute/second precision. Empty dates retain a zero value, clearing the retained clock.

Compatible behavior: `Value` / `SetValue` for ranges still only read/replace dates; Value / SetValue for single-date WithTime include the time. OnChange contains the saved moment when the user selects a date, completion range, or selects a preset. The default range text displays only the date; specify the full Format when a time of day is required. `DefaultTime` sets the clocks on both ends simultaneously. The time is combined according to the local clock face of the time zone where the date belongs. If daylight saving time does not exist/repeated times follow the processing rules of Go time.Date.

`FirstWeekday(time.Monday)` separately configures the starting day of the week in the bullet calendar; `ResetFirstWeekday()` resumes following the current locale, and illegal values are ignored. It can be modified during opening, the selected date and range are retained in draft, and the table header, date grid, Home/End and scroll positioning of the focus date are updated at the same time.
