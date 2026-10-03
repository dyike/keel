# Spinner

`kit.Spinner().Size(20).Label("加载中")` 显示不确定进度。Label 为空时只画图形，默认名称为“加载中”。Agent 角色 progressbar，value 为 indeterminate。没有键盘操作。

旋转相位来自 cx.Now，cx.Animating 请求下一帧，不启动 goroutine。theme.SetReducedMotion(true) 后保持静态。固定图形尺寸，文本按父容器约束换行。

验证：`go run ./examples/components -section spinner -theme dark`，省略 theme 查看浅色；示例按钮切换减少动画。像素测试注入不同帧时间，检查旋转和静止。

`Icon(kit.IconSettings)` 将圆环替换为旋转图标，`Icon(kit.IconNone)` 恢复圆环；`VectorIcon` 接受自定义 Gio 图标，传 nil 恢复圆环。`Color(color.NRGBA{...})` 设置图形颜色，标签仍使用主题的 Muted 色。未设置颜色时随主题使用 PrimaryText。自定义图标与圆环均为每秒一周，减少动画时静止；尺寸与 Agent 语义保持一致。
