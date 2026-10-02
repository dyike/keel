# 架构

Keel 的代码分两组：`ui/*` 负责界面，`native/*` 负责 Gio 做不到的系统能力。两组内部都按职责拆成小模块，每个模块一个目录、一个 README。界面建立在一条约定上：所有窗口的渲染和所有回调都在同一把锁下串行执行。

## 模块

```
github.com/dyike/keel
├── ui/
│   ├── core/             地基：Widget 接口、回调、线程规则
│   ├── theme/            颜色、字号、字体
│   ├── locale/           框架自己显示的文字：确定、复制、关闭……
│   ├── el/               GPUI 风格：视图、链式样式元素、flexbox
│   ├── kit/              组件：Button、Input、Table、Dialog、Chart …（一个组件一个文件）
│   ├── window/           窗口：Open、Main、快捷键、截图
│   ├── markdown/         Markdown 渲染，针对 AI 流式输出
│   └── internal/         loop（帧锁）、editorstyle（输入绘制）、imageload（图片加载）、uitest（测试工具）
├── native/
│   ├── permission/       权限检查与申请
│   ├── screen/           显示器列表、截图
│   ├── input/            合成鼠标、键盘事件
│   ├── hotkey/           全局快捷键
│   ├── internal/sys/     cgo 绑定，所有 Objective-C 代码只在这里
│   └── native.go         共用的错误值
├── cmd/
│   └── keel-mcp/         MCP server：Agent 用它对应用做端到端测试
├── internal/deps/        模块边界检查
├── examples/
└── docs/
```

每个模块目录下都有 README，写明它做什么、依赖谁、被谁依赖。

### 依赖方向

```
ui:
  kit ───────► el ──┐
  kit ───────► base（组件行为，不依赖其他模块）
  markdown ──► el   ├──► theme、locale
  window ───────────┤
                    └──► core

  theme、locale ──► internal/loop（切换主题或语言时通知所有窗口重绘）
  el ──► internal/editorstyle（输入框绘制）
  markdown ──► internal/imageload（图片解码与占位）

native:
  permission ─┐
  screen     ─┼──► internal/sys ──► native（错误值）
  input      ─┤
  hotkey     ─┘
```

规则只有三条：

1. **依赖只往下走。** 下层不知道上层存在：`core` 只依赖内部帧锁，`theme` 仅引用内部帧循环以通知主题重绘；`el` 不知道有 `kit`；`window` 只认 `core.Widget` 接口，不知道具体有哪些组件。
2. **同层之间不互相引用。** `kit`、`markdown` 和 `window` 互不引用；四个 `native` 模块互不引用。
3. **`ui` 和 `native` 互不引用。** 不需要窗口的程序（后台截图、全局快捷键）只引用需要的 `native/*`，不会带进 Gio。

`markdown` 用 `el` 排版，图片走 `internal/imageload`（加载、解码限制、占位）。`el` 的输入框用 `internal/editorstyle` 绘制光标和选区；这个内部包只负责 Gio 输入绘制和字形测量，不依赖其他 Keel 模块。

`cmd/keel-mcp` 不引用任何 Keel 包，也不引用 Gio：它只通过 socket 上的 JSON 协议驱动 `ui/window` 的自动化模式，见 [Agent 端到端测试](automation.md#原理)。

这些规则由 `internal/deps` 的测试强制执行：它把每个模块允许依赖的包写成一张表，越界或者新增目录没登记，`go test ./...` 就失败。改架构时先改那张表，再改代码。

### 每层放什么

| 模块 | 放什么 | 不放什么 |
| --- | --- | --- |
| `ui/core` | 所有界面模块都要遵守的接口和函数 | 任何具体组件、颜色 |
| `ui/theme` | 视觉参数、全局调色板切换、局部主题作用域与重绘通知 | 组件 |
| `ui/locale` | 框架自己显示或报告给 Agent 的文字，运行时切换语言 | 应用自己的文案、翻译系统 |
| `ui/base` | 组件的行为：键盘导航、首字母跳转、多选、打开状态 | 任何绘制、颜色、Gio 以外的 Keel 依赖 |
| `ui/kit` | 基于 el 和 base 的组件 | Gio 输入和浮层基础设施、窗口管理 |
| `ui/window` | 与窗口绑定的东西：生命周期、快捷键、根视图、截图 | 具体组件 |
| `ui/el` | 元素、样式、布局引擎、元素状态、视图 | 业务组件（它们在应用里写成函数或视图） |
| `ui/internal/loop` | 跨窗口共享的可变状态：帧锁、更新队列 | 任何 Gio 类型 |
| `ui/internal/editorstyle` | 输入框的光标、选区绘制和字形测量 | 具体 Keel 组件、窗口和主题 |
| `ui/internal/imageload` | 图片来源解析、异步加载、尺寸限制、加载中和失败占位 | 组件外观、点击等交互 |

### 什么时候新建模块

先看它是不是现有模块的职责。是，就在那个模块里加文件：新下拉框是 kit 组件，就是 `ui/kit/select.go`；新的布局能力（比如换行）属于 el，就是 `ui/el/layout.go` 里的改动。

只有现有模块都装不下、而且它有自己清楚的职责时，才新建目录。例如剪贴板：不属于四个现有能力中的任何一个，也有只用它的场景，所以是新模块 `native/clipboard`。

新建模块时，写 README，在上面的模块图和 `internal/deps` 的表里登记。

## 一帧里发生了什么

Gio 是即时模式框架：每一帧都从头调用一遍所有组件的 `Layout`，组件在 `Layout` 里处理自上一帧以来的输入事件，同时输出绘制指令。Keel 的组件对象只是保存状态（文本、勾选、输入框内容），不保存绘制结果。

每个窗口有自己的 goroutine，收到 `FrameEvent` 后执行：

```
lockFrame()
drainUpdates()               // 执行 core.Update 排队的函数
window.handleShortcuts(gtx)  // 窗口快捷键
root.Layout(gtx, content)    // 背景 + 滚动 + 24dp 内边距 + 组件树
    └─ 各组件 Layout：处理事件 → 必要时 core.Call(回调) → 画自己
unlockFrame()
e.Frame(ops)                 // 提交给 GPU，在锁外执行
```

回调发生在布局中途，排在按钮前面的组件这一帧已经画完了，看到的是旧值。所以 `core.Call` 执行回调后会请求所有窗口再画一帧。实际效果是回调的修改晚一帧上屏，60Hz 屏幕上约 16ms，人眼察觉不到。

## 线程规则

**一句话：回调里直接改组件；其他 goroutine 改组件要包进 `core.Update`。**

`ui/internal/loop` 里有一把全局帧锁。每个窗口画一帧时持有它，组件回调都发生在画帧过程中，所以回调执行时一定持有这把锁。组件本身不加锁，它们的并发安全全靠这把锁。

| 代码在哪里运行 | 能否直接改组件 | 做法 |
| --- | --- | --- |
| 按钮、输入框、复选框的回调 | 能 | 直接改字段、调 `SetValue` 等 |
| `window.Options.Shortcuts` 的回调 | 能 | 直接改 |
| `window.Options.OnClose` | 能 | 直接改 |
| 你启动的 goroutine、`time.AfterFunc` | 不能 | `core.Update(func() { ... })` |
| `hotkey.Register` 的回调 | 不能 | `core.Update(func() { ... })` |
| `main` 里 `window.Main()` 之前 | 能 | 窗口还没开始画，没有竞争 |

`core.Update(fn)` 把 `fn` 放进队列并唤醒所有窗口；下一个开始画帧的窗口先执行队列，再布局。它在任何地方调用都安全，包括回调内部。

后台任务的写法：

```go
kit.Button("刷新", func() {
    v.status = "加载中…"
    go func() {
        data, err := fetch()           // 耗时操作在锁外
        core.Update(func() {             // 结果回到界面
            if err != nil {
                v.status = err.Error()
                return
            }
            v.status = data
        })
    }()
})
```

这套模型的代价：

- **所有窗口串行渲染。** 一个回调卡 2 秒，所有窗口都卡 2 秒。回调里只做改状态这类微秒级的事，I/O 和计算放 goroutine。
- **`core.Update` 是异步的。** 调用返回时 `fn` 还没执行。在回调里等 `core.Update` 的结果（比如用 channel 等它执行完）会死锁：回调持有锁，`fn` 要等锁。
- **没有窗口在画帧，队列就不会被执行。** 程序开窗前 `core.Update` 的函数会在第一个窗口的第一帧执行。

为什么选一把大锁而不是给每个组件加锁，见[设计决策](decisions.md#一把全局帧锁)。

## 不能在锁内等待主线程

这是给维护 `ui/window` 和写组件的人的规则：**持有帧锁时，不能调用任何会等主线程执行完才返回的 Gio 窗口方法**。在 Gio 里就是 `Window.Perform`、`Window.Option`、`Window.Run`（窗口创建之后调用时）。

原因是一个三方循环等待。以"设置窗口已打开，回到主窗口按 ⌘+,"为例：

```
主窗口 goroutine   持有帧锁，在快捷键回调里调 settings.Raise()
                   └─ Gio Perform 等主线程执行它
主线程             正在给设置窗口派发事件（焦点变了），等设置窗口画完这一帧
设置窗口 goroutine  要画帧，等帧锁  ← 被主窗口 goroutine 持有
```

三方互相等待，界面卡死。这个 bug 真实出现过，`Raise`、`Close` 现在都用 `go w.win.Perform(...)` 放到锁外执行。`win.Invalidate()` 不等主线程，可以在锁内调用。

新增窗口方法（设置位置、置顶、改标题）时照同样的办法处理，并在 `ui/window/testdata/raise` 这类真实窗口测试里覆盖，见[测试](testing.md#真实窗口测试)。

## 窗口生命周期

- `window.Open` 立即返回；原生窗口在它自己的 goroutine 里异步创建。
- `Close`、`Raise` 会等窗口画出第一帧后才真正执行。Gio v0.10.3 在 macOS 上有个 bug：原生窗口还没建好就被关闭，进程会崩溃。人手点不了这么快，Agent 可以，所以 Keel 在这里等一下。回归测试是 `ui/window/testdata/reopen`。
- 关闭窗口（用户点关闭，或调用 `w.Close()`）后，`OnClose` 在锁内执行，`w.Closed()` 变成 `true`。关闭的窗口不能重新打开，需要时重新 `window.Open`。
- 最后一个窗口关闭后，进程调用 `os.Exit(0)` 退出。`main` 里 `window.Main()` 之后的代码不会执行，要做清理放进 `OnClose`。
