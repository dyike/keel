# Marker

`kit.Marker(kit.MarkerDiamond).Color(theme.Info).Size(12)` 创建纯图形标记，支持 MarkerDot、MarkerSquare、MarkerDiamond。默认 8dp，颜色每次 Render 读取 theme.Text；显式 Color 保持固定。

Marker 不向 Agent 暴露语义，也没有键盘操作。用于图例、列表项、图表，需可读标签时由调用方在旁边放 Text。父容器约束会限制绘制尺寸。

验证：`go run ./examples/components -section marker -theme dark`，省略 theme 查看浅色。
