# ui/theme

[English](README.md) | 简体中文

颜色、尺寸刻度、字体和系统动态效果偏好。用户用法与参数表见 [主题](../../docs/theme.zh-CN.md)。

| 文件 | 职责 |
| --- | --- |
| `theme.go`、`gofonts.go` | 默认文字样式、Gio Material 主题与兜底字体 |
| `palette.go` | 调色板副本、Apply / Scope、版本号与语义色 |
| `registry.go`、`themes/` | 内置主题、注册表、JSON 解析 |
| `watch.go` | 主题目录轮询和当前主题热重载 |
| `scale.go` | 间距、圆角、字号和阴影刻度 |
| `fonts.go`、`fonts_*.go`、`fetch_*.go` | 字体加载、Android 系统 CJK 字体与浏览器字体下载 |
| `text_paint.go` | `GlyphPainter` 为已经排版的单行文字复用矢量片段；缓存有界，复杂字形回退到整段绘制 |
| `glyph_atlas.go` | 按需启用字形图片图集；Metal 下各颜色共享覆盖掩码，其他构建保留彩色页面；缓存有界，绘制前准备，关闭时释放 |
| `motion.go` | 系统减少动画与应用覆盖值、滚动条偏好 |

依赖 Gio 和 `ui/internal/loop`。`Apply` 保留 Material 指针、字体和排版器，同步调色板、递增 `Revision()`，并请求所有窗口重绘；它不自行取得帧锁。`Scope` 临时切换颜色，不重绘、不递增版本号。

`el`、`kit`、`window`、`markdown` 读取主题。颜色和版本号的读写遵守 [线程规则](../../docs/architecture.zh-CN.md#线程规则)，组件缓存使用 `cx.Cache` 或把主题版本纳入键。

验证入口：`go run ./examples/components -section theme -theme dark`。
