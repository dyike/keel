# ui/internal/imageload

[English](README.md) | 简体中文

`ui/markdown` 的图片：调用 `core.DecodeImage` 解析来源（文件路径、`file://`、`data:`、`http(s)://`），在后台 goroutine 加载解码，限制文件大小和像素数，加载中和失败时画占位。

- **依赖**：`ui/core`、`ui/theme`、`ui/locale`、`ui/internal/loop`。
- **被谁依赖**：`ui/markdown`（公开为 `markdown.ImageLoader` 和 `markdown.DecodeImage`）。

`Load` 返回 `*Asset`，加载完成后通过 `core.Update` 写回并重绘；`Revision` 变化表示状态变了。`View` 画图片，Agent 看到的角色是 `image`，值是 loading / loaded / error。kit 的 `Image` 直接接收 `image.Image`，不经过这里。
