# ui/highlight

English | [简体中文](README.zh-CN.md)

Code highlighting, optional modules. The syntax coloring of CodeEditor, TextView and Markdown code blocks all comes from here; when not introduced, the code is displayed as plain text and other functions are not affected.

```go
import _ "github.com/dyike/keel/ui/highlight"
```

- **Why is it packaged separately**: It uses [chroma](https://github.com/alecthomas/chroma), built-in rules for hundreds of languages, all registered at startup, cannot be deleted by the linker, and occupies about 4 MB. Applications that do not display codes do not need to be brought.
- **Dependencies**: `ui/core` (implements `core.Highlighter` and `core.SetHighlighter` in `init`) and chroma.
- The language name supports aliases and extensions (`go`, `golang`, `Go`); the default color matching is `github` / `github-dark`, which can be selected according to the depth of the code background. Markdown can use `markdown.CodeStyle` to specify the chroma style name.

To change to another highlighting implementation, implement `core.Highlighter` and call `core.SetHighlighter`.
