# AppShell

[English](app_shell.md) | 简体中文

桌面窗口布局：左侧为可折叠侧栏，顶部为 56dp 标题栏，右侧为主内容，底部可固定输入区。将 AppShell 保存在应用状态中，不要每帧重建。

```go
shell := kit.AppShell("工作区").
    Page(kit.AppPage{
        ID: "projects", Title: "项目", Icon: kit.IconHome,
        Create: func() el.View { return kit.Input("项目名称") },
    }).
    Page(kit.AppPage{
        ID: "settings", Title: "设置", Icon: kit.IconSettings,
        Create: func() el.View { return kit.Switch("通知", true) },
    })
window.Open(window.Options{Title: "工作区", Width: 1000, Height: 680, Frameless: true, NativeTrafficLights: true,
    TrafficLightLayout: &window.TrafficLightLayout{Height: kit.AppShellTitleBarHeight, Left: 16, Spacing: 20},
    Content: el.Root(shell)})
```

`Create` 在首次访问时调用一次。每页保存独立的元素根节点，保留控件和滚动状态。未显示的页面不渲染、不运行动画，也不接收输入或出现在自动化快照中。页面缓存随应用壳保留，适合有限的导航目的地，不适合为每份文档创建一个页面。

默认选中第一个可用页面。空 ID 和重复 ID 被忽略。`SetValue(id)` 切换到已有的可用页面，不触发回调；`OnChange` 在用户导航时触发。`Value` 返回当前页面 ID。

`AppPage.Actions` 设置本页操作区，`Actions(view)` 设置常驻操作区，`Footer(view)` 设置侧栏底部内容。`SidebarWidth(dp)` 设置展开宽度；`SetCollapsed` 和 `Collapsed` 控制折叠状态。窗口窄于 640dp 时自动收成图标栏，变宽后恢复用户原来的折叠选择。`CollapseBelow(dp)` 修改阈值，零关闭自动折叠。没有指定图标的页面使用文件图标。

页面默认带 24dp 内边距和纵向滚动。编辑器、分栏等需要自行管理尺寸和滚动的内容设置 `AppPage.Fill`。空工厂显示空白页面。

切换平级页面时，旧页立即退出，新页在 150ms 内淡入，避免两页文字叠加。`Transition(duration)` 修改时长，零禁用动画；减少动态效果设置会跳过动画。首次显示不淡入，也没有后台计时器。这是平级应用布局，尚不包含历史路由和共享元素过渡。

运行 `go run ./examples/app-shell` 查看完整应用和输入保留。运行 `go run ./examples/components -section motion_pages` 查看切页动效，使用 `-section motion_presets` 查看弹簧对照。运行 `go run ./examples/components -section app_shell` 查看组件示例。

## 自定义区域

| 方法 | 作用 |
| --- | --- |
| `SidebarHeader(view)` | 侧栏上方工具栏，如新建、搜索按钮 |
| `Sidebar(view)` | 替换默认页面导航；自行管理滚动，读取 `Collapsed()` 适配图标栏 |
| `Header(view)` | 替换默认标题与说明，可放模型选择器等交互控件 |
| `Actions(view)` | 标题栏右侧常驻按钮；`AppPage.Actions` 放本页按钮 |
| `Content(view)` | 使用完全自定义的主内容，自己管理内边距和滚动；nil 恢复页面模式 |
| `ContentFooter(view)` | 固定在主内容底部，如聊天输入区 |
| `Footer(view)` | 侧栏底部，如用户资料入口 |

这些区域都是 `el.View`，无需修改 AppShell 的渲染代码。`Content` 在配置时建立独立根节点，应保留视图状态并只在需要替换内容时调用。侧栏宽度只影响导航和正文；标题栏可独立保留窗口按钮所需空间。

## 桌面窗口行为

设置 `window.Options{Frameless: true}` 使用完整应用标题栏。Windows 与 Linux 使用相同的窗口按钮尺寸、顺序与右上角位置；macOS 的红黄绿固定在窗口左上角，折叠侧栏不会移动它们，新建等自定义按钮始终排在窗口按钮之后。设置 `NativeTrafficLights: true`，并将 `TrafficLightLayout.Height` 设为 `kit.AppShellTitleBarHeight`，即可使用 AppKit 的原生关闭、最小化和全屏按钮。AppShell 自动预留位置，不再重复绘制；外观、悬停、操作和辅助功能由系统管理。不启用此选项时仍保留旧的绘制按钮。侧栏折叠按钮使用侧栏图标，是 Keel 控件，不是 AppKit 按钮。

标题和空白区域可拖动窗口，双击遵循平台标题栏行为。自定义标题区、操作按钮和窗口按钮不在拖动区域内，正常接收点击与输入。使用系统标题栏时，应用标题栏保留业务内容，但不重复显示窗口按钮。

含宽选择器或多组按钮的应用应设置合适的 `MinWidth`，并在自定义视图里适配紧凑布局。示例最小尺寸为 560×420dp。视觉布局与按钮回调有跨平台分支测试；跨平台编译不能代替 Windows、Linux 真机验证。

侧栏与正文之间有竖向分隔线，可左右拖动调宽。`SidebarLimits(min, max)` 设置宽度范围，默认 200–420dp；`SidebarSize()` 返回当前展开宽度，自定义导航应使用这个值。聚焦分隔线后可用方向键、Home、End 调整。折叠按钮在窗口控件旁，折叠再展开会保留拖动后的宽度。

`HideSidebarOnCollapse(true)` 可将折叠行为改为完全隐藏侧栏：导航、底部资料区和竖分隔线一起移除，正文占满宽度；顶部展开按钮保留，重新展开后恢复原宽度。默认 `false` 保留图标栏。此配置也适用于窄窗口自动折叠。标题栏底部有 1dp 横分隔线：展开时从侧栏右边开始，折叠后贯穿整个窗口。`examples/app-shell` 默认启用完全隐藏；右上角的“折叠方式”选择器可随时切换“保留图标栏”和“完全隐藏”，折叠状态下也能直接对比。

标题栏高度为最小值：大字体可撑高控件及其容器。Workspace 示例还会根据
字体缩放调整下拉框宽度，并启用 AutoSize 和普通窗口尺寸记忆（见 ui/window README）。
