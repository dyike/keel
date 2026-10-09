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

- `gio/gpu/internal/metal`：空闲清理等待最后提交的 GPU 工作完成，再释放命令缓冲区、上传/读回暂存缓冲区和四边形实例缓冲区。仍在使用的纹理和管线保留，临时缓冲区在下次绘制或读回时重建，降低空闲常驻内存而不请求新帧。无窗口回归测试覆盖重复清理、读回和恢复绘制。

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

- `gio/op/paint`、`gio/gpu`、`gio/gpu/internal/metal`：Metal 构建提供可选的 `ImageOp.AddTinted` 扩展，将不可变图片与每次绘制的线性颜色相乘。Keel 字形图集只保存一份白色覆盖掩码，不再为每种文字颜色复制像素（benchmark 网格从六张 1 MiB 页面降为一张）。Metal 在上传纹理前预留整帧的实例缓冲，避免共享纹理在绘制过程中退回不支持着色的路径。路径裁剪、透明度图层或批绘制不可用时，将颜色烘焙进独立缓存的纹理。其他构建和上游 Gio 通过接口检测保留原来的彩色图集。测试覆盖纹理共享、缓存淘汰、兼容路径的颜色、透明度、变换和裁剪；8 位覆盖掩码的量化可能使边缘像素略有变化。
在 Apple M4 上各做三轮交替测量，进程总内存中位数在网格场景从 205 降到 145 MB，控件场景从 191 降到 148 MB。240 帧离屏网格测试的每帧耗时中位数从 2.776 降到 2.452 ms，分配量从 1.18 降到 1.04 MB。这些是启用两种图集后的本机结果，不代表其他应用的内存上限或收益保证。

- `gio/app`（macOS）：停止绘制一秒后，通过可取消的单次回调释放 GPU 纹理、
  路径和临时缓冲区，并将 Metal drawable 池缩至 1×1。Core Animation 保留已显示
  的画面，下次绘制前恢复尺寸；已编译的管线继续保留。原生窗口测试覆盖多次
  空闲、恢复和回调尚未触发时关闭窗口。原生调用使用 purego v0.11.1，
  因此此分支的 Go 最低版本升至 1.25。
- `gio/gpu`：颜色、纹理、渐变的回退管线及模板渲染器按需初始化，避免纯矩形
  场景加载未使用的着色器资源。离屏像素测试覆盖回收前后的绘制和读回。
- `gio/io/input`：收集输入时删除相邻的纯绘制叶子裁剪节点，保留输入处理、
  光标、语义和窗口动作节点。随机测试与原收集器逐点比较命中结果。
- `typesetting/font`：仅为实际测量过的字形分配 extents 缓存页；CFF 字体保留
  已校验的 charstring INDEX 偏移，省去每个字形的 Go 切片头。公开的展开式
  CFF 解析接口保留原行为。测试覆盖缓存边界、畸形 INDEX 和轮廓等价性。
- `typesetting/fontscan`：共享内容相同的只读系统字体字符覆盖集合，哈希后仍
  比较完整内容；查询缓存持有字体族切片副本，避免调用者复用切片破坏缓存。
- `gio/text`、`typesetting/shaping`、`typesetting/harfbuzz`：缓存已解析字体族，
  为符合条件的文本跳过多余换行和分段，复用规范化临时空间，避免复制整份
  shaping-plan 键。快速路径保留 Unicode 强制换行，并配有回归测试。
