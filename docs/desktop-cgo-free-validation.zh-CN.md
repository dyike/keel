# 桌面无 CGO 实现验证记录

[English](desktop-cgo-free-validation.md) | 简体中文

2026-10-13 在 macOS 27.0.1、darwin/arm64、Go 1.27.1 上验证。
单模块步骤 D1 已实现并通过本地验收。Desktop CGO-Free 尚未完成：
macOS/Linux 的 UI 绑定仍需 CGO。本轮画面对比基线为 `8d59576`。

## 根级目录迁移复验

两个基础库实际移到 gio/ 和 typesetting/，导入前缀分别为
`github.com/dyike/keel/gio` 和 `github.com/dyike/keel/typesetting`。
`go list -json ./gio/layout ./typesetting/shaping` 的两个结果都以
`github.com/dyike/keel` 为 Module.Path，仓库只剩根 go.mod/go.sum。
旧 third_party 基础库路径只保留在双语迁移表中，明确标注为未发布的中间路径。
上游核心路径用于历史来源、迁移说明或拒绝检查；shader 保持独立外部模块。
站点扫描排除迁移后的上游文档。

迁移后重新执行格式、vet、完整 test、完整 race、站点测试、Windows
amd64/arm64 无 CGO 门禁和真实源码外部发布夹具，全部通过。
夹具还拒绝归档中的过时基础库源码/模块，并通过原生 macOS 打包后的应用 ID
探针验证新链接符号。原生 arm64 ABI 探针通过，hello 截图字节一致。

完整检查并发运行时，第一次真实窗口检查在 reopen/update/now 超时。
单独复跑全部真实窗口夹具通过，用时 16.443 秒。超时原因尚未确定，
未修改生命周期实现。下文 792 张矩阵截图对比属于根级迁移前的单模块验收。

## 单模块实现（D1）

Keel、Gio 和 typesetting 共用根 go.mod/go.sum。两个基础库使用 Keel 前缀导入，
随根模块归档发布；shader 保持独立。源码检查覆盖所有平台文件的 import，
包括当前主机 build tag 未选中的文件，并约束基础库的依赖方向。

脚手架只写一个 Keel require，本地检出模式只写一个 Keel replace。
精确版本的离线生成不需要发布元数据或网络；未缓存的版本查询明确失败。
本地路径支持空格，必须指向模块根目录。暂存生成、tidy 失败回滚和并发文件保护
仍有测试。此前未发布的独立 Fork 锁清单、解析器、backend 模式和 deps 命令
已被本实现取代。

维护的 gogio v0.10.0 副本按迁移后的 Gio 前缀设置应用 ID、启动参数，
识别权限和查找资源。Run 选项提供工作目录、环境与输出流，串行调用隔离上游
包级状态；默认产物写入指定工作目录。Windows 资源和 iOS 模拟器 shader
适配仍由 CLI 处理。架构、迁移说明、补丁记录和 CI 均已同步单模块规则。

## 命令及结果

| 检查 | 结果 |
| --- | --- |
| `gofmt -l .`、`git diff --check`、`go mod tidy -diff` | 通过，无需修改格式、空白或依赖文件。 |
| `go vet ./...` | 通过，包含两个基础库。 |
| `go test ./...`、`go test -race ./...` | 通过，包含 MCP 测试和基础库回归。 |
| 后续修改后的 CLI、打包器、边界、站点定向测试与 race 检查 | 通过。 |
| `python3 scripts/check-module-release.py` | 真实源码发布夹具通过，详见下节。 |
| 六个桌面 OS/架构组合的 `check-desktop-cgo-free.sh native` | 交叉构建通过，CC/CXX 指向不存在的编译器。 |
| Windows amd64/arm64 的 `check-desktop-cgo-free.sh all` | 关闭 CGO 的完整根模块交叉构建通过。 |
| darwin/arm64 的 `CGO_ENABLED=0 go build ./examples/hello` | 仍在 Gio GL 报 `undefined: Functions`，属于后续迁移门禁。 |
| `CGO_ENABLED=0 go test ./gio/internal/nativeabi -count=1` | 原生 arm64 通过。 |
| darwin/amd64 的无 CGO ABI 测试二进制 | 交叉编译通过，未在原生 Intel Mac 执行。 |
| `KEEL_DESKTOP=1 go test -run RealWindows ./ui/window -count=1` | 原生 macOS 主机通过。 |

这些是本地命令结果，本轮未在远程 runner 执行更新后的 CI。
交叉构建不代表原生运行支持；不存在的编译器路径也不能替代无编译器隔离环境。

## 发布与打包夹具

文件代理把真实工作区归档作为 `v0.0.0-fixture` 提供。
使用全新模块缓存和 `GOWORK=off` 验证：

- 外部 UI 应用只 require Keel，没有 replace。两个基础库的 Module.Path
  都是 Keel，真实源码包含在归档中，依赖图没有嵌套清单或上游核心模块。
- 外部组件用迁移后的 Gio layout、unit、clip、paint 包实现 core.Widget，
  并编译 Windows 版本。
- 普通按版本 `go install` 能从该归档安装 CLI；默认生成选择 CLI 自身版本。
  精确版本离线生成在发布尚未缓存时成功；本地生成只有一个 replace。
  应用、检出目录和模块缓存路径均覆盖空格。
- 安装后的 CLI 打包 Windows amd64/arm64 和 WebAssembly；Web 产物包含
  main.wasm、wasm.js 和 index.html。
- 原生 macOS 上，维护的打包器使用现有 CGO 后端生成真实 app bundle。
  执行文件输出 `dev.keel.releaseprobe`，并同时保留原生命令行参数与链接时
  注入的启动参数。

定向测试还覆盖 WebAssembly 应用 ID/启动参数链接、HTML 元数据，
以及新路径下 Android/iOS 资源和 Android 权限包的发现。
本轮未执行完整 APK/iOS 打包：主机只有 Command Line Tools，没有完整移动端
SDK 工具链。没有发布生产标签；最终发布的远程安装仍是发布门禁。
仅合成夹具关闭校验服务器验证。

## 画面与窗口验收

Hello 的前后 PNG 字节一致。两份完整组件矩阵各有 792 张 PNG：
99 个 section，浅色/深色、两种宽度、两种缩放。

动画组件在不同墙钟时间采样会产生差异。临时 Go overlay 只把两份检出目录的
截图帧上下文固定为 `time.Unix(0, 0)`，未修改生产源码。
使用同一帧时间后，全部 792 张 PNG 和矩阵索引字节一致。
macOS 上基线 overlay 使用解析后的 `/private/tmp` 源码路径。
真实窗口生命周期及多窗口夹具也通过。

## M1 进展：Keel 的 CoreText 绑定（2026-10-24）

`ui/theme` 通过 `coretext_cgo_darwin.go`（`cgo && !purego`）调用 CoreText 栅格化字形；不使用 cgo 时改用 `coretext_purego_darwin.go`（`!cgo || purego`）。字体缓存、留边和线性光转换为两者共用。与 cgo 代码一样，位图上下文绘制到一块 Go 缓冲区，其生命周期限定在一个函数内；改由上下文自己持有像素时实测慢三倍。整数和指针调用通过 `SyscallN` 按地址调用；两种文字颜色是一次性创建的 CGColor；只有边界查询在 amd64 上保留 `RegisterFunc` 包装，因为它返回的矩形通过隐藏指针传递。

| 检查 | 结果 |
| --- | --- |
| `go test ./ui/theme`，带与不带 `-tags purego`；`-race -tags purego` | 通过。 |
| 组件截图：11 个文字密集页面 × 浅色/深色 × 1x/2x，以及 `hello` | 两种绑定的 44 张和 1 张 PNG 逐字节相同。 |
| 12 次 emoji 栅格化（Apple Color Emoji，三种尺寸）及浅色/深色文字遮罩 | 两种绑定的 SHA-256 相同。 |
| `BenchmarkCoreTextMask`（一次字形缓存未命中），Apple M4 | cgo 2.7 µs、4 次分配；purego 3.3 µs、19 次分配。稳定帧不做栅格化。 |
| `CGO_ENABLED=0 go vet ./ui/theme` | 仍在 Gio 的 `internal/gl` 失败：`ui/theme` 经由无头渲染器依赖它，是 M1 的后续步骤。 |

在 macOS 默认切换之前，cgo 绑定仍是默认值，切换后删除。未在原生 amd64 上运行。

## M1 进展：Gio 的 GL 绑定（2026-10-24）

`gio/internal/gl` 新增面向 macOS 和 Linux 桌面的 purego 绑定，通过 `CGO_ENABLED=0` 或 `-tags purego` 选用；iOS、Android 和 BSD 仍使用 `gl_unix.go`，`go list` 已逐个配置确认。借此，`CGO_ENABLED=0 go vet ./gio/internal/gl` 在 darwin 与 linux 的 amd64、arm64 上均通过。在 macOS 上使用 `-tags "nometal purego"` 时，无头与绘制测试经由 OpenGL 渲染并通过；`gio/gpu` 在 `nometal` 下无论哪种绑定都无法构建，与 `main` 相同（其着色测试使用仅 Metal 提供的 `AddTinted`）。`hello` 的无 cgo 构建现在止于 `gio/app`。Linux 上 GL 的性能与运行情况留待 L1 测量。

## M1 进展：Gio 的 Metal 后端（2026-10-24）

`gio/gpu/internal/metal` 新增面向 macOS 的 purego 绑定，通过 `CGO_ENABLED=0` 或 `-tags purego` 选用；iOS 仍使用 cgo 文件，`go list` 已对 iOS、带 cgo 的 macOS 和不带 cgo 的 macOS 分别确认。在 Apple M4（darwin/arm64）上对比两种绑定：

- 像素：hello 逐字节相同。组件矩阵 793 张截图中 721 张逐字节相同；其余 72 张是动画页面（spinner、progress、skeleton、status marker 及包含它们的页面），cgo 绑定连续两次运行时同样是这 72 张不同。
- 测试：`gio/...` 和 `ui/...` 在 `-tags purego` 下通过，包括真实窗口测试（`KEEL_DESKTOP=1`）。
- 每帧分配：旋转路径 6 次（cgo 为 7 次），重复裁剪 4 次（cgo 为 5 次）。purego 绑定自身的调用中只有阻塞等待会分配。
- CPU：交替运行四轮，6000 帧无头渲染的用户加系统时间 purego 为 1.97–2.13 s，cgo 为 2.01–2.07 s，差异在噪声范围内。这些帧的绘制调用很少；单次调用约 55 ns，未合批的数千次绘制可能更慢。
- 泄漏：两种绑定在约 15000 帧中按类统计的 Objective-C 对象数保持不变；编码器调用不使用自动释放池。

amd64 只做了编译和 vet（未安装 Rosetta），其经 `RegisterFunc` 按值传递结构体的调用尚未运行。`hello` 的无 cgo 构建仍止于 `gio/app`；`gio/gpu/headless` 仍使用 cgo。

## 保留的 ABI 工作与待验收项

D0 编译门禁和 M0 无 CGO Objective-C 探针保留在根模块中。
原生 arm64 探针覆盖 range、寄存器耗尽、嵌套 rect、可空原生出参、
super 调用和复制 block；标量浮点回调返回仍明确不支持。
原生 amd64 执行与精确的 NSTextInputClient 协议行为还需验证。

单模块评审后继续 M0/M1：迁移 AppKit、cocoainit、Metal/GL/headless
及 Keel CoreText 绑定，再在两种 Mac 架构验证 IME、生命周期、GPU 和性能。
Linux X11、Wayland、Vulkan 迁移，原生 GPU/IME 验收、无编译器隔离环境、
移动端打包与生产发布均待完成。本次结构调整未将 macOS/Linux 后端提升为无 CGO。
