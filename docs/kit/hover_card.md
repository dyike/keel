# HoverCard

指针停在某个视图上时，弹出一张预览卡片，比如人员信息、链接摘要。

```go
card := kit.HoverCard(nameView, profileView).Width(300)
```

- 指针停留 `kit.HoverCardOpenDelay`（700ms）后打开；指针离开 `kit.HoverCardCloseDelay`（300ms）后关闭。
- 指针从目标移到卡片上时，卡片保持打开，因此卡片里可以放链接和按钮。
- 按 Esc 或点击外部会立即关闭。
- 卡片不移动焦点。只能用键盘操作的用户无法打开卡片，所以不要把唯一的操作入口放在卡片里。

Agent：卡片的角色是 `dialog`，里面的元素单独列出。

验证：`go run ./examples/components -section hover_card`。
