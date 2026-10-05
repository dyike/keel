# internal/site

生成 Keel 的文档站：`docs/` 下的全部文档和各模块、示例的 README 转成静态 HTML，组件页嵌入在线组件库（`examples/components` 编译成的 WebAssembly）。由 `.github/workflows/site.yml` 在每次推送到 main 后构建并发布到 GitHub Pages。

本地预览：

```sh
go run gioui.org/cmd/gogio@v0.10.0 -target js -tags osusergo -o /tmp/demo ./examples/components
python3 internal/site/subset_font.py . NotoSansSC-Regular.otf /tmp/demo/font.otf
go run ./internal/site -out _site -demo /tmp/demo
python3 -m http.server --directory _site
```

不带 `-demo` 时只生成文档，组件页里的在线示例无法加载。`-version v0.0.5` 在页头显示版本号并链接到 GitHub 的版本说明；CI 取最近的 `v*` 标签，推送新标签时也会重新发布。

`docs/reports/` 下的进度报告不发布：它们是工作记录，不是文档。其他文档里指向它们的表格行或列表项，在站点上会一并去掉。

导航直接读取 Markdown 索引：`docs/README.md` 的二级标题决定指南分组与阅读顺序，`docs/kit.md` 的二级标题决定组件分类。源码和示例 README 归入“源码导览”。新增文档必须加入相应索引；遗漏组件或重复分类会使构建失败。侧栏默认展开“开始使用”和当前页面所在分类。

快速开始统一说明脚手架开发、运行与打包，旧的 `docs/cli.html` 保留跳转，不占用导航或搜索入口。每个组件页在在线展示下方提供可展开、可复制的完整示例代码，构建时直接读取注册该 section 的源码文件，避免维护第二份代码。

正文底部的“上一页 / 下一页”沿用侧栏顺序，包含组件分类和源码导览。第一篇只显示下一页，最后一篇只显示上一页，站点首页不参与翻页。窄屏上两个入口上下排列。

| 文件 | 内容 |
| --- | --- |
| `main.go` | 收集页面、生成站点、搜索索引、相对链接 |
| `navigation.go` | 从文档索引生成分组导航，检查遗漏和重复分类 |
| `render.go` | Markdown 渲染：链接改写、与 GitHub 一致的中文锚点、chroma 代码高亮 |
| `demo.go` | 发布组件库，带下载进度的加载页 |
| `subset_font.py` | 中文字体子集：仓库里出现过的字加 GB2312 常用字 |
| `assets/` | 页面模板、样式、搜索和复制按钮脚本 |

`go test ./internal/site` 会生成整站，检查站内链接、锚点和分类覆盖，并验证当前分类的展开行为与翻页顺序。
