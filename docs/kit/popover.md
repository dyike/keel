# Popover

English | [简体中文](popover.zh-CN.md)

Displays a non-modal panel next to the trigger element for things like filters, short forms, etc.

```go
filters := kit.Popover(form).Width(280).Offset(12)
filters.Trigger(kit.Button("Filters", filters.Toggle).Variant(kit.ButtonSecondary))
```

- The triggering element is responsible for opening itself: pass `Toggle` to its click callback. Popover only wraps a layer of non-interactive anchor points outside the triggering element, and there are no additional tab stops. The keyboard behavior is completely determined by the triggering element.
- The following actions close the panel: clicking the triggering element again, pressing the mouse outside the panel and triggering element, or pressing Esc. That click outside the panel will continue to be passed to the following elements.
- The panel does not move focus. If there is an input box in the content, the user needs to click in it or use Tab to enter it.
- `Value()` returns whether it is open; `SetValue(bool)` is opened or closed by program and does not trigger a callback; `OnChange(fn)` is only called when the user operates.
- `Placement(side, align)` sets the position of the panel relative to the trigger element. The default is `el.Bottom, el.Start`; it will automatically flip to the opposite side if it cannot be placed.
- Requires `el.Root` (see [el · overlay](../el.md#overlay-e4e5)).

Agent: The role of the panel is `dialog`, and the elements inside are listed separately; the rest of the page remains visible.

Verify: `go run ./examples/components -section popover`, add `-theme dark` to check the dark theme.

The width and height of the panel are constrained by the window, and long content can be scrolled. `Width(0)` Restores content-width; negative numbers and non-finite widths are ignored. `SetDisabled(true)` Closes the panel and disables the trigger area. Program opening and toggle will not bypass the disabling. The overlay will close when the parent container is disabled or the anchor disappears.

You can continue to place floating-layer components such as Menu and Select in the content. The parent panel is registered first, and the internal overlays are displayed at the top; Esc closes each layer from the innermost layer. The sample "Select Preset" verifies this path.

`Offset(dp)` sets the spacing between the trigger element and the panel, the default is 4dp; 0 means fit, negative numbers allow overlap, and NaN/infinity values are ignored. Modifications when opened will be repositioned on the next frame and OnChange will not be triggered. Still flipping or constraining position near window edge, actual spacing may be limited by available space.

`Appearance(false)` Removes default background, borders, rounded corners, shadow, and padding; on by default. Positioning, scrolling, and closing behaviors are retained.

```go
filters.Appearance(false).PanelStyle(func(panel *el.DivEl) {
    panel.Bg(theme.Surface).Border(1, theme.Primary).Rounded(theme.RadiusMd).P(12)
})
```

`PanelStyle` Called every frame after default appearance and Width, nil Remove custom styles; do not keep element references. The panel ID, dialog role, window size limit and scrolling are set last by the component. Modifying the appearance during opening retains content status and focus; the theme color read in the callback can be updated as the theme switches.

`RightClick(true)` allows right-clicking in the trigger area to directly call Toggle, and right-clicking again to close it; false restores the default behavior. The left click and keyboard still execute the trigger element's own callback, without adding additional tab stops. The triggering element should no longer register the right-click callback for calling Toggle to avoid executing it twice. Right-clicking when disabling itself or its parent container will not open the panel.

```go
info := kit.Popover(details).RightClick(true)
// Left-click and keyboard can also be opened as an alternative entrance for right-click operations.
info.Trigger(kit.Button("Details", info.Toggle))
```

It is only hoped that when the right mouse button is turned on, different left-click callbacks can be provided for the trigger button; the keyboard replacement entrance is provided by the application. Other mouse buttons are configured using MouseButton.

`MouseButton(pointer.ButtonPrimary/Secondary/Tertiary)` Select to automatically switch panels when the left, right or middle button is pressed. The default is 0, which is opened by the trigger element itself; `MouseButton(0)` restores this mode. Illegal values are ignored, and pressing multiple keys at the same time does not trigger. `RightClick(true/false)` is equivalent to selecting right-click/return to manual mode respectively, and the last setting takes effect.

```go
info.MouseButton(pointer.ButtonTertiary) // github.com/dyike/keel/third_party/gio/io/pointer
```

Listeners will not swallow events that trigger the element itself. When selecting the left button to open automatically, do not bind Toggle to the button left button callback, otherwise it will switch once when pressed and released. Automatic listening does not add tab stops, and the keyboard replacement entrance is still provided by the trigger element.

`Arrow(true)` Shows an arrow pointing to the trigger, off by default. Arrows are 6dp deep, panels have a corresponding 6dp spacing, and Offset is measured to the tip. The direction flips to follow the actual positioning, aligning with Start/Center/End along the edges of the panel and avoiding rounded corners; narrow panels shrink arrows. Arrows may be clipped at the edge of the window. The arrows scroll independently of the content, and clicking on the arrows does not close the panel.

The arrow uses the solid color background set by PanelStyle, which takes the current theme Surface when not set; no shadows or borders are drawn separately, and gradient backgrounds are not sampled. Appearance(false) does not automatically turn off the arrow. Arrow(false) can be called independently.
