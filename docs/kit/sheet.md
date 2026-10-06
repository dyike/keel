# Sheet

English | [简体中文](sheet.zh-CN.md)

A modal panel attached to one side of the window, used for details, settings, and long forms.

```go
details := kit.Sheet(el.Right, "订单详情").Body(view).Size(400)
details.SetValue(true)
```

- Directions: `el.Right`, `el.Left`, `el.Top`, `el.Bottom`.
- Size: `Size(dp)` is the width in the left and right directions and the height in the up and down directions. The default is 360; it will not exceed the window.
- Open by sliding in from the edge with `kit.SheetSlide` (200ms). Appears directly at the final position when reduced motion is turned on, which is the default in automation mode.
- The closing method is the same as Dialog: Esc, click on the mask, close button on the title bar, and call `OnClose(fn)` when closing.
- The content area can be scrolled. The focus is limited to the panel and returns to its original position after closing.
- `Value()` / `SetValue(bool)` reads or sets whether to open, `SetTitle` modifies the title. You can also call `sheet.Show(cx)` in the callback. The drawer is hung directly to the root of the window without being placed in the view tree. It will be automatically removed after closing; the same instance should not be rendered at the same time. `el.Root` REQUIRED.

Agent: The role is `dialog`, the name is the title; the name of the close button is "Close".

Verification: `go run ./examples/components -section sheet`.

The body of Sheet can contain overlays such as Menu and Popover; the parent layer is registered first, Esc closes from the innermost layer, and finally the Sheet is closed. `SetDisabled(true)` closes and prevents reopening. Disabling or hiding ancestors will also close the modal layer. Restoring enablement will not automatically reopen. The user close callback is executed at most once.

The sliding distance uses the actual width and height constrained by the window. Ignored when the size is NaN or infinite; reopening immediately after closing will restart the animation, and does not require the closed state to be rendered once in the middle. The example "Order Action" validates the nested menu.

`Footer(views...)` sets an independent footer, copies incoming slices, supports multiple operation buttons and line breaks; the text continues to scroll independently, and the footer does not scroll with the text. When there is no text, the footer will be at the bottom of the panel, empty parameters will clear the footer, and nil sub-items will be ignored. The footer is suitable for small operations and requires sufficient panel height for the title and footer.

`Keyboard(bool)`, `Overlay(bool)`, `OverlayClosable(bool)`, and `CloseButton(bool)` respectively control the Esc, mask color, external click close and title close buttons. They are all turned on by default. Hidden masks still maintain modal blocking and focus constraints; close buttons are not subject to Keyboard/OverlayClosable constraints. The configuration can be updated during opening, the procedure SetValue(false) is always available.

`MarginTop(dp)` reserves space at the top of the panel, e.g. `MarginTop(32)` avoids the title bar. All four directions are laid out within the remaining window area; the top panel slides in from the lower edge of the reserved area, the bottom panel remains attached to the bottom, and the left and right panels shorten their height. The drawing and clicking areas are cropped simultaneously, and the animation will not cover the reserved area. Default is 0, negative numbers and non-finite values are ignored; when the window height is exceeded, the panel is completely cropped and can still be closed by pressing Esc. Masks and modal blocking continue to cover the entire window, and click on the top reserved area to handle it as OverlayClosable.

`PanelStyle(func(*el.DivEl))` Configure panel color, text color, border, rounded corners, shadow, padding and content spacing after default appearance; `PanelStyle(nil)` restore default. The callback receives new elements every frame and should not save references. The panel identity and Size/window constraints are set after the callback, and the body and actions continue to be provided through the Body/Footer. Subcomponents with explicitly set colors retain their own colors, and unset text inherits the panel color.

```go
details.PanelStyle(func(panel *el.DivEl) {
    panel.Bg(theme.Surface).Border(1, theme.Border).P(24).Gap(theme.SpaceLg)
})
```

`Resizable(bool)` Controls the 6dp adjustment knob on the inside edge, on by default. Drag the width of the left and right panels, drag the height of the upper and lower panels; the handle supports Tab focus, the arrow keys move 16dp along the corresponding axis, and Home/End is adjusted to the minimum/maximum size. The user adjustment range is from 80dp to the available window size; when the window is less than 80dp, it is limited to the actual space, and MarginTop is deducted vertically.

`PanelSize()` returns the requested size; `OnResize(func(float32))` notifies the user of the adjustment result, and the program call Size is not triggered. Adjustments start from the actual size on the screen; canceling pointer dragging returns to the starting size (still constrained by the current window). Turning off, disabling resizing, or resizing the program ends the current drag. Apps can save dimensions in callbacks; dimensions are retained when reopened. By default, adding a new handle will add a keyboard dock. Resizable(false) can be used to restore the fixed size behavior.
