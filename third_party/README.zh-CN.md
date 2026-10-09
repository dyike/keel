# third_party

[English](README.md) | 简体中文

Keel 自带它绘制界面所用的库的副本，这样可以直接修复和优化，不必等上游发版；应用只要依赖 Keel 就能拿到这些修改（Keel 的 `go.mod` 里写 `replace` 传不到应用）。

| 目录 | 上游 | 基线 | 许可 |
| --- | --- | --- | --- |
| `gio/` | [Gio](https://gioui.org) `gioui.org` | v0.10.3 | Unlicense 或 MIT（`gio/LICENSE`） |
| `gio/cmd/gogio/` | Gio 的打包器 `gioui.org/cmd/gogio` | v0.10.0 | Unlicense 或 MIT |
| `typesetting/` | [go-text](https://github.com/go-text/typesetting) | v0.3.5 | Unlicense 或 BSD-3-Clause（`typesetting/LICENSE`） |

`gioui.org/shader` 仍是普通依赖：只有预编译的着色器，没有出现在任何 API 里的类型。

导入路径是 `github.com/dyike/keel/third_party/gio/...` 和 `github.com/dyike/keel/third_party/typesetting/...`。之前直接导入上游包的应用运行一次 `keel migrate`。

## 规则

- 这两个目录保持上游风格；只改 Keel 需要的，不做纯格式修改。每处修改都记在下面并写明原因，换到更新的上游时照着重做。
- 不复制上游的测试和测试数据。补丁自带测试，放在补丁旁边。
- 升级上游：复制发布版本，去掉 `*_test.go` 和 `testdata`，改写导入路径（`gioui.org/`，`gioui.org/shader` 除外；`github.com/go-text/typesetting/`），再重做下面的补丁。

## 补丁

- 导入路径改为本模块；修复 `gio/internal/f32` 和 `gio/app/internal/ibus` 里 `go vet` 报的无字段名结构体字面量。
- `gio/cmd/gogio`：查找 `github.com/dyike/keel/third_party/gio/app` 而不是 `gioui.org/app`。`keel build` 从应用所依赖的 Keel 模块编译它。
- `typesetting/font/opentype`：`Shared` 资源（`NewShared`）给出的表是切片而不是副本，也从不写入调用方传入的缓冲区。`typesetting/fontscan` 以只读方式映射系统字体文件（`mmap`、`MapViewOfFile`）并按 `Shared` 解析，字体的表是可回收的文件页，不占 Go 堆：hello 在 macOS 上空闲时的内存占用从 153 MB 降到 103 MB。`gio/font/opentype.ParseCollectionShared` 对不会变的字节做同样的事，`gio/font/gofont` 和 Keel 的主题使用它。测试用只读映射解析、描述、排版并取轮廓全部系统字体（`fontscan/openfont_test.go`）。
- `typesetting/harfbuzz`：附着链指向缓冲区之前时像 HarfBuzz 一样直接返回，不再访问 `pos[-1]`；原来排版 macOS 的 Farisi.ttf 会崩溃。
