# internal/gallerybuild

[English](README.md) | 简体中文

默认把 Components Gallery 构建成**一个 Go 运行时与细粒度按需叶子模块及共享依赖模块**。在仓库根目录运行，需要 Python 3 和 Binaryen **version_133**（仅构建工具；应用构建设置 `CGO_ENABLED=0`）：

```sh
BINARYEN_DIR=/path/to/binaryen/bin go run ./internal/gallerybuild -out /tmp/demo-new
# 等价参数：-binaryen-dir /path/to/binaryen/bin
# 仅依赖 Go 的回退：完整应用及独立分类应用。
go run ./internal/gallerybuild -mode independent -out /tmp/demo-independent
# 开发用分类子集（隐式选择 independent）。
go run ./internal/gallerybuild -categories general,inputs -out /tmp/demo-subset
```

共享模式输出目录必须为空或不存在。构建在临时目录完成链接器归属分析、拆分，以及完整回退包和所有共享模块的同等 `-Oz --strip-debug` 优化。最终只复制带内容哈希的运行资源、`sizes.json` 和最后写入的 `shared-build.json`，排除分析中间文件和独立分类清单。Binaryen 缺失、版本错误或任何阶段失败都会报错，不会静默改变模式。随后提供 `font.otf`，运行 `go run ./internal/site -demo /tmp/demo-new -out /tmp/site-new` 发布站点。站点的 `?bundle=all` 使用同等优化的完整回退包。

AST 扫描器识别字面量形式的 `registerSection(name, category, builder)`。临时 `-overlay` 源码保留选定分类的构建器，把其他注册替换为仅元数据条目；恒假分支保持导入有效但不保留链接器根。不会修改源码和模块清单。`-analysis` 保留 Go 符号、统一可选导入并生成不含示例构建器的基线；默认流水线内部使用它。分析模式拒绝分类子集。

独立模式每个分类各有一个运行时。内部 `keel_gallery_small` 白名单在无 Markdown/代码示例的分类中省略语法高亮；示例变化时需要更新名单。普通桌面构建不受影响。

CI 在 Pages 发布前检查归属和加载器测试、全部组件首帧、输入保留、回退、重试/历史导航，以及受控 Chrome 加载性能。阈值、复现方式和部署限制见[共享 WASM](../../docs/shared-wasm.zh-CN.md)。压缩体积本身不能证明实际启动或内存改善。
