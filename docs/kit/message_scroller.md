# MessageScroller

English | [简体中文](message_scroller.zh-CN.md)

Conversation scrolling area virtualized by stable message ID, only visible messages and nearby pre-read areas are built.

```go
sc := kit.MessageScroller(ids, 120, func(cx *el.Context, i int) el.Element {
    return renderMessage(cx, messages[i])
}).OnReachTop(loadOlder)
// Update ID order after inserting history, adding, deleting or rearranging messages:
sc.SetKeys(ids)
// Locate unread messages after the data is loaded; it can also be called before the first rendering:
found := sc.ScrollToMessage(unreadID)
_ = found // If the ID does not exist, return false and retain the original positioning request.
// When the user sends a message, explicitly jump to the latest:
sc.ScrollToEnd()
```

IDs are arranged from oldest to newest and must be non-empty and unique. Duplicate values will panic before modification. Both construction and SetKeys copy slices. The second parameter is the estimated height (dp) of the unmeasured message; the true height is determined by the content, supporting answers and images of different lengths.

- When it is at the bottom, it automatically follows new messages and the flow increases; after scrolling up to read, the current message and its screen position are retained, and "Back to Latest" is displayed.
- SetKeys When inserting history, the reading position is retained according to the stable ID; the message above the visible area is increased and the scroll position is also corrected. No need to call HistoryPrepended anymore.
- The height of the visible message is checked every frame; `Invalidate(ids...)` is called after changing the off-screen content. If the ID is not passed, all height caches are cleared.
- `ScrollToMessage(id)` displays the message with minimum scrolling; it can also be requested before the first rendering, and the message that exceeds the viewport height is displayed at the beginning. Message jump and `ScrollToEnd` are based on the last valid request. The unread mark is maintained by the application.
- `IsScrolledUp(cx)` queries whether there is content below the recently drawn viewport; `IsFollowingTail(cx)` queries the actual automatic following status. They are false before the first drawing and the default is true, and the pending message jump will be paused to follow.
- `SetFollow(false)` turns off automatic following; closing it immediately after construction will read from the beginning and retain the position when it is displayed. Explicit `ScrollToEnd` still jumps to the end, but does not modify this switch. When following is enabled, manually scrolling back to the bottom will resume following.
- `OnReachTop` loads a batch of history when it reaches the top, and will not be called repeatedly when it stops at the top. After loading, the trigger must be at least one viewport away from the top to avoid repeated requests caused by small batch insertion and scroll wheel inertia.
- `SetDisabled` disables scrolling, message content and "back to latest", also suspends the top load callback. The component fills the space given by the parent container.

The Markdown in the message is still saved as a Doc by the caller, and the selection range and streaming parsing state are not reconstructed by virtualization. Drag and select within a single answer to the edge of the viewport to continue scrolling and release it to stop; text selections for different messages are independent of each other.

Agent: Container role log, currently visible messages are listed one by one. Verification: `go run ./examples/components -section message_scroller`; complete flow and selection process with `go run ./examples/chat`.

The "Back to latest" button is enabled by default, `LatestButton(false)` hides it without changing the scrolling state; `LatestLabel` sets the text and accessible name, and the empty string restores the current language. `LatestRenderer` receives the default Button created every frame, can modify the variant, icon, size, Content and Appearance, and can also return another Button. The component copies the return value and retains the internal ID and jump action; nil configuration or return nil to restore default. Custom content is limited to display elements.

```go
sc.LatestLabel("View new messages").LatestRenderer(func(b *kit.ButtonView) *kit.ButtonView {
    return b.Variant(kit.ButtonPrimary).Outline(true).Size(32)
}).LatestTransition(250 * time.Millisecond)
```

The button uses 150ms to fade in and out by default; `LatestTransition(0)` switches immediately, negative duration is ignored, and reduced motion takes priority. Interaction is immediately prohibited during exit and removed after completion. The text button in the lower right corner is retained by default, which is different from the round icon button appearance of GPUI.
