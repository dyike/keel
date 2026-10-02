# HoverCard

指针停在某个视图上时，弹出一张预览卡片，比如人员信息、链接摘要。

```go
card := kit.HoverCard(nameView, profileView).Width(300)
```

- 指针停留 `kit.HoverCardOpenDelay`（700ms）后打开；指针离开 `kit.HoverCardCloseDelay`（300ms）后关闭。
- 指针从目标移到卡片上时，卡片保持打开，因此卡片里可以放链接和按钮。
- 按 Esc 或点击外部会立即关闭；指针和焦点离开后才重新允许打开，避免停在原处又自动弹出。
- 可聚焦的目标获得焦点时立即打开；焦点在目标或卡片内部时保持打开，卡片不会主动移动焦点。纯文字目标不增加 Tab 停靠点，需要键盘入口时传入 Button 或其他可聚焦 View。

Agent：卡片的角色是 `dialog`，里面的元素单独列出。

验证：`go run ./examples/components -section hover_card`。

卡片宽高受窗口约束，长内容可滚动；非有限宽度被忽略。`SetDisabled(true)` 关闭并禁用目标区域。内容可以包含 Menu 等浮层，子菜单打开期间关闭延迟暂停，不会在操作中误收起父卡片。
