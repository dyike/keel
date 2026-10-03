# AvatarGroup

把多个头像叠放成成员组，限制可见人数后用 `+N` 显示隐藏人数。

```go
team := kit.AvatarGroup(
    kit.Avatar("张三").Status(kit.AvatarOnline),
    kit.Avatar("Ada Lovelace"),
    kit.Avatar("Bob"),
).Size(40).Limit(2)
```

`Size(dp)` 统一组内直径，默认 40dp，限制在 16–256dp；不修改成员头像自身尺寸。头像重叠四分之一，以主题表面色描边区分。窄窗口提供横向滚动。

`Limit(n)` 限制实际显示的头像数量，溢出标记另外占一格。默认负数表示不限，0 只显示总隐藏人数。`Ellipsis(true)` 把可见的 `+N` 换成省略号，Agent 仍可读取隐藏人数。

`SetAvatars` 复制成员切片并过滤 nil；`Avatars()` 返回切片副本，头像实例仍然共享，可以继续更新姓名、图片和状态。空组不占高度。

Agent：外层为 `group`，可见成员为 `avatar`，保留原姓名和状态；溢出头像名称为本地化的“更多 N”，value 为隐藏人数。组为展示内容，不占用键盘焦点。

验证：`go run ./examples/components -section avatar_group`，加 `-theme dark` 查看深色。
