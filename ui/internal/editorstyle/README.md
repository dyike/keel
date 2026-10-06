# ui/internal/editorstyle

English | [简体中文](README.zh-CN.md)

`ui/el` Drawing of the input box: Draw the selection background and cursor according to the visible glyph, keeping the cursor aligned with the placeholder text. Input events, selection ranges, scrolling, undo, and input methods are still handled by Gio's `Editor`.

- **Dependencies**: Gio, Go standard image and time library, `golang.org/x/image/math/fixed`; does not depend on other Keel modules.
- **Used by**: `ui/el` (the input box of kit is used by el).

`Caret.Layout` is only responsible for drawing; the caller must process `Editor.Update` first. The test covers both editing interaction and 1× / 2× actual rendered pixels.

The selection follows the horizontal range and baseline of `Editor.Regions`, and then calculates the height according to the glyph range. When drawing, first record the original editor operation, draw the selection background, and then play back the text; you cannot modify the selection by moving the text as a whole. For positioning steps, see [Troubleshooting text and selection alignment](../../../docs/troubleshooting.md#the-input-box-text-is-attached-to-the-upper-edge-of-the-selection-and-the-space-below-the-selection-is-too-large).

`Composition` draws a combined underline using Regions typed by the editor and returns the combined range within the viewport; `Caret.InputMethodCaret` provides a baseline and glyph height consistent with the visible cursor. The input adaptation layer that consumes input method events by itself is called after Layout and is responsible for sending these coordinates to the platform. This internal layer does not maintain editing transactions.

InlineFont Constructs a whitespace font containing only private characters for a quoted block. The glyph's advance participates in Gio Editor's line wrapping, selection and cursor calculations, and the UI can draw additional content in the corresponding Regions. Fonts are only added to the shaper of a single editor and are not registered globally; the caller is responsible for selecting private characters that do not conflict with normal text, measuring widths, and layout caching. Input/Textarea reference typesetting has been connected; outlineless glyphs retain non-zero boundaries to prevent line endings from being ignored as whitespace.

Fonts are constructed in OpenType [cmap](https://learn.microsoft.com/en-us/typography/opentype/spec/cmap) format 12 and [hmtx](https://learn.microsoft.com/en-us/typography/opentype/spec/hmtx), including sfnt checksum; normal characters must fall back to normal fonts. The existing go-text uses signed advance, and the design unit limit for a single glyph is 32767; the caller needs to select UnitsPerEm according to the font size and maximum width, and should not directly use dp as the design unit.
