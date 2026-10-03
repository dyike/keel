# ShimmerText

保留可读文字，让高光扫过字形；背景和字间空白不受影响。

```go
loading := kit.ShimmerText("正在生成内容……").Duration(2*time.Second).Spread(.3)
loading.Reverse(true).Once(true)
loading.Restart()
```

默认从左向右循环，周期 2 秒，高光半宽为文字框宽度的 30%。`Duration` 接受正时长；`Spread` 接受 (0,1]，非法值忽略。`Reverse` 反向；`Once` 在一个周期后恢复普通文字并停止请求动画帧，`Restart` 重新开始。调整时长不重置起点，需要时显式 Restart。

`Enabled(false)` 显示普通文字，再启用从头开始；减少动画设置下同样显示普通文字。`SetText` 更新正文，`Size` 设置字号（0 继承），`MaxLines` 限制行数（0 不限）；字体、字重、行高和颜色默认从 el 容器继承。`Color` 覆盖文字底色，`Highlight` 覆盖高光颜色，默认高光使用主题 PrimaryText。

彩色位图字形（例如部分 emoji）保留原色，不参与高光着色。组件保持文本语义与布局尺寸，不承担加载任务，也不提供文字选择。底层 `el.Text(...).Shimmer(phase, spread, color)` 只负责绘制，便于应用自定义时间控制。

运行示例：`go run ./examples/components -section shimmer_text`。
