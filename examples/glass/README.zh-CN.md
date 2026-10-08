# Glass Studio

[English](README.md)

运行 macOS 原生玻璃示例：

```sh
go run ./examples/glass -backdrop
```

侧栏和顶部透出原生背景，主内容区保持实色。可以点击导航、切换明暗外观，或打开 Clear、毛玻璃对比窗口。`-style clear`、`-style frosted`、`-dark`、`-opaque` 控制启动状态；`-backdrop` 在背后打开彩色参考窗口。

Liquid Glass 需要 macOS 26+ 和 macOS 26+ SDK；旧版 macOS 或 SDK 使用传统毛玻璃，其他平台保留实色背景。原生效果需在桌面窗口查看，离屏截图无法包含。API 见[窗口文档](../../docs/app.zh-CN.md#macos-玻璃背景)。
