# 组件示例

```sh
go run ./examples/components                    # 全部组件，从上到下排列
go run ./examples/components -section select    # 单个组件，名字与 docs/kit/<名字>.md 一致
go run ./examples/components -section inputs    # 一类：controls、inputs、overlays、data
go run ./examples/components -theme dark        # 深色
```

每个 kit 组件对应一个 `<名字>.go`，用 `registerSection` 注册同名 section，展示常用状态、边界和浅深色。`ui/kit/conventions_test.go` 检查每个组件都有示例。

生成截图：

```sh
go run ./examples/components -section chart -screenshot /tmp/keel-chart.png
```

另有几个验证 el 基础能力的 section：`theme`（运行时切换浅深色）、`focus`（Tab / Shift+Tab、子树禁用）、`time`（定时关闭与取消）、`overlay`（非模态点击穿透、模态遮罩、Esc 关闭与焦点恢复）。

组件的交互测试在 `ui/kit/*_test.go`，Agent 快照测试在 `ui/window/kit_*_test.go`。

`-section scrollable` 验证 el 横向滚动、宽内容裁剪和程序定位。

`-section variable_list` 验证 10 万行自然高度列表：定位、插入历史、展开内容与窗口宽度变化。

`-section layout` 验证换行布局和简单网格，缩窄窗口可检查换行、行列间距和最小尺寸。
