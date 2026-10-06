# TimeField

English | [简体中文](time_field.zh-CN.md)

Enter the time of day. Supports whole string input and independent editing of hours, minutes and seconds.

```go
start := kit.TimeField("开始时间").Segmented()
start.SetValue(9*time.Hour + 30*time.Minute)
precise := kit.TimeField("结束时间").Seconds().Hour12(false)
precise.SetValue(17*time.Hour + 45*time.Minute + 30*time.Second)
```

- `Segmented()` turns on segmented input; `Seconds()` enables both seconds segments and second precision. Tab / Shift+Tab Switch between sections, press Enter after input or move out of the section to submit; illegal numbers revert to submitted values.
- In segment mode, ↑ ↓ adjusts the current segment by one unit, and PageUp / PageDown adjusts by ten units. Hours, minutes and seconds carry linkage, looping past midnight; one key press only triggers `OnChange` once.
- The segmented format follows `locale.Current().Clock12` by default: Chinese is 24-hour format, English is 12-hour format. `Hour12(bool)` can be overridden explicitly. The 12-hour clock shows localized AM/PM buttons; switching periods retain minutes and seconds, showing 12 for both midnight and noon. Switching formats at runtime discards uncommitted text and retains the value.
- When the above configuration is not called, the entire string of `HH:MM` is retained. Input: accepts `9:30`, `0930`, `930`; ↑ ↓ adjusts minutes, PageUp / PageDown adjusts hours.
- `Value()` returns `time.Duration` starting at midnight; defaults to minutes and `Seconds()` to seconds. `SetValue` takes the remainder of 24 hours and does not trigger a callback.
- `Size(kit.TimeFieldSizeXSmall / TimeFieldSizeSmall / TimeFieldSizeMedium / TimeFieldSizeLarge)` synchronously adjusts the frame, font size, clock icon, segment width and padding, applicable to full string and segmented modes. By default, Medium uses the theme control height; the minimum heights of the other three levels are 24, 28, and 40dp. Labels and error text retain the form's standard font size; switching sizes does not submit a draft or change values. Double-rate size validation for whole strings and segments passed.
- `SetDisabled`, `SetError`. Discard drafts when component or ancestor is disabled, show committed values when restored.

Agent: Each segment is a separate `textbox`, whose name consists of the localized "hour/minute/second" and field label; `FocusID()` points to the small segment. AM/PM are buttons with localized "slot" names.

Verify: `go run ./examples/components -section time_field`, add `-theme dark` to check the dark theme.

`SegmentKeys(true)` enables fast segmented editing and automatically turns on segmented mode; pass false to restore the original draft editing. Compared to default mode:

- When the focus is obtained, the entire paragraph is selected, and the left and right arrows submit the current draft and move it to the adjacent paragraph; moving the first paragraph to the left and the last paragraph to the right will not leave the control, and Tab can still leave.
- Enter or paste a valid two-digit number to submit immediately and focus the next segment; the focus period button is after the last digit segment of the 12-hour clock. Invalid two-digit numbers will not be automatically submitted or skipped.
- ↑ ↓. PageUp / PageDown only circulates within the current segment range and does not carry to other segments. The 12-hour period is 1–12, the 24-hour period is 0–23, and the minutes and seconds are 0–59.
- Backspace / Delete resets the current segment to 0; resets the 12-hour segment to 12 and retains the current AM/PM.
- The a/p of the 12-hour clock is set to am/pm; the time period button supports left and right segmentation, up and down switching, and deletion to return to am.

Continuous input, paste skips, independent loops, delete reset, period shortcuts, disabled and double size verified; Enter/focus shift commit and carry behavior still retained in default mode.
