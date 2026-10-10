# Elements and views (ui/el)

English | [简体中文](el.zh-CN.md)

`ui/el` is written in GPUI style: the interface is a view (ordinary Go struct), and its `Render` is called every frame, returning an element tree built in a chain style. The status is the field of the struct. The event callback directly changes the field and automatically redraws the next frame.

```go
type Counter struct{ n int }

func (c *Counter) Render(cx *el.Context) el.Element {
    return el.Div().P(24).Gap(12).Items(el.Start).Child(
        el.Text(fmt.Sprintf("Clicked %d times", c.n)).TextSize(20).Bold(),
        el.Div().Px(16).Py(8).Rounded(6).Bg(theme.Primary).TextColor(theme.OnColor).
            CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.PrimaryHover) }).
            OnClick(func() { c.n++ }).
            Child(el.Text("+1")),
    )
}

window.Open(window.Options{Title: "Counter", Content: el.Root(&Counter{})})
window.Main()
```

Compared with writing Gio directly:

| | Gio | el |
| --- | --- | --- |
| Interactive state | Each component declares fields such as `widget.Clickable` by itself | The frame is automatically saved according to the element position (or `ID`) without declaration |
| Layout | Nested layer by layer `layout.Flex{}.Layout(gtx, layout.Rigid(...))` | flexbox: inner and outer margins, spacing, scaling, alignment, percentage size, scrolling, absolute positioning |
| Style | Draw handwriting every time | Each element can be set in a chain, and there will be style variations when hovering and pressing |
| Event results | The processing order depends on the layout order | Process the event first and then render, and the same frame will be drawn |
| Agent semantics | Manual declaration | Automatic: `OnClick` is a button, text is text, `Input` is an input box |

The ready-made components are at [ui/kit](kit.md), and they are all el views.

## What happens in a frame

```
1. Dispatch events   Clicks registered in the previous frame → call the element’s OnClick
2. Render            view.Render(cx) → element tree (rebuilt each frame from current state)
3. Layout            The flexbox engine calculates element positions and sizes
4. Paint             Backgrounds, borders, text; register click areas; generate semantics
5. Collect           Remove state for elements absent from this frame
```

Render will be called every frame, so it should be kept cheap: only build trees based on status, no I/O, no large calculations. Put time-consuming things into goroutine, and use `core.Update` to change the status after completion. Threading rules are the same as other modules, see [Architecture · Threading Rules](architecture.md#threading).

### Numeric label image cache

Views with many changing numeric labels can opt into a shared glyph atlas:

```go
atlas := new(theme.GlyphAtlas)
root := el.Root(view)
root.SetTextAtlas(atlas)
window.Open(window.Options{Content: root, OnClose: atlas.Release})
```

Use one atlas per root. The root must be drawn at an integral pixel translation, without extra scaling or rotation; font sizes still follow the display metric. Leave `SubpixelPhases` at zero to preserve glyph positions. The root prepares visible, short, opaque, single-line labels containing digits, commits the changed image pages once, then paints them. Text, font, constraints and color must still match at paint time. Fractional label origins, complex text, decorated or scroll subtrees, overlays, rich text and editors retain the existing drawing path. Input handling and accessibility labels are preserved.

Preparation retains at most 2048 labels and 32768 glyphs for one frame; additional labels use the existing path. Images and masks follow [GlyphAtlas's budgets](theme.md#glyph-image-cache), with extra memory for GPU copies. The first frame must build its cache, so measure startup, CPU and physical memory for your view before enabling it. Pass `nil` to `SetTextAtlas` to disable it, and call the caller-owned atlas's `Release` when it is no longer needed.

## Element

| Construction | Description |
| --- | --- |
| `el.Div()` | Box, the only element that can have child elements. Default child elements are arranged from top to bottom |
| `el.Text(s)` | Text, wrap to available width |
| `el.Input()` / `el.TextArea()` | Input box, with default border style, the border turns blue when it gets focus. The default height of the single-line box is `theme.ControlHeight`, and the text is vertically centered, consistent with the field of kit; the point padding will also be focused |
| `el.Widget(w)` | Embed any `core.Widget`, such as a Gio layout wrapped with `core.Func` |

Input box:

```go
el.Input().ID("q").Placeholder("Search").Bind(&v.query).OnChange(func(s string) { v.refresh() }).OnSubmit(v.search)
```

Clicking on an empty space will cause the input box to lose focus.

`Bind(&text)` two-way binding: user input will be written into variables. If the program changes the variables, the input box of the next frame will also change. `Password()` obscures content. `MaxLen(n)` limits the number of characters, `Filter("0123456789")` only accepts these characters (both input and paste are filtered), and `ReadOnly(true)` allows selection for copying but not editing. After the single-line input box is set to `OnKey`, ↑ ↓ PageUp PageDown is first handed over to it for processing, and the editor no longer receives these keys (in the single-line box, they can only move the cursor to the beginning or end); combinations with modifier keys such as Shift are still owned by the editor and are used to expand the selection. The return value does not affect the result, these keys are always taken away. The input box cannot be edited when it is located in the subtree of `Disabled(true)`, and the same is true for the Gio code embedded in `el.Widget`.

When the input box is surrounded by another layer of boxes (a search box with icons and buttons), set `ID` and `FocusOnPress(inputID)` to the outer box: clicking on a place in the outer box where there are no sub-elements will focus the input box, and the mouse will display the text cursor; the outer box will not turn into a button, and the tab order will not be entered.

## Style method

All elements share the same set of methods (`Styled[T]` generic implementation, chain call returns the original type):

| Classification | Method |
| --- | --- |
| Orientation and alignment | `Row()`, `Col()` (default), `Wrap()` wrap, `Grid(columns)` equal width column grid, `ColSpan(n)` span columns, `Gap(dp)`, `Justify(Start/Center/End/SpaceBetween/SpaceAround)`, `Items(Start/Center/End/Stretch)`, `Center()` |
| Scaling | `Grow()` is equal to CSS's `flex: 1`: the initial size is calculated as 0 and the remaining space is shared; other elements shrink proportionally when there is not enough space, `NoShrink()` is prohibited from shrinking |
| Size | `W(l)`, `H(l)`, `Size(l)`, `MinW/MinH/MaxW/MaxH(l)`, `WFull()`, `HFull()`; use `el.Dp(40)` for length, `el.Sp(40)`, `el.Frac(0.5)`, `el.Full` for font size scaling |
| spacing | `P`, `Px`, `Py`, `Pt`, `Pb`, `Pl`, `Pr` (padding), `M`, `Mx`, `My`, `Mt`, `Mb`, `Ml`, `Mr` (margin), unit dp |
| Scrolling and positioning | `ScrollX()` horizontal scrolling (needs to constrain the width), `ScrollY()` vertical scrolling (requiring a certain height), `StickToBottom()` follows to the bottom, `ScrollToEndOn(v)` jumps to the bottom when v changes; `Absolute()` + `Top/Right/Bottom/Left` absolute positioning, and the width will be stretched to the left and right at the same time |
| Appearance | `Bg(c)`, `Border(dp, c)`, `Rounded(dp)` (use `theme.RadiusSm/Md/Lg/Xl/Full`), `RoundedCorners(topLeft, topRight, bottomRight, bottomLeft)` sets the four corners respectively, for example, the button group only rounds the outer corners, `Shadow(theme.ElevationSm/Md/Lg)` draws the shadow outside the element without changing the size, `Opacity(a)` sets the entire subtree 0–1 transparency (0 Not drawn at all, layout and interaction are still retained), `CursorPointer()`, `Hidden(b)`, `IsHidden()` (read the hidden value declared by this element, excluding ancestors) |
| Text (downward inheritance) | `TextColor(c)`, `TextSize(sp)` (use `theme.TextXs` ... `theme.TextHeading`), `Bold()`, `Medium()`, `Mono()` monospaced font (`theme.MonoFace`), `Italic()`, `LineHeight(multiplier)`, `MaxLines(n)` |
| State variations | `Hover(func(*el.Style))`, `Active(func(*el.Style))`: color changes when hovering and pressing, the background gradually fades within 120ms; switch directly when the reduced motion is turned on (`theme.SetReducedMotion`, automation mode is turned on by default) |
| Interaction | `OnClick(fn)`, `OnDoubleClick(fn)` |
| Structure | `ID(s)`, `Child(...)`, `Children(slice)`, `When(cond, func(*T))` |
| Agent Semantics | `Role(s)`, `Name(s)`, `Value(s)`, `Selected(b)` |

The layout rules follow the intuition of flexbox, which is different from CSS in several ways:

- `Div` is arranged vertically by default, and the sub-elements fill the width (like block-level elements). `Row()` is changed to horizontal.
- The width of horizontally arranged sub-elements takes the width of the content; when there is not enough space, they are shrunk proportionally and the text is wrapped accordingly.
- The viewport of `ScrollY` is a padding box, and the padding scrolls with the content.
- Text styles are inherited to child elements just like CSS.

## Status and ID

The internal state of the element (hover, pressed, scroll position, input box content and cursor) is saved by the framework, and the key is the path of the element in the tree: for each level, take its `ID`, if not, take its sequence number among siblings.

**List items that will be added, deleted, and rearranged must be given to `ID`**, otherwise the status will be mismatched according to position (for example, if the first line is deleted, the input box content of the second line will move to the first line):

```go
el.Div().Children(el.Map(v.rows, func(i int, r Row) el.Element {
    return el.Div().ID(r.ID).OnClick(func() { v.open(r) }).Child(el.Text(r.Name))
}))
```

If the element does not appear in a certain frame, its state is deleted. Input fields will be cleared when hidden and then shown; to retain the content, use `Bind` to place the content in the view's field.

## Views and combinations

A view is any type that implements `Render(*el.Context) el.Element`. Splitting the interface means splitting the struct, and calling the Render of the subview in the Render of the parent view:

```go
type Page struct{ list *OrderList; form *OrderForm }

func (p *Page) Render(cx *el.Context) el.Element {
    return el.Div().Row().Gap(16).Child(
        el.Div().W(el.Dp(300)).Child(p.list.Render(cx)),
        el.Div().Grow().Child(p.form.Render(cx)),
    )
}
```

`el.ViewFunc(func(cx *el.Context) el.Element { … })` Adapts the function to a View, suitable for the content slot passed to the kit. The slot holds the View, and Render is called every frame; the theme color is read within the function, do not save the Element when constructing the view.

`cx.Shortcut("mod+s", fn)` Binds a shortcut key during view rendering.

### Action and key table

If you want the user to change keys, don't hardcode the keys in the view, instead use named actions:

```go
core.Bind("editor.save", "mod+s")      // Set the default key at startup, which can be given to multiple
core.LoadKeymap(userJSON)              // Overlay user's {"editor.save": ["ctrl+alt+s"]}

func (v *editor) Render(cx *el.Context) el.Element {
    cx.Action("editor.save", v.save)   // Every key bound to this action will fire
    ...
}
```

- The key table is global. `core.Bind` replaces all the keys of an action. If the key is not transmitted, it will be unbound. If the key is written incorrectly, an error will be returned and nothing will be changed.
- After the key is changed, all windows are redrawn and take effect from the next frame; `cx.Action` is registered according to the current key table every frame, without restarting.
- `cx.Perform(targetID, action)` directly executes an action in the callback, and the effect is the same as pressing its shortcut key when the focus is on targetID: first find the innermost `cx.ActionAt` that wraps the element, and then find `cx.Action`; no binding keys are required. The menu's `ActionItem` and the command panel use this shared command to implement, and the return value indicates whether there is a processor running.
- `core.Bindings(name)` looks up the keys for an action, `core.Keymap()` returns all, and can be used as a shortcut key setting page.
- `kit.KbdFor(name)` and `Menu.ActionItem` are used to display the keys, which change according to the key table.

### Cache unchanged parts

Most of the content in long lists and chat history does not change every frame. `cx.Cache(key, build)` directly reuses the elements and layout of the previous frame when the key remains unchanged, and `build` will not be called:

```go
for _, m := range v.msgs {
    m := m
    list.Child(cx.Cache(msgKey{m.id, m.version}, func() el.Element { return v.message(m) }))
}
```

The keys must be comparable, and the appearance of the element is determined only by the key: when the appearance changes, the key must also change (for example, with a version number). Cache entries that are not used in a certain frame will be deleted. This is how `ui/markdown` caches written blocks.

`theme.Apply` causes the elements of `cx.Cache` to be rebuilt on the next visit when the theme switches. The self-built cache needs to contain `theme.Revision()`; the theme color should be read within the Render or cache construction function, and the fixed color will not be automatically converted.

### Copy to clipboard

`el.WriteClipboard(text)` is called in the callback and the current frame is written to the system clipboard.

## Focus, Keys and Disable (E1 / E2)

```go
el.Div().ID("save").Focusable(true).
    OnClick(save).
    FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
    Child(el.Text("Save"))
```

`Focusable(true)` allows the element to accept click focus and participate in Tab / Shift+Tab navigation in drawing order, sharing the native focus order with `Input` and `TextArea`. Nodes that are hidden, removed, and completely rolled out of the drawing area do not participate in navigation. Elements with OnClick are focusable by default; Focusable(false) explicitly exits the tab order.

When the focused element receives Space / Enter without modifier keys, `OnClick` is called once in the matching key release event; keys to be activated are not retained after losing focus. `OnKey(func(el.KeyEvent) bool)` Receives press and release events, bubbling from the focused element to its handler ancestor. Returning `true` stops bubbling and deactivates the default. `KeyEvent` is el's own structure, including Name, KeyPress/KeyRelease status and key.Modifiers of string type. Tab retains native navigation behavior; global shortcut keys continue to use `cx.Shortcut`.

`FocusStyle(func(*el.Style))` is a drawing style that can change the background, border color and text color without changing the size. When an ordinary element is actively focused via Tab, arrow keys, Space/Enter, or program, the 2dp Primary focus border is used by default. Mouse clicks retain actual focus and keyboard capabilities, but do not draw focus styles; synchronized `cx.Focus` in click callbacks also follows this rule. The input box always uses its own frame when focused. The text color is passed to child elements whose color is not explicitly set, and is restored after defocusing.

`cx.Focus("save")` requests focus after drawing this frame, and also supports `Input` / `TextArea` with ID. The ID should be unique within the current root; when repeated, the first drawn matching target is selected. Retains original focus when the target does not exist, is hidden, or is completely outside the viewport; `cx.Focus("")` clears focus. Can only be called in Render or its event callback.

The current `OnKey` bubbling source is a focusable ordinary element; the input box editing button is still processed by the Gio editor and does not go through this bubbling chain. The focus trap is reserved for the overlay stage. Can run `go run ./examples/components -section focus` verification interface.

`cx.Focused(id)` reads the current focus, and the query supports ordinary elements and input boxes. `Disabled(true)` prohibits clicks, hovers and keystrokes on the subtree from top to bottom, releases the current focus, prohibits program focus, and marks the Agent semantics as disabled; it can be refocused after being lifted. `DisabledStyle(func(*el.Style))` sets the disabled appearance, and the default text is Muted. Explicit disabling is handled separately from measuring without an input source, and continuous measurements do not clear the interaction state.

## Time and Reduction Animation (E3)

`cx.Now()` returns the current frame time; the animation only calculates phase from it. `cx.Animating()` requests the next frame without creating a goroutine. `el.ReducedMotion()` queries application preferences; `theme.SetReducedMotion(true)` switches within frame locks. There is currently no native system preference bridge, default is false.

`cx.After(key, duration, callback)` declares a one-time timer, the key must be comparable and unique within the current root. During the survival period, the same key is declared for each Render; if a frame is not declared, it will be canceled, regardless of the calling position and the order of declaration. Restart when duration changes. After triggering, the continuous declaration will not be executed repeatedly, and it will be restarted only if one frame is omitted and then declared again. The callback is executed through core.Call within the frame lock after the current tree is drawn, and notifies all windows to redraw.

```go
type noticeTimerKey struct { ID string }
if visible {
    cx.After(noticeTimerKey{noticeID}, 3*time.Second, func() { visible = false })
}
return el.Div().Hidden(!visible).Child(el.Text("Saved"))
```

Similar components use their own stable IDs to form keys. Do not declare `After` in the constructor of `cx.Cache`: the constructor will not be executed when the cache is hit, and timers that are not declared again will be cancelled. After should be placed on the Render path of each execution, and the element tree should be cached separately.

## Scroll state and anchoring

`ScrollX()` causes child content to stretch horizontally and clip to the viewport. Can be combined with `ScrollY()`. Supports horizontal scroll wheel and touchpad horizontal gestures; the horizontal and vertical axes consume corresponding scroll amounts respectively. `cx.ScrollStateX(id)` returns the offset, viewport width, and content width (dp); `cx.ScrollIntoViewX(id, left, right)` minimum scrolls to display the target range. The horizontal and vertical scroll bars support dragging the slider, clicking on the track to position, and switching between light and dark themes. After adding `Focusable(true)` to the scrolling container, you can use Tab to focus, and then use the direction keys, PageUp/PageDown, and Home/End to scroll; Shift+PageUp/PageDown and Shift+Home/End of the dual-axis container operate the horizontal axis. The component itself `OnKey` is processed first (such as table row selection). Example: `go run ./examples/components -section scrollable`.

`Scrollbars(el.ScrollbarAlways / el.ScrollbarHover / el.ScrollbarScrolling)` sets the display strategy of a single scroll container, shared by both axes; the default is System, which follows the platform preference. Always stays visible whenever content overflows. Hover is displayed when the pointer enters the entire viewport; Scrolling is displayed after the actual offset changes, hidden after stopping for 900ms, and continues to be displayed while the mouse drags the scroll bar. Program positioning and keyboard scrolling will also be displayed; stopping at the boundary and keeping the offset will not restart the timer. After hiding, the scroll bar click area is not retained, and the content can still receive pointer events; the scroll wheel and keyboard scrolling are not affected by the policy. `ScrollOffset`'s controlled mode still hides all scroll bars. In Hover and Scrolling modes, the scroll bar fades in 120ms when it appears and fades out in 200ms when it is hidden (directly displayed or hidden when the reduced motion is turned on). It does not receive clicks during the fade-out process; Always does not fade, and it is hidden immediately when switching to other modes from Always.

`el.ScrollbarSystem` Follow the system settings: macOS reads "Show scroll bars" (automatic/scrolling → Scrolling, always → Always, real-time updates after system settings are changed), Windows reads "Auto-hide scroll bars" (read once at startup), other platforms press Always. Containers that do not have a separate mode set use the default value set by `el.SetScrollbarDefault(mode)`. The default is System. To keep all overflowing bars visible, call `el.SetScrollbarDefault(el.ScrollbarAlways)`. Rounded scroll containers inset the tracks at their ends so the thumbs stay clear of the corner clipping; this does not change the content layout. Built-in selectable rows and horizontal component strips reserve 12dp for the 10dp scrollbar hit area and 2dp clearance. Custom `ScrollX`/`ScrollY` content can reserve the same space with padding or row margins. Showing or hiding a bar does not resize the content. `el.SystemScrollbars()` returns the system preferences read. Verified horizontal and vertical double ratio interaction, click penetration after idle hiding, continued dragging when dragging out of the viewport, and window pixels for display strategy switching.

`cx.ScrollState(id)` returns the scroll offset, visual height, and content height of the previous frame of the `ScrollY` element with ID, in dp; the three values are all 0 before the first drawing. This is used by the virtual list to decide which rows to build. `cx.ScrollIntoView(id, top, bottom)` makes the `[top, bottom]` section of the content visible with minimal scrolling, and will take effect the next time it is drawn.

In the frame where `KeepBottomOn(version)`: `version` changes, the distance to the bottom remains unchanged, and the screen will not jump when content is inserted above (such as loading earlier chat records). Increment version in the same callback where content is inserted, the usage is the same as `ScrollToEndOn`.

`Flex(w)` takes up the same remaining space as `Grow`, but is allocated by weight: `Flex(2)` gets twice as much space as `Flex(1)` or `Grow`.

## Drag

```go
el.Div().ID("track").W(el.Dp(240)).H(el.Dp(20)).OnDrag(func(e el.DragEvent) {
    v.value = clamp(e.X / e.W) // It will be called when pressing, moving and releasing
})
```

`DragEvent.Kind` is `DragStart` (pressed), `DragMove` (press and hold to move, the pointer continues to report when it moves out of the element), `DragEnd` (released or canceled). `X` and `Y` are the dp relative to the upper left corner of the element, `W` and `H` are the element dimensions, so `X/W` is the proportion in the horizontal direction. Focusable elements are focused when pressed. Disabled elements cannot be dragged.

## Overlay (E4/E5)

Call `cx.Overlay(key, layer)` in Render to declare the overlay. The key must be comparable and unique within the current root. The open state is saved by the view, and Render is declared every time during the opening period; if a certain frame is omitted, it will be closed. Do not declare overlays in Cache's constructor. The overlay declared later is on the upper layer, and Esc only requests to close the uppermost layer.

```go
if v.open {
    cx.Overlay("filters", el.Anchored("filter-button",
        el.Div().W(el.Dp(280)).P(16).Bg(theme.Surface).Child(
            el.Text("筛选条件"),
            el.Input().ID("query").Bind(&v.query),
        ),
    ).Placement(el.Bottom, el.Start).OnDismiss(func() { v.open = false }))
}
```

`Anchored(anchorID, content)` uses the anchor point position of this frame. The anchor point can be in the main tree or the overlay declared first. The direction of Placement is Bottom / Top / Left / Right, the alignment is Start / Center / End, and the default is Bottom / Start; the default of Offset is 4dp. If it cannot be placed in the specified direction and can be placed on the opposite side, flip it over and then move the position to the root; the excess part will be cropped according to the root. MatchAnchorWidth makes the overlay have the same width as the anchor point (the drop-down box has the same width as the trigger), and the content is wrapped or truncated according to this width. When the anchor point does not exist or is hidden, it is not drawn and OnDismiss is called once.

Press events outside the non-modal overlay and not on the anchor point will request closing and continue to be passed to the following elements. `.Modal()` makes the anchor overlay intercept external clicks; `el.Modal(content)` creates a modal overlay that is centered by default, with its own mask and focus constraints. The mask is read during drawing `theme.Scrim` and updated with the runtime theme switching; `.Scrim(false)` only hides the mask color and still intercepts input. During the modal period, the background does not respond to hovers and clicks, and the Agent snapshot does not list the occluded main tree and lower overlays.

`layer.Owner(id)` Binds the overlay life cycle to the enabled element in this frame tree. Can be used for modal components: return a zero-size Absolute element with ID as the owning tag, and then set the Owner to Modal; when the ancestor disables, hides or removes the tag, it will request to close. The Owner does not change the positioning, nor does it treat the modal occlusion of the overlay itself as disabled.

`.TrapFocus()` limits Tab / Shift+Tab to the overlay, focusing on the first focusable element when it appears; `cx.Focus(id)` in the same frame can specify the target within the overlay. Restores the previous focus after closing, and clears focus when the original target has been removed. A non-modal overlay without focus constraints does not move the focus. OnDismiss is executed within the frame lock and only notifies the caller to update the open status and does not save open for the caller.

`el.Modal(content).Placement(side, align)` sticks the modal content to a certain edge of the root instead of centering it. Used for side drawers: `Placement(el.Right, el.Start)` sticks to the right and aligns the top.

Esc is passed to the top layer** to set the overlay of OnDismiss**. Overlays without OnDismiss (such as notification stacks) will not swallow Esc, and the following dialog box will be closed as usual.

`cx.FocusWithin(id)` Query whether this element or any of its descendants received focus in the previous frame, used by Tooltip to display a prompt when the keyboard is focused.

`cx.Hovered(id)` queries whether the most recently processed pointer position is within the element. Elements that are disabled or obscured by a modal layer return false. Ordinary elements with IDs can also be queried without adding a click callback. HoverCard combines the Hovered results of anchors and cards.

### Hang to the root of the window

`cx.Mount(key, view)` Hang a view on the current root: after that, the root view is rendered first in each frame, and then these views are rendered in the mounting order. The returned elements are put into the root element and do not participate in its layout (overlay declaration, Absolute or hidden elements should be returned). Mounting the same key again will replace the original view. `cx.Unmount(key)` is removed and `cx.Mounted(key)` / `cx.MountedView(key)` is queried. When the root element is a leaf such as text or input box, it will automatically be wrapped with a container. Kit's `Dialog.Show`, `Sheet.Show` and `kit.WindowNotifier` are built on it, corresponding to the dialog box, drawer and notification layer that come with the GPUI root view.

Full float capability requirements `el.Root`. `el.Embed` uses the maximum constraint when embedding, and delays drawing through `op.Defer`, which is a best-effort support; its available space is not necessarily equal to the window size. Overlays do not span windows. Frames without input sources are uniformly processed as read-only frames, including measurement and parent component disabling. They reuse the real store/cache and retain the input content and scroll position; they do not distribute events, do not trigger close callbacks, do not increase or decrease the overlay life cycle, do not advance the timer, and do not clean up the state or overwrite the focus recovery record.

Verification: `go run ./examples/components -section overlay`, add `-theme dark` to check the dark color; switch the overlay, open the modal, edit the input box, and use Tab / Shift+Tab / Esc to check the focus.

## Make reusable components

The component is the function that returns `el.Element`, and the parameters are the data and callback it needs:

```go
func button(label string, onClick func()) el.Element {
    return el.Div().ID(label).Px(16).Pt(10).Pb(6).Rounded(6).
        Bg(theme.Primary).TextColor(theme.OnColor).TextSize(14).
        CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.PrimaryHover) }).
        OnClick(onClick).Child(el.Text(label))
}
```

Components that need their own state and whose state needs to be saved across frames are written as views (struct + Render).

## Local topic

`cx.Themed(palette, view)` uses another set of palettes to render a view: it uses this set of colors when rendering and drawing, so the kit components inside and the content you draw change accordingly, and the rest of the window still uses the global theme. Suitable for dark sidebars and theme previews in light-colored windows. The returned box will stretch the child elements and set a background to cover it with color.

```go
nord, _ := theme.Named("nord")
cx.Themed(nord, sidebar).Bg(nord.Bg)
```

Elements cached with `cx.Cache` are invalid according to the global theme version, and the content in local themes should not be reused and cached across themes.

## Put into window and embed Gio

- **Use el for the entire window**: `window.Options{Content: el.Root(view)}`. `Root` occupies the window, and the window no longer adds margins and outer scrolling; when the page needs to be scrolled, add `ScrollY()` to the root `Div`. Overlays (dialog boxes, menus) are declared with `cx.Overlay` and do not require `window.Options.Overlay`.
- **Put Gio code in el**: `el.Widget(w)` embeds any `core.Widget`, such as a Gio layout wrapped in `core.Func`.
- **Put el into Gio layout**: `el.Embed(view)` Get a `core.Widget` sized by content.

## What can Agent see?

No need to write additional code:

| Element | What the Agent sees |
| --- | --- |
| `Text` | `text`, the name is the word |
| `Div` with `OnClick` | `button`, the name is the text inside |
| `Input` | `textbox`, the name is `Name` or placeholder text, the value is the current content |
| `.Role("tab").Selected(true)` | `tab`, selected |
| `.Role("progressbar").Name("Import").Value("40%")` | `progressbar`, worth 40% |

Content outside the scroll container will not appear in the element list.

## Known limitations

- Layout is a subset of flexbox: supports wrap and simple grid; no `align-self`, min-content yet. Shrinkage is distributed proportionally to the content width.
- The child elements in `ScrollY` are laid out every frame (those that are invisible are not drawn). For parts with unchanged content, use `cx.Cache` to skip reconstruction and rearrangement; for more than a few hundred lines, use `kit.VirtualList` or `kit.Table` to lay out only visible lines.
- There is no package for transition animation, so you need to calculate it yourself using `Now` / `Animating`.
- The overlay is only fully supported in `el.Root`, and `el.Embed` is supported as best as possible according to embedding constraints.
- Overlays do not support cross-windows.


Available within `Decorate` `cx.LayoutSize(element)` reads the final width and height (dp) of an element in the same tree, including clipped rows. Not readable before layout. The virtual list can use `cx.ScrollTo(id, offset)` to set the vertical offset the next time it is drawn, and crop it according to the new content size. It is used to maintain the anchor point after the content changes; it does not take effect before the first draw and in read-only layout.

`cx.AfterEnabled(id, key, delay, fn)` Binds the timer to the specified element: it runs when the element is visible and not disabled; it pauses after the element or ancestor is disabled, hidden, or blocked by the modal layer, and waits for the full delay again when resuming. As with `After`, it is declared per frame, omitting the declaration will cancel it.

`cx.Enabled(id)` Query whether the recently declared element can receive input, including ancestor disabling and modal layer blocking; returns false if the ID is not found. What is queried in `Render` is the previous round of statements, which is consistent with the timing of the focus query.


`Wrap()` is arranged from left to right. If the width is insufficient, start a new line. `Gap` works on both rows and inline elements; `Grow/Flex` allocates remaining width within their respective rows, `Justify` aligns each row, and `Items` aligns elements of different heights within the same row. There will be no line wrapping when there is no width constraint.

`Grid(columns)` fills the specified number of columns row by row, and `Gap` sets the spacing between rows and columns. Columns are equal-width by default, but will first satisfy the fixed width and minimum width of child elements; if the sum of all minimum widths exceeds the available space, the minimum width will be retained and overflow. The height of each row is determined by the tallest element, and elements with automatic height are stretched to the row height by default. `ColSpan(n)` allows child elements to span columns, limited to 1 to the number of parent grid columns, and starts from a new row if it cannot fit; the minimum width of the span is allocated to the covered columns. Hidden and absolutely positioned elements occupy no grid cells. Spanning rows or named ranges is not supported. `Row`, `Col`, `Wrap`, `Grid` will switch the layout mode.

Verify: `go run ./examples/components -section layout`, adjust window width to check wrapping and grid.

`PinLeft(offset)` / `PinRight(offset)` draws the element at the offset dp of the corresponding edge of the nearest `ScrollX` viewport, retaining the layout placeholder. Fixed elements are drawn last; ordinary sibling elements are clipped between fixed elements on both sides, and the clipping constrains clicks and semantic areas at the same time. When the width of both sides exceeds the viewport, the left side takes precedence. Maintain the normal layout when there is no horizontal scrolling ancestor, used for scenarios such as frozen columns in tables.

`cx.ClickModifiers()` Returns the Shift, Ctrl, Command, etc. modifier keys for this event only in pointer click/double-click callbacks; zero outside callbacks. Keyboard events use `KeyEvent.Modifiers` directly.

`OnContextMenu(fn)` is called when the secondary key is pressed and does not swallow the primary key operation; the callback can be `ClickModifiers`. It will also be triggered by pressing a single finger on the touch screen for about 500ms and moving no more than 8dp. Letting go will no longer count as a click; it will be canceled when the finger moves, is pressed by a second finger, or is taken over by gestures such as scrolling. Therefore, the menus of Input, TextArea, Table, and Sidebar can all be opened by long pressing. Keyboard entries are declared with `OnKey`, for example Shift+F10. An anchored overlay is closed after the anchor point or its ancestor is disabled or hidden, and the input blocking of the background by the overlay itself is not considered disabled.

`DragEvent.Canceled` differentiates between cancellation and normal release when dragging ends. Components that need to commit changes after letting go should discard the staging results on cancellation.

`DragAccept(func(dx, dy float32) bool)` works with OnDrag to decide whether to take over when the initial move reaches 3dp. The parameter is the dp displacement of the pointer relative to the pressed position, not the scroll increment; the function should only judge and not modify the state. When false is returned, a Canceled DragEnd is reported and the outer gesture processor has a chance to take over; after returning true, the direction of this drag will not be re-judged. When pressed from the visible area of a clickable, draggable, input, scrollable container or embedded Widget subtree, priority is given to the child component and dragging of this layer is not started. Pass nil to resume normal dragging. Suitable for carousels that give way to gestures when dragging at the starting boundary or across axes; native drag scrolling of outer ScrollX/ScrollY primarily receives touch by Gio rules, desktop mice still use a wheel or scroll bar.

`cx.ViewportSize()` returns the available width and height (dp) of the root viewport in the Render stage, which is used to limit the height of overlays in windows such as the command panel to prevent the keyboard scroll target from being outside the window.

`cx.Countdown(id, key, duration, paused, fn)` Statement retains a one-time countdown of remaining time. Explicitly paused, pause when the element to which it belongs is invisible/disabled/modally blocked, and resume for the remaining time; change duration to restart, and omit the cancellation statement. Different from the semantics of `AfterEnabled`, which waits for the complete delay again after recovery, the notification countdown uses Countdown, and the hover prompt delay continues using AfterEnabled.

`element.Reveal(fraction)` reveals the natural height according to the ratio of 0–1, retains the complete layout of child elements, and crops the drawing and input areas; when 0, it does not occupy the height and cannot obtain focus. For collapse animations, animation timing is still driven by the component based on `cx.Now()`. NaN Press 0, limiting out-of-bounds values to 0–1.

### Mouse press monitoring

`OnMousePress(button, fn)` To observe left, right, or middle button presses, use `pointer.ButtonPrimary/Secondary/Tertiary`. Listeners override interactive child elements but do not prevent them from receiving events or adding tab stops; disabling the container disables the listener. 0 clears the monitoring, ignores illegal key values, and does not trigger when pressing multiple keys at the same time. It shares the same processor with `OnContextMenu(fn)`, which is equivalent to selecting the right button, and the last setter takes effect. Keyboard operations continue using `OnKey` or child component callbacks.

The anchor overlay can be used to draw a 6dp indicator arrow using `el.Anchored(...).Arrow(true)`, following the actual pop-up direction; Offset is measured to the tip of the arrow. Arrows take the panel's solid color background, use the theme Surface when not set, and gradients, borders, and shadows do not extend to the arrows. Modal does not display arrows.

### Explicit tab order

`TabStop(false)` skips sequential traversal and retains mouse/program focus; `TabIndex(n)` sorts in ascending order, the same value keeps the tree order, negative values skip, default 0. When explicit configuration occurs, el root handles Tab/Shift+Tab, looping within the current modal or TrapFocus float; otherwise the Gio native order is used. Input boxes also participate in sorting. Disabled, hidden and undrawn nodes are skipped.

Scope is limited to a single el root and does not span standalone Embed or native Gio controls. Calling Router.MoveFocus directly bypasses this rule and should use the normal Tab event; the control retains its operational behavior when explicitly consuming Tab.

`WrapFit()` supports line breaks like Wrap; when the automatic width is used, it is tightened according to the content of each line, which is suitable for containers such as button capsules that need to fit the content. Explicit or stretched widths still use regular row alignment and Grow allocation. Calling Wrap() restores the default behavior of filling the available line width.

### Controlled scrolling and previous frame size

`ScrollX/ScrollY` can be used with `ScrollOffset(x, y)` to use absolute dp offset; when drawing, it is limited by the content boundary, and when disabled, the specified position will still be displayed. Disables the default scroll gesture and hides the scroll bar when specifying an offset. The container can use OnDrag by itself; omit it to restore normal scrolling; non-finite values are ignored. The target location should be provided continuously by the application state.

`cx.LastSize(id)` Reads the dimensions (dp) of the last drawn element with root in Render, returning zero when first or clipped; contains geometry updates that disable frames. `cx.LayoutSize(element)` is used in Decorate to read the current layout size. After the window size changes, layouts that rely on LastSize usually need to draw another frame to converge.

`cx.PixelScale()` Returns the current number of physical pixels per dp, can be used in Render for layout-consistent pixel rounding, or 1 when not set.

### Custom scroll events

`OnScroll(xRange, yRange, fn)` Receives a ScrollEvent in dp units, scoped using `el.ScrollRange{Min: ..., Max: ...}`. Zero range does not receive this axis, and displacements outside the range are routed by Gio to the outer layer; subscroll areas take precedence. nil Remove callback, disable/hide ancestor blocking events. The range is converted to pixel scale, non-finite endpoints are treated as 0, and the endpoints are limited to ±1,000,000dp.

This event does not expose the wheel/trackpad type or gesture end phase. Can be used in controlled ScrollOffset containers; application state is updated during processing, and the offset is applied by the next frame.

### Specify content bottom alignment

Use `Items(el.ContentBottom)` for horizontal containers to align children to the bottom edge of the specified descendants. The child uses `.ContentBottom(target)` to specify the normal flow descendant in the current rendering tree; when it is not specified, the target is hidden or not in the subtree, it falls back to the bottom of the child itself. This is geometric alignment, not font baseline.

```go
body := el.Div().Child(header, content, footer).ContentBottom(content)
row := el.Div().Row().Items(el.ContentBottom).Child(avatar, body)
```

The layout uses the current frame dimensions to calculate the required space above and below the alignment line, supporting Row and Wrap for each row. The target does not accept absolutely positioned nodes; the height limit is still respected when the container has an explicit height limit. This mode is not used with vertical containers or Grids.

`el.Input().SelectOnFocus(true)` Selects all content when focused. `CaptureKeys(names...)` lets OnKey receive the specified unmodified keys before the editor, for single-line inputs and text areas alike; the callback fully owns these keys, and returning false does not hand them back. Keys with modifiers are not affected, so a `TextArea` capturing `⏎` to send still breaks lines with Shift+Enter; an input method composing text consumes Enter itself. `cx.SelectInput(id, start, end)` Sets rune selection after next Bind sync, without changing focus or text; endpoints are limited to valid range by editor, missing or disabled input ignored. The above interface is used for fast segmented editing of TimeField, and the editing behavior of Chinese rune selection, range restrictions and modifier keys has been verified.

`el.Input().TransformEdit(func(before, after el.InputEdit) el.InputEdit)` provides the text before and after editing and the rune selection at the same time during editing normalization, which is suitable for format masks to determine the deletion direction. It is mutually exclusive with Transform and takes effect after the configuration is completed; undo and redo to save the normalized text and selection, and the program Bind update will still clear the history. Regressions have been redone with masked deletes and undos.

`cx.InputSelection(id)` Returns the most recent editor text and rune selection. `cx.InputAction(id, el.InputCopy / InputCut / InputPaste / InputSelectAll)` Queue edit commands to the next draw of this input; missing/disabled input ignored, read-only rejects cut/paste, password input rejects command copy/cut. Paste accesses the asynchronous system text clipboard, and the editor continues to complete filtering and Transform; the command itself does not change the focus, and the menu caller can use cx.Focus to restore the input focus. Returned with menu clipping, focus restoration, and restricted input.

`ContainerContentSize(element)` reads the pixel dimensions of the container's content after layout in `Decorate`, calculated before the container's own min/max size limits and `Reveal`; text, input, and widget leaves return zero.

`ElementBounds(root, target)` returns the border position of the target relative to the root element after layout, including off-screen elements, without superimposing scroll offset; returns false for targets that are hidden or do not belong to the tree. `Decorate` can be combined with `ScrollTo` to achieve off-screen positioning.

### Ordinary container focus loop

`el.Div().FocusTrap(true)` Confines Tab / Shift+Tab to this subtree after focus enters it; `FocusTrap(false)` releases it. When nested, the trap ancestor closest to the current focus takes effect, and multiple parallel areas cycle individually. The order follows TabIndex/TabStop, skipping disabled, hidden and undrawn nodes; the input box participates in the loop. The modal overlay takes priority, and the background area will not intercept the overlay Tab.

It only constrains sequential navigation. Mouse clicks and `cx.Focus` can switch areas; mounting will not grab focus, and removal will not automatically restore. When it is necessary to focus when entering or return when exiting, the application calls `cx.Focus(id)`; the overlay continues to use `Layer.TrapFocus`'s automatic focus/recovery. Scope is limited to elements within the same el root, and independently embedded core.Widgets manage focus themselves.

The component library `focus` page provides opening, entering, and exiting demonstrations. Automatic testing covers parallel/nested areas, two-way loops, input boxes, TabIndex/TabStop, dynamic disabling/unlocking/removal, mouse exit, and modal overlay priority and return; native keyboard acceptance is still to be completed.

`Translate(x, y)` Press dp to move the drawing, click area and overlay anchor point of elements and subtrees without changing the original layout space or scrolling content length. It can be used for carousel tracks to reuse the same item; the visible range after shifting determines whether to draw, and the parent container cropping still takes effect.
