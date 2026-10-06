# CodeEditor

English | [简体中文](code_editor.zh-CN.md)

Code editor: line numbers, syntax highlighting, multi-cursor and column selection, find and replace, code folding, bracket matching, undo and redo, clipboard, input method. Only visible rows are laid out, and rows are stored in blocks: about 2 ms per frame for a 200,000-row file (`BenchmarkCodeEditorLargeFileFrame`), with `TestCodeEditorLargeFileFrames` checking for editing and page turning in it.

```go
ed := kit.CodeEditor(src).Language("go").Name("main.go").Height(400).
    OnChange(save).
    OnComplete(func(line, col int, prefix string) []kit.CodeCompletion { return lsp.Complete(line, col) }).
    OnHover(func(line, col int) string { return lsp.Hover(line, col) }).
    OnDefinition(func(line, col int) { ed.SetCursor(lsp.Definition(line, col)) })
ed.SetDiagnostics(lsp.Diagnostics())
```

No built-in language server protocol: The application connects the language server's diagnosis, completion, hover, and jump definitions to `SetDiagnostics`, `OnComplete`, `OnHover`, and `OnDefinition`.

## Highlight

Highlighting is an optional feature: `_ "github.com/dyike/keel/ui/highlight"` is introduced in the application to have color. When it is not introduced, the code is displayed as plain text, and editing, completion, folding, and search are performed as usual (it occupies about 4 MB, see [ui/highlight](../../ui/highlight/README.md)). After importing, press the name of `Language` to select the syntax. Immediately after editing, re-highlight dozens of lines near the changes, and then recalculate the entire file in the background; until the background results are returned, the remaining lines maintain their original colors. Highlight strings and comments at the same time, and bracket matching determines the position accordingly. No use for Tree-sitter: it requires cgo in Go and breaks browser versions and cross-compilation.

## Keyboard

| Operations | macOS | Windows / Linux |
| --- | --- | --- |
| Move and delete by word | Option+←/→, Option+⌫ | Ctrl+←/→, Ctrl+⌫ |
| Start of line, end of line | ⌘+←/→, Home/End (Home switches between the first non-blank column and the 0th column) | Home/End |
| The beginning and end of the file | ⌘+↑/↓, ⌘+Home/End | Ctrl+Home/End |
| Add cursor up and down | ⌘+Option+↑/↓ or Option+Shift+↑/↓ | Ctrl+Alt+↑/↓ or Alt+Shift+↑/↓ |
| Select the next same text | ⌘+D | Ctrl+D |
| Leave only the main cursor | Esc | Esc |
| Find, Find and Replace | ⌘+F, ⌘+Option+F | Ctrl+F, Ctrl+H |
| Next, previous match | ⌘+G, ⌘+Shift+G, F3, Shift+F3 | Ctrl+G, Ctrl+Shift+G, F3, Shift+F3 |
| Collapse and expand the area | ⌘+Option+[, ⌘+Option+] | Ctrl+Alt+[, Ctrl+Alt+] |
| Jump to definition | F12 | F12 |
| Completion | Ctrl+Space | Ctrl+Space |
| Undo, redo | ⌘+Z, ⌘+Shift+Z | Ctrl+Z, Ctrl+Shift+Z |

Enter keeps the indentation, indents one more level after `{ ( [` (Python also recognizes `:`), and puts the right bracket on the next line when it is between a pair of brackets. Tab/Shift+Tab Indent and un-indent selected lines. Copy and cut the entire line when there is no selection. When the number of pasted lines and the number of cursors are the same, each cursor gets one line. Press Esc then Tab to leave the editor.

## Mouse

Click to locate, drag to select, double-click to select a word, triple-click to select a line, click on the line number to select the entire line (select the entire area when folded). Option/Alt+click to add the cursor, Option/Alt+Shift to drag to select a column. The identifier under the pointer is underlined while holding down ⌘/Ctrl, and click to jump to the definition. The arrow to the left of the line number collapses and expands.

## Options and APIs

- Bracket pairing: `AutoClose(bool)`, enabled by default. Typing an opening parenthesis, bracket, brace, quote, or backtick inserts its closing partner. Typing a matching closer immediately before an existing one skips it. Backspace removes an empty pair; with a selection, typing an opener wraps the selection. A closing bracket on an otherwise blank line dedents to its matching opener. Pairing is suppressed inside strings and comments, and an apostrophe after a word is not paired.
- Fold by indent: A line and the deeper indented line after it form a region, and the right bracket line returned to the same level of indentation remains displayed. `Fold`/`Unfold`/`FoldAll`/`UnfoldAll`/`Folded`; when the cursor moves into the fold area, it will automatically expand; when editing above the area, the fold will move accordingly.
- `SoftWrap(bool)` Soft line wrap: Long lines wrap to the next line at the width of the editor and no longer scroll horizontally. Break after a space first, and break in the middle if a word cannot fit. The up and down direction keys move according to the display line and maintain the horizontal position; the line number is only marked on the first segment of a line; folding, search, multi-cursor, decoration and diagnostic wavy lines are all drawn according to the display line. Home/End still presses the entire line.
- `TabSize(n, hard)` sets the tab stop width and whether Tab inserts tab characters or spaces; the default width is 4, determined by the existing indentation of the file. `ShowWhitespace(bool)` Use dots and arrows to mark spaces and tabs.
- Search panel: case-sensitive, whole-word matching, regular expressions (can cross lines, use `$1` in replacement), matching items are marked in the text and next to the scroll bar, up to 10,000 statistics are counted. Replace all is an undo. `OpenSearch(replace)`, `CloseSearch`, `SearchMatches`, `Searchable(bool)`; read-only editor can only search.
- `Value`/`SetValue`, `Cursor`/`SetCursor`, `Cursors`, `Selection`, `Lines`, `SetReadOnly`, `SetDisabled`, `Focus`, `Fill`. Multi-cursor editing is counted as one undo.

Agent: Editor role `textbox`, the name is `Name`, the value is the full text (the number of lines when more than 2000 lines); the completion item is independent `option`; the hover prompt is `tooltip`; the search panel role `search`, the input boxes and buttons inside are listed separately; the folding arrow is a button, with a name such as "Collapse 6".

Verification: `go run ./examples/components -section code_editor`.

## Custom search session

SetSearchQuery(query, CodeSearchOptions{MatchCase, WholeWord, Regex}) starts the search and draws matches, without opening the built-in panel or grabbing focus; Searchable(false) only closes the built-in entrance, and you can still use the application search bar. SearchSession Returns a copy of the query, options, panel state, InvalidPattern, Truncated, Current, and Matches; Current starts at 0 and is -1 if no match happens to be selected.

NextSearchMatch and PreviousSearchMatch jump in a loop, and SelectSearchMatch(index) jumps to the specified result and expands and collapses it. ReplaceCurrentSearchMatch(text) only replaces the exactly selected match, and ReplaceAllSearchMatches(text) returns the number of replacements in the entire document. One operation corresponds to one undo and one OnChange. Both substitutions return false/0 when read-only or disabled by itself. CloseSearch ends highlighting; the focus is returned to the editing area only when the built-in panel is closed.

Normal text matches within a line. Regular expressions match the full text and can span lines: `\n`, `\s+`, and `[^;]*` can all skip line breaks; when multi-line mode is turned on, `^` and `$` still match line breaks, but `.` still does not match line breaks. Matches across rows are highlighted segmentally on each row, and selections and replacements are performed on the entire range. Zero-length matches are not supported. The list can be kept up to 10,000 entries, and extra results will make Truncated true; replace all and still traverse all matches. Regular replacement supports `$1`/`${name}` expansion of Go regexp. Large document searches and replace-all are performed simultaneously, and applications should incorporate high-frequency input; this interface is not a background LSP search.

## Trace decoration collection

```go
marks := ed.Decorations(
    kit.CodeDecoration{Range: kit.CodeRange{Line: 2, Col: 0, EndLine: 2, EndCol: 8}, Style: kit.CodeDecorationFill},
    kit.CodeDecoration{Range: kit.CodeRange{Line: 4, Col: 0, EndLine: 4, EndCol: 6}, Style: kit.CodeDecorationFrame},
)
marks.Append(other)
tracked := marks.Get()
marks.Clear()   // Still reusable
marks.Dispose() // The operations on this collection will no longer take effect thereafter.
```

CodeRange uses 0-based line numbers and rune columns, half-open ranges; unlike upstream's UTF-8 byte offsets. Each collection is independent and supports Frame, Fill, Text (foreground color), and Underline; when Color is nil, it follows CodeText, and low transparency is used for filling. Both the color and the returned result are copied. Discarding the handle will not remove the decoration; retaining the handle will also retain the editor reference and release it with Dispose.

Insertion does not expand the range at both ends, and internal insertion expands the range; replacement converges the internal anchor points to the new range, and deletion removes the entire paragraph. Undo/redo also transforms the current position, deleted entries are not resurrected; decorations that need to be restored should be reconstructed from the semantic data. SetValue transforms intermediate changes based on the longest common prefix/suffix, still clearing the edit history.

Visible rows are queried through the interval index, and the folded content is not drawn; update/append rebuilds the sorting, and the editor maintains the index linearly in the existing order. Fills under the border/underline, then draws the selection and text; collections created after the same type and items appended later overwrite the previous style. Decoration does not reserve space or intercept input. Geometric decorations are drawn on the display line and follow soft line breaks. A cross-row Frame is a continuous outline: each row only draws upper and lower edges where the upper and lower adjacent rows are not connected. Cross-line Fill and Frame cover one more space at the end of the line like a selection, so the blank lines in the middle are also connected together; Underline and Text are only drawn where there are glyphs. Text decoration changes the color, and you can also use `Weight` (such as `font.Bold`) and `Italic` to change the font weight and italics; when only setting the font weight or italics and not setting Color, the color of the syntax highlight is retained. The font is positioned according to the regular font width, and the cursor and click positions remain unchanged.

## Language editing rules

```go
err := kit.SetCodeLanguageRules("template", kit.CodeLanguageRules{
    Brackets: []kit.CodePair{{Open: "{{", Close: "}}"}},
    AutoClosingPairs: []kit.CodePair{{Open: "{{", Close: "}}", NotIn: []kit.CodeSyntaxContext{kit.CodeSyntaxString, kit.CodeSyntaxComment}}},
    AutoCloseBefore: ";,}",
    Increase: `\{\{\s*$`, Decrease: `^\s*\}\}`,
})
ed.Language("template").AutoClose(true).SmartIndent(true)
```

SetCodeLanguageRules is registered by language, can be replaced in the same event, and will be used immediately for the next edit. Names recognized by Chroma are assigned to their language names, and unknown names retain their case; ClearCodeLanguageRules returns to default. SetEditingRules(&rules) installs instance overwriting, nil restores the registry; illegal regular, empty/cross-line/more than 64 rune delimiters reject the entire configuration and retain the old rules.

Brackets control the structure indentation of Enter; Brackets are used when AutoClosingPairs is nil, and non-nil empty slices turn off automatic pairing. Input supports multi-character matching, spanning existing end strings, single-line selection wrapping and empty matching Backspace. AutoCloseBefore limits following characters, whitespace and end-of-line are always allowed. NotIn defaults to Chroma's code/string/comment classification, and SyntaxContext can be replaced by the application; unknown languages without syntax classification are treated as Code. Highlighting continues to be provided by Chroma, not Tree-sitter.

Increase/Decrease match on the text before and after Enter respectively, without formatting the existing line or pasting. When no rules are provided, indentation is based on structural brackets. By default, Python also recognizes end-of-line colons. AutoClose and SmartIndent are independent; turning off SmartIndent still copies the current line's leading whitespace, but does not increase/split the indent. Language switching does not reset these two preferences.

`OnPaste(func(core.ClipboardData) bool)` Deliver text, encoded image, and file path before text insertion; true means the application has received it, false to continue with normal text pasting. `PasteReader(core.ClipboardReader)` Configures asynchronous rich clipboard reading; nil uses the Gio text path. When the native read fails, `OnPasteError` is used to report and the text is rolled back. If the text fails to be read or exceeds 16MiB, the paste is rejected.

If the document version, any selection, or the main cursor changes during reading, the result will not be inserted; disabling/read-only will cancel the waiting request, and continuous pasting will only accept the latest request. The default insertion continues to use the multi-cursor paste rules and the same undo record. If the callback modifies the document or selection by itself, the default insertion will no longer be executed. The callback is executed on the UI thread; when the platform read is completed, the component dispatches it back to the UI.

The component library has reused the native/clipboard adaptation of the Input example. Images and files for macOS, Windows, and Linux (Wayland and X11) are handed over to the example callback, and the text is rolled back when the reading fails. Automatic testing covers multi-cursor paste and overall undo, file consumption, document/selection change rejection, native failure fallback, read-only and delayed arrival of expired text. The macOS native bridge has actually read the picture snapshot; the picture/file pasting of the real window is still pending acceptance.
