# 在浏览器里运行

`ui/*` 可以编译成 WebAssembly，同一份代码在浏览器里运行。`native/*` 调用的是系统 API，Web 上没有。

## 构建

用 Gio 的 `gogio` 工具打包，它会生成 `index.html`、`wasm.js` 和 `main.wasm`：

```sh
go run gioui.org/cmd/gogio@latest -target js -tags osusergo -o web ./examples/hello
cp /path/to/NotoSansSC-Regular.ttf web/font.ttf
python3 -m http.server --directory web
```

两点必须做：

- **`-tags osusergo`**：Go 1.26 的 `os/user` 在 js/wasm 下没有实现，Gio 的字体扫描引用了它。不加这个标签就编译不过。`internal/deps` 的 `TestUIBuildsForWebAssembly` 会检查这一点。
- **提供中文字体**：浏览器不把系统字体交给 WebAssembly 程序，没有中文字体时中文显示成方框。用 `theme.LoadFonts(data)` 加载字体文件，族名最好是 `theme.Face` 里列出的，比如 Noto Sans SC。`theme.FetchFonts("font.ttf")` 在浏览器里下载字体文件并加载，`examples/hello/font_js.go` 就是这样做的。中文字体通常有十几 MB，建议用子集化后的字体。

## 和桌面版的差别

- 没有 `native/*`：权限、截屏、全局快捷键、合成输入都不可用。
- 只有一个窗口，`window.Open` 多次打开时都画在同一个页面里。
- 自动化模式（`KEEL_AUTOMATION`）和 keel-mcp 只在桌面版可用。

## 文档站

Keel 的文档站就是这样发布的：`examples/components` 编译成 WebAssembly，嵌进每个组件的文档页，和 [GPUI Kit](https://gpui-kit.com) 的做法一样。组件库解压后约 43 MB，GitHub Pages 用 gzip 传输，约 10 MB，之后由浏览器缓存。

在浏览器里，组件库从网址读参数：`demo/?section=dock` 只显示 Dock，`&theme=dark` 用深色主题。中文字体是子集化的 Noto Sans SC，约 1.9 MB。

生成器和本地预览方法见 [internal/site](../internal/site/README.md)，发布流程是 `.github/workflows/site.yml`。
