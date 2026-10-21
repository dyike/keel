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

### 覆盖表解码与图标回调复用（本地后续优化）

磁盘索引在解码时就复用相同字符覆盖表，不再先分配重复切片。去重表只属于一次读取，命中哈希后完整比较页号和位图；解码结果不引用复用的记录缓冲。命名实例元数据仍独立，版本 7 缓存继续有效。
IconView 只缓存一个不可变的渲染器与颜色快照，更换缓存不会修改旧元素捕获的颜色；复制视图会建立自己的渲染器。元素句柄和 Context.Cache 生命周期不变。参照的是本地 mygo 保留图标栅格回调的做法，不借用帧 scratch，也不改 Keel 的指针 API。

```sh
go test ./typesetting/fontscan -run 'Test(IndexDecode|RuneSetDecode)'
go test ./ui/kit -run 'TestIcon(RenderReuses|CachedElement)'
go test ./typesetting/fontscan -run '^$' -bench BenchmarkFontIndexDeserialize -benchmem -count=3
go test ./ui/kit -run '^$' -bench 'Benchmark(IconRender|RetainedIconFrame)' -benchmem -count=3
```

Apple M4、Go 1.27.1，以 main `afe213c` 为基线，各跑三次 500ms，取中位数：64 文件、每文件四字体、512 页共享覆盖表的合成索引，分配从 4,959,613 B / 888 次降到 259,418 B / 633 次，CPU 时间 2.01ms → 1.48ms。
所有覆盖表都不同时，去重表带来代价：4,959,578 → 4,996,580 B、886 → 899 次，2.05ms → 2.29ms；不能只报告共享场景的收益。
稳定保留的 IconView.Render 从 944 B / 3 次降到 920 B / 2 次；每次新建视图从 992 B / 4 次降到 984 B / 4 次。这些数字包含独立元素、节点，不只是回调，完整帧仍非零分配。
792 张固定时间 GPU 截图全部逐字节一致。短基准中三个场景的字节回退，在交错的五轮独立进程复测中未持续复现。微基准不能证明真实窗口 footprint、峰值或 FPS 改善。

相同卡片客户端以 main `afe213c`、尚未发布的本地后续优化、最新本地 mygo `fefd36a5` 各跑三轮独立进程，交错顺序。窗口为 800×600dp、原生 2×（1600×1200px），保留已有 v7 字体索引。滚动、缩放各 10 秒，空闲 12 秒；每 100ms 采样，启动为 50ms，空闲取最后三秒。仅测试客户端关闭空闲主动回收，与 mygo 对齐；GOGC/GOMEMLIMIT 和库默认值不变。强制 GC 全部为零，profile、截图另开进程。

| 原生窗口三轮中位数 | 改前 | 本地后续优化 | Mygo |
| --- | ---: | ---: | ---: |
| 启动 TotalAlloc / MiB | 26.74 | 21.62 | 20.03 |
| 启动 HeapSys / MiB | 27.22 | 19.25 | 23.22 |
| 滚动分配 / MiB/s | 4.88 | 4.07 | 45.33 |
| 滚动分配次数 / 绘制调用 | 693 | 93 | 2,568 |
| 滚动 GC 次数 | 7 | 6 | 38 |
| 缩放分配 / MiB/s | 2.33 | 1.93 | 15.30 |
| 缩放 GC 次数 | 3 | 3 | 12 |
| 空闲末三秒 footprint / MiB | 58.80 | 57.24 | 73.30 |
| 内核全程峰值 / MiB | 159.10 | 164.06 | 163.67 |

峰值未改善：改前 157.86–162.45 MiB、改后 158.10–165.38 MiB，全部进程在缩放时达到峰值，中位数观察值升高约 5 MiB。三轮范围重叠，空闲小幅差异也不是已确认收益。单轮启动 GC 可能在分配减少后反而增加。Mygo 的无障碍工作显著影响分配速度；这不是框架排名或端到端 FPS，也不能证明 controls 分类峰值改善。独立 profile 中 IconView.Render 滚动自分配为零采样，指针 frameHandles 仍占约 57% 采样字节；启动 RuneSet 解码采样约 1 MiB，而非原来的约 6 MiB。采样和 GC 统计滞后使 profile 总量不能替代阶段实际分配或物理存活内存。

### 绘制分配后续优化（2026-10-15）

本地后续改动保留颜色指针 API 和元素生命周期。Material 标签改用新增的颜色值布局方法，可选中文本自己保存颜色。anchors 在绘制开始时 clear，Render 仍能通过 LastSize 读取上一帧；超过 4,096 项的 map 在下一次绘制时丢弃。描边只有连续两帧出现才准入缓存，候选表只存哈希；正式条目不借用可变 Ops 或描边 scratch。参考本地 mygo 的有界 scratch 和清理边界，不改其对应的公开句柄 API。本次不改 drawable、分辨率、GC 或默认空闲回收。

99 个场景分别启动进程，每个场景测三次、每次 50 个预热后帧。合计从 1,402,961 降到 1,086,720 B/帧，分配次数从 11,144 降到 9,876，没有场景退步。它只测布局和操作生成，不含 GPU 执行。普通 Material 标签从 128 B、1 次降到零；可选中标签从 133 B、2 次降到 5 B、1 次。512 个 anchors 从 111,288 B、17 次降到零。792 张定时 GPU 截图逐字节一致。

真实 controls 对 main `3b88f26`、字体+描边（`af8d7dd` + `b45e325`，现已合入 main `6ab9dae`）、本地后续优化各跑五轮，交错顺序，每轮新进程。窗口 800×600 像素、原生 1×；滚动、缩放各 10 秒，随后停止输入 12 秒。controls 最后一阶段仍有动画，不能称为静态空闲。

| controls，五轮中位数 | 原 main | 字体+描边 | 本地后续优化 |
| --- | ---: | ---: | ---: |
| 启动 TotalAlloc / MiB | 245.54 | 131.75 | 123.01 |
| 滚动分配 / MiB/s | 22.88 | 14.40 | 10.51 |
| 滚动 GC 次数 | 9 | 6 | 4 |
| 缩放分配 / MiB/s | 86.80 | 39.53 | 28.41 |
| 缩放 GC 次数 | 32 | 14 | 10 |
| 无输入最后 3 秒 footprint / MiB | 162.38 | 162.33 | 162.47 |
| 内核进程峰值 / MiB | 187.39 | 175.72 | 188.36 |

峰值范围分别为 180.60–193.42、170.08–187.38、175.70–209.03 MiB。后续优化降低了垃圾，未降低物理峰值：中位数反而高于单独的字体+描边版本。本轮后续优化五个进程均在首次滚动时定下峰值，不在缩放时；原 main 五个进程均在启动时定峰。客户端和自然 GC 时机不同，不能沿用上轮“缩放定峰”解释所有场景。

现有分类快照中，三个 controls 版本均约为 51 MiB unmapped graphics、8.3 MiB IOAccelerator graphics、5.7 MiB IOSurface。启动末 Untagged 中位数约为 70→87→98 MiB，无输入时回到约 73–74 MiB。Untagged 不等于 Go 堆；但本地后续优化第 1 轮的 gctrace 显示，3.501 秒时堆涨到 101 MB 才自然 GC（goal 107 MB），回收后存活量 29 MB。这不支持把峰值变化全部归给 GPU，也不能据此认定泄漏。分类工具有取整，快照不一定刚好捕捉到短暂峰值。

footprint 每 100ms 采样，启动每 50ms，单列内核进程峰值。所有正式阶段 forced GC 为零；仅测试客户端关闭 Keel 空闲主动回收以匹配 mygo，GOGC/GOMEMLIMIT 未设置。正式轮次不同时构建、跑测试、剖析或额外采分类；使用已有字体索引。字体、描边现已合入主线，不归功于某个已发布版本；本轮后续改动仍只在本地。

相同 600 张卡片的客户端对各版本交错跑三轮，包含本地 mygo `7e868828`，forced GC 同样全部为零：

| 卡片，三轮中位数 | 字体+描边 | 本地后续优化 | 本地 mygo |
| --- | ---: | ---: | ---: |
| 启动 TotalAlloc / MiB | 26.65 | 26.58 | 19.20 |
| 启动末 footprint / MiB | 47.13 | 47.66 | 54.52 |
| 滚动 footprint / MiB | 105.17 | 103.31 | 124.91 |
| 缩放 footprint / MiB | 107.47 | 107.63 | 126.25 |
| 空闲最后 3 秒 footprint / MiB | 50.10 | 50.19 | 63.88 |
| 内核进程峰值 / MiB | 115.97 | 115.83 | 131.33 |
| 滚动分配 / MiB/s | 4.91 | 4.88 | 45.21 |
| 缩放分配 / MiB/s | 2.29 | 2.42 | 15.42 |
| 滚动 / 缩放 GC 次数 | 7 / 3 | 7 / 3 | 38 / 12 |

卡片客户端各进程均在缩放时定峰。两版 Keel 的峰值范围重叠：114.42–116.08 与 115.56–116.74 MiB。卡片很少涉及本轮 controls 热点，缩放分配速率没有改善；相同每秒 20 次尺寸请求下，实际帧事件数仍有波动，不能当作端到端 FPS。本机 mygo 的 macOS accessibility 工作显著影响分配量，不据此给框架做普遍排名。

独立进程按 64 KiB 采样的 allocs profile 中，LabelStyle 自身分配和 strokeCache.put 已不再是热点。controls 缩放样本约 36% 来自文字 shaping 与 toGioGlyphs；卡片滚动约 57% 是 frameHandles。controls 滚动中的 writeVertCache 占 15.18 MiB，但只有四次采样到的扩容，不是每帧泄漏。profile 记账可能滞后自然 GC 约两个周期，阶段总量以 MemStats 差值为准，占比只作近似。

另跑的一对帧统计诊断不混入正式内存数据：首帧布局+提交 124.67→141.49ms，去掉前 20 帧后 P95 为 8.56→7.57ms。首帧停顿仍在，单对进程不能证明端到端 FPS 或稳定的首帧变化。两帧准入可能让新稳定的几何计算两次，哈希碰撞也可能阻止准入；这是命中率和 CPU 的代价，不是借用不安全内存的捷径。

### 字体启动分配检查（2026-10-15）

字体初始化和索引基准单独运行：

```sh
go test ./typesetting/font -run '^$' -bench BenchmarkGlyfInitialization -benchmem -benchtime=500ms -count=3
go test ./typesetting/fontscan -run '^$' -bench 'BenchmarkFontIndex(RecordRead|Serialize)' -benchmem -benchtime=200ms -count=3
```

Apple M4、darwin/arm64、Go 1.27.1，三个独立样本的中位数：

| 操作 | 改前字节 | 改后字节 | 改前分配次数 | 改后分配次数 |
| --- | ---: | ---: | ---: | ---: |
| Go Regular glyf 初始化 | 191,033 | 30,720 | 2,125 | 2 |
| 读取 3,143,610 字节索引记录 | 8,388,283 | 3,178,546 | 17 | 3 |
| 编码 64 个合成文件，各含四组 512 页字符覆盖 | 28,489,518 | 6,009,804 | 293 | 278 |

字体文件未变化时跳过索引序列化；冷启动或更新时保留版本 7 格式，逐文件流式编码。测试对照旧解析器的字形头、可变字体点数、轮廓及度量；畸形记录模糊测试检查校验结果一致。792 张固定时钟组件截图与 main 逐字节一致。

解锁后的原生对比使用相同的 600 张非虚拟化卡片，800×600 实际像素、原生 1×，各版本交错启动三个新进程。基线为 main `3b88f26`；本地优化构建通过 Go overlay 合用字体补丁 `af8d7dd` 和描边补丁 `b45e325`。测量时尚未发布；之后已分别通过 PR #35、#36 合入主线，不归功于某个已发布版本。字体索引已存在。本地 mygo 为 `7e868828`。

| 原生指标，三轮中位数 | Keel 改前 | 本地优化构建 | 本地 mygo |
| --- | ---: | ---: | ---: |
| 启动 TotalAlloc / MiB | 95.38 | 26.65 | 19.17 |
| 启动 GC 次数 | 10 | 5 | 4 |
| 启动 HeapSys / MiB | 83.22 | 27.22 | 23.22 |
| 全程内核进程峰值 / MiB | 115.36 | 115.55 | 132.41 |

启动分配下降 72.1%，但物理峰值没有下降证据。Keel 三轮内核峰值范围重叠：改前 114.75–124.24，改后 112.70–115.58 MiB。启动阶段内核峰值中位数为 106.86 → 106.00 MiB，后续缩放仍会推高全程峰值。HeapSys 包含已交还系统的页，不是物理 footprint。卡片滚动仍约 4.91 MiB/s，GC 为 6 → 7；启动分配减少不等于每个阶段 GC 都减少。

每个进程依次测启动/加载、10 秒滚动、10 秒连续缩放和 12 秒空闲。物理 footprint 每 100ms 采样（启动每 50ms），与内核累计峰值分开记录；空闲值取最后三秒。全部阶段 forced GC 为零。只在测试客户端关闭 Keel 空闲主动回收以匹配 mygo，库默认值及 GOGC/GOMEMLIMIT 不改。锁屏试跑、profile 和截图进程均不计入这些中位数。原生滚动、缩放流程完成，没有无响应超时。

另起开启帧统计的单轮诊断，首帧布局加提交耗时为 104.00 → 60.89ms，优化后仍有首帧停顿。这不是首次实际呈现延迟、三轮耗时中位数或端到端 FPS。

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

macOS 上，`TestRealWindowsResize` 分别打开普通窗口和磨砂玻璃窗口，各执行
60 次 Gio 尺寸请求，再保持窗口中心不动，通过原生 frame 改变尺寸 60 次。
随后通过 `ToggleMaximize` 执行 4 次原生最大化/还原，等 AppKit 的缩放结束帧
完成后再检查尺寸固定的滚动。测试检查 NSView 采用 `duringViewResize` 重绘和
`topLeft` 内容放置策略，让 AppKit 在动画各步重绘，而非缩放缓存画面。
它分别记录内容图层的 `anchorPoint` 和 gravity，检查模型图层 X/Y 缩放值
始终为 1，并检查图层采用不拉伸的左上锚定（按 `contentsAreFlipped`
选择 gravity）、drawable 尺寸等于图层 bounds × backing scale，以及玻璃
窗口缩放帧在主线程启用 `presentsWithTransaction` 提交。每个窗口还通过程序
连续改变 60 次文字滚动位置。尺寸不变的帧保持异步提交，包括空闲 drawable
回收后的恢复；静态窗口必须停止绘帧。
探针在回收后采样两秒，报告进程空闲 CPU。
600ms 超时用于发现卡顿，不代表显示 FPS。这些属性和调度检查没有读取合成器
的中间帧或系统动画使用的整窗快照；还需拖动缩放、最大化和还原
`examples/hello` 和 `examples/glass`，确认文字没有
变形、锚点正确，并在 GPU 完成后检查空闲 CPU。修复位于 `ui/window`，没有
新增 cgo，也没有修改 Gio。

默认跳过，因为它会在屏幕上弹出窗口，而且需要图形界面环境。改动 `ui/window`、`ui/internal/loop` 里和窗口、锁相关的代码后必须跑一次。新增窗口方法时，把它加进 `testdata/raise` 的步骤里。

## 需要手动验证的部分

无界面测试覆盖不到这些，改动相关代码后手动跑一遍：

| 场景 | 怎么验证 |
| --- | --- |
| 真实窗口的打开、关闭、置前 | `go run ./examples/multiwindow -lang zh-CN`，点"打开设置窗口"两次，应该只有一个设置窗口 |
| 最后一个窗口关闭后退出 | 关掉所有窗口，终端里进程应结束 |
| 窗口快捷键 | 在主窗口按 ⌘+, |
| 全局快捷键 | `go run ./examples/hotkey -lang zh-CN`，切到别的应用按 ⌘⇧K，计数增加 |
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

微基准只测 CPU 顶点生成，不代表整帧或 FPS。解锁后的原生对比使用实际 controls 分类，800×600 像素、原生 1×，各版本交错启动三个新进程。main `3b88f26` 为基线；优化二进制通过 overlay 合用本地字体补丁 `af8d7dd` 与描边补丁 `b45e325`。测量时尚未发布，之后已通过 PR #35、#36 合入主线。

| 原生 controls 阶段 | 改前 MiB/s | 改后 MiB/s | GC 改前 → 改后 |
| --- | ---: | ---: | ---: |
| 10 秒滚动及组件动画 | 22.91 | 14.41 | 9 → 5 |
| 10 秒连续缩放 | 86.17 | 39.55 | 32 → 14 |
| 12 秒停止输入，动画继续 | 23.11 | 14.50 | 10 → 6 |

这不是静态空闲 CPU 检查，也不是与 mygo gallery 的同内容对比。原生滚动、缩放流程完成，没有无响应超时；该客户端没有帧计数器，不据此判断 FPS 或输入延迟。全部阶段 forced GC 为零；只在测试客户端关闭空闲主动回收，以匹配 mygo。着色器、drawable 数量、库的 GC 和回收默认值不变。

分配改善，物理内存没有改善证据：滚动 footprint 中位数为 157.45 → 159.86 MiB，缩放为 160.99 → 161.22 MiB。内核全程峰值中位数上升到 183.58 → 190.85 MiB；三轮范围重叠，为 176.45–194.60 → 173.08–200.31 MiB。报告保留这个观测代价。

另起不强制 GC 的 profile 显示，滚动剩余样本中 LabelStyle 约占 13.3%、anchors 插入约占 9.3%；缩放时文字 shaping 加 Gio 字形转换约占 28%，新缓存的独立所有权复制约占 15.8%。尺寸持续变化的路径仍需研究缓存准入或缓冲复用，不能改成不安全地借用 scratch。profile 账目可能滞后约两个 GC 周期，比例只用于定位，样本总量不是精确阶段字节量。

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

原生毛玻璃除了透明 GPU 像素，还需要检查系统合成结果。在 macOS 14.4+ 上运行 `KEEL_DESKTOP=1 go test -tags keelnativeqa -run TestNativeMaterialComposition ./ui/window`：用例把自己的红蓝参考背景和毛玻璃窗口一起捕获，再断言两侧颜色都能透出。捕获使用 ScreenCaptureKit 的当前进程内容接口，不申请录屏权限。过滤器保留系统桌面层，否则窗口背后的材质会被截成灰色；其他应用的内容窗口被排除，图片只保存测试窗口所在区域。
