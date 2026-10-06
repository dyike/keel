# Troubleshooting

English | [简体中文](troubleshooting.zh-CN.md)

## Chinese is displayed as a box

The Go font that comes with Gio does not have Chinese characters, and Chinese characters fall back to the system fonts. Automatic fallback sometimes selects fonts with missing characters, for example, the "group" in a bold title is displayed as a box.

`theme.Face` Fixed font priority: Apple → Holly Helvetica → Microsoft Yahei → Noto Sans CJK. A box is displayed on Linux, indicating that none of these are installed. Just install the Noto CJK font (Debian/Ubuntu: `fonts-noto-cjk`). To use a different font, change `theme.Face` and put the font family name first.

## Chinese in the title bar of Linux (WSLg, GNOME Wayland) is displayed as a box

Some Wayland synthesizers do not draw the title bar and leave it to the program to draw it, such as WSLg and GNOME. When Gio encounters this situation, it will draw one by itself, but the theme it uses does not contain the font loaded by keel, so the Chinese title becomes a square box and the color does not change with the theme.

Keel will now turn off the title bar drawn by Gio himself. When the compositor does not provide a title bar, keel will draw it using the current theme and font: the title is typeset with keel's font, and the background is `theme.Subtle`. Minimize, maximize, close buttons, drag to move, and double-click to maximize are all available. Environments that draw title bars, such as X11 and KDE, are not affected; the `Frameless` window is still drawn by the application itself (such as `kit.TitleBar`).

If the Chinese characters in the title bar are still square, it means there are no Chinese fonts in the system. Install `fonts-noto-cjk`, or use `theme.LoadFonts` to load the font file.

## The lower half of g and y are cut off on Windows

Pure text elements (without customized Role, Name, or selected state) no longer draw the cropping area by themselves. Gio's Label reports the semantics and leaves margins according to the glyph excess, so the descending part will not be cut. The previous writing method was based on the fixed downward movement amount (font size × 0.22) adjusted by macOS Apple. There was a lot of blank space below the Apple square line box, so there was no problem. Microsoft Yahei line box in Windows adhered to the font. After moving down, the g and y exceeded the text box and were cut off by the semantic cropping area of the text box. Regression test `TestTextShiftKeepsDescendersInBox` Reproduces this compact line box on any system using Go fonts.

## The Chinese in the button is on the upper side

The logical line frame of the font contains ascent/descent and white space, and the visible glyph is not necessarily in the center of the line frame. The text of el is measured and adjusted to the drawing baseline according to the representative glyph (`ui/el/paint.go`'s `textShift`): Use the actual line frame of the same font size Label and the glyph position of "国" and "A" to calculate the downward movement, and do not allow the descending part of g to move out of the line frame; the numerical corner label is centered according to the actual numerical glyph. First distinguish whether the container is in the wrong position or the glyph is offset within the container, and then change the corresponding layer. Don't directly add the pixel difference measured in a certain screenshot to all components.

## The input box cursor extends downward than the text

By default, Gio draws the cursor according to the ascent/descent of the font. The descent of some Chinese fonts contains more white spaces. `ui/internal/editorstyle` draws the cursor uniformly for two sets of input components: the position uses the actual coordinates of the editor, and the height uses the visible range of Chinese glyphs, Latin capitals, and descenders. Do not move the cursor alone to compensate, otherwise click positioning, multi-line, and scrolling will be misaligned.

## The input box text is attached to the upper edge of the selection, and the space below the selection is too large.

This is a problem of mixing logical line boxes and visible glyphs. In a 2× Chinese rendering, glyphs occupy pixels 2–28, and Gio’s default selection occupies pixels 0–44. Just changing the cursor height cannot correct the selection; just moving the text will make the text deviate from the coordinates used by clicking, word selection and input methods.

The fix entry is [`ui/internal/editorstyle`](../ui/internal/editorstyle/README.md), which is called by `el.Input`, `el.TextArea` (and `kit.Input`, `kit.TextArea` based on them). Processing order:

1. The caller first consumes `Editor.Update`, then uses the original editor to complete typesetting and record the drawing operation. The native selection is made transparent.
2. Use `Editor.Regions` to get a selection range that takes wrapping, bidirectional text, and scrolling into account. The lateral ranges are inherited directly; the baseline for each zone is `Bounds.Max.Y - Baseline`.
3. Measure the top and bottom edges of the visible glyph of the selected text, draw the selection relative to the baseline, and leave 1dp padding. Purely white selections use the height of the representative glyph; password boxes measure the mask. Remeasure when text or font changes.
4. Crop the selection to the editor viewport, draw the background first, and then play back the original text. Retains the original editor's text positioning, line spacing, input, selection range, and scrolling behavior.

When encountering text alignment problems again, troubleshoot in this order:

- **Reproduce state first**: fill in content, focus, select text and scroll. Screenshots of empty input boxes cannot verify selection and scrolling.
- **Check coordinates separately**: Compare component bounds, logical line boxes, actual glyphs, cursors and selections. The container only reserves space for the corner mark on one side, which will cause the entire button to shift; for this kind of problem, the container should be repaired.
- **Use real fonts and proportions**: Simultaneously test Chinese, numbers, English descending letters, mixed layout and passwords, covering 1×/2×, blank lines, soft line breaks, and partially visible lines.
- **Verify actual pixels**: Draw text and selections in different colors, check visible range and center, not just the returned size or font parameters. The test must provide `input.Router.Source()`, otherwise the context is disabled and the color is faded.
- **Verify interaction**: After selection, scroll, replace, undo and input method combination events are all retained. When simulating text replacement, `key.EditEvent` should include a selection range, and then send `key.SelectionEvent` to update the cursor.

Return entry:

```sh
go test ./ui/internal/editorstyle ./ui/el ./ui/kit -count=1
go run ./examples/components -section textarea
go run ./examples/components -section input
```

Pixel regression in `ui/internal/editorstyle/selection_test.go`, cursor and input method regression in `caret_test.go`. In the TextArea example, click "Fill in multiple lines" and select all, check the selection of each line, and then select scroll; in the Input example, check numbers and Chinese.

## The automatic height is set to four lines, but the fourth line is cut off

Gio's `LineHeightScale` of 0 uses the default row height of 1.2x. If the viewport is calculated based on `rows × measured line height`, and then only `LineHeight` is set, the text layout will still be multiplied by an additional 1.2, and the last line will not fit.

`Field.AutoHeight` uses the measured line height and explicitly sets `LineHeightScale = 1` so that the typesetting and viewport use the same leading. The regression cannot only compare whether the input box becomes taller, but also checks whether the four lines of `Editor.Regions` all fall within the viewport, covering the blank lines and 1× / 2×; the corresponding test is `TestTextAreaAutoHeightFitsEveryVisibleLine`. It should be scrolled after exceeding the maximum number of lines, and the content cannot be truncated.

## -lobjc warning when linking

```
ld: warning: ignoring duplicate libraries: '-lobjc'
```

Both Gio and `native/internal/sys` declare the link `libobjc`. The linker prompts for duplication and automatically removes the duplication. It does not affect the results and can be ignored.

## The interface is not refreshed, or -race reports data competition.

Most of the time the component is changed directly in the goroutine. Modifying components outside the callback must include `core.Update`:

```go
go func() {
    data := fetch()
    core.Update(func() { label.SetText(data) })   // Don't label.SetText directly
}()
```

`go test -race` and `go run -race` can detect such problems.

## The program is stuck and all windows are unresponsive

A callback took too long to execute. All windows share a lock (see [Architecture · Threading Rules](architecture.md#threading)), a callback is stuck, and all windows are stopped. Common reasons:

- In the callback, network requests, large file reading, and `screen.Capture` are made: put into goroutine.
- The result of waiting for `core.Update` in the callback: `core.Update` will not be executed until the current callback ends. Waiting for each other is a deadlock.

## Stuck after pressing shortcut keys or clicking buttons when using multiple windows

If it is a newly added window method: it calls the Gio method (`Perform`, `Option`, `Run`) that will wait for the main thread in the lock, forming a loop waiting with another window. For solutions and principles, see [Architecture · Cannot wait for the main thread](architecture.md#avoid-waiting-for-the-main-thread-while-holding-the-lock) in the lock. `Raise`, `Close` have fixed this issue.

## Crash when the window just opened is closed immediately

Gio v0.10.3 on macOS, the window is closed before it is completely created, and a segfault occurs at `cascadeTopLeftFromPoint`. Keel's `Close` and `Raise` already wait until the first frame of the window is drawn before executing. If you bypass Keel and call Gio's `app.Window.Perform` directly, you have to wait for the first `FrameEvent` yourself.

## Window shortcut keys are not responding

- Only takes effect when the window has focus.
- When the input box has focus, it will first process editing-related key combinations (⌘A, ⌘C, ⌘V, ⌘X, ⌘Z, etc.), and the same key combinations registered as window shortcut keys will not be triggered. Change the combination.
- If the shortcut key string is written incorrectly, it will panic at `window.Open` and will not fail silently.

## Global shortcut keys are not responding

- The program on macOS must be run in `window.Main()`, and shortcut key events are dispatched by the event loop of the main thread.
- An X11 session is required on Linux, and under Wayland it can only be received when the XWayland program has focus.
- When the key combination is occupied by other programs, `hotkey.Register` returns `native.ErrConflict` and checks the return value.
- The callback is indeed triggered but the interface has not changed: Use `core.Update` to change the interface in the callback.

## Permissions are always false

- When `go run` is executed in the terminal, macOS usually records the authorization under the name of the terminal program. Go to System Settings → Privacy & Security and look for Terminal (or iTerm, VS Code), not your program name.
- After granting screen recording permission, restart the program.
- Every time you recompile, the binary signature will change, and macOS may think it is a new program and require re-authorization. Wrapping it into a signed `.app` avoids this problem.

## After closing the window, the code after window.Main() in main is not executed.

When the last window is closed, the process directly `os.Exit(0)`, `window.Main()` will not return, and `defer` will not be executed. The cleanup work is placed in `OnClose` of the last window.

## Window cannot be hidden

Gio does not support hiding and showing windows. For windows that need to be "closed and reopened", close and reopen `window.Open`, and the status is saved in your own variables. See the costs listed in [Design Decisions](decisions.md#choose-gio).

## Components overlap in the first frame and are restored after one interaction

First check whether the size is recorded in the Decorate/drawing stage, and Render has used the size of the previous frame to determine the structure. Simply calling Frame repeatedly and asserting again can easily hide this problem.

The corresponding processing method is: Toolbar first uses "More" to carry the operation when the width of the command area is unknown, and then expands after measurement. The command area is cropped to prevent covering the left and right slots; Dock uses the current root viewport to initialize the docking size. If the actual embedded size is different, redraw correction is requested. Don't hide errors by taking two screenshots first.

Regression should include the first unsettled frame, the first frame with narrowed windows, 1×/2×, and check for occupied adjacent areas and operational reachability. Corresponds to `TestToolbarFirstFrameKeepsActionsAccessibleBesideSlots`, `TestDockFirstFrameAndResizeKeepCenter`. `examples/components -matrix` intentionally only outputs the first frame, and the dynamic overlay and scrolling still pass the real input routing test.
