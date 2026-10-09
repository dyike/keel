# third_party

[English](README.md) | 简体中文

Keel 自带它绘制界面所用的库的副本，这样可以直接修复和优化，不必等上游发版。副本保留原来的模块路径，通过 Keel `go.mod` 里的 `replace` 接入：

```
replace (
	gioui.org => ./third_party/gio
	github.com/go-text/typesetting => ./third_party/typesetting
)
```

| 目录 | 上游 | 基线 | 许可 |
| --- | --- | --- | --- |
| `gio/` | [Gio](https://gioui.org) `gioui.org` | v0.10.3 | Unlicense 或 MIT（`gio/LICENSE`） |
| `typesetting/` | [go-text](https://github.com/go-text/typesetting) `github.com/go-text/typesetting` | v0.3.5 | Unlicense 或 BSD-3-Clause（`typesetting/LICENSE`） |

`gioui.org/shader` 和打包器 `gioui.org/cmd/gogio` 仍用上游。

`replace` 只对写它的模块生效，模块压缩包也不包含嵌套模块。所以：

- Keel 自己的构建、测试和示例使用副本。
- 用 `keel new -replace <Keel 检出目录>` 创建的项目会得到同样的 `replace`。其他应用要使用副本，需自行加上这几行并指向 Keel 检出目录；不加则基于上游 Gio 和 go-text 构建。
- 因此 `third_party` 之外的 Keel 代码也必须能基于上游构建：补丁不新增 Keel 调用的 API，或只用上游原样接受的方式（如下面的 `Bytes` 方法）。检查方法：去掉 `replace` 后构建，`go mod edit -dropreplace gioui.org -dropreplace github.com/go-text/typesetting && go build ./...`，再恢复 `go.mod`。

## 规则

- 这两个目录保持上游风格；只改 Keel 需要的，不做纯格式修改。每处修改都记在下面并写明原因，换到更新的上游时照着重做。
- 不复制上游的测试和测试数据。补丁自带测试，放在补丁旁边。每个副本是独立模块，在它的目录里测试（`cd third_party/typesetting && go test ./...`）。
- 升级上游：复制发布版本，去掉 `*_test.go` 和 `testdata`，保留 `go.mod`（gio 的 `go.mod` 还把 go-text 替换为 `../typesetting`），再重做下面的补丁。

## 补丁

- 修复 `gio/internal/f32` 和 `gio/app/internal/ibus` 里 `go vet` 报的无字段名结构体字面量。
- `typesetting/font/opentype`：能给出自身字节的资源（`Shared`，或任何有 `Bytes() []byte` 方法的资源，这样调用方也能基于上游构建）给出的表是切片而不是副本，也从不写入调用方传入的缓冲区。`typesetting/fontscan` 以只读方式映射系统字体文件（`mmap`、`MapViewOfFile`）并按 `Shared` 解析，字体的表是可回收的文件页，不占 Go 堆：hello 在 macOS 上空闲时的内存占用从 153 MB 降到 103 MB。`gio/font/opentype.ParseCollectionShared` 对不会变的字节做同样的事，供 `gio/font/gofont` 使用；Keel 的主题传入带 `Bytes` 方法的字体文件。测试用只读映射解析、描述、排版并取轮廓全部系统字体（`fontscan/openfont_test.go`）。
- `typesetting/harfbuzz`：附着链指向缓冲区之前时像 HarfBuzz 一样直接返回，不再访问 `pos[-1]`；原来排版 macOS 的 Farisi.ttf 会崩溃。
- `gio/gpu`：路径渲染器的覆盖纹理（stencil 和 intersection 帧缓冲，打包了所有圆角图形，常常比窗口还大）每帧都会重画，所以一帧提交后把它们标为可回收（`driver.Volatile`，即 Metal 的 purgeable 状态），下一帧调整尺寸、准备使用时再标回。窗口空闲时最大的几块 GPU 内存因此不计入占用：hello 在 macOS 上从 72 MB 降到 64 MB。其他后端暂未实现。
动画中的窗口若每帧都标记，就要每帧付一次内核调用，所以只在某一帧之后 100 ms 内不再需要新帧时才标记（`app.Window` 调用 `gpu.Idle`，判断用不会消耗唤醒标记的 `input.Router.PeekWakeup`）。每个纹理记住自己的状态，状态不变时不做调用。
- `gio/gpu/internal/metal`：GPU 缓冲按 2 的幂大小分级复用，不再每帧创建、释放（动画窗口的 CPU 有五分之一花在 `newBuffer` 和 `CFRelease` 上）。一帧中释放的缓冲要等到下一次 `BeginFrame` 等待上一个命令缓冲完成后才复用；缓冲池有上限（32 MB，每级 64 个），`gpu.Idle` 时清空。
