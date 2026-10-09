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
- `gio/gpu`：裁剪路径按内容缓存（路径数据的哈希，加上描边宽度、轮廓标志和变换），而不是按记录它的 op；后者在 `Ops` 重置后就变了。Keel 每帧重新记录，所以每个边框和圆角图形每帧都要重新细分并上传。偏移不在键里，移动的内容（如滚动）也能复用路径。动画中的组件库每 10 秒分配从 881 MB 降到 470 MB（GC 从 51 次降到 29 次）。所有静态组件截图逐字节一致。
- `gio/app`（macOS、iOS）：`displayLink.Start` 和 `Stop` 只在请求的状态改变时才发给显示链接的 goroutine。窗口在每个动画帧都调用 `Start`，原来这个无缓冲发送每次都让主线程等一次 goroutine 交接。
- `gio/app`（macOS）：窗口的 `CAMetalLayer` 最多保留两个 drawable。渲染器在下一帧前会等待上一帧完成，第三个只是多占一块窗口大小的表面（640x512 pt 时 5 MB，4K 时 33 MB）；hello 空闲时的内存在各次运行之间因此相差 5 MB。
- `gio/gpu`、`gio/gpu/internal/metal`：裁剪为普通矩形、填充为纯色或纹理的连续 op 合并为一次实例化绘制（`driver.QuadBatcher`，每批最多 8 个纹理），不再一个 op 一次绘制调用：终端大小的文字网格原来每帧 3,300 次绘制。Metal 着色器在运行时编译，计算与 Gio 的 blit 着色器相同；其他后端保持逐个绘制。所有静态组件截图逐字节一致。
- `typesetting/harfbuzz`：AAT 排版（`morx`、`kerx`，Menlo 等苹果字体使用）把每个子表的字形类别缓存留在字体的加速器上，与 HarfBuzz 一致；上游把它复制进每次调用的上下文，填好后就丢弃。上下文和子表驱动的状态放在复用的 `Buffer` 里，不再每次排版分配 1.3 KB 以上。239 个带 `morx` 的 macOS 字体排版结果完全一致；benchmark 网格场景的文字分配减半。
- `typesetting/fontscan`：`FontMap.ResolveFace` 对 ASCII 字符查当前查询的表，查询的字体族哈希每次 `SetQuery` 只算一次，不再每个字符算一次。查询或文字系统没变时，`SetQuery` 和 `SetScript` 保留缓存；Gio 排版每段文字都会调用这两个方法。测试检查带缓存的和全新的 FontMap 解析出相同的字体。
- `gio/app`（macOS）：没人看得见的窗口（被遮住、在别的桌面、锁屏或屏幕休眠时）在第一帧之后不再绘制，并停掉显示链接；再次可见时重画（`windowDidChangeOcclusionState:`）。AppKit 只报告状态变化，所以视图挂到窗口上时也读一次状态。Gio 原来对隐藏的窗口照样画每一帧动画：benchmark 网格场景锁屏时每秒用 0.49 秒 CPU，现在 0.001 秒。
- `gio/app`（macOS）：所有显示器都休眠时，创建显示链接失败，随后创建窗口在 `gio_onDestroy` 里崩溃（视图还没有句柄）。现在显示链接退回到主显示器，没有句柄的视图跳过 `gio_onDestroy`。
