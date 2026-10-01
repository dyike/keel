# 组件交互示例

```sh
go run ./examples/components
```

- 滑块：点击轨道，拖出左右边界，按方向键和 PageUp / PageDown 调整；Home / End 跳到两端。温度滑块使用 0.5 步长，音量使用 5 步长；勾选禁用后不再响应。
- 折叠面板：个人资料和通知设置互斥展开。在输入框填写内容，收起再打开，内容仍保留。禁用标题不响应点击。
- 键盘：Tab 聚焦标题，Enter / 空格切换展开，↑ ↓ 在可用标题间移动，Home / End 跳到首尾。收起内容里的输入框不应再获得焦点。
- 多项展开：打开下面两个问题，确认可以同时显示答案。
- 图片：点击色块图片，下方文字变为“已点击图片”。异步加载和失败占位见 `examples/chat -sample=images`。

生成截图：

```sh
go run ./examples/components -screenshot /tmp/keel-components.png
```

交互回归测试在 `ui/widget/slider_test.go`、`accordion_test.go`、`image_test.go`，自动化角色和键盘焦点测试在 `ui/window/automation_test.go`。

主题验证：`go run ./examples/components -section theme`，点击“浅色”“深色”来回切换；可用 `-theme dark` 指定启动配色，并与 `-screenshot /tmp/keel-theme.png` 组合检查截图。

M1 展示组件可用 `-section icon|alert|empty|avatar|tag|description_list|group_box|status_bar|marker|spinner|skeleton` 分别运行（任选一个值），加 `-theme dark` 检查深色。`-section focus` 验证 Tab / Shift+Tab、Space / Enter 激活、祖先按键冒泡、程序聚焦输入框和子树禁用。`-section time` 验证定时关闭与提前取消。Spinner/Skeleton 示例可切换减少动画。
