# 测试

[English](testing.md) | 简体中文

界面测试大部分不需要开窗口：`ui/internal/uitest` 测试工具通过 Gio 真实的输入路由送入点击、按键、文字，组件走的代码和真实窗口完全一样。`go test -race ./...` 在本机跑完大约 10 秒。

## 桌面无 CGO 编译门禁

```sh
GOOS=windows GOARCH=amd64 bash scripts/check-desktop-cgo-free.sh
GOOS=linux GOARCH=arm64 bash scripts/check-desktop-cgo-free.sh native
```

默认 `all` 模式先检查 native 能力，再检查包含 Keel Gio 和 go-text 的完整模块。脚本关闭 CGO 和 workspace，发现依赖包错误、选中的 CgoFiles 或构建错误即失败。`native` 模式仅检查 `native/*`，不能视为桌面 UI 支持验收。
六目标 CI 矩阵在每次 push/PR 检查 native 能力及完整 Windows 构建。完整 macOS/Linux 门禁仍待迁移，可通过 workflow 的 `full` 参数执行；不使用静默容忍失败。将 CC/CXX 设为不存在的路径可防止意外调用 C 编译器，但不能代替隔离的无编译器/头文件发布验证。
编译通过不等于运行、IME、原生 GPU、打包或性能验收；详见[桌面交付设计](desktop-cgo-free.zh-CN.md)。

Ubuntu 的 `distribution` job 在脚手架测试前安装 Linux 图形开发库。`TestNewProjectCompiles` 对生成的 GUI 项目执行宿主平台 vet；Linux 上会开启 cgo，仍需要 Gio 的 Wayland、X11、xkbcommon、EGL/GLES 和 Vulkan 头文件。独立的 `compile` 矩阵继续关闭 cgo，并将 C/C++ 编译器路径设为不存在的文件。

`python3 scripts/check-module-release.py` 把工作区打包成本地模块代理中的测试用 Keel 发布，用 `keel new` 创建只 require Keel 的应用，确认其 Gio 和 go-text 包来自 Keel 模块（而非上游 `gioui.org` 或 go-text），编译外部自定义组件，并从该发布 `go install` CLI。它检查默认、离线、本地生成及含空格路径，无 CGO 打包 Windows amd64/arm64 和 WebAssembly；macOS 上还用现有 CGO 后端运行打包后的应用 ID/启动参数探针。外部依赖使用当前配置的 Go 代理；合成发布由本地代理提供，使用全新模块缓存并关闭 workspace。

## 交互测试

`ui/` 下的模块都可以引用它（它在 `internal` 里，外部项目引用不到）：

```go
func TestInputSubmit(t *testing.T) {
    var got string
    f := Input("").OnSubmit(func(s string) { got = s })   // 在 package kit 里
    h := uitest.New(el.Root(f)) // 布局一帧
    h.Click(20, 10)      // 点击获得焦点
    h.Type("你好")        // 输入文字
    h.Key("⏎", 0)        // 回车
    if got != "你好" {
        t.Fatalf("OnSubmit saw %q", got)
    }
}
```

| 函数 | 作用 |
| --- | --- |
| `uitest.New(w)` | 把组件放在左上角 (0,0) 布局一帧 |
| `uitest.NewFunc(fn)` | 用任意帧函数，比如 `window` 包的测试用 `w.layout` 带上根视图和快捷键 |
| `h.Click(x, y)` | 在 (x, y) 按下并松开左键，然后画一帧 |
| `h.Key(name, mods)` | 按下并松开一个键，然后画一帧。`name` 是 Gio 的键名：字母大写 `"A"`，回车 `"⏎"`，也可以用 `key.NameReturn` 等常量 |
| `h.Type(s)` | 向有焦点的输入框插入文字，然后画一帧 |
| `h.Frame()` | 执行 `core.Update` 队列并画一帧 |

坐标规则：测试视口是 400×300，1dp = 1 像素。`uitest.New` 不带根视图，组件从 (0,0) 开始；用窗口布局测试时，内容从 (24,24) 开始。

每个 `h.Click` 等操作后面自带一帧，但回调修改了状态、要看重绘结果时，多调一次 `h.Frame()`。

## 模块边界测试

`internal/deps` 检查每个模块只引用了允许的包（见[架构 · 模块](architecture.zh-CN.md#模块)）：

- 下层模块引用了上层（比如 `el` 引用 `kit`），同层互相引用，或者 `native/*` 引用了 `ui`、Gio，测试失败；
- 新增了模块目录却没在允许表里登记，测试失败。

它随 `go test ./...` 一起运行。

## 什么该测

- 回调触发次数：点一次触发一次，禁用时不触发。
- 回调参数：`OnChange` 收到的是新值。
- 程序调用和用户操作的区别：`SetValue` 不触发 `OnChange`，打字会触发。
- 解析类函数（快捷键字符串）：合法输入和每一种非法输入。
- `native` 包：参数校验路径（不需要权限，也不会真的动鼠标）。

## Agent 端到端测试

`cmd/keel-mcp` 让 Agent 通过 MCP 驱动应用：读元素、点击、输入、滚动、截图。应用在内存里渲染，不弹窗口。用法见 [Agent 端到端测试](automation.zh-CN.md)。

仓库里有两层相关测试，都随 `go test ./...` 运行：

- `ui/window/automation_test.go`：进程内测试自动化模式本身，包括滚动、Tab 移动焦点、回调里关闭窗口、禁用和勾选状态的报告。
- `cmd/keel-mcp/main_test.go`：以 MCP 客户端身份启动 `keel-mcp`，把 multiwindow 示例完整走一遍，并测试连接用户自己启动的应用（`attach`）。改了协议、工具或组件的语义信息后，它会第一个失败。

## 截图对比

改了主题、间距、字体后，渲染前后截图，确认只有预期的地方变了：

```sh
go run ./examples/hello -screenshot /tmp/before.png
# 修改代码
go run ./examples/hello -screenshot /tmp/after.png
cmp /tmp/before.png /tmp/after.png && echo 完全一致
```

纯重构应该输出"完全一致"。截图由 GPU 离屏渲染，同一台机器上结果稳定；不同机器、不同系统字体下像素可能不同，不要把截图提交成跨机器的基准文件。

## 真实窗口测试

有些 bug 只在真实窗口里出现，比如多窗口之间的死锁。`ui/window/desktop_test.go` 在子进程里运行 `ui/window/testdata/raise`：打开两个窗口，在界面代码里（持有帧锁）调用 `Raise`、`Close`，15 秒内没走完就判定为死锁。

```sh
KEEL_DESKTOP=1 go test -run RealWindows ./ui/window
```

这组测试也覆盖窗口首次居中、立即关闭后重开。多显示器切换、各平台窗口管理器和系统偏好变化仍需手动验证。

默认跳过，因为它会在屏幕上弹出窗口，而且需要图形界面环境。改动 `ui/window`、`ui/internal/loop` 里和窗口、锁相关的代码后必须跑一次。新增窗口方法时，把它加进 `testdata/raise` 的步骤里。

## 需要手动验证的部分

无界面测试覆盖不到这些，改动相关代码后手动跑一遍：

| 场景 | 怎么验证 |
| --- | --- |
| 真实窗口的打开、关闭、置前 | `go run ./examples/multiwindow`，点"打开设置窗口"两次，应该只有一个设置窗口 |
| 最后一个窗口关闭后退出 | 关掉所有窗口，终端里进程应结束 |
| 窗口快捷键 | 在主窗口按 ⌘+, |
| 全局快捷键 | `go run ./examples/hotkey`，切到别的应用按 ⌘⇧K，计数增加 |
| 后台 `core.Update` | hotkey 示例里的时钟每秒走 |
| 权限、截图、合成输入 | 需要授权，按 [原生能力](native.zh-CN.md) 的说明手动测 |
| 中文输入法 | 在输入框里用拼音输入，候选框位置正确，上屏后内容正确 |

## 全组件截图矩阵

```sh
go run ./examples/components -matrix /tmp/keel-component-matrix
# 单个组件：窄窗口、1×、深色
go run ./examples/components -section button -width 320 -scale 1 -theme dark -screenshot /tmp/button.png
```

矩阵按已注册的每个组件生成浅 / 深色、320 / 680dp 宽、1× / 2× 共八张首帧截图，并输出可浏览的 `index.html`。每个案例重新构造组件，避免上一个案例的布局缓存影响结果。大图通过索引中的图片链接打开；截图只包含当前视口，滚动后的内容和浮层仍需交互测试。

`window.Screenshot` 保持默认 2×；需要其他缩放时用 `ScreenshotAtScale`。尺寸以 dp 表示，PNG 尺寸按缩放四舍五入到物理像素。非法尺寸或缩放返回错误。

组件示例中的固定展示宽度必须同时设置 `MaxW(el.Full)`；否则测试到的是被裁掉的宽画布，而非组件的窄布局。并列操作用 Wrap，保留需要验证横向滚动的内容宽度。Sidebar 示例在 600dp 以下初始折叠，仍可手动展开；Dock 的多列工作区需要足够宽度，不把 320dp 截图当成手机布局承诺。

`cmd/keel-mcp` 的示例级回归包括：订单筛选、详情和删除、新建订单纯键盘流程；聊天流式输出、代码/表格和中断；设置页主题、搜索和语言切换；Dock 移动面板、切换后保留输入、菜单 Esc 与关闭面板。`TestOrdersKeyboardOnly` 不点击控件，从 Mod+N 开始用 Tab、方向键、Space、Enter 完成保存。

## 托盘与驻留回归

`go test -race ./native/tray ./ui/internal/loop` 检查菜单快照、旧事件、串行回调、关闭后丢弃排队回调，以及没有窗口时更新仍执行。Linux 的 SNI/DBusMenu 协议测试使用纯 Go 的序列化与菜单模型，不要求桌面会话。

```sh
KEEL_DESKTOP=1 go test -run RealWindows ./ui/window
go run ./examples/tray -smoke
```

真实窗口用例从无窗口启动，关闭最后一个窗口后重新打开，再检查退出回调。托盘 smoke 额外创建、更新和移除系统图标。菜单实际点击、明暗外观、高 DPI、Windows Explorer 重启和 Linux Host 重启需要分别在对应桌面上验证；跨平台编译不代表运行验收。

菜单回归覆盖图标快照、非法更新的原子性、启用/禁用模板缓存（预热后图标布局零分配）、固定时间的动画进度、减弱动画和结束后的帧调度。Agent 测试检查菜单项标签、禁用/勾选状态、操作路由和控制器取消。Windows 目标平台测试验证位图挂载、预乘像素、保持宽高比的透明填充，以及替换/失败时的资源释放；交叉编译测试不代表已执行。macOS 测试创建真实 AppKit 图片和弹出菜单跟踪循环。

```sh
python3 scripts/check-components.py --offline --desktop --only checks interactive
```

## 一键组件验证

```sh
python3 scripts/check-components.py
# 离线运行：依赖和 Go 工具链须已缓存。
python3 scripts/check-components.py --offline
# CI 不显示原生窗口；仍需 GPU 后端和图形开发库。
python3 scripts/check-components.py --offline --headless
```

标准流程串行执行：脚本回归测试、格式检查、vet、全仓库 Go 测试、race、Windows 无 CGO 构建、Linux native/CLI vet；随后运行独立进程的全组件和大数据基准、固定时间的完整 GPU 截图矩阵；最后启动 `go run ./examples/components`，遍历实际注册的全部场景并验证下面的交互。`--desktop` 还会在检查阶段运行真实窗口生命周期测试，包含 macOS 菜单图片替换和两种原生取消模式；同时运行 macOS 原生托盘测试及 `examples/tray -smoke`。桌面托盘 smoke 需要可用的通知区域 / SNI Host。需要 Python 3 和 Go；Python 部分只使用标准库。macOS 需要 Xcode 命令行工具；Linux 仍需 Gio 图形开发库和可用的图形/显示后端。

默认产物保存在新的 `work/component-checks/<时间戳>/`：`summary.json`、由 Go 注册表导出的 `manifest.json`、各命令日志、原始基准和中位数 `benchmarks.csv`、`matrix/index.html`，以及交互报告和截图。失败立即停止并返回非零退出码；已有内容的结果目录不会被覆盖。每个基准场景启动独立进程，检查、GPU 渲染和交互不会与性能测量并行。

```sh
# 只跑指定阶段；默认每个示例场景测量 3 次，每次 200ms。
python3 scripts/check-components.py --only benchmarks --count 3 --benchtime 1s
python3 scripts/check-components.py --only matrix interactive --offline
# 同机先保存基线，再修改代码并比较。
python3 scripts/check-components.py --output work/component-baseline
python3 scripts/check-components.py --baseline work/component-baseline
# 快速验证流程能否跑通；1x 样本不能证明性能收益。
python3 scripts/check-components.py --offline --count 1 --benchtime 1x --large-benchtime 1x
```

`--baseline` 要求场景集合一致，矩阵 PNG 变化会失败，并生成 `comparison.csv`（百分比为正表示回退）。基准百分比供评审使用，不自动设置耗时或分配门槛。比较时保持机器、字体、Go 版本和基准参数一致。截图时间通过生成的 Go overlay 固定，不修改生产代码。视觉变化符合预期时，应评审后建立新基线。

交互检查覆盖复选框及禁用状态、开关、中文及多行输入、弹窗保存/Escape/重开、搜索和万条选项多选、十万项虚拟列表、表格排序和滚动、日历键盘跳过禁用周日、日期选择器重开后的选中状态，以及应用菜单图标、禁用项、操作路由和 Escape 关闭。日历读取当前展示的日期，不硬编码月份。自动化操作的是 Gio 的影子输入路由；默认应用同时显示原生窗口，`--headless` 隐藏原生窗口。两种模式均不能验证系统输入法候选框或原生鼠标事件投递，GPU 呈现延迟和动画性能需要单独 profile。

`--offline` 禁用 Go 模块、VCS 和工具链下载，未缓存的依赖会明确报错。它不会让截图变成无 CGO 构建，也不会禁用示例自身的可选网络图片请求。

## 全组件帧分配基准

组件示例的基准自动遍历全部注册场景，覆盖控件、输入、浮层、导航、图表和虚拟化内容：

```sh
go test ./examples/components -run '^$' -bench '^BenchmarkComponentFrame$' -benchmem -benchtime=200ms
# 单独测一个场景；第二段表达式也加首尾锚点，避免匹配名称后缀。
go test ./examples/components -run '^$' -bench '^BenchmarkComponentFrame$/^calendar$' -benchmem -benchtime=1s -count=3
go test ./ui/kit -run '^$' -bench '^BenchmarkTableFrame300k$' -benchmem -benchtime=1s -count=3
go test ./ui/kit -run '^$' -bench '^BenchmarkCodeEditorPlainLargeFileFrame$' -benchmem -benchtime=1s -count=3
```

`BenchmarkComponentFrame` 在 680×1040dp、1× 缩放下预热场景，使用真实 Gio 输入路由并固定动画时间。它测量 CPU 布局和输入注册，不包含 GPU 提交或后台更新投递。关闭浮层的静态帧不能代表打开菜单、弹窗、滚动、输入或动画的延迟；交互回归和 GPU 截图矩阵需要同时验证。

稳态帧分配次数是回归指标。普通测试在预热三帧后检查组件示例中的 label（35）、list（250）、button（260）、calendar（1,050）、table（550）、input（500）六个分配上限。上限是防回退门禁，不是性能目标；每次分配优化仍要用基准同时比较前后的字节数和次数。

设置 `KEEL_FRAME_STATS=all` 可记录每个原生窗口帧的布局、提交耗时，以及该时段进程的分配次数、字节数和完成的 GC 轮数。`KEEL_FRAME_STATS=1` 只记录超过 8ms 的帧；其他非负数字表示毫秒阈值。正常运行时不要设置：读取 runtime metrics 本身是诊断开销。分配计数覆盖这一帧期间的整个进程，并发后台工作也可能计入。

比较时，用同一份基准代码分别构建基线和修改后的测试程序，每个场景启动独立进程。可选的高亮、图片加载会影响同进程中后续场景的分配计数。可以用 Go overlay 选择原始生产代码，无需撤销工作区修改。普通编辑器基准不设置语言；已有的大文件高亮基准仍包含整文件后台高亮。

### 绘制路径分配检查，2026-10-15

路径坐标系描边缓存使用独立的渲染器基准：

```sh
go test ./gio/gpu -run '^$' -bench 'Benchmark(Rotating|Changing)StrokeVertices' -benchmem -benchtime=500ms -count=3
```

Apple M4、darwin/arm64、Go 1.27.1，三个独立样本的中位数：

| 三次曲线描边顶点生成 | 改前耗时 | 改后耗时 | 字节：改前 → 改后 | 分配次数：改前 → 改后 |
| --- | ---: | ---: | ---: | ---: |
| 相同路径，仅旋转变换 | 17.88 µs | 5.29 µs | 42,648 → 0 | 103 → 0 |
| 路径内容变化 | 18.53 µs | 16.44 µs | 42,939 → 8,232 | 103 → 2 |

第一项已预热缓冲；第二项每帧改变控制点，包含缓存未命中和独立所有权的复制。scratch 不与缓存曲线或 Ops 共用可写数组。缓存最多保留 128 组轮廓及 1 MiB 路径/曲线数据，淘汰未使用项，超大路径不缓存；渲染器清理或释放时丢弃全部缓存。斜切、反射等仿射变换的顶点和边界与原算法完全一致。旋转、圆角祖先裁剪和透明度下的冷热 GPU 像素一致；792 张固定时钟组件截图与 main 逐字节一致。

这只测 CPU 顶点生成，不代表整帧、端到端 FPS 或原生内存。原生滚动、缩放和四阶段 footprint/GC 对比仍待桌面解锁后验收。不改变着色器、drawable 数量、GC 或空闲回收。

`BenchmarkCategoryFrame` 使用实际 `-section` 分类组合，窗口为 1360×1720
像素（680×860dp，2×），接入真实输入路由，动画时间按 60 Hz 推进。
每个分类单独启动进程测量，期间不运行其他基准：

```sh
go test ./examples/components -run '^$' -bench '^BenchmarkCategoryFrame$/^controls$' -benchmem -benchtime=1s -count=3
```

Apple M4、darwin/arm64、Go 1.27.1 上，每个分类运行三轮，每轮一秒，
稳态帧中位数如下。基线已包含节点 arena；两个二进制使用相同基准。
修改版增加有界图标缓存、借用字形缓冲，并保留状态标记与主题示例的视图对象。

| 分类 | 修改前 B/帧 | 修改后 B/帧 | 修改前分配次数/帧 | 修改后分配次数/帧 |
| --- | ---: | ---: | ---: | ---: |
| controls | 334,888 | 152,289 | 2,169 | 2,050 |
| inputs | 484,933 | 183,947 | 2,176 | 2,021 |
| overlays | 175,744 | 38,832 | 558 | 487 |
| data | 726,864 | 599,584 | 3,262 | 3,191 |
| shell | 384,267 | 94,172 | 1,339 | 1,182 |

该基准只测 CPU 布局和绘制指令生成，不包含 GPU 执行、原生窗口 footprint、
滚动延迟或端到端帧率；没有修改 GC 和 drawable 策略。图标缓存以有界保留
换取更少的栅格化和上传：每个图标最多八种样式、1 MiB 像素，超大当前图片只保留自身。

设置 `KEEL_STATE_STATS=1`，按根节点与帧记录状态新增、淘汰数量，附带前八个
新增元素的 ID、角色和名称。固定的根节点编号可区分根节点重建和元素 key 不稳定。
日志仅按需启用，不请求重绘；新增可见行本来就会创建状态，不能仅凭新增计数判断错误。
这次追踪发现 controls 的状态标记和主题示例分别在每个预热帧替换 21、4 个状态，
原因都是在渲染时创建以指针标识身份的视图。保留视图对象后，后续追踪中不再出现这两处
状态替换，主题示例的输入状态也能跨渲染保留。

### 后续分配优化，2026-10-15

Apple M4、darwin/arm64、Go 1.27.1 上，以完成节点 arena 后的代码为基线，99 个组件场景的分配字节和次数全部进一步下降。逐场景降幅中位数分别为 42.2% 和 44.2%，不加权的场景平均分配从 57,243 降到 35,420 字节/帧。每个场景独立进程运行三次、每次 200ms；前后都没有调整 GC 参数。这些归档测量使用了当时工作区的三 drawable 和关闭空闲回收实验；两项策略都不属于 arena PR，Metal 保持 main 的策略，空闲堆回收仍默认开启。

| 场景 | 分配字节/帧，前 → 后 | 分配次数/帧，前 → 后 |
| --- | ---: | ---: |
| Label | 9,808 → 3,336 | 51 → 26 |
| Button | 106,583 → 64,989 | 463 → 202 |
| Calendar | 413,218 → 190,688 | 2,344 → 842 |
| Table | 149,992 → 79,336 | 817 → 424 |
| List | 58,464 → 29,848 | 374 → 200 |
| Input | 178,293 → 122,612 | 806 → 400 |

本轮把可选样式值内联，只有实际执行回调时才分配私有绘制样式，按需缩小独立句柄批次，复用帧内测量条目和编辑器 filter，并按不可变内容缓存文档映射。按钮只在有界缓存中共享纯数值样式函数，不缓存用户回调或帧 Context；日历每个月共用一个禁用样式函数。样式仍区分未设置与显式透明色、正常字重、零偏移；回调副本和指针句柄都不会在后续帧覆盖复用。

30 万行表格滚动基准，前后各三次、每次 300ms：368,397 → 297,361 字节/帧，1,622 → 1,197 次/帧，布局耗时中位数 525µs → 506µs。组件布局耗时的逐场景改善中位数为 5.0%，小幅变化仍受主机负载影响。初次短测中 Bubble 和大文件普通编辑器的耗时回退，前后各五次、每次 1s 后未复现：中位数分别为 57.2µs → 55.0µs、2.50ms → 2.33ms。这些 CPU 帧数据不能证明端到端帧率、物理内存、GPU 内存或启动延迟。兼容现有指针 API 仍需分配新的校验句柄，文字塑形、组件创建和用户回调也仍有分配，并未达到零分配。

第一轮截图发现局部 `Themed` 子树的默认文字色回退：内联值提前捕获了全局颜色。修复后，默认色在绘制时解析，显式透明色仍保留，并补了 root 级回归。此后 792 张矩阵截图逐字节一致，99 个场景导航和交互组全部通过；原生窗口生命周期、空闲恢复和连续 60 次缩放检查也通过。新增测试还覆盖零值样式继承、回调副本身份、旧指针句柄、哈希冲突、文字引用释放、编辑器内容及对象模式变化。arena 不新增强制 GC，也不调整空闲回收阈值。

### 节点 arena 测量，2026-10-15

Apple M4、darwin/arm64、Go 1.27.1 上，99 个组件场景的稳态帧分配字节数和次数全部下降。每个场景独立进程运行三次，每次 200ms；基线已经包含语义 op 直接写入、焦点 filter、单元格 ID 和可见列缓存。arena 带来的逐场景降幅中位数为：字节数 42.4%，次数 37.7%；不加权的场景平均分配从 106,662 降到 57,243 字节/帧。

| 场景 | 分配字节/帧，前 → 后 | 分配次数/帧，前 → 后 |
| --- | ---: | ---: |
| Button | 181,059 → 106,583 | 699 → 463 |
| Calendar | 918,307 → 413,218 | 3,861 → 2,344 |
| Table | 558,217 → 149,992 | 2,056 → 817 |
| List | 162,097 → 58,464 | 679 → 374 |
| Input | 218,070 → 178,293 | 935 → 806 |

30 万行表格滚动基准，前后各运行三次、每次 300ms：分配从 1,121,028 降到 368,397 字节/帧，次数从 3,930 降到 1,622，布局耗时中位数为 590µs → 525µs。Radio 和 Button 在短基准中的耗时回退，延长到前后各五次、每次 1s 后未复现：中位数分别为 30.7µs → 29.8µs、152.6µs → 147.6µs。小幅耗时变化仍受主机负载影响。

792 张 GPU 矩阵截图逐字节一致，组件库 smoke 通过全部 99 个场景导航及输入、浮层、列表、表格、日历交互。arena 回归覆盖槽位复用前后的旧句柄、同帧重建、root 归属、缓存逃逸、废弃节点引用清理、一次性 root 和嵌套 root。预热后只构建 101 个节点，最多分配六批小句柄；完整组件帧仍有样式、回调和文字处理分配。这些 CPU 帧数据不代表原生窗口物理内存、GPU 内存或启动延迟。

原生窗口检查（`RealWindows`、`TestMetalIdleWindow`、`TestMetalResizeWindow`）也通过：三次空闲恢复耗时 28.8～31.8ms，60 次原生缩放最慢 27.4ms。这些检查不测量端到端帧率。

### 更早的分配测量，2026-10-13

2026-10-13，在 darwin/arm64、Apple M5 Pro 上，99 个场景的每帧分配字节数和分配次数均下降。逐场景降幅的中位数分别为 8.7% 和 16.2%（每个场景独立进程运行三次，每次 200ms）。短测试中出现耗时回退的五个场景，延长到五次、每次 1s 后未复现稳定回退。小幅耗时差异仍受主机负载影响；这些数据是同机观察值，不是跨机器性能门槛。

| 场景 | 分配字节/帧，前 → 后 | 分配次数/帧，前 → 后 |
| --- | ---: | ---: |
| Calendar | 1,139,345 → 959,473 | 7,085 → 5,184 |
| Form | 92,141 → 82,192 | 467 → 348 |
| Chart | 292,385 → 267,760 | 1,445 → 1,229 |
| Tree | 118,003 → 104,320 | 644 → 520 |
| VirtualList | 142,285 → 111,931 | 645 → 452 |
| MessageScroller | 1,057,741 → 682,160 | 2,385 → 1,590 |

大数据场景单独运行三次、每次 1s。表格前后使用相同的 CPU/内存 profile 参数；普通编辑器前后均不启用 profile：

| 场景 | 耗时/帧，前 → 后 | 分配字节/帧，前 → 后 | 分配次数/帧，前 → 后 |
| --- | ---: | ---: | ---: |
| 30 万行表格滚动 | 702µs → 608µs | 1,515,575 → 1,138,768 | 6,524 → 4,522 |
| 20 万行普通编辑器移动光标 | 2,594µs → 2,319µs | 540,302 → 250,920 | 1,745 → 1,077 |

公共路径的改动包括：仅缓存节点分配布局 memo，仅输入框创建 Gio 编辑器，跳过未使用的快捷键和浮层索引，共享不可变快捷键祖先路径，小容器的布局子列表使用栈空间，构建子列表时一次预留容量，状态 ID 哈希不再拼接字符串。状态哈希保持原有身份兼容。组件自动 ID 和日历日期 ID 直接写入字节缓冲，保持原有格式；微基准中前者从 43ns 降到 19ns，后者从 94ns、3 次分配降到 43ns、1 次分配。编辑器复用字形临时缓冲，最多保留 4,096 个字形；超大行的缓冲不保留。

回归覆盖缓存节点的宽度、DPI、继承字体变化，兄弟动作作用域，无处理器的快捷键提示，稳定的编辑器焦点地址和中文输入。792 张 GPU 矩阵截图与基线逐字节一致；前后均通过测试专用 overlay 固定截图时间。实际运行 `go run ./examples/components`，还验证了全部 99 个场景的导航、输入、复选框和开关状态、弹窗焦点/保存/重开、万条选项搜索、十万项虚拟列表滚动和表格排序，并检查日历跳过禁用日期的键盘导航和日期选择器重开后的选中状态。示例分类检查防止注册组件遗漏在侧栏之外。
