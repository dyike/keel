# ui/theme

English | [简体中文](README.zh-CN.md)

Color, size scale, font and system dynamics preferences. For user usage and parameter list, see [Topic](../../docs/theme.md).

| Documentation | Responsibilities |
| --- | --- |
| `theme.go`, `gofonts.go` | Default text style, Gio Material theme and pocket font |
| `palette.go` | Palette copy, Apply / Scope, version number and semantic color |
| `registry.go`, `themes/` | Built-in theme, registry, JSON parsing |
| `watch.go` | Theme directory polling and hot reloading of the current theme |
| `scale.go` | Spacing, corner rounded corners, font size and shading scale |
| `fonts.go`, `fonts_*.go`, `fetch_*.go` | Font loading, Android system CJK fonts and browser font download |
| `text_paint.go` | `GlyphPainter` reuses vector fragments for caller-shaped single-line text; bounded caches, whole-run fallback for complex glyphs |
| `glyph_atlas.go` | Opt-in glyph image pages with bounded masks and page storage; Metal shares coverage across colors, other builds retain colored pages; prepare before painting and release on close |
| `motion.go` | The system reduces animation and application overlay values, scroll bar preferences |

Depends on Gio and `ui/internal/loop`. `Apply` retains the Material pointer, font, and typesetter, synchronizes the palette, increments `Revision()`, and requests a redraw of all windows; it does not acquire the frame lock itself. `Scope` temporarily switches the color without redrawing or incrementing the version number.

`el`, `kit`, `window`, `markdown` Read topics. Colors and version numbers are read and written according to [threading rule](../../docs/architecture.md#threading), component caching uses `cx.Cache` or the theme version is included in the key.

Verification entrance: `go run ./examples/components -section theme -theme dark`.
