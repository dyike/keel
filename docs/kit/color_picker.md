# ColorPicker

English | [简体中文](color_picker.zh-CN.md)

Color picker.

```go
picker := kit.ColorPicker().Alpha().Swatches(presets...).OnChange(func(c color.NRGBA) { … })
picker.SetValue(theme.Primary)
```

- Drag in the box to select saturation and brightness, select hue in the hue bar, and add an opacity bar after `Alpha()`. Blocks and individual bars can be focused, and can be fine-tuned using the arrow keys.
- The hexadecimal input box accepts `#RGB`, `#RRGGBB`, and `#RRGGBBAA`. It takes effect when entering or losing focus. It restores the original value when the input is illegal.
- `Swatches(...)` displays a preset color block below, wraps according to the available width, and copies the preset slices.
- Press HSV to save the color internally: After dragging the color to gray or black, the hue will not be lost, and it will still be the original hue when you drag it back.
- `Popup(true)` uses built-in trigger buttons and anchored elastic layers; it is inline by default, but you can still combine `kit.Popover` by yourself.

Agent: container role `group`, the name is the hexadecimal value of the current color; the role of the square and each bar is `slider` (named "Saturation and Brightness", "Hue", "Opacity"); the input box is named HEX; the default color block is a button named with a hexadecimal value.

Verify: `go run ./examples/components -section color_picker`, add `-theme dark` to check the dark theme.

`SetDisabled(true)` Disables swatches, sliders, presets, and HEX input, and cancels unsubmitted HEX drafts. Drafts will not be submitted if the parent container is out of focus. `SetValue` can still update colors while disabled and does not call `OnChange`.

The default width is 240dp, which will shrink to the available width when placed in a narrow container. The slider indicator ring is positioned according to the actual layout, and the black and white double strokes ensure legibility in both light and dark areas. The hue and transparency bars have a hit height of 24dp, and the focus points have independent borders.

Arrow keys for fine-tuning, Shift + arrow keys accelerate ten times; the hue/transparency bar supports ten-step adjustments from Home/End to the endpoint and PageUp/PageDown. The selected state of the color block is expressed by the check icon and Agent `selected` at the same time. The icon is selected as black or white according to the synthesized background. Repeatedly clicking the current color block will not trigger the callback repeatedly. Dragging to cancel retains the last valid color without overwriting it with the coordinates of the cancel event.

`Label(text)` Displays a label above a selector or trigger and serves as the accessible name of a built-in button. The button uses the current HEX when there is no label. `Icon(name)` replaces the color block in the trigger, `IconNone` restores the color block; the button always displays the current HEX. Color modification continues to use the original OnChange.

`Size(ColorPickerSizeXSmall/Small/Medium/Large)` simultaneously adjusts the elastic layer/inline panel width, color palette height, HEX field and trigger button size; the panel width is 192/216/240/280dp, and the button height is 24/28/32/40dp. By default, Medium retains the original inline layout, and illegal gears are ignored.

`SetOpen(bool)` program sets the pop-up layer, and `IsOpen()` queries it; it is only valid in Popup mode and does not trigger color callback. Focus moves to the swatch when the user opens it; Esc or clicks outside to close, and focus returns to the trigger button. Turn off or disable canceling uncommitted HEX drafts, committed colors are retained. Switching back to inline also turns off the elastic layer. Keyboard, disabling, draft cancellation, size switching and value retention have passed the 1×/2× automatic test; the feel of the real device elastic layer is still to be accepted.

## Color format

`Format(kit.ColorRGB)` Select the display and input format: `ColorHex` (default, `#2563EB`), `ColorRGB` (R, G, B three input boxes of 0–255), `ColorHSL` (H 0–360, S, L percentage). When `Alpha()` is turned on, there is an additional opacity percentage input box for RGB and HSL. The HEX / RGB / HSL switch button in the panel allows users to change formats at any time, and `CurrentFormat()` reads back the current format.

- The input box takes effect when you press Enter or loses focus; it restores the original value when it is not a legal value, and numbers outside the range will be cut off at the boundary.
- The text on the pop-up trigger follows the format, such as `rgb(37, 99, 235)`, `hsla(221, 83%, 53%, 0.8)`; `Text()` returns the same string.
- `kit.FormatColor(c, format, alpha)` can be used alone to format colors.
- The hue entered in HSL is preserved even if the color itself is already gray at 0 saturation.

## Elastic layer position

`Placement(el.Top, el.End)` specifies on which side of the trigger the elastic layer opens and how to align it, the same as Popover. Only valid for `Popup(true)`'s color picker.
