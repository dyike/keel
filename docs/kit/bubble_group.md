# BubbleGroup

English | [简体中文](bubble_group.zh-CN.md)

Arrange continuous bubbles into vertical groups with default spacing of `theme.SpaceMd` (8dp). Grouping is determined by the application; the component does not infer the sender or change the appearance, left-right alignment, or reaction slots of individual bubbles.

```go
group := kit.BubbleGroup(first, second).Name("Support messages").Gap(6)
group.SetItems(first, second, third)
```

`SetItems` copies the slice and ignores nil, empty parameters are cleared; `Items` returns a copy of the slice. Bubble instances should be reused, with each instance appearing only once within the group. When other items are inserted, deleted, or rearranged, the internal input state and focus of the retained instance continue to be retained; the deleted control no longer participates in interaction. Ordinary custom views need to provide a stable ID by themselves.

`Gap` accepts non-negative finite dp, illegal values are ignored. `Style(func(*el.DivEl))` configures spacing, borders, background, etc. after the default layout. nil returns to default; the callback is only used for the current frame element. `Name` sets the semantic name of the group, `SetDisabled` disables all children but does not modify the state of the child components themselves. Components maintain group identities, roles, names, and disabled states.

The group takes up the available width, normal Bubble still lays out its own 75% cap, and Ghost uses the entire row. Use Message when you need an avatar, author, or message status.

Run: `go run ./examples/components -section bubble_group`, add `-theme dark` to check the dark theme.
