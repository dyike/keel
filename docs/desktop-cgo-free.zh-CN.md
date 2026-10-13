# 桌面端无 CGO 交付设计

[English](desktop-cgo-free.md) | 简体中文

状态：交给实现步骤的设计，不能作为当前平台支持声明。
以下单模块修订取代原独立 Fork 分发方案，先完成[模块合并](single-module.zh-CN.md)。
基线：`b2696a6`；2026-10-12 在 darwin/arm64、Go 1.26.4 上核验。

## 目标与边界

让应用可靠使用 Keel 的 Gio、go-text 补丁，并在没有 C 编译器和原生开发头文件的环境中通过 `CGO_ENABLED=0` 构建桌面可执行文件。保留原生窗口、Native GPU、组件行为、渲染算法和已有性能补丁；模块合并按迁移指南修改 Gio/go-text 的 import 路径。系统框架、动态库及 GPU 驱动仍是运行时依赖。

目标为 macOS arm64/amd64、Windows amd64/arm64、Linux amd64/arm64 的 X11 与 Wayland。交叉编译与实际运行分开验收。iOS/Android 的 CGO 迁移延期，已有构建不能回退；WebAssembly 保留现有路径。签名、公证、图标转换、移动端打包工具与桌面可执行文件的编译要求分开说明。

## 历史基线

| 范围 | 证据与影响 |
| --- | --- |
| 分发 | 根 `go.mod` replace 两个嵌套模块；`cmd/keel/new.go` 仅在 `new -replace` 时补齐它们。普通生成应用使用上游；根模块 zip 不包含嵌套模块。 |
| macOS 依赖扫描 | hello 引入 Gio 的 `internal/gl`、`gpu/internal/metal`、`internal/cocoainit`、`gpu/headless`、`app`，以及 Keel 的 `ui/theme/coretext_darwin.go` 和 `runtime/cgo`。 |
| macOS 构建 | `CGO_ENABLED=0 go build ./examples/hello` 在 Gio `internal/gl/util.go` 报 `undefined: Functions`。GPU Backend 注册会将 GL 带进默认 Metal 的编译链。 |
| Native API | 本机 `CGO_ENABLED=0 go build ./native/...` 通过，静态搜索未发现 native 的 `import "C"`；这不代表跨平台运行验收完成。 |
| Linux 源码选择 | `GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go list -e -deps` 选中 Gio 的 X11/Wayland、xkbcommon、GL、EGL、Vulkan 的 CGO 文件。标准库 net/os/user 和 purego 可选 CGO loader 也会出现，需另测关闭 CGO 后的替代实现。本次仅扫描依赖，未执行 Linux 构建或运行。 |
| CI | 已有 Windows 原生窗口和 Linux Wayland 生命周期任务；Linux 安装开发头文件。未发现桌面无 CGO 构建矩阵。 |

`AGENTS.md` 原有清单已经修正。以上为历史基线，后续工作见[验证记录](desktop-cgo-free-validation.zh-CN.md)；其中独立 Fork 的 D1 结果不代表单模块合并通过。

## 决策一：统一 Keel 模块

执行[单模块设计](single-module.zh-CN.md)：Gio、typesetting 成为 Keel 的普通包，改用 Keel 前缀导入，共享根 go.mod、go.sum 和发布版本。取代原 Fork 清单、独立发布、应用 replace 和上游兼容模式；shader 保持独立外部依赖。同步适配包括 gogio 在内的包路径相关工具，保留上一轮有价值的事务生成测试。

用户已明确选择此结构，取代旧的保留模块路径规则。组件行为保持稳定，直接使用 Gio 类型的应用迁移 import。删除嵌套清单使两个库进入根模块 zip；移除根 replace 后可恢复常规指定版本 CLI 安装，需通过外部模块夹具验证。模块合并本身不移除 CGO。

## 决策二：以双 Backend 过渡迁移绑定

每个正在迁移的桌面平台采用明确的选择规则：

| CGO_ENABLED | `purego` tag | 过渡期 Backend |
| --- | --- | --- |
| 1 | 无 | 原 CGO |
| 1 | 有 | Pure Go |
| 0 | 有或无 | Pure Go |

macOS 专用旧文件为 `darwin && !ios && cgo && !purego`，新文件为 `darwin && !ios && (!cgo || purego)`。Go 文件、原有 `.m`/`.c`、汇编、共用定义和测试必须一致选择；共享 Darwin 代码显式分离 iOS，不能让桌面实现误入移动端。不增加任何 C/Objective-C 源文件或 C 辅助函数。旧文件仅用于迁移对照。

每个平台通过下述验收后，切换为默认 Pure Go，再以独立改动移除旧绑定。过渡期通过 `CGO_ENABLED=1 -tags purego` 对新实现执行 race 检查；不能把 `-race` 与关闭 CGO 组合，也不能把 race runtime 的 CGO 等同于应用绑定要求。

### macOS 工作包

1. **先验证 ABI。** 在 amd64、arm64 实机测试 NSRect/NSPoint/NSSize/NSRange、浮点参数、出参、objc super、block 的正向调用与原生回调，重点覆盖 NSTextInputClient 签名。验证 purego 0.11.1 的真实能力，不能因为支持 RegisterClass 就假定支持所有结构体回调。必要时使用 Go/汇编适配并附原生测试；若无法在不增加 CGO 的前提下实现，带限制退回设计。
2. **公共绑定与完整首帧。** 在 Gio 自己的 internal 边界增加最小辅助代码，不能反向导入 Keel 的 `ui/internal/appkit` 或 `native/internal/sys`。迁移 cocoainit、AppKit 启动、窗口/view、DisplayLink、Metal layer/context、GPU Metal 绑定（含 pool/quads）、headless；同时迁移 Keel theme 的 CoreText/CoreGraphics。用真实 GL 绑定解决编译依赖，保留已有 `nometal` 路径，不用空 stub 冒充实现。保持 shader、命令调度、图集行为及资源策略。
3. **交互等价。** 保留现有窗口操作、键盘/修饰键、鼠标滚动、焦点、剪贴板、已有拖放行为、Retina/屏幕变化、全屏、遮挡、休眠唤醒。中文 IME 覆盖 marked text、替换/选区、提交、取消、代理对的 UTF-16 转换、缩放后的候选框定位。保留 AppKitViewEvent、无效 view 事件和销毁次序。
4. **默认切换。** 两种架构的 ABI、真实窗口、文字和 GPU 验收全部通过后才切换。首帧成功不等于完整平台支持。

Objective-C 类和回调只注册一次，使用整数 token 或对象映射路由；销毁时先使路由失效，再释放对象。不让原生侧保留 Go 指针。位图/纹理内存使用原生持有缓冲，或明确合法的有限调用生命周期。所有自有对象、复制 block 及错误路径配对释放，锁定线程内使用 autorelease pool。Metal 初始化失败通过原有生命周期返回明确错误，不静默更换渲染器。

整数/指针热调用缓存函数地址与 selector，使用 `SyscallN`；浮点/结构体使用 ABI 正确的类型适配，不能为速度强行使用整数调用约定，开销用 benchmark 衡量。原生回调不能同步获取 Keel 帧锁；保留 Gio 事件顺序，UI 修改经现有更新机制投递。不用轮询代替 DisplayLink、窗口事件或显式唤醒。

### Linux 工作包

先通过 purego 迁移现有 Xlib/X11，包含 Xcursor/Xfixes、xkbcommon、EGL/GL。保持现有协议和事件模型，不为移除 CGO 另换 X11 栈。验证现有实现支持的 XIM/IBus 行为、键盘布局、剪贴板、DPI、缩放、关闭、唤醒。

再迁移 Wayland 协议元数据、代理构造/封送、监听器/dispatcher、xdg-shell/decoration/text-input、cursor、EGL。借鉴单 dispatcher 模式，不直接引用 Keel 私有包。保持 native/clipboard 所用 display 共享和私有队列、线程约束、configure 确认、原生对象销毁前的 view 失效通知。libwayland 保留的协议对象/签名须有合法原生生命周期；回调注册数量固定，不能每个 proxy/请求分配 FFI 回调。

Vulkan 也在默认 Linux 依赖图中，需迁移 loader/surface 绑定，保留渲染器。X11/EGL 中间验收可以用 `nowayland,novulkan`，但最终必须在无额外 tag 的默认构建下支持两个窗口系统，并覆盖原有 `nox11`、`nowayland`、`novulkan` 配置。桌面约束隔离 Android/BSD，避免破坏旧路径。缺库/缺符号给明确错误，沿用已有 Backend 回退规则；运行时不需要编译器、pkg-config 或头文件。

## 实施顺序与文件归属

按以下依赖顺序拆为可独立审查的改动：

| 步骤 | 主要位置 | 交付/完成条件 |
| --- | --- | --- |
| D0 | `AGENTS.md`、测试脚本、`.github/workflows/` | 修正清单，建立可复现基线与无 CGO 矩阵。 |
| D1（修订） | 根清单、Fork 导入、依赖规则、CLI/打包器、发布夹具 | 完成单模块设计，外部应用仅 require Keel 即可得到两个库。 |
| M0 | Gio internal 原生辅助代码、ABI 测试 | 两种 Mac 架构原生结构体回调通过。 |
| M1 | Gio app/Metal/GL/headless/cocoainit、Keel theme | 无 CGO hello、离屏截图、原生首帧；仍可与 CGO 对照。 |
| M2 | Gio app、`ui/window` 现有夹具、theme 测试 | IME、生命周期、GPU/性能等价，macOS 切为默认。 |
| L1 | Gio X11/xkb/EGL/GL | X11 运行验收，中间阶段可使用已说明的临时 tag。 |
| L2 | Gio Wayland/Vulkan、native clipboard 测试 | Linux 默认无 CGO 构建及 X11/Wayland 运行验收，切为默认。 |
| R | CLI build/doctor/run、发布夹具、双语文档 | 新应用至打包产物全链路通过；等价验收后才删除桌面旧绑定。 |

先完成修订 D1，保留已有 D0/M0 工作，再继续绑定迁移。设计完成不代表实现通过。L2 或任何 Mac 架构未验证时，不能声称 Desktop CGO-Free 整体完成。每项 Fork 补丁附测试和 `third_party/README*` 双语记录。若新增 Keel 包，先更新架构/依赖表，保持 ui/native 隔离。支持状态变化时同步 README、getting-started、testing、troubleshooting、CLI 的双语文档。

## 验收证据

1. **分发：** 单模块外部归档夹具、全新模块缓存、`GOWORK=off`。应用仅 require Keel，两个库包树归属根模块；验证常规 CLI 安装、本地/离线生成与打包器新路径。不再验收 Fork 锁定清单或独立子模块发布。
2. **编译：** 所列目标均执行关闭 CGO 的 `./...`、`./native/...` 构建，根检查覆盖 Gio/typesetting。用 `go list -deps -json` 确认无选中 CgoFiles 或包错误。仅诊断扫描/race 开启 CGO；无编译器/头文件环境中 Go 可独立链接。这是后续 CGO 门禁，不是单模块合并的前置完成条件。
3. **兼容：** 覆盖过渡选择和渲染 tag；Windows 保留当前实现。验证外部自定义组件迁移后的 import、旧上游核心包不再进入依赖图（允许 shader）。保持移动端/Web 构建与打包。随模块合并撤销旧的删除 replace 后编译上游测试。
4. **运行：** macOS arm64/amd64、Windows、Linux X11/Wayland 原生运行。执行 `KEEL_DESKTOP=1 go test -run RealWindows ./ui/window` 和 Gio app 生命周期测试；覆盖立即关闭/重开、回调内 raise/close、全屏、焦点、销毁、剪贴板。中文 IME 必须验收真实预编辑和候选框，不能仅注入最终文本。记录 OS/GPU/缩放/输入法；软件渲染 CI 不能代替原生 GPU 验收，没有测试的硬件明确待验。
5. **画面与性能：** 改造前同机保存 hello 与组件矩阵截图；不改光栅化时要求输出一致，有差异须解释。执行 CoreText/emoji、Metal idle/resize、readback、资源重建回归。至少五轮交替前后 benchmark，使用 `-benchmem` 与 benchstat，包含 `BenchmarkGlyphRendererScrolling`；记录 idle、滚动、resize、文字场景的帧延迟 p50/p95、CPU、分配、GPU/进程内存。可重复且超出噪声的 CPU/延迟/分配恶化超过 5% 时阻止默认切换，需定位解决。空闲窗口不能新增周期唤醒，不预设 FFI 更快。
6. **CLI 与发布：** `keel run`、桌面 build 尊重所选 Backend 和 CGO 设置。Linux 显式设置 GOOS 并证实打包可用后解除仅本机构建限制。审计固定版 gogio 是否强制 CGO，必要时适配桌面调用或直接编译可执行文件。doctor 区分编译工具和 `.app`/签名工具。独立验收 `go build`，再验收真实 `keel new → run → build` 产物。完成 `AGENTS.md` 要求的格式、vet、test、race ，包括已合入根模块的库目录及更新后的 `AGENTS.md` 规则。

实现报告必须区分已执行命令、失败、跳过的桌面验收和待发布步骤。交叉编译成功、模拟事件以及设计文档都不能当成原生运行通过。
