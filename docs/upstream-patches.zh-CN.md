# Keel 对 Gio 与 go-text 的改动

[English](upstream-patches.md) | 简体中文

Keel 保留 Gio 的窗口、布局、输入和 GPU 架构，针对资源复用、文字绘制和窗口
生命周期问题增加补丁；go-text 副本主要优化字体存储与排版。
本文按行为说明“上游是什么、Keel 改了什么、去哪看源码和测试”；
[完整补丁记录](../third_party/README.zh-CN.md)保留历史测量和维护说明。

## 上游基线与模块交付

| 副本 | 上游基线 | 当前源码与导入前缀 |
| --- | --- | --- |
| Gio | `gioui.org` v0.10.3 | `gio/`；`github.com/dyike/keel/gio` |
| go-text | `github.com/go-text/typesetting` v0.3.5 | `typesetting/`；`github.com/dyike/keel/typesetting` |
| gogio 打包器 | `gioui.org/cmd/gogio` v0.10.0 | `internal/gogio/`；CLI 内部实现 |

两个基础库是 Keel 根模块中的普通包。删除各自的 go.mod/go.sum 后，依赖统一
维护在根目录，内部导入、链接符号、资源查找均使用新前缀。应用只 require Keel
即可获得补丁源码，无需复制 Keel 原来的 replace。`gioui.org/shader` 保持独立外部模块。

改导入路径会改变 Go 命名类型的身份。上游 layout.Context 即使字段相同，也不能
直接替代 Keel 的 layout.Context。应用和第三方组件须迁移 import，见[单模块](single-module.zh-CN.md)。
通过接口检测某个绘制扩展是否存在，只是能力回退机制，不代表不同包的类型可以混用。

副本保留许可证和上游来源。未整批复制上游测试与测试数据；下文列出的测试是
Keel 的补丁回归。本次说明已对照本机缓存中的基线源码，将纯导入改写与实现改动分开核对。

## Gio：GPU 绘制与资源生命周期

| 范围 | 上游行为与 Keel 改动 | 源码与验证依据 |
| --- | --- | --- |
| 路径缓存 | 每帧重置 ops 会改变记录对象身份。Keel 按几何内容、描边和变换缓存路径，让重复或平移的形状复用细分结果与 GPU 上传。 | [gpu.go](../gio/gpu/gpu.go)、[caches.go](../gio/gpu/caches.go)；补丁记录中的同机组件截图历史对比。 |
| Metal 缓冲池 | 反复创建、释放缓冲改为按二次幂尺寸分级复用。一帧中退役的缓冲，要等下一次 BeginFrame 等待此前 GPU 工作完成后才能复用。池容量有上限，空闲时清理。 | [pool_darwin.go](../gio/gpu/internal/metal/pool_darwin.go)、[metal_darwin.go](../gio/gpu/internal/metal/metal_darwin.go)；[空闲恢复与读回测试](../gio/gpu/headless/trim_darwin_test.go)。 |
| 矩形批绘制 | 符合条件的连续纯色/图片填充和普通矩形裁剪，由逐操作绘制改为 Metal 实例化批绘制，每批最多八张纹理。其他后端和不符合条件的操作保留逐操作路径。 | [quads_darwin.go](../gio/gpu/internal/metal/quads_darwin.go)、[driver 接口](../gio/gpu/internal/driver/driver.go)；历史画面对比及现有 [GPU 生命周期测试](../gio/gpu/lazy_pipeline_test.go)。 |
| 图片着色 | Metal 构建提供可选 ImageOp.AddTinted 扩展，白色字形覆盖掩码只保存一份，颜色在每次绘制时传入。路径裁剪、透明度或批绘制不可用时，使用缓存的预着色纹理。覆盖掩码量化可能使边缘像素略有变化。 | [tint_darwin.go](../gio/op/paint/tint_darwin.go)、[纹理共享与回退测试](../gio/gpu/tint_test.go)、[像素测试](../gio/op/paint/tint_darwin_test.go)。 |
| 空闲资源 | 没有近期帧请求时，Metal 的临时覆盖纹理可设为可回收。空闲清理释放缓存和临时缓冲；macOS 一秒未绘制后，还把备用 drawable 表面缩为 1×1，下次绘制恢复尺寸。编译好的 pipeline 保留。 | [path.go](../gio/gpu/path.go)、[metal_idle_macos.go](../gio/app/metal_idle_macos.go)、[恢复绘制测试](../gio/gpu/headless/trim_darwin_test.go)、[原生空闲测试](../gio/app/metal_idle_macos_test.go)。 |
| Pipeline 创建 | 纯色、图片、渐变和 stencil 资源首次使用时才初始化，减少简单场景用不到的 shader 分配。 | [gpu.go](../gio/gpu/gpu.go)、[延迟初始化生命周期测试](../gio/gpu/lazy_pipeline_test.go)。 |
| 可变图片 | 上游把每个 ImageOp 都视为不可变：像素一变就要新建 ImageOp，渲染器随之新建并整张上传纹理。paint.NewMutableImageOp 让一张图片只对应一张纹理；调用 Invalidate(rect) 后，各渲染器只上传自上次上传以来变化的区域，落后超过 32 次变化时整张上传。paint.FilterLinearNoMipmap 线性采样但不生成 mipmap，适合频繁变化、且不会缩到一半以下的图片。 | [paint.go](../gio/op/paint/paint.go)、[dirty.go](../gio/internal/dirty/dirty.go)、[gpu.go](../gio/gpu/gpu.go)；[局部上传、着色与过滤测试](../gio/gpu/mutable_image_test.go)。 |
| Surface 图片 | 上游只能绘制自己上传的像素。paint.NewSurfaceOp 显示平台 GPU 表面（Apple 平台上是 IOSurface），由持有者用 Surface.Set 替换，例如硬件视频解码器的输出；渲染器通过 driver.SurfaceTextureDevice 把每个新画面包成纹理，不复制像素。Metal 通过 purego 实现；不支持的设备画一个透明像素。 | [surface.go](../gio/internal/surface/surface.go)、[gpu.go](../gio/gpu/gpu.go)、[surface_darwin.go](../gio/gpu/internal/metal/surface_darwin.go)；[渲染器](../gio/gpu/surface_image_test.go)、[Metal 像素](../gio/gpu/internal/metal/surface_darwin_test.go)与[无头绘制](../gio/gpu/headless/surface_darwin_test.go)测试。 |
| 不透明清屏修复 | 整窗清屏时，把被覆盖的透明度图层与其绘制操作一起丢弃。此前遗留的图层索引可能替换后续前景操作，使浮层中的字形消失。 | [gpu.go](../gio/gpu/gpu.go)、[清屏回归测试](../gio/gpu/opaque_clear_test.go)。 |

这些补丁扩展 Gio 的渲染器和资源策略。Metal 的批绘制、着色、可回收存储不代表
所有后端都实现了相同能力。释放临时资源也不等于释放活跃纹理、字体 face 或进程全部内存。

## Gio：文字、编辑与输入

文字 shaping 和布局仍由 Gio 处理。Keel 增加光栅化挂钩，让 `ui/theme` 在
macOS 提供 CoreText 字形像素；Gio 基础库不导入 Keel 的 theme 或 native 包。

| 改动 | 原行为与新行为 | 源码与回归测试 |
| --- | --- | --- |
| 平台字形光栅化 | Shaper.GlyphFile 暴露文件字体的路径、集合序号、字号和字形 ID。SetRasterHook/Raster 提供可选的纯色字形绘制挂钩；Label、Editor、Selectable 和 material 样式传入颜色及横向小数位置。挂钩不存在或返回 false 时保留轮廓绘制。 | [raster.go](../gio/text/raster.go)、[shaper.go](../gio/text/shaper.go)、[Label 光栅化测试](../gio/widget/label_raster_test.go)；CoreText 适配位于 [ui/theme](../ui/theme/coretext_cache_darwin.go)。 |
| 共享字体集合 | ParseCollectionShared 和内置 Go 字体使用不可变底层字节，避免复制字体表。Font.FaceSource 提供文件/集合身份，供平台光栅化使用。 | [font API](../gio/font/font.go)、[OpenType 解析](../gio/font/opentype/opentype.go)、[共享字体表测试](../typesetting/fontscan/openfont_test.go)。 |
| Typeface 解析缓存 | 重复排版复用最近解析的 typeface 和字体族列表，不再每个文本 run 都重新解析。 | [gotext.go](../gio/text/gotext.go)；现有文字回归与历史分配测量。 |
| Editor 视口 API | ScrollBounds、ScrollOffset、ScrollTo 暴露多行编辑器的视口控制。移动视口保留选区，后续光标移动或编辑恢复光标可见性跟随。 | [editor.go](../gio/widget/editor.go)、[滚动回归测试](../gio/widget/editor_scroll_test.go)。 |
| 输入裁剪剪枝 | 输入收集不再累积相邻且仅用于绘制的叶子裁剪；含事件处理器、光标、语义或窗口操作的裁剪仍保留。 | [pointer.go](../gio/io/input/pointer.go)、[随机路由等价测试](../gio/io/input/clip_prune_test.go)。 |

内存字体和设置变体的 face 不提供可用的 GlyphFile 文件身份，所以 CoreText
适配仍需回退。挂钩没有替换 shaping，也不代表 macOS 文字绑定已经去掉 CGO。

## Gio：原生窗口事件

| 平台或范围 | 改动与原因 | 源码与回归测试 |
| --- | --- | --- |
| Darwin 帧调度 | Display-link 的 Start/Stop 只在请求状态改变时通知 goroutine，避免每个动画帧都同步切换 goroutine。 | [os_darwin.go](../gio/app/os_darwin.go)；原生空闲/缩放测试与历史 CPU 测量。 |
| macOS 可见性与休眠 | 被遮挡、位于其他 Space 或屏幕休眠时，窗口首帧后停止动画绘制，重新可见时重绘。Display-link 创建可回退到主屏；销毁时跳过尚未关联 handle 的 view。 | [os_macos.go](../gio/app/os_macos.go)、[os_macos.m](../gio/app/os_macos.m)；原生窗口检查，休眠/Space 切换仍需人工覆盖。 |
| macOS 缩放 | 绘制时保留三个 CAMetalLayer drawable，因为 Core Animation 在缩放时可能持有两个已提交表面。空闲时仍会缩小备用表面。 | [metal_macos.go](../gio/app/metal_macos.go)、[原生缩放测试](../gio/app/metal_resize_macos_test.go)。 |
| Wayland 销毁顺序 | 原生销毁前先返回无效 WaylandViewEvent，等待消费者下一次调用 Event 再销毁，使消费者能在连接仍有效时释放 proxy 和私有队列。 | [deferred_teardown.go](../gio/app/deferred_teardown.go)、[os_wayland.go](../gio/app/os_wayland.go)、[销毁回归测试](../gio/app/deferred_teardown_test.go)。 |
| Wayland Window.Run | 回调排队到原生事件线程执行，不再在后台调用方直接执行。退出时停止唤醒并释放等待者，避免重置 frame ops 与 GPU 读取竞争。 | [event_funcs.go](../gio/app/event_funcs.go)、[队列测试](../gio/app/event_funcs_test.go)、[Wayland 生命周期测试](../gio/app/os_wayland_lifecycle_test.go)。 |
| 原生 ABI 可行性 | 新增无 CGO 测试，验证 Objective-C 结构体回调、super 调用和复制 block。这些是后续绑定迁移的探针，不是替换后的 AppKit 后端。 | [ABI 探针](../gio/internal/nativeabi/abi_darwin_test.go)。 |

另有 [internal/f32](../gio/internal/f32/f32.go) 和
[app/internal/ibus](../gio/app/internal/ibus/ibus.go) 的结构体字段键名修复，用于通过 vet。

## go-text：字体存储、查找与排版

| 范围 | 上游行为与 Keel 改动 | 源码与回归测试 |
| --- | --- | --- |
| 字体表存储 | 基线读取字体表时复制数据。Shared 资源及实现 Bytes() 的资源改为借用不可变表切片；系统字体在可用时用只读 mmap/MapViewOfFile，失败时回退普通文件。 | [shared.go](../typesetting/font/opentype/shared.go)、[reader.go](../typesetting/font/opentype/reader.go)、[映射与共享表测试](../typesetting/fontscan/openfont_test.go)。 |
| 字体缓存体积 | 字形 extents 按每页 128 个字形、实际测量时分配。CFF 保留校验过的 charstring INDEX 偏移，不再每个字形保存切片头；展开解析 API 仍保留。不可变字符覆盖集合在完整内容比较后共享存储。 | [cache.go](../typesetting/font/cache.go)、[缓存页测试](../typesetting/font/cache_pages_test.go)、[CFF 测试](../typesetting/font/cff/compact_test.go)、[覆盖集合测试](../typesetting/fontscan/share_coverage_test.go)。 |
| 字体查找 | 每个 query 缓存 ASCII face 与字体族哈希，query/script 不变时保留缓存；SetQuery 持有字体族切片的副本，防止调用方复用切片破坏缓存。macOS 还扫描 PingFang 所在的系统 asset 字体目录。 | [fontmap.go](../typesetting/fontscan/fontmap.go)、[query/缓存测试](../typesetting/fontscan/resolve_cache_test.go)、[PingFang 测试](../typesetting/fontscan/scan_darwin_test.go)。 |
| AAT 与 shaping 临时状态 | 苹果字体的字形类别缓存留在字体加速器上；AAT 和 normalization 工作状态复用 Buffer 存储。Shaping-plan 缓存命中时避免构造、复制完整 plan key。 | [aat_layout.go](../typesetting/harfbuzz/aat_layout.go)、[buffer.go](../typesetting/harfbuzz/buffer.go)、[shape.go](../typesetting/harfbuzz/shape.go)；历史字体等价与分配对比。 |
| 分段与折行 | 符合条件的 LTR 段落跳过不必要的 bidi 分段；多个 run 能放入一行时走快速折行路径，Unicode 强制换行、bidi 顺序和空白处理仍与通用路径一致。 | [input.go](../typesetting/shaping/input.go)、[wrapping.go](../typesetting/shaping/wrapping.go)、[LTR](../typesetting/shaping/simple_ltr_test.go)、[折行](../typesetting/shaping/fast_wrap_test.go)及[强制换行](../typesetting/shaping/mandatory_break_test.go)测试。 |
| Attachment-chain 修复 | 字形 attachment 指向缓冲区之前时返回，不再访问 pos[-1]，避免排版 Farisi.ttf 时出现的 panic。 | [ot_layout_gpos.go](../typesetting/harfbuzz/ot_layout_gpos.go)、[attachment 测试](../typesetting/harfbuzz/ot_layout_gpos_keel_test.go)。 |

共享字节在解析后的 face 使用期间必须保持存活且不可变。成功映射的字体有意
保留到这些 face 的进程生命周期结束。Go 堆分配减少，不代表不存在映射内存，
更不代表可以修改借用的字体数据。

## 打包器适配与兼容边界

复制的 gogio 提供串行 Run 入口，每次运行传入工作目录、环境、参数和输出流。
GioRoot 提供应用 ID/启动参数的链接符号、Android 权限前缀和移动端资源发现路径，
默认输出也遵守传入的工作目录。见 [Run](../internal/gogio/run.go)
和[打包回归测试](../internal/gogio/run_test.go)。

Windows 可执行文件资源由 Keel 单独生成。iOS 模拟器 shader 适配及 SDK/签名调度
位于 CLI。[真实模块发布夹具](../scripts/check-module-release.py)验证外部应用、
自定义组件、按版本安装 CLI、Windows/Web 打包和 macOS 包内 ID/启动参数。

macOS 空闲辅助代码已使用 purego，但 AppKit、Metal/GL/headless、CoreText 仍含 CGO
绑定。Linux 的 X11/Wayland/EGL/GL/Vulkan 迁移也尚未完成。ABI 探针目前证明
原生 arm64 上的可行性，原生 amd64 执行与完整 IME 协议等价仍待验证。
当前支持状态见[验证记录](desktop-cgo-free-validation.zh-CN.md)，不能把补丁清单当作
Desktop CGO-Free 整体验收。

## 验证与升级上游

在仓库根目录运行：

```sh
go test ./gio/... ./typesetting/... ./cmd/keel/... ./internal/deps
go test ./internal/site
```

macOS 桌面上的原生 Metal 检查需要显式启用：

```sh
KEEL_DESKTOP=1 go test ./gio/app -run 'TestMetalIdleWindow|TestMetalResizeWindow' -count=1
```

部分测试依赖系统字体、GPU 或特定平台，在其他环境会跳过。补丁记录中的性能数字
对应当时记录的机器与负载，不是本轮新跑的 benchmark，也不是跨平台保证。

升级上游时，对比记录中的基线与候选版本，把导入/模块迁移和行为补丁分开，
再逐项重做行为补丁及其回归。保留许可证，同步双语补丁记录，验证外部模块交付，
并按受影响路径执行画面与原生检查。
