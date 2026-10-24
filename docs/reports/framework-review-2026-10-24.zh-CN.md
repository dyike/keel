# 框架评审 · 2026-10-24

[English](framework-review-2026-10-24.md) | 简体中文

从跨平台应用开发者的角度评审 macOS 性能、易用性和脚手架。测量环境为 Apple M4，外接 60 Hz 4K 显示器。标注“未验证”的条目是风险，仍需在对应平台确认。

## 测量方法

Go 的 `pprof` 在 macOS 上会把 cgo 和内核时间算错：它显示约 10% 的 CPU 在 Metal `newBuffer`、约 23% 在 `pthread_cond_signal`，而 Instruments Time Profiler 显示分别是 2–3% 和不到 1%。原生归因请用 Instruments（`xcrun xctrace record --template 'Time Profiler' --attach <pid>`），统计全部采样而不是截断的排行榜，并在相同帧率下对比。CPU 百分比与帧数不成线性：负载变轻后任务跑在能效核上，频率也更低。

## 基线

| 场景 | 空闲内存 | 启动峰值 | 空闲 CPU |
| --- | --- | --- | --- |
| `examples/hello` | 37 MB | 96 MB | 0% |
| `examples/chat` | 44 MB | 103 MB | 0% |
| `examples/components` Button 页（3 个转圈） | 149 MB | 149 MB | 单核 27–29% |

空闲应用不绘制、不占 CPU。开销出现在任何东西开始动的时候：每一帧都重跑整个窗口（Render、布局、绘制、输入和 GPU 收集、一次 Core Animation 提交）。在 Button 页上，Keel 自己的布局和绘制约占进程 CPU 的 45%，Gio 的帧处理约 20%，AppKit/Core Animation 约 25%。

## 已做的改动

| 改动 | 状态 | 对 Button 页的效果 |
| --- | --- | --- |
| 循环指示器共用 30 Hz 节奏（`cx.AnimatingAmbient`） | 已合并，#50 | 进程 CPU 219 → 114 ms/s；能耗 −17% |
| 一串字形合成一个操作（`paint.ImageOp.PaintQuads`） | 已合并，#51 | 进程 CPU −10%；Gio `gpu.collect` −33%；像素逐字节一致 |
| 不可变 Metal 缓冲池化 | 已准备，未提交 | 去掉 `newBuffer`/`dealloc`，约 2–3% CPU |

#51 之前，这个页面每帧约 5300 个操作里有约 4000 个是文字。一个把 Gio 事件循环每帧 4 次 channel 交接减为 3 次的改动，每帧只省约 0.3 µs，没有保留。

## 剩余性能工作

按收益与工作量之比排序。

1. **缓存未变化 label 的字形。** Keel 缓存了测量结果，但绘制时每帧仍调用 `Shaper.LayoutString`（对字符串做哈希），再用 `NextGlyph` 逐个重建字形。约 8% CPU。
2. **`ui/el` 每帧的固定记账。** 即使没按 Tab，Tab 导航每帧也要遍历两遍树（`focusedContainerTrap`、`collectTabTargets`）；每个节点的状态每帧要在 `store.states` 里查 5 次以上（assignKeys、prepareKeys、Tab 遍历、blockTree、paint）。约 8–10%。
3. **纯绘制帧。** 已在 `feat/desktop-next` 完成：只推进绘制时动画（`el.AmbientPhase`；Spinner 和不确定状态的 ProgressCircle）的帧，窗口根元素直接用上一帧的元素树和布局重绘。这类帧在 `el` 中耗时从 359 µs 降到 228 µs（−37%）；真实窗口中只有一个 Spinner 和 60 行内容在显示时，进程 CPU 约减少 5.5%（每 10 秒 1.48 s 对 1.57 s）。线性进度条、骨架屏、扫光和状态标记仍在 Render 中动画。
4. **保留子树或局部重绘。** 收益最大，滚动也能受益，属于架构级改动；约束见[渲染内存](../rendering-memory.zh-CN.md)。
5. **启动峰值。** `hello` 稳定在 37 MB，但首帧期间峰值 96 MB；活动监视器先看到的是峰值。
6. **120 Hz 屏。** ProMotion 上每帧节省都会翻倍；尚未在 120 Hz 屏上实测。

## 易用性与生产可用性

### 影响生产使用

1. **交叉编译。** macOS 和 Linux 的 UI 构建仍需要 cgo（Gio 的窗口和 GPU 绑定），只有 Windows 能用 `CGO_ENABLED=0` 构建。CI 需要每个桌面系统各一台机器。见[交付设计](../desktop-cgo-free.zh-CN.md)。
2. **无障碍。** 没有对接 NSAccessibility、UIAccessibility、UI Automation 或 AT-SPI，读屏软件无法读取 Keel 窗口。Agent 自动化使用的语义树（角色、名称、值、边界）正好可以作为这些桥接的数据来源。
3. **macOS 以外的文字**（未验证）。只有 macOS 通过平台渲染（CoreText）。Windows 和 Linux 用 go-text 光栅化，hinting、子像素渲染和字体回退（中文、emoji）可能与原生应用不同。
4. **移动端。** `native/*` 在 iOS 和 Android 上返回 `native.ErrUnsupported`；iOS 只在模拟器上验证过；示例没有移动端布局。对外公布的支持矩阵应写明这些。

### 开发体验

5. **两种子元素。** `el.Text(...)` 是 `el.Element`，kit 组件却要调用 `.Render(cx)`，因为 `el.Element` 带有未导出的方法。有状态的 kit 视图（`kit.Input`）必须存在字段里，无状态的（`kit.Button`）在行内创建；类型上看不出区别。
6. **线程。** 在 goroutine 里直接改视图字段是数据竞争，必须经过 `core.Update`；只有 `-race` 能发现。脚手架和指南需要一个“加载数据，然后更新视图”的示例，也许还需要一个辅助函数。
7. **性能分析。** 没有内置的帧统计层（帧耗时、每秒帧数、操作数、字形数），而 `pprof` 在 macOS 上会误导；见[测量方法](#测量方法)。
8. **重载。** `keel run` 保存后重新编译并重启，界面状态会丢失。
9. **体积与工具链。** `hello` 为 17 MB（strip 后 11.6 MB）。最低要求 Go 1.26.1。

### 脚手架与文档（随本报告修复）

10. 入门指南和 `cmd/keel` README 写着 `go install …@latest` 需要尚未发布的单模块版本；v0.3.8 已经是单模块版本。
11. `keel new` 生成的 `app.go` 只有中文文字。
12. 生成的应用提示保存后重新运行 `keel run`，而 `keel run` 默认会监听文件并自动重启。

### 应保留的优势

单一模块依赖、`keel doctor`、`keel run` 生成带图标的开发 app bundle、`uitest` 交互测试，以及让 agent 端到端驱动真实窗口的 `keel-mcp`。
