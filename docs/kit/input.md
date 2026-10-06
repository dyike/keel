# Input / TextArea

English | [简体中文](input.zh-CN.md)

Labeled text box, you can add suffixes, clear buttons and error prompts. `kit.TextArea` creates a multi-line version.

```go
search := kit.Input("搜索").Placeholder("客户或单号").Clearable().Prefix(searchIcon)
price := kit.Input("单价").Filter("0123456789.").Suffix(yuan)
note := kit.TextArea("备注").Rows(4)
message := kit.TextArea("消息").AutoGrow(2, 8)
```

- `Value()` / `SetValue`; `OnChange` is called after each edit, and `OnSubmit` is called when Enter is pressed in the single-line box.
- `Password()` masks the content, `MaxLength(n)` limits the number of characters, and `Filter(chars)` only accepts these characters (filtered for both input and paste).
- `Clearable()` displays a clear button when there is content. After clearing, the focus remains in the input box.
- `SetError(msg)` displays errors below and turns the border red, which is automatically cleared when the user edits again; `Form` uses it to display the verification results.
- `SetDisabled`, `SetReadOnly`; when read-only, you can select and copy, but cannot edit.
- The text box will fill the width given by the parent container. `FocusID()` returns the element ID of the text box, which can be passed to `cx.Focus`.

Agent: role `textbox`, the name is the label (it is placeholder text when there is no label, and it is the row label in Form), `value` is the content; the clear button is named "Clear Label".

Verify: `go run ./examples/components -section input`, add `-theme dark` to check the dark theme.

After the user modifies a single line or multiple lines of input, the current error will be cleared to facilitate re-verification; programmatic assignment does not implicitly clear server-side errors, and input when disabled will not clear errors or trigger callbacks.

`AutoGrow(minRows, maxRows)` The number of lines after formatting according to the text will automatically increase, including soft line breaks; after exceeding the upper limit, scrolling in the editor, and deleting text will shrink to the minimum height. Line height is calculated with font and display scaling, and there is no limit on text length. The parameters must satisfy `minRows > 0` and `maxRows >= minRows`, and illegal parameters are ignored; if the two values are equal, the number of visible rows can be fixed. `Rows(n)` restores the original minimum height mode and cancels AutoGrow. Single-line Input ignores AutoGrow. Mode switching retains input focus and content; long placeholder text does not participate in automatic heightening. Explicit height or space constraints from the parent still take precedence.

`Mask(pattern)` sets a template mask for single-line input: `#` is a required ASCII number, `9` is an optional number, `A` is a Unicode letter, `*` is a letter or number; the backslash escapes the next character, and the remaining characters are fixed text. Fixed text is displayed with filled slots, no placeholders are used to fill empty slots; empty input remains an empty string. A single backslash at the end is considered an illegal configuration and the old mask is retained.

```go
phone := kit.Input("电话").Mask("(###)-###-####")
phone.SetValue("1234567890")
// Value(): (123)-456-7890；UnmaskedValue(): 1234567890
amount := kit.Input("金额").NumberMask(',', 2)
amount.SetValue("1234567.89") // 1,234,567.89
```

`NumberMask(separator, fraction)` Groups the integer part into three digits, allowing a leading negative sign, with a period as the decimal point. When separator is 0, there is no grouping; when fraction is -1, there is no limit to the number of decimal places. When 0, it is truncated when encountering a decimal point. Only the integer part is retained, and a positive fraction limits the number of decimal places; the excess part is truncated, not rounded, and floating point numbers are not converted. Only negative signs or endings with a decimal point remain in draft. An illegal separator or fraction does not change the configuration.

`Value()`, `OnChange`, `OnSubmit` use formatted text; `UnmaskedValue()` returns the actual characters filled in the template or numeric text without grouping characters. `MaskComplete()` checks whether the required slot is filled. The number mode requires at least one digit and cannot end with a decimal point; business rules such as date validity and number attribution are not verified. Only empty mask values for optional slots may be complete and should be supplemented with required validation.

When the mask is enabled, Filter constrains the actual input characters, MaxLength limits the number of runes after formatting, and automatically inserted template text or grouping characters are not counted. Setting a mask, length, or Filter reformats the current value but does not trigger a callback; `SetValue` also formats it. `Mask("")` Removes the mask and retains the current text. TextArea ignores Mask and NumberMask.

The editor synchronizes the formatted text and selection through el.TransformEdit, and uses the undo and redo records of the editing layer; when deleting a single automatically inserted delimiter, press the Backspace/Delete direction to delete adjacent editable characters to avoid deletion due to repeated insertion of delimiters. Automatic testing has covered template/number grouping, Unicode, filtering/length, delimiter deletion, selection, undo redo and menu cut; fixed prefixes with numbers are recognized as complete fragments to avoid swallowing the original input.

Input and TextArea provide right-click editing menus by default (open by long-pressing the input box on the touch screen), including copy, cut, paste, and select all; copy/cut is disabled when there is no selection, cut/paste is disabled when read-only, and menu copy/cut is disabled in the password box. The menu is anchored below the input box, and the editing command returns focus to the input box after closing; the selection operation uses the editor selection retained before opening the menu.

`ContextMenu(menu)` uses a custom Menu, passing nil to restore the default menu; `ContextMenuEnabled(false)` closes the default and custom menus at the same time. The menu instance belongs to this input, its Trigger is not used, and the entries, submenus, position, width, etc. are still configured by the Menu. The menu is closed by overlay ownership rules when the component or ancestor is disabled/hidden. Custom menus can be used to invoke editing commands with `cx.InputAction(field.FocusID(), el.InputCopy / InputCut / InputPaste / InputSelectAll)` and to restore focus with `cx.Focus(field.FocusID())`. Create a Menu once and reuse it to avoid losing the open state caused by replacement every frame.

Menu editing commands and keyboard paste share filtering, formatting and editing callbacks. See below for image/file interception. Automatic testing covers right-click hits, touch screen long presses, focus recovery, read-only/password restrictions, custom menus, and mask cutting.

`OnPaste(func(core.ClipboardData) bool)` intercepts keyboard or menu pasting: return true to indicate that the application has received it and no longer insert text; false to hand over Data.Text to the normal filtering/masking/callback process. Images contains MIME and encoding data, Files is the path, and the component will not open the file by itself. `PasteReader(core.ClipboardReader)` Provided by the application for asynchronous reading; callback to receive Gio text paste when not configured. `OnPasteError` reports an error in the UI thread and falls back to Gio text reading after rich reading fails; rejects insertion if Gio text reading exceeds 16MiB or fails.

The component is responsible for sending asynchronous completion back to the UI thread; input text or selection changes while waiting, discarding old results when the component is disabled/read-only. If the paste is initiated repeatedly, the latest request shall prevail. The platform reader must call the completion callback and cannot wait synchronously on the platform main thread in the UI thread.

Component library notes example has been adapted to `native/clipboard.Read`: macOS supports text, PNG/TIFF and file URLs, Windows supports Unicode text, PNG/DIB and file paths; Linux reads through the window's own connection under Wayland (the example calls `clipboard.UseWaylandDisplay(mainWindow.WaylandDisplay())`), and uses an independent connection under X11. They all support UTF-8 text, encoded images and native file URIs; Wayland paths have not yet been run on real devices. After receiving the image/file, display the quantity and consume and paste; fallback text for platforms that are not yet supported. Automatic testing covers single-line/multi-line consumption and rollback, old text/selection rejection, repeated completion, read-only, over-limit and stream closing. The bridge has been run on the main thread of macOS and the current image clipboard has been successfully read; file URL, PNG/TIFF types and real window pasting have not been accepted one by one. CodeEditor equivalent hooks have verified multi-cursor and undo. The Windows format parsing and cross-compilation, X11 format and chunked transmission tests have been passed. The system clipboard and real window pasting of the two platforms are awaiting acceptance by the real device.

## Size

`Size(InputSizeXSmall/Small/Medium/Large)` is available for Input and TextArea. Medium retains the original theme size; other files adjust the input font size, minimum height of the box, and padding. The TextArea's fixed Rows height changes with font size, AutoGrow continues to be measured by the actual text; labels, suffixes, and custom content retain their own styles. Within the InputGroup, the group still controls the outline and padding. Double size, focus and value preservation, AutoGrow grow/shrink and fixed Rows font size scaling verified.

## Atomic reference

```go
field := kit.Input("引用")
draft, err := kit.NewInputContent("查看 docs/input.md", kit.InputTokenSpan{
    Range: kit.InputRange{Start: len("查看 "), End: len("查看 docs/input.md")},
    Token: kit.InputToken{ID: "input-doc", Text: "docs/input.md", Label: "输入组件文档"},
})
if err == nil {
    err = field.SetContent(draft)
}
field.OnTokenActivate(func(token kit.InputToken) { /* 应用打开 token.ID */ })
```

`Content()` returns a separate draft containing commit text and citation metadata; ranges use UTF-8 bytes, must fall on grapheme boundaries, and must not overlap. Text must be consistent with the text in the range, and ID cannot be empty; the same ID can appear multiple times. Text is displayed when Label is omitted. The display name and submission text can be different. `Value()`, selection copy, and Form value all use the submission text.

`SetContent` Restore draft and clear undo; `SetValue` remove all references and clear undo even if the text is the same. `ReplaceWithToken` Replaces the current selection, records the undo and triggers OnChange; returns an error during disabled, read-only, or input method combination input. Normal typing or pasting will not automatically recognize the same text as a quote; replacing the quote with the same displayed text will also remove the quote ID. Deletions and non-empty selections that overwrite any part of a reference are treated as the entire reference, and undo restores metadata and selections.

Clicking on a reference selects it and calls `OnTokenActivate`; drag selection, shift click, and disabled state do not, and read-only state allows viewing. Applications can bind `ActivateToken()` to the shortcut key to activate the fully selected reference. The display background reuses the theme color; Agent can read the reference name and field submission text.

Inputs within Input, TextArea and InputGroup can use this editing path. Fields with password, mask, Filter, or MaxLength reject SetContent; subsequently enabling these modes exits quote editing, retaining the current commit text. Rich paste still uses PasteReader/OnPaste and rejects late paste results after restoring the draft.

References now participate in editor layout as a single object with independent width, and are moved to the next line as a whole during soft wrap, including references at the end of the text. Displays the label background block by default; when wider than the input viewport, content is constrained to the available width and is not split into multiple references. Real Gio event testing covers delete, undo, cursor, combined input transactions, asynchronous paste and click/drag selection; native input method candidate position, combined underline and cross-platform interaction have not yet been accepted.

Atomic reference input now maintains the combination range, draws the combination underline according to the actual typesetting position, and reports the visible combination boundary and cursor baseline to the platform; single-line horizontal scrolling, multi-line wrapping/cropping, and 1×/2× have automatic checks. Combo End, Defocus, Disable, Turn Read-Only, and Reset Draft all clear the old range. The final display of the platform candidate window still requires acceptance by the native window.

`TokenRenderer(func(gtx core.C, token InputToken) core.D)` Custom reference content, can draw icons and text, returns pixel dimensions and baseline from bottom. The callback first disables context measurement and then draws according to the measured size; it must comply with the Constraints and does not modify the application state or register the interaction processor. Reference activation still goes through OnTokenActivate. Pass nil to restore the default label block. The example uses el.Embed to construct a passive view of icons and text, and also displays the entire line wrapping under TextArea.AutoGrow.

The editor's local font is only responsible for object geometry and is not registered to the global font; ordinary characters and loaded fonts retain the fallback path. Internal placeholder characters avoid existing characters in the original text and labels; the input method reports readable labels, copy/Form uses the original text, and the platform editing interval and combination range will be converted into corresponding layout coordinates. 1×/2× tests cover custom sizes, narrow widths, end of text, private character conflicts, combined input and undo after quotes; the native IME has not yet been accepted on real devices.
