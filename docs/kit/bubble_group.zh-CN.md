# BubbleGroup

[English](bubble_group.md) | 简体中文

将连续气泡排成纵向组，默认间距为 `theme.SpaceMd`（8dp）。分组由应用决定；组件不推断发送人，也不改变各气泡的外观、左右对齐或反应槽。

```go
group := kit.BubbleGroup(first, second).Name("客服消息").Gap(6)
group.SetItems(first, second, third)
```

`SetItems` 复制切片并忽略 nil，空参数清空；`Items` 返回切片副本。应复用 Bubble 实例，每个实例在组内只出现一次。插入、删除其他项或重排时，保留实例的内部输入状态与焦点继续保留；被删除的控件不再参与交互。普通自定义 View 需自行提供稳定 ID。

`Gap` 接受非负有限 dp，非法值忽略。`Style(func(*el.DivEl))` 在默认布局后配置间距、边框、背景等，nil 恢复默认；回调只用于当帧元素。`Name` 设置组的语义名称，`SetDisabled` 禁用所有子项但不修改子组件自身状态。组件保持组身份、角色、名称和禁用状态。

分组占满可用宽度，普通 Bubble 仍按自己的 75% 上限布局，Ghost 使用整行。需要头像、作者或消息状态时使用 Message。

运行：`go run ./examples/components -section bubble_group`，加 `-theme dark` 检查深色。
