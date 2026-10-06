# Chat example

English | [简体中文](README.zh-CN.md)

The preset answer simulates AI streaming output and supports stopping, message area follow-through and Markdown interactive verification. No model API required.

Start from the root directory of the repository. The picture example uses the relative path within the repository:

```sh
go run ./examples/chat
```

There are six samples listed on the blank page. Click on any one to display them in a streaming format. You can also send the following keywords in the input box.

| Sample name | Keywords | Checking method |
| --- | --- | --- |
| `selection` | Select, span, drag and select | Drag and select across titles, paragraphs, quotes, lists, codes, and tables; Cmd/Ctrl+A selects all current answers, Cmd/Ctrl+C copies; scroll up during output to select text, and check whether the selection is retained after appending and ending |
| `math` | Formulas, Mathematics, LaTeX | View fractions, radicals, subscripts, upper and lower summation bounds, nested matrices, spanning macros, and auto-scaling brackets for inline and independent formulas; dollar signs in code should remain intact |
| `code-scroll` | Long code, long lines, horizontal direction | View "plain text" and language titles, icon hover prompts and copy feedback; use the trackpad to slide horizontally, or drag, click, or scroll the wheel to the END mark on the bottom scroll bar to switch to automatic word wrapping; each block is set independently, and long lines should not be inserted into line breaks after copying |
| `click-selection` | Double-click, triple-click, word selection, paragraph selection | Double-click on Chinese and English words, words next to punctuation, and `rendering` across bold boundaries; triple-click on automatically wrapped paragraphs, copy and check boundaries |
| `references` | Footnotes, citations, links | Click across blocks for complete, collapsed, and abbreviated citations; view repeated citations, multiple footnotes, and return locations; same-block links as controls |
| `images` | Image | View local PNG's three color blocks, scale, linked image, empty alt text, and load failure placeholder |
| `table` | Table | View column alignment, cross-cell selection and copy |
| `downloader` | Other News | View the original Concurrent Downloader answer, including code, tables, quotes, and task list |
| `all` | All, TODO, verification examples | View the first six examples in one answer |

Cross-paragraph selection, double-click word selection, triple-click paragraph selection, automatic scrolling by edge drag selection, commonly used mathematical formulas, horizontal scrolling of code blocks, and line wrapping switching have been implemented. Footnote jumps and returns, cross-block references, asynchronous pictures, mathematical macros, nested matrices and telescopic brackets have also been implemented, and are checked according to the actions in the table.

You can directly display the complete sample, start browsing from the top, and skip waiting for streaming:

```sh
go run ./examples/chat -sample=all
go run ./examples/chat -sample=math
go run ./examples/chat -sample=images
```

`-sample` Use the sample name from the first column of the table. View fast or slow streaming processes:

```sh
go run ./examples/chat -delay=0
go run ./examples/chat -delay=80ms
```

"Copy full text" copies the Markdown source text; press Cmd/Ctrl+C after selecting the text to copy plain text; the code block button only copies the block of code. Clicking the link invokes the macOS `open` command.

```sh
go test ./examples/chat -count=1
```

The test checks keyword routing, coverage of all samples, block reference link comparison, and PNG file validity. Renderer selection and copy regression tests are in `ui/markdown/selection_test.go` and `ui/markdown/selection_units_test.go`, code block line wrapping, horizontal scrolling, post-scroll selection and streaming state tests are in `ui/markdown/code_test.go`; formula parsing, baseline, line height, streaming fallback and source code copy tests are in `ui/markdown/math_test.go` and `math_extensions_test.go`; footnote and reference tests are in `references_test.go`, image loading and layout tests are in `references_test.go` `images_test.go`, `ui/internal/imageload/decode_test.go`.
