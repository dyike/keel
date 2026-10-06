# ui/markdown

English | [简体中文](README.zh-CN.md)

Render Markdown into `ui/el` elements, optimized for streaming output of AI chat: ordinary documents only re-parse the block being written; when containing footnotes, reference definitions or macros, the entire parsing context is shared, unclosed syntax is temporarily completed, and completed blocks reuse elements and layout.

- **Dependencies**: `el`, `core`, `theme`, `locale`, as well as internal `ui/internal/imageload` (picture) and `ui/internal/editorstyle` which is indirectly dependent on el; third party: goldmark (parsing), golang.org/x/net/html (HTML), chroma (code highlighting), Gio text.Shaper (font typesetting).
- **Used by**: Application code.

| File | Responsibility |
| --- | --- |
| `markdown.go` | `Doc`: slicing, incremental parsing, streaming completion |
| `plugins.go` | Document-level Goldmark extension, block view and inline atomic control factory |
| `parse.go` | goldmark syntax tree → intermediate structures (paragraphs, code blocks, lists, tables...) |
| `render.go` | Intermediate structure → el element; rich text, code highlighting, block caching |
| `code_extensions.go` | Code block operation slot and replacement display by language, retain the source document |
| `code.go` | Code cards, language and action icons, hover prompts, line breaks and horizontal scrolling |
| `math_parse.go` | Math delimiters and common TeX subset analysis, source code rollback |
| `math_more.go` | Extended TeX: more symbols and functions, mathematical alphabet, accents, binomials, brackets, colors, boxes, more environments |
| `html.go` | Inline HTML tag style, HTML block conversion to Markdown block |
| `math_macros.go`, `math_structures.go` | Document macros, nested matrices and pair delimiters |
| `math_delimiters.go` | Automatic scaling separator drawing |
| `references.go` | Document-level parsing, footnote jump and return |
| `images.go` | Asynchronous image resource sharing and layout cache invalidation |
| `math_layout.go` | Formula box layout, fractional radicals, superscripts and subscripts and matrices |
| `text.go` | Rich text layout, glyph coordinates, decorations, links and selection drawing |
| `stream_fade.go` | Incremental text and style fade-in, independent clip timing and reduced motion |
| `preview.go` | Full line height budget, full line cropping and truncation status |
| `ranges.go` | Rendered text snapshots, UTF-8 range highlighting, change migration and minimal vertical positioning |
| `selection.go` | Document coordinates, cross-block selection, automatic edge scrolling, entire selection and plain text copy |
| `selection_units.go` | Unicode word selection, three-click selection, formulas and code line boundaries |

Usage and Design: [Markdown](../../docs/markdown.md).

Atomic input references are connected to `ui/internal/inputcontent` by `el.InputDocument`, which only saves text, reference range, selection and editing transactions, and does not rely on Gio or other Keel modules. kit and markdown only depend on it indirectly via el.
