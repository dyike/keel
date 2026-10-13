# third_party

[English](README.md) | 简体中文

Keel 自带它绘制界面所用的库的副本，这样可以直接修复和优化，不必等上游发版。源码位于根级 gio/ 和 typesetting/ 目录，本目录保存来源与补丁记录。副本是 Keel 模块里的普通包，没有自己的 `go.mod`，以 Keel 的路径导入：

| 上游导入前缀 | Keel 导入前缀 |
| --- | --- |
| `gioui.org`（Gio 根模块） | `github.com/dyike/keel/gio` |
| `github.com/go-text/typesetting` | `github.com/dyike/keel/typesetting` |

| 目录（相对仓库根目录） | 上游 | 基线 | 许可 |
| --- | --- | --- | --- |
| `gio/` | [Gio](https://gioui.org) `gioui.org` | v0.10.3 | Unlicense 或 MIT（`gio/LICENSE`） |
| `typesetting/` | [go-text](https://github.com/go-text/typesetting) `github.com/go-text/typesetting` | v0.3.5 | Unlicense 或 BSD-3-Clause（`typesetting/LICENSE`） |
| `cmd/keel/internal/gogio/` | Gio 打包器 `gioui.org/cmd/gogio` | v0.10.0 | Unlicense 或 MIT（`cmd/keel/internal/gogio/LICENSE`） |

`gioui.org/shader`（含 `gioui.org/shader/gio`）是独立模块，仍用上游。

两个副本随 Keel 的模块压缩包发布、共用 Keel 的版本，所以任何 require `github.com/dyike/keel` 的应用都会用上它们，不需要 `replace`。改导入路径会改变 Go 类型身份：与 Keel 一起直接使用上游 `gioui.org/...` 或 go-text 类型的应用或组件，需要改用上表的前缀（见 [docs/single-module.zh-CN.md](../docs/single-module.zh-CN.md)）。

## 规则

- 这两个目录保持上游风格；只改 Keel 需要的，不做纯格式修改。每处修改都记在下面并写明原因，换到更新的上游时照着重做。
- 副本是基础库，不是 Keel 模块：Gio 只导入自身和 typesetting，typesetting 只导入自身，二者都不导入 Keel 其他包或上游 Gio/go-text（`internal/deps` 检查）。
- 不复制上游的测试和测试数据。补丁自带测试，放在补丁旁边；在根目录运行（`go test ./gio/... ./typesetting/...`）。
- 升级上游：复制发布版本，去掉 `go.mod`、`go.sum`、`*_test.go` 和 `testdata`，改写两个导入前缀（不动 `gioui.org/shader`），再重做下面的补丁。

## 补丁

按行为整理的改动、源码与测试链接见[与上游的差异](../docs/upstream-patches.zh-CN.md)。
下文性能数字是历史本机测量，不是本轮新跑的 benchmark 或跨平台应用保证。

- 根级迁移：两个基础库实际移到 gio/ 和 typesetting/，导入前缀与目录对应。链接符号、资源发现、CI 和发布夹具同步新路径；源码检查拒绝过时的基础库导入。
- 模块合并：删除副本的 `go.mod`/`go.sum`，把导入改写为 Keel 前缀，不再依赖应用自己写的 `replace`（Go 不会传递 replace）才能用上副本。
- `cmd/keel/internal/gogio`：打包器在 `keel` 进程内运行，而不是 `go run gioui.org/cmd/gogio`，因为上游 gogio 写死了 Gio 的导入路径。现在从 `GioRoot` 取路径（用于 `app.ID`、日志应用 ID、额外参数、权限和 `app/` 包查找）；参数、工作目录、环境变量和输出改为每次运行的选项（`Run`），上游包级状态通过串行调用隔离，默认产物也写入指定工作目录。WebAssembly 打包回归检查应用 ID/启动参数链接、HTML 标题和产物文件；发布夹具从真实根模块归档编译外部自定义组件。未复制 Windows 目标和上游端到端测试：Keel 自己生成 Windows 资源。`cmd/keel/internal/gogio/cmd` 可直接运行它，用于没有 `keel.json` 的构建，例如站点的组件库。

- `gio/internal/nativeabi`：无 CGO macOS ABI 探针通过系统真实 `objc_msgSend` 调用 Go Objective-C IMP，覆盖 range 参数对/寄存器耗尽、嵌套 rect 参数与返回、浮点参数、可空的原生分配出参及哨兵值。类和回调只注册一次，并明确记录 purego 0.11.1 不支持标量浮点回调返回。这是 M0 可行性测试，不是已迁移的窗口后端。扩展探针覆盖结构体 super 调用及复制 block 的原生调用，arm64 实机通过；amd64 实机和精确协议签名仍须验收。

- `gio/widget.Editor`：增加可选的 `ScrollBounds`、`ScrollOffset`、`ScrollTo`，让多行输入组件绘制并拖动滚动条。移动视口保留选区，后续编辑或移动光标恢复自动跟随。扩展不存在时，通过能力检测保留原生滚轮滚动；导入路径兼容性仍以迁移指南为准。1×/2× 回归覆盖边界、选区、光标可见性和内容替换。

- `gio/gpu/internal/metal`：空闲清理等待最后提交的 GPU 工作完成，再释放命令缓冲区、上传/读回暂存缓冲区和四边形实例缓冲区。仍在使用的纹理和管线保留，临时缓冲区在下次绘制或读回时重建，降低空闲常驻内存而不请求新帧。无窗口回归测试覆盖重复清理、读回和恢复绘制。

- 修复 `gio/internal/f32` 和 `gio/app/internal/ibus` 里 `go vet` 报的无字段名结构体字面量。
- `typesetting/font/opentype`：能给出自身字节的资源（`Shared`，或任何有 `Bytes() []byte` 方法的资源）给出的表是切片而不是副本，也从不写入调用方传入的缓冲区。`typesetting/fontscan` 以只读方式映射系统字体文件（`mmap`、`MapViewOfFile`）并按 `Shared` 解析，字体的表是可回收的文件页，不占 Go 堆：hello 在 macOS 上空闲时的内存占用从 153 MB 降到 103 MB。`gio/font/opentype.ParseCollectionShared` 对不会变的字节做同样的事，供 `gio/font/gofont` 使用；Keel 的主题传入带 `Bytes` 方法的字体文件。测试用只读映射解析、描述、排版并取轮廓全部系统字体（`fontscan/openfont_test.go`）。
- `typesetting/harfbuzz`：附着链指向缓冲区之前时像 HarfBuzz 一样直接返回，不再访问 `pos[-1]`；原来排版 macOS 的 Farisi.ttf 会崩溃。
- `gio/gpu`：路径渲染器的覆盖纹理（stencil 和 intersection 帧缓冲，打包了所有圆角图形，常常比窗口还大）每帧都会重画，所以一帧提交后把它们标为可回收（`driver.Volatile`，即 Metal 的 purgeable 状态），下一帧调整尺寸、准备使用时再标回。窗口空闲时最大的几块 GPU 内存因此不计入占用：hello 在 macOS 上从 72 MB 降到 64 MB。其他后端暂未实现。
动画中的窗口若每帧都标记，就要每帧付一次内核调用，所以只在某一帧之后 100 ms 内不再需要新帧时才标记（`app.Window` 调用 `gpu.Idle`，判断用不会消耗唤醒标记的 `input.Router.PeekWakeup`）。每个纹理记住自己的状态，状态不变时不做调用。
- `gio/gpu/internal/metal`：GPU 缓冲按 2 的幂大小分级复用，不再每帧创建、释放（动画窗口的 CPU 有五分之一花在 `newBuffer` 和 `CFRelease` 上）。一帧中释放的缓冲要等到下一次 `BeginFrame` 等待上一个命令缓冲完成后才复用；缓冲池有上限（32 MB，每级 64 个），`gpu.Idle` 时清空。
- `gio/gpu`：裁剪路径按内容缓存（路径数据的哈希，加上描边宽度、轮廓标志和变换），而不是按记录它的 op；后者在 `Ops` 重置后就变了。Keel 每帧重新记录，所以每个边框和圆角图形每帧都要重新细分并上传。偏移不在键里，移动的内容（如滚动）也能复用路径。动画中的组件库每 10 秒分配从 881 MB 降到 470 MB（GC 从 51 次降到 29 次）。所有静态组件截图逐字节一致。
- `gio/app`（macOS、iOS）：`displayLink.Start` 和 `Stop` 只在请求的状态改变时才发给显示链接的 goroutine。窗口在每个动画帧都调用 `Start`，原来这个无缓冲发送每次都让主线程等一次 goroutine 交接。
- `gio/app`（macOS）：绘制时保留三个 `CAMetalLayer` drawable。Core Animation 在窗口缩放时可能同时持有两个已提交的表面，限制为两个会让主线程在 `nextDrawable` 中间歇性等待约一秒。空闲时仍会缩小备用表面。原生回归测试 `TestMetalResizeWindow` 连续放大、缩小 60 帧，并拒绝这种合成器停顿（`KEEL_DESKTOP=1 go test ./gio/app`）。
- `gio/gpu`、`gio/gpu/internal/metal`：裁剪为普通矩形、填充为纯色或纹理的连续 op 合并为一次实例化绘制（`driver.QuadBatcher`，每批最多 8 个纹理），不再一个 op 一次绘制调用：终端大小的文字网格原来每帧 3,300 次绘制。Metal 着色器在运行时编译，计算与 Gio 的 blit 着色器相同；其他后端保持逐个绘制。所有静态组件截图逐字节一致。
- `typesetting/harfbuzz`：AAT 排版（`morx`、`kerx`，Menlo 等苹果字体使用）把每个子表的字形类别缓存留在字体的加速器上，与 HarfBuzz 一致；上游把它复制进每次调用的上下文，填好后就丢弃。上下文和子表驱动的状态放在复用的 `Buffer` 里，不再每次排版分配 1.3 KB 以上。239 个带 `morx` 的 macOS 字体排版结果完全一致；benchmark 网格场景的文字分配减半。
- `typesetting/fontscan`：`FontMap.ResolveFace` 对 ASCII 字符查当前查询的表，查询的字体族哈希每次 `SetQuery` 只算一次，不再每个字符算一次。查询或文字系统没变时，`SetQuery` 和 `SetScript` 保留缓存；Gio 排版每段文字都会调用这两个方法。测试检查带缓存的和全新的 FontMap 解析出相同的字体。
- `gio/app`（macOS）：没人看得见的窗口（被遮住、在别的桌面、锁屏或屏幕休眠时）在第一帧之后不再绘制，并停掉显示链接；再次可见时重画（`windowDidChangeOcclusionState:`）。AppKit 只报告状态变化，所以视图挂到窗口上时也读一次状态。Gio 原来对隐藏的窗口照样画每一帧动画：benchmark 网格场景锁屏时每秒用 0.49 秒 CPU，现在 0.001 秒。
- `gio/app`（macOS）：所有显示器都休眠时，创建显示链接失败，随后创建窗口在 `gio_onDestroy` 里崩溃（视图还没有句柄）。现在显示链接退回到主显示器，没有句柄的视图跳过 `gio_onDestroy`。

- `gio/op/paint`、`gio/gpu`、`gio/gpu/internal/metal`：Metal 构建提供可选的 `ImageOp.AddTinted` 扩展，将不可变图片与每次绘制的线性颜色相乘。Keel 字形图集只保存一份白色覆盖掩码，不再为每种文字颜色复制像素（benchmark 网格从六张 1 MiB 页面降为一张）。Metal 在上传纹理前预留整帧的实例缓冲，避免共享纹理在绘制过程中退回不支持着色的路径。路径裁剪、透明度图层或批绘制不可用时，将颜色烘焙进独立缓存的纹理。未提供扩展的构建通过接口检测保留原来的彩色图集。测试覆盖纹理共享、缓存淘汰、兼容路径的颜色、透明度、变换和裁剪；8 位覆盖掩码的量化可能使边缘像素略有变化。
在 Apple M4 上各做三轮交替测量，进程总内存中位数在网格场景从 205 降到 145 MB，控件场景从 191 降到 148 MB。240 帧离屏网格测试的每帧耗时中位数从 2.776 降到 2.452 ms，分配量从 1.18 降到 1.04 MB。这些是启用两种图集后的本机结果，不代表其他应用的内存上限或收益保证。

- `gio/app`（macOS）：停止绘制一秒后，通过可取消的单次回调释放 GPU 纹理、
  路径和临时缓冲区，并将 Metal drawable 池缩至 1×1。Core Animation 保留已显示
  的画面，下次绘制前恢复尺寸；已编译的管线继续保留。原生窗口测试覆盖多次
  空闲、恢复和回调尚未触发时关闭窗口。原生调用使用 purego v0.11.1，
  合并后模块使用的 Go 版本以根 go.mod 为准。
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
- `typesetting/fontscan`（macOS）：额外扫描 `/System/Library/AssetsV2/com_apple_MobileAsset_Font*`，当前 macOS 把苹方等系统字体放在这里。否则"PingFang SC"永远解析不到，中文回退到冬青黑 W3，且没有 Medium、Semibold。`fontscan/scan_darwin_test.go` 验证三种字重都能解析。
- `gio/text`、`gio/font`、`gio/widget`、`gio/widget/material`：平台字形光栅化。`Shaper.GlyphFile(id)` 返回字形所在的字体文件、集合内序号、字号和字形 id（实现 `font.FaceSource` 的字体保留路径；内置字体、可变字体实例和设置了变体坐标的字体返回 false）。`Shaper.SetRasterHook` 注册绘制一行纯色字形的函数；`Label`、`Editor`、`Selectable` 新增可选的 `Color`（由 `material.LabelStyle` 和 `EditorStyle` 设置），有挂钩时每行按整像素平移并把小数部分交给挂钩，挂钩返回 false 时回退到轮廓。Keel 的 theme 在 macOS 上设置挂钩，用 CoreText 绘制，通过接口断言检测可选能力；这不代表可以混用上游与 Keel 的导入路径。
- `gio/gpu`：不透明全窗口背景清除已覆盖的绘制操作时，同时清除已完成的透明度图层。旧图层索引否则会替换前景中无关的绘制操作，导致全屏覆盖页面缺字。回归测试覆盖旧图层清理、后续新图层、局部覆盖和活动透明度组内的填充。

## Wayland 销毁确认

Wayland 后端在返回无效的 `WaylandViewEvent` 后，等调用方下一次调用 `Event`
才销毁原生窗口和连接。事件入队不代表调用方已经处理：Keel 必须先在连接仍有效时
释放滚动监听器的 pointer、seat 和私有队列。此前 `close` 立即断开连接，导致 Linux
上的 `wlScrollWatch.close` 访问已释放的对象。
`deferred_teardown_test.go` 覆盖通知、确认与只清理一次的顺序。

- `gio/app`（Wayland）：`Run` 回调排队到原生事件线程执行。此前后台最大化、配置窗口与收集帧指令并发，可能在 GPU 读取时重置录制指令。关闭时先停止队列唤醒，再关闭原生管道，并让等待中的 `Window.Run` 返回。回归测试在执行排队重置前解码当前帧。
