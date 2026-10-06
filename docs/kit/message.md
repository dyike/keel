# Message

English | [简体中文](message.zh-CN.md)

A message in a conversation: avatar, content, action bar.

```go
kit.Message("Me", text).User()                  // User message: The bubble on the right does not display the avatar.
kit.Message("AI assistant", answerDoc).Actions(copy) // Others: avatar + full width content + action bar
```

- Non-user messages take up the full width, suitable for long Markdown answers. The avatar displays the author's initials.
- `Actions(views...)` appears below content, such as copy and retry buttons. When the streaming output has not ended, no operation is usually performed.

Agent: Each message is `article`, the name is the author, the content and action buttons are listed separately.

Verify: `go run ./examples/components -section message`, add `-theme dark` to check the dark theme.

`SetState(MessageSending, "")` shows sending; `SetState(MessageFailed, reason)` shows failure reason. `OnRetry(fn)` displays the retry button when it fails. Click to change to sending first, and then call a business callback; after the business is completed, `MessageReady` is set, and if it fails again, `MessageFailed` is set.

Both user and assistant messages support `Actions`, incoming slices will be copied, and empty slots will be ignored. Whether the action bar is displayed during streaming output is determined by the application. `SetDisabled(true)` disables actions, reactions, and retries within the message.

```go
msg.Reactions(kit.MessageReaction{Name: "Helpful", Count: 2}).
    OnReaction(func(index int, active bool) { /* Save to the server */ })
```

The reaction uses a toggle button, clicks to update the current user's selected status and count, and then calls the callback. `Reactions` copies data; subsequent server results can be overwritten by calling it again. Reactions with no callback provided are read-only. Agent can read the `sending` / `failed` status of `article` and the selected status of the response button.

`Avatar(view)` replaces the avatar, pass nil to hide; `DefaultAvatar()` restores the default: the assistant displays the initials of the name, and user messages do not display the avatar. When the user message explicitly configures the avatar, it is placed on the right side. The avatar is aligned according to the bottom edge of the text container, with the head at the top, and status, operations, reactions, and tails continue to be arranged below the text column; positioning uses the current frame layout, and the taller avatar will expand the space above. When the text is hidden or does not exist, the bottom edge of the text column is used as the fallback.

`Header(view)` is placed above the text, and `Footer(view)` is placed after the status, operation and reaction areas; pass nil to clear. The head and tail support any View and interactive controls. The trumpet and weakened text styles are inherited by default, and user messages are placed on the right. `Content(view)` replaces the body independently, nil clears the body but retains other partitions. Message disabled state covers all slots.

```go
msg.Avatar(kit.Avatar("Alice").Size(32)).
    Header(kit.Label("Alice · 10:24")).
    Footer(kit.Button("Reply", reply))
```

When inserting or deleting avatars or headers and tails, the identity of the text remains stable, and the input content and focus remain unchanged. The head and tail of ordinary user bubbles and explicit bubbles use `theme.SpaceLg` horizontal indentation by default; the text of ordinary assistants remains unindented. Full message lines are available in the [MessageGroup](message_group.md) grouping.

`Bubble(surface)` installs explicit bubbles to prevent User messages from wrapping another layer of bubbles. The component is rendered as a copy, and the bubble alignment follows the message; the variant that modifies the original bubble will be reflected in the next frame, and the original instance will not be modified in reverse. Pass nil to clear the text; `Content(view)` restores the normal text mode (User automatically includes bubbles).

```go
surface := kit.Bubble(answer).Variant(kit.BubbleGhost)
msg.Bubble(surface).Header(kit.Label("System message")).Footer(kit.Label("Just now"))
msg.HeaderInset(true).FooterInset(false)
msg.ResetContentInsets() // Restore automated rules
```

Explicit Ghost bubbles automatically cancel head and tail indentation; `HeaderInset` and `FooterInset` override inheritance rules respectively. Custom Views inside ordinary Content do not participate in Ghost detection. This entrance accepts a single Bubble; multiple bubbles and attachments are mixed using [MessageContent](message_content.md), and the direct bubbles participate in Ghost inheritance.

`Alignment(el.Start/el.End)` independently controls the left and right positions, overwriting the User's default right layout; `ResetAlignment()` restores the default. Avatars, heads, tails, states, actions, and reactions follow the position, as do explicit bubbles. The user still determines the default bubble color and whether to display the default avatar, and changing the alignment will not change these settings. Other Align values are ignored. When the normal text is positioned to the right, it is laid out according to its own width, and controls that fill the width still occupy the text column.

`PartStyle(part, func(*el.DivEl))` configures the partition style after the default layout of each frame, nil returns to default, and illegal parts are ignored. Optional partitions: Root (outer layer), Stack (text column), Avatar, Header, Content, Footer, Status, Actions, Reactions, the corresponding constants all start with `MessagePart`. Header and footer styles are executed after automatic indentation and explicit overrides, allowing further adjustment of padding.

```go
msg.PartStyle(kit.MessagePartRoot, func(e *el.DivEl) {
    e.P(theme.SpaceLg).Bg(theme.Subtle).Rounded(theme.RadiusMd)
}).PartStyle(kit.MessagePartStack, func(e *el.DivEl) {
    e.Gap(theme.SpaceSm)
})
```

The callback only modifies the current frame element and should not retain it or append children. The component retains the internal ID, name/state of the outer article, and overall disablement; partitions can additionally disable themselves and cannot bypass ancestor disabling. The Content style acts on the body container, and the bubble surface is still controlled by Bubble's style interface.

Custom Items for `MessagePartRoot` can override the default bottom alignment; ContentBottom for `MessagePartStack` can select another descendant as the alignment target. It is the application's responsibility to change these layout rules. The default avatar is still 28dp. The assistant automatically displays the avatar and the User automatically packs the main color bubble convention. This is different from the upstream default of no slot.
