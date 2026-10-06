# MessageGroup

English | [简体中文](message_group.zh-CN.md)

Arrange complete message lines into vertical groups, retaining the avatar, header, body, status, action, reaction, and trailer of each line. The default is to fill the available width with a spacing of `theme.SpaceMd` (8dp). Grouping is determined by the application, with no inference of sender and no automatic hiding of avatars or metadata for consecutive messages.

```go
first := kit.Message("客服", answer).Header(kit.Label("客服 · 10:24"))
second := kit.Message("客服", details).Footer(kit.Label("已送达"))
group := kit.MessageGroup(first, second).Name("客服消息").Gap(6)
group.SetItems(second, first)
```

`SetItems` copies the slice and ignores nil, empty parameters are cleared; `Items` returns a copy. Message instances should be reused, each appearing only once within the group. The input and focus of the retained message are retained when other items are inserted, deleted, or reordered. Custom Views can also be added, but you need to provide a stable ID yourself.

`Gap` accepts non-negative finite dp, illegal values are ignored. `Style(func(*el.DivEl))` adjusts the background, border, spacing, etc. after the default layout; nil restores the default, and the callback is only used for the current frame element. Components retain internal IDs, group roles, names, and disabled states. `SetDisabled` disables all sub-items and does not modify the status of the sub-component itself.

The difference from BubbleGroup is the combination object: BubbleGroup is used for the bubble surface, and MessageGroup is used for the complete message line including avatar, metadata, and operations. Neither changes the alignment of the child.

Run: `go run ./examples/components -section message_group`, add `-theme dark` to check the dark theme.
