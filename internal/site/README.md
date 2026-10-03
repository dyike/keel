# internal/site

生成 Keel 的文档站：`docs/` 下的全部文档和各模块、示例的 README 转成静态 HTML，组件页嵌入在线组件库（`examples/components` 编译成的 WebAssembly）。由 `.github/workflows/site.yml` 在每次推送到 main 后构建并发布到 GitHub Pages。

本地预览：

```sh
go run gioui.org/cmd/gogio@v0.10.0 -target js -tags osusergo -o /tmp/demo ./examples/components
python3 internal/site/subset_font.py . NotoSansSC-Regular.otf /tmp/demo/font.otf
go run ./internal/site -out _site -demo /tmp/demo
python3 -m http.server --directory _site
```

不带 `-demo` 时只生成文档，组件页里的在线示例无法加载。

`docs/reports/` 下的进度报告不发布：它们是工作记录，不是文档。其他文档里指向它们的表格行或列表项，在站点上会一并去掉。

| 文件 | 内容 |
| --- | --- |
| `main.go` | 收集页面、导航分组、搜索索引、相对链接 |
| `render.go` | Markdown 渲染：链接改写、与 GitHub 一致的中文锚点、chroma 代码高亮 |
| `demo.go` | 发布组件库，带下载进度的加载页 |
| `subset_font.py` | 中文字体子集：仓库里出现过的字加 GB2312 常用字 |
| `assets/` | 页面模板、样式、搜索和复制按钮脚本 |

`go test ./internal/site` 会生成整站，并检查站内每个链接和锚点都存在；文档里写错链接时测试失败。
