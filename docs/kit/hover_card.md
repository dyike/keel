# HoverCard

指针停在某个视图上时，弹出一张预览卡片，比如人员信息、链接摘要。

```go
card := kit.HoverCard(nameView, profileView).Width(300).
    OpenDelay(200*time.Millisecond).CloseDelay(500*time.Millisecond).
    Placement(el.Right, el.Start).Offset(8)
```

- 默认指针停留 700ms 后打开，离开 300ms 后关闭；`OpenDelay` / `CloseDelay` 按实例配置，负数按零处理。修改正在等待的延时会从修改时重新计时。
- 指针从目标移到卡片上时，卡片保持打开，因此卡片里可以放链接和按钮。
- 按 Esc 或点击外部会立即关闭；指针和焦点离开后才重新允许打开，避免停在原处又自动弹出。
- 可聚焦的目标获得焦点时立即打开；焦点在目标或卡片内部时保持打开，卡片不会主动移动焦点。纯文字目标不增加 Tab 停靠点，需要键盘入口时传入 Button 或其他可聚焦 View。

Agent：卡片的角色是 `dialog`，里面的元素单独列出。

验证：`go run ./examples/components -section hover_card`。

卡片宽高受窗口约束，长内容可滚动；非有限宽度被忽略。`SetDisabled(true)` 关闭并禁用目标区域。内容可以包含 Menu 等浮层，子菜单打开期间关闭延迟暂停，不会在操作中误收起父卡片。

`Placement(side, align)` 设置首选方向与对齐，`Offset(dp)` 设置锚点间距（默认 4dp，忽略非有限值）。窗口边缘仍自动避让；已打开时修改定位会在下一帧生效。延时只影响鼠标悬停，键盘聚焦仍立即打开。
