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
| `internal/gogio/` | Gio 打包器 `gioui.org/cmd/gogio` | v0.10.0 | Unlicense 或 MIT（`internal/gogio/LICENSE`） |

`gioui.org/shader`（含 `gioui.org/shader/gio`）是独立模块，仍用上游。

两个副本随 Keel 的模块压缩包发布、共用 Keel 的版本，所以任何 require `github.com/dyike/keel` 的应用都会用上它们，不需要 `replace`。改导入路径会改变 Go 类型身份：与 Keel 一起直接使用上游 `gioui.org/...` 或 go-text 类型的应用或组件，需要改用上表的前缀（见 [docs/single-module.zh-CN.md](../docs/single-module.zh-CN.md)）。

## 规则

- 这两个目录保持上游风格；只改 Keel 需要的，不做纯格式修改。每处修改都记在下面并写明原因，换到更新的上游时照着重做。
- 副本是基础库，不是 Keel 模块：Gio 只导入自身和 typesetting，typesetting 只导入自身，二者都不导入 Keel 其他包或上游 Gio/go-text（`internal/deps` 检查）。
- 不复制上游的测试和测试数据。补丁自带测试，放在补丁旁边；在根目录运行（`go test ./gio/... ./typesetting/...`）。
- 升级上游：复制发布版本，去掉 `go.mod`、`go.sum`、`*_test.go` 和 `testdata`，改写两个导入前缀（不动 `gioui.org/shader`），再重做下面的补丁。

## 补丁

- `typesetting/font`：初始化时校验 TrueType 轮廓并收集可变字体点数，不再先为所有轮廓分配点数组后丢弃。完整轮廓仍按需解码，沿用共享缓存的锁。测试对照旧解析器的字形头、点数、轮廓和度量；模糊测试检查畸形记录的校验结果。
- `typesetting/fontscan`：文件路径和修改时间未变时不再重写磁盘索引；更新时逐文件流式编码，读取时复用定长记录缓冲，避免 `bytes.Buffer` 为探测记录 EOF 扩容。版本 7 格式、字体身份和回退有效性检查不变。测试覆盖失效、格式兼容、缓冲复用和截断记录。不改变 GC 策略。
- `gio/gpu`、`gio/internal/stroke`：在路径坐标系中缓存描边轮廓，再执行原有仿射变换和顶点分割。旋转不再重建相同轮廓；线宽或路径内容变化仍单独计算，命中时比较完整输入以防哈希碰撞。缓存独占数据，未使用一帧后淘汰，最多 128 项、1 MiB；超大路径不缓存，渲染器清理和释放时丢弃全部缓存。描边 scratch 的底层数组互相独立，可复用并丢弃超大数组，减少动态路径未命中的分配。测试逐字节对照旧算法的曲线和仿射顶点，覆盖所有权、淘汰、分配上限，以及旋转、嵌套圆角裁剪和透明度下的冷热 GPU 像素。不改变着色器或 GC。

- `gio/widget.Icon`：按尺寸、颜色保留最多八张不可变图片，使用 LRU 淘汰，
  每个图标的像素预算为 1 MiB。超大图片和原来一样只保留自身。
  交替使用不同样式不再反复栅格化或更换纹理身份；测试覆盖预热后零分配命中、
  淘汰顺序和引用释放。
- `gio/text`、`gio/widget`：标签和编辑器绘制从 shaper 借用一个 32 字形的临时缓冲。
  CoreText 显式接受同步、只读、不保留切片的 raster hook 契约；原有
  `SetRasterHook` 调用仍获得独立缓冲。重叠借用和更换 hook 不会复用尚在使用的
  缓冲。测试覆盖旧 hook 保留切片、借用失效和预热后零分配复用。
- `typesetting/fontscan` 测试用独立的可变 PingFang UI 字体集合验证命名实例与
  不可变表共享；系统回退合法地可能优先选用 MobileAsset 中的静态字体。
  CoreText 测试分别覆盖指定的可变字体和系统实际解析出的字重，保留精确
  PostScript 身份与非空字形断言。

- 同帧相同裁剪掩码共享图集位置，不新增 GPU 纹理；保留绝对变换输入和父级交集合成，已通过原生像素对照。见[渲染内存](../docs/rendering-memory.zh-CN.md)。

- `gio/gpu/internal/metal`：统一内存设备直接初始化共享的只读几何缓冲和首次完整上传的采样图片，避免普通 UI 上传触发额外 GPU 复制工作；本机最小 Metal 实验中，这条工作路径增加约 57 MiB。CPU 生成完整 mip 链，保留 sRGB 线性光过滤；原生像素对照覆盖奇数尺寸、行间距和每级 mip，观察到最大通道差值 2/255。局部/重复上传、可变缓冲、渲染目标、非统一内存设备和分配失败保留有序 GPU 路径。不改变帧率、drawable 数量或每帧 GC。范围和限制见[渲染内存与复用](../docs/rendering-memory.zh-CN.md)。

按行为整理的改动、源码与测试链接见[与上游的差异](../docs/upstream-patches.zh-CN.md)。
下文性能数字是历史本机测量，不是本轮新跑的 benchmark 或跨平台应用保证。

- 根级迁移：两个基础库实际移到 gio/ 和 typesetting/，导入前缀与目录对应。链接符号、资源发现、CI 和发布夹具同步新路径；源码检查拒绝过时的基础库导入。
- 模块合并：删除副本的 `go.mod`/`go.sum`，把导入改写为 Keel 前缀，不再依赖应用自己写的 `replace`（Go 不会传递 replace）才能用上副本。
- `internal/gogio`：打包器在 `keel` 进程内运行，而不是 `go run gioui.org/cmd/gogio`，因为上游 gogio 写死了 Gio 的导入路径。现在从 `GioRoot` 取路径（用于 `app.ID`、日志应用 ID、额外参数、权限和 `app/` 包查找）；参数、工作目录、环境变量和输出改为每次运行的选项（`Run`），上游包级状态通过串行调用隔离，默认产物也写入指定工作目录。WebAssembly 打包回归检查应用 ID/启动参数链接、HTML 标题和产物文件；发布夹具从真实根模块归档编译外部自定义组件。未复制 Windows 目标和上游端到端测试：Keel 自己生成 Windows 资源。`cmd/gogio` 可直接运行它，用于没有 `keel.json` 的构建，例如站点的组件库。

- 打包器迁移：独立入口放在 `cmd/gogio`，共享实现放在 `internal/gogio`，两个命令共用打包器，避免在 `cmd/keel` 内嵌套命令。资源发现测试适配新的目录深度，CI 和发布夹具改为调用独立入口。

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
- `gio/app`（macOS）：绘制时保留三个 `CAMetalLayer` drawable。Core Animation 在窗口缩放时可能同时持有两个已提交的表面，限制为两个会让主线程在 `nextDrawable` 中间歇性等待约一秒。空闲时仍会缩小备用表面。原生回归测试 `TestMetalResizeWindow` 连续放大、缩小 60 帧，并拒绝这种合成器停顿（`KEEL_DESKTOP=1 go test ./gio/app`）。原生 Metal 测试的窗口置顶；若仍被遮挡（被覆盖、显示器休眠或锁屏），测试会跳过而不是卡住，因为被遮挡的窗口不绘制。
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

- `gio/gpu`：按图片绘制操作记录是否位于透明度图层内，无关图层不再让整帧禁用着色纹理共享；路径裁剪和图层内图片仍使用烘焙颜色。纹理数量和原生像素测试覆盖无关图层与嵌套透明度。CoreText 页面改为 256×256，保留 32 MiB 像素预算，减少回退路径按颜色复制整页的成本；缓存测试覆盖淘汰和当前帧正在使用的页面。

- `typesetting/fontscan`（macOS）：扫描系统 FontServices 的 `Resources/Reserved` 目录。当前 macOS 将 PingFangUI.ttc 放在这里，漏扫会静默回退到其他中文字体。原生字体解析测试覆盖 Helvetica Neue 与 PingFang SC。

- `ui/theme` 原生字形绘制：通过 purego 按 PostScript 身份寻找所选 TTC 字体。CoreText 展开命名实例后，描述符下标不等于 TTC 字体编号。原生测试检查实际创建的苹方简体 Medium 和 SF Regular 对象，防止将前者画成香港 Thin；未新增 cgo 绑定。

- `typesetting/font/opentype/tables`：修复 `fvar` 命名实例记录步长，并限制单条记录的读取范围，避免缺失的 PostScript 名称读到下一条记录。`fontscan` 收录其他轴为默认值的命名字重，加载对应变体坐标，并把磁盘缓存升到版本 7。`gio/text.Shaper.GlyphFileInstance` 向 CoreText 传递所选实例；原生测试覆盖 Regular 和 Semibold 从排版到绘制的实际身份。

### 2026-10-14：帧缓冲与可变字体表的保留内存

- `gio/gpu/path.go`：帧缓冲浪费比例改为“保留面积 / 请求面积”。原来的反向比例使大帧之后的小帧无法缩小缓冲。`gio/gpu/fbo_memory_test.go` 验证缩小、同尺寸复用和释放。
- `typesetting/fontscan`：同一文件、同一 TTC 索引的命名实例共享只读字体表，每个实例仍保留独立的 `Font` 身份和 `Face`，保持位置元数据与变体坐标正确。Darwin PingFang 实例测试同时检查 GSUB 底层存储共享和字重区分。
- `typesetting/font`：加载后保留字形头和变体所需点数，仅在请求轮廓时缓存完整 TrueType 轮廓，避免原生文字绘制长期保留全部 CJK 轮廓。共享缓存加锁，Face 的临时状态仍独立。测试逐一对照 Go Regular 的全部轮廓，以及 PingFang 四个字重的中英文字形，并验证并发访问。

### 2026-10-14：带覆盖纹理的批量绘制与保守裁剪内区

- Metal 将路径/交集裁剪下的纯色和图片合入批量绘制，覆盖纹理共用现有八个纹理槽，保持原来的 `min(abs(value), 1)` 覆盖率运算。裁剪内的着色字形可以共享图集；透明度图层和不支持此能力的设备仍走原回退路径。测试覆盖纹理槽耗尽、顺序、着色、分配失败回退，以及圆角/嵌套裁剪的原生像素。
- 圆角矩形路径操作携带保守内区。只有所有路径祖先均完整包含绘制像素边界时，才跳过覆盖合成；额外一像素余量保护抗锯齿边缘。未知路径、描边、旋转和不支持的变换矩形保留原路径。收集测试覆盖这些情况；内部裁剪编码增加 16 字节，所有读取者使用统一更新的长度。
- OneCatch 的 5016×2820 像素对照最大单通道差值为 1/255。实际窗口加载和滚动仍超过 300 MiB；像素测试和较低的帧提交耗时都不能证明内存达标或端到端帧率。

- `typesetting/shaping`：行尾空白裁剪先核对原始字簇是否全为空白。轮廓尺寸读取失败（本机苹方 UI 字体可复现）也会得到零墨迹宽度，不能据此清零行末可见字形的前进宽度。多字体单行快速路径和换行路径共用检查；新增 LTR、RTL、纵排合成测试及混合字体标签尺寸、像素回归。
- `typesetting/shaping`：换行测量候选行宽也使用同一检查。`Output.advanceSpaceAware` 曾把行末零墨迹宽度的字形当作空白、不计入行宽，苹方（读不到轮廓尺寸）下每行中文因此多排一个字，最后一个字被裁切。现在只有原文是空白时才扣除（与裁剪共用 `whitespaceCluster`）。合成测试覆盖换行行数和行宽检查。

### 2026-10-14：呈现阶段的 GPU 设备丢失恢复

- `gio/app`：移植上游 Gio 提交 [bf9f995](https://github.com/gioui/gio/commit/bf9f995d86c22d555d7ce2db154395bca94541b7)。`Present` 返回 `gpu.ErrDeviceLost`（包括包装后的错误）时，释放 GPU 上下文并立即安排下一帧；已经确认交付的帧不能重试。Keel 仅在呈现成功且上下文仍加锁时执行空闲 GPU 资源回收。
- 回归测试覆盖成功、直接与包装的设备丢失错误、其他错误、帧确认次数和上下文加锁/释放顺序。Windows 交叉编译通过，尚未进行 Windows 真机验证。Metal 当前不会返回呈现错误，因此该补丁不降低 macOS 内存峰值。
