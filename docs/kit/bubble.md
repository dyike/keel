# Bubble

English | [简体中文](bubble.zh-CN.md)

Chat bubble.

```go
kit.Bubble(text).Mine()  // Posted by myself: Right, main color
kit.Bubble(text)         // Posted by others: left, light-colored background
```

- The width is up to 75% of the parent container. The content is an arbitrary View.
- If you want the bubble itself, use Bubble; if you want a complete message with an avatar and action bar, use Message.

Verify: `go run ./examples/components -section bubble`, add `-theme dark` to check the dark theme.

`Variant` Separate appearance and alignment: BubbleFilled (primary color), BubbleSecondary (secondary background color), BubbleMuted (weakened text), BubbleTinted (theme tint), BubbleOutline (stroke), BubbleGhost (frameless full line), BubbleDestructive (error color). The default BubbleAuto retains its original usage: Mine uses the primary color, and the rest use secondary background colors. Alignment(el.Start/el.End) can switch alignment, Mine is equivalent to End; explicit Variant is not affected by Mine, and illegal enumeration values are ignored.

The content and reaction area of ordinary bubbles are limited to 75% of the parent width, retaining Keel's original diameter; Ghost uses the entire line width and removes the default background, padding, border and rounded corners. Use SpaceLg/SpaceMd for normal content padding instead. Ghost does not add cropping areas, content can still use its own cropping or shadowing.

`Reactions(view)` receives an independent View, nil clears this slot; `ReactionSide(BubbleReactionTop/Bottom)` sets the position, defaults to the bottom; `ReactionAlignment(el.Start/el.End)` sets the alignment of the reaction area, defaults to the right. The reaction area uses Surface, Border and capsule rounded corners, and the content manages counting, selection and callback by itself, and buttons or pop-up triggers can be placed. The reaction area uses a flowing layout right next to the edge of the content.

`PartStyle(BubblePartRoot/Content/Reactions, fn)` adjusts the entire row, content surface, and reaction surface respectively after default styling; nil returns to default. The callback receives new elements each frame and should not retain references. Button and input box disabling, keyboard, and semantics are handled by subcomponents; the entire group can be disabled via Root's Disabled. Switching appearance, alignment and reactive position preserves child control identity. Bubble instances should be created and reused outside the frame.

```go
b := kit.Bubble(content).Variant(kit.BubbleOutline).
    Reactions(kit.Button("赞同", like).Variant(kit.ButtonGhost).Size(24)).
    ReactionSide(kit.BubbleReactionTop)
```

Continuous bubbles are available in the [BubbleGroup](bubble_group.md) combination, supporting spacing, styling, updating and group-level disabling.

`ReactionActions(buttons ...*ButtonView)` sets the direct button list, copies the list, ignores nil, and clears empty parameters. The button is still held by the application, and subsequent SetText/SetLoading/SetDisabled will take effect in the next frame; the variant, size, icon, custom content, and callbacks maintain the original configuration. Bubble only sets the rounded corners of these buttons to RadiusFull in this drawing, without modifying the button instances. When drawing independently, the default rounded corners of Button are still used.

As long as there is a direct button, the reaction area will remove the default padding and arrange it in rows, wrapping when the width is narrow. `PartStyle(BubblePartReactions, ...)` still executes after the default configuration and can explicitly restore padding or adjust spacing. The normal content of `Reactions(view)` is displayed before the direct button. The two can coexist and be cleared separately. Buttons, Popovers, etc. in the normal slot will not automatically change their rounded corners. The same direct button instance appears only once in the list, and reused instances retain focus during reflow.

```go
like := kit.Button("赞同 · 2", onLike).Variant(kit.ButtonGhost).Size(24)
copy := kit.Button("复制", onCopy).Variant(kit.ButtonGhost).Size(24)
b.ReactionActions(like, copy)
```
