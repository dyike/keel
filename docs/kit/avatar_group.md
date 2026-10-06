# AvatarGroup

English | [简体中文](avatar_group.zh-CN.md)

Stack multiple avatars to form a member group, limit the number of visible people and then use `+N` to display the hidden number of people.

```go
team := kit.AvatarGroup(
    kit.Avatar("Alex Chen").Status(kit.AvatarOnline),
    kit.Avatar("Ada Lovelace"),
    kit.Avatar("Bob"),
).Size(40).Limit(2)
```

`Size(dp)` Unifies the diameter within the group, defaults to 40dp, limited to 16–256dp; does not modify the size of member avatars themselves. The avatars overlap by a quarter and are distinguished by the theme surface color strokes. Narrow windows provide horizontal scrolling.

`Limit(n)` limits the number of avatars actually displayed, and the overflow mark occupies an additional space. By default, a negative number means no limit, and 0 only displays the total number of hidden people. `Ellipsis(true)` Replace the visible `+N` with an ellipsis, and the Agent can still read the hidden number of people.

`SetAvatars` copies the member slice and filters out nil; `Avatars()` returns a copy of the slice. The avatar instance is still shared and you can continue to update the name, picture and status. Empty groups do not occupy height.

Agent: The outer layer is `group`, the visible member is `avatar`, and the original name and status are retained; the overflow avatar name is the localized "More N", and the value is the hidden number of people. The group is for display content and does not occupy keyboard focus.

Verify: `go run ./examples/components -section avatar_group`, add `-theme dark` to see dark colors.
