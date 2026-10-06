# 在浏览器里运行

[English](web.md) | 简体中文

`ui/*` 可以编译成 WebAssembly，同一份代码在浏览器里运行。`native/*` 调用的是系统 API，Web 上没有。

## 构建

脚手架项目使用 `keel build -target js` 构建浏览器版本。它已处理 Go 1.26 的 `osusergo` 标签，并通过 Gio 的打包工具生成 `index.html`、`wasm.js` 和 `main.wasm`；应用继续使用脚手架的入口和 `app.go`。

浏览器不把系统字体交给 WebAssembly，中文字体需要由应用加载。在项目中新增 `font_js.go`（只在 js 平台编译）：

```go
//go:build js

package main

import (
    "log"

    "github.com/dyike/keel/ui/theme"
)

func init() {
    if err := theme.FetchFonts("font.ttf"); err != nil {
        log.Print(err)
    }
}
```

在项目目录中构建、复制字体并启动静态服务器：

```sh
keel build -target js
cp /path/to/NotoSansSC-Regular.ttf dist/web/font.ttf
python3 -m http.server --directory dist/web
```

族名最好是 `theme.Face` 列出的字体，例如 Noto Sans SC；中文字体通常有十几 MB，建议使用子集化后的字体。也可以用 `theme.LoadFonts(data)` 加载嵌入的字体文件。

## 和桌面版的差别

- 没有 `native/*`：权限、截屏、全局快捷键、合成输入、系统通知、富剪贴板都不可用。
- 只有一个窗口，`window.Open` 多次打开时都画在同一个页面里。
- 自动化模式（`KEEL_AUTOMATION`）和 keel-mcp 只在桌面版可用。

## 文档站

Keel 的文档站就是这样发布的：`examples/components` 编译成 WebAssembly，嵌进每个组件的文档页，和 [GPUI Kit](https://gpui-kit.com) 的做法一样。组件库解压后约 43 MB，GitHub Pages 用 gzip 传输，约 10 MB，之后由浏览器缓存。

在浏览器里，组件库从网址读参数：`demo/?section=dock` 只显示 Dock，`&theme=dark` 用深色主题。中文字体是子集化的 Noto Sans SC，约 1.9 MB。

生成器和本地预览方法见 [internal/site](../internal/site/README.zh-CN.md)，发布流程是 `.github/workflows/site.yml`。
