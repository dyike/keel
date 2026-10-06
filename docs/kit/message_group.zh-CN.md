# MessageGroup

[English](message_group.md) | 简体中文

将完整消息行排成纵向组，保留各行的头像、头部、正文、状态、操作、反应和尾部。默认占满可用宽度，间距为 `theme.SpaceMd`（8dp）。分组由应用决定，不推断发送人，也不自动隐藏连续消息的头像或元数据。

```go
first := kit.Message("客服", answer).Header(kit.Label("客服 · 10:24"))
second := kit.Message("客服", details).Footer(kit.Label("已送达"))
group := kit.MessageGroup(first, second).Name("客服消息").Gap(6)
group.SetItems(second, first)
```

`SetItems` 复制切片并忽略 nil，空参数清空；`Items` 返回副本。应复用 Message 实例，每个实例在组内只出现一次。插入、删除其他项或重排时，保留消息的输入与焦点继续保留。自定义 View 也可加入，但需自行提供稳定 ID。

`Gap` 接受非负有限 dp，非法值忽略。`Style(func(*el.DivEl))` 在默认布局后调整背景、边框、间距等；nil 恢复默认，回调仅用于当帧元素。组件保留内部 ID、group 角色、名称及禁用状态。`SetDisabled` 禁用全部子项，不修改子组件自身状态。

与 BubbleGroup 的区别是组合对象：BubbleGroup 用于气泡表面，MessageGroup 用于含头像、元数据、操作的完整消息行。两者都不改变子项的对齐方式。

运行：`go run ./examples/components -section message_group`，加 `-theme dark` 检查深色。
