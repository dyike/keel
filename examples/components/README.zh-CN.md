# 组件示例

[English](README.md) | 简体中文

```sh
go run ./examples/components                    # 组件库应用：左侧导航和搜索，右侧是选中的组件
go run ./examples/components -section select    # 只显示一个组件，名字与 docs/kit/<名字>.md 一致
go run ./examples/components -section inputs    # 一类：controls、inputs、overlays、data、shell
go run ./examples/components -theme dark        # 深色启动；应用右上角也能切换深浅色和中英文
```

默认使用英文；加上 `-lang zh-CN` 使用中文。浏览器示例通过 `?lang=en` 或 `?lang=zh-CN` 选择语言。`demoText(english, chinese)` 选择示例标签、消息和数据；框架控件使用 `ui/locale`。切换组件库语言会重建示例，保留当前组件和搜索词。

应用的侧栏按"基础能力、基础组件、输入、浮层、数据、应用外壳"分组，顶部搜索框过滤组件名。每个组件第一次打开时创建，切走再回来状态保留。整窗口的组件（Dock、Settings 等 el.Root）占满右侧内容区，自带的浮层也限制在内容区里。

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
