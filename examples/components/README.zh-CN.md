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

## 窄屏与浏览器分类包

宽度低于 720 dp（随无障碍文字缩放调整）时，组件库显示可搜索的分类列表。点击组件进入占满宽度的示例页，44 dp 返回按钮回到列表。同一运行实例内，搜索条件和已创建的示例状态在返回、桌面/窄屏切换时保留。浏览器使用历史记录管理组件库页面，支持 `#component=button` 深链接。原生组件库仍是完整的单一应用。

Web 发布可使用独立链接的分类程序（`go run ./internal/gallerybuild -out demo`）。完整组件库（包括 `#component=input` 深链接）默认加载同一个完整程序，切换分类保留示例状态。嵌入式 `?section=button` 预览选择较小的分类程序。显式指定 `?bundle=general` 可使用独立分类程序；这种模式跨分类仍会重载并重置实例内状态。`?bundle=all` 显式选择默认的完整程序。这是独立 Go 运行时，不是共享运行时的 WASM 动态链接。

浏览器入口不包含截图矩阵导出和文件主题监听。没有代码/Markdown 示例的分类不加载可选 Go 语法高亮提供者。中文字体使用共享资源地址。浏览器安全区跟随 CSS 环境边距，在尺寸变化时更新；Gio 已提供可视视口的键盘边距。实体手机键盘、输入法和手势仍需真机验收。

Layout 示例在 480 dp 以下使用一列，720 dp 以下两列，桌面三列；断点随文字缩放调整。Dialog 和 Sheet 示例的操作按钮支持换行，窄屏高度为 44 dp。移动端交互测试覆盖 320／390 dp 下 Dialog、Sheet、Popover、Select 的打开与关闭，以及缩短视口后编辑、保存并重新进入对话框。这些离屏测试不能替代实体手机键盘和输入法验收。

组件深链接优先于过期的 `bundle` 参数；显式指定 `bundle=all` 仍保留完整运行时。修改地址片段和浏览器前进后退会恢复选中的组件，不额外压入历史记录。当前运行时缺少该组件时，会把当前地址替换为对应分类地址。

### 尚未闭环的对照工作

与 NSH 对照后仍有五类工作，不能理解为五个独立小 bug：

- 逐个验收组件在窄屏、大字体和触控下的表现。目前覆盖的是组件库外壳和部分布局、输入、弹层示例。
- 跨分类运行时重载后的状态保留。目前完整包可以在同一运行实例内保持状态。
- 研究共享 WASM 运行时。目前每个分类程序都包含自己的 Go 运行时。
- 实际浏览器、真机及部署验收：键盘／输入法、手势、安全区、历史记录、缓存命中与协商压缩。离屏／Node 测试和生成 gzip 文件不代表这些验收已经通过。
- 建立逐组件能力差距清单并确定新增优先级。组件库导航与加载优化尚未完成组件丰富度扩充。

运行 `KEEL_MOBILE_AUDIT=/tmp/keel-mobile KEEL_MOBILE_AUDIT_IMAGES=1 go test ./examples/components -run '^TestGalleryMobileAudit$'` 可生成全量移动端首帧盘点，包括语义候选项和 320 dp 下正常／放大文字截图；裁剪后的语义边界不能证明布局正确。分类包通过 HTTP 提供后，运行 `node scripts/verify-gallery-web.cjs http://127.0.0.1:8766/demo/`，并确保 Node 能找到 Playwright（例如通过 `NODE_PATH`）。脚本默认使用已安装的 Google Chrome 进行触控模拟，将截图和请求报告保存到 `work/gallery-web`；不代表实体手机键盘／输入法或部署响应头已验收。

窄屏下，Resizable 面板上下排列并分别保存横纵尺寸，Dock 使用紧凑标签切换而不改写桌面布局。ToggleGroup 支持横向滚动，Radio 卡片和操作栏换行，StaticTable 横向滚动以保持订单编号、金额可读。这些适配保留已有组件接口；示例负责选择响应式展示方式。

移动端标题按组件名称的单词边界换行，不缩小文字。长操作文案使用简短标签，分隔面板按钮描述当前可执行的操作；StatusMarker 的富内容操作支持换行。NavStack 示例展示模态导航和左边缘滑动返回，窗口测试覆盖 320／390／600 dp 及 1.6 字体比例。实体设备键盘／输入法仍需单独验收。
