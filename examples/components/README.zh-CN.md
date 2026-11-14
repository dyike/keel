# 组件示例

[English](README.md) | 简体中文

```sh
go run ./examples/components                    # 组件库应用：左侧导航和搜索，右侧是选中的组件
go run ./examples/components -section select    # 只显示一个组件，名字与 docs/kit/<名字>.md 一致
go run ./examples/components -section inputs    # 一类，例如 inputs、charts、navigation、motion
go run ./examples/components -theme dark        # 深色启动；应用右上角也能切换深浅色和中英文
```

默认使用英文；加上 `-lang zh-CN` 使用中文。浏览器示例通过 `?lang=en` 或 `?lang=zh-CN` 选择语言。`demoText(english, chinese)` 选择示例标签、消息和数据；框架控件使用 `ui/locale`。切换组件库语言会重建示例，保留当前组件和搜索词。

应用的侧栏按“通用、表单输入、数据展示、列表与表格、图表、消息与对话、反馈、导航、浮层、布局与外壳、动效、基础能力”分组，顶部搜索框过滤组件名。每个组件第一次打开时创建，切走再回来状态保留。整窗口的组件（Dock、Settings 等 el.Root）占满右侧内容区，自带的浮层也限制在内容区里。

组件文档页的“示例代码”直接读取注册该 section 的源码，显示在在线展示下方；修改示例后重新生成站点即可。复用到脚手架应用时保留组件构造和 Render 写法，注册函数与共享辅助函数属于组件库。

每个 kit 组件对应一个 `<名字>.go`，用 `registerSection` 注册同名 section，展示常用状态、边界和浅深色；注册后自动出现在应用侧栏里。`ui/kit/conventions_test.go` 检查每个组件都有示例。

生成截图：

```sh
go run ./examples/components -section chart -screenshot /tmp/keel-chart.png
```

另有几个验证 el 基础能力的 section：`theme`（运行时切换浅深色）、`focus`（Tab / Shift+Tab、子树禁用）、`time`（定时关闭与取消）、`overlay`（非模态点击穿透、模态遮罩、Esc 关闭与焦点恢复）。

组件的交互测试在 `ui/kit/*_test.go`，Agent 快照测试在 `ui/window/kit_*_test.go`。

`-section scrollable` 验证 el 横向滚动、宽内容裁剪和程序定位。

`-section variable_list` 验证 10 万行自然高度列表：定位、插入历史、展开内容与窗口宽度变化。

`-section layout` 验证换行布局和简单网格，缩窄窗口可检查换行、行列间距和最小尺寸。

运行 `go run ./examples/components -section motion -lang zh-CN`，查看姿态反向、移动中的点击区域与浮层、列表进退场、先退后进、布局缩放，以及组件共用的过渡。

截图和矩阵命令开启减少动态效果，避免首帧截图恰好捕获透明的进入状态；交互运行保留动画。

Motion 页面还演示文字变形、数字滚动、图标与图片替换，以及图片在缩略图和详情父容器之间的共享转场。

Motion 还演示渐变、阴影过渡，以及在容器上用 MatchText 自动匹配文字，无需逐个标签声明共享身份。


分类标题使用更明显的正文颜色和半粗体。拖动侧栏分隔线可调整宽度（180–420dp），也可聚焦分隔线后用方向键、Home 和 End 调整。本次会话中，切换主题和语言会保留调整后的宽度。

原 examples/motion 中的演示已统一合并到这里，可从“动效”分类选择，或指定分区：

- motion_values：弹簧、缓动和关键帧。
- motion_gestures：悬停、按压、拖拽和惯性。
- motion_layout：进退场和重排。
- motion_scroll：滚动进度联动。
- motion_presets：Smooth / Snappy / Bouncy 对比。
- motion_pages：切页淡入与输入保留。
- motion_pose：移动点击区域及锚定浮层。
- motion_timeline：重播、暂停、倒放、跳转及滚动控制。
- motion_effects：指针光晕、波纹和订单彩纸。
- rising_glow、shimmer_text：装饰粒子及文字动效。

例如：go run ./examples/components -section motion_effects -lang zh-CN。

原 frameless 示例已合并到 TitleBar。在“布局与外壳 → TitleBar”点击 **打开无边框窗口** 可打开原生预览；重复点击会唤起已打开的窗口，关闭预览不会关闭组件库。标题栏内保留搜索、命令和侧栏开关。

原 hotkey 示例已合并到“基础能力 → 全局快捷键与显示器”（section hotkey）。打开预览后显式启用/停用全局快捷键，按需刷新显示器信息。macOS 使用 Cmd+Shift+K，其他平台使用 Ctrl+Shift+K。关闭预览会注销快捷键；平台错误会显示出来，不启动后台时钟或轮询。原生调用在帧锁外执行，回调通过 core.Update 更新界面。

“布局与外壳 → 多窗口”（section multiwindow）展示设置子窗口：重复打开复用现有窗口，关闭后重开创建新设置，主页面输入独立保留。页面激活时 Ctrl/Cmd+, 可打开设置。原 orders 应用不再作为公开示例，MCP 回归夹具移至 cmd/keel-mcp/testdata/orders；表格、表单、对话框等演示仍在组件库中。
