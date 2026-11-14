# internal/gallerybuild

[English](README.md) | 简体中文

构建完整 Components WebAssembly 应用及独立链接的分类应用。只依赖 Go 标准库，调用 Keel 的 gogio CLI。请在仓库根目录执行：

```sh
go run ./internal/gallerybuild -out demo
# 开发时仅构建完整应用和指定分类。
go run ./internal/gallerybuild -out /tmp/demo -categories general,inputs
```

AST 扫描器读取使用字面量的 `registerSection(name, category, builder)` 调用。临时 `-overlay` 源码保留选中分类的构造函数，将其他注册改为仅保留元数据。原调用置于恒假的分支中，使导入仍合法，同时不再成为链接保留入口。不修改工作区文件或模块清单。构建失败时，不发布新的 `groups.json` 清单。

每个分类有自己的运行时。这减少首次分类下载，但跨分类会重复下载公共代码，切换程序会重置应用状态。它不是共享运行时的动态链接。站点发布器读取 `groups.json`，生成哈希文件名、共享字体地址，并通过 `?bundle=all` 提供完整程序。CI 构建所有分类，分类子集参数用于开发。

没有 Markdown/代码示例的浏览器分类使用内部 `keel_gallery_small` 标签排除语法高亮，桌面版本保留原提供者。示例增加语法高亮需求时，请更新分类白名单。

测试覆盖注册解析和裁剪。请构建全部分类来验证实际编译、链接兼容性。原始/压缩大小与冷/热缓存启动时间应分别测量；分包本身不能证明总下载量降低或帧率提升。
