# ui/internal/imageload

English | [简体中文](README.zh-CN.md)

Image loading for `ui/markdown`: `core.DecodeImage` resolves file paths, `file://`, `data:`, and `http(s)://` sources. A background goroutine loads and decodes the image with byte and pixel limits; the view draws placeholders while loading or after failure.

- **Dependencies:** `ui/core`, `ui/theme`, `ui/locale`, and `ui/internal/loop`.
- **Used by:** `ui/markdown`, exposed as `markdown.ImageLoader` and `markdown.DecodeImage`.

`Load` returns an `*Asset`. On completion, `core.Update` applies the result and requests a redraw; a change in `Revision` indicates a state change. `View` draws the image. Agents see role `image` and value loading / loaded / error. The kit `Image` also accepts an `image.Image` directly without using this package.
