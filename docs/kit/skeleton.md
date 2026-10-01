# Skeleton

`kit.Skeleton().W(el.Dp(240)).H(el.Dp(16))` 创建灰色加载占位，默认铺满宽度、高 16dp。Circle 绘制居中的圆形，常与相同宽高搭配。Shimmer 启用扫光，默认使用明暗脉冲。

动画基于 cx.Now 并通过 cx.Animating 请求帧；减少动画时保持静态。Skeleton 是装饰，不暴露 Agent 语义，不处理键盘。Shimmer 是选项，不是独立组件。

验证：`go run ./examples/components -section skeleton -theme dark`，省略 theme 查看浅色。示例按钮切换减少动画；像素测试注入帧时间验证两种动画及静止状态。
