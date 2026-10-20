# Button

English | [简体中文](button.zh-CN.md)

el-based action button that supports mouse, keyboard, icon and loading state. Default Primary appearance, 32dp height, 6dp rounded corners; color reads current theme in Render.

```go
save := kit.Button("Save", saveOrder).Icon(kit.IconPlus)
deleteButton := kit.Button("Delete", deleteOrder).Variant(kit.ButtonDanger)
```

Variant Accepts ButtonVariant: ButtonPrimary (zero value), ButtonSecondary, ButtonGhost, ButtonDanger, ButtonLink, ButtonText, ButtonSuccess, ButtonWarning, ButtonInfo. Only Variant is used to configure the appearance, and shortcut entries such as Danger are not provided. Size(float32) specifies the height dp, 28 / 32 / 40 is recommended; non-positive numbers and non-finite numbers are ignored. Narrow containers are limited to a single row, respecting the available width.

Outline(true) is a compatibility name for the borderless tinted style: the background is transparent, the text uses the variant color, hover uses `Subtle`, and active uses `SubtleHover`. Compact(true) reduces horizontal padding while retaining height. Link has no horizontal padding by default, and Text retains padding. Buttons with only icons and no copy are square, and use Name to set the accessible name.

Content(el.View) replaces the visible text and icon, and returns it by passing nil. The content should be a display element, without nested buttons or input boxes; the copy (or Name) when constructed should still be used as an accessible name. Hide the drawing of content while loading and preserve the layout, showing a progress ring in the center.

Appearance(func(ButtonAppearance) ButtonAppearance) modifies the current variant coloring every Render, including normal, hover, pressed, and keyboard-focus colors. `ButtonAppearance.Border` is deprecated and ignored. Pass nil to restore the theme color; custom color matching does not use PrimaryGradient. Success, warning, and information variants choose black or white text according to the background color. Disabled buttons remain borderless and use the theme disabled colors.

Icon(IconName) configures the front icon. Loading(bool) sets the loading state, SetLoading can be updated in the callback. Reuse the spinner's animation when loading: replace the icon when there is an icon, retain the measured size of the copy when there is no icon and draw the spinner centered on it, so the button size remains unchanged before and after switching. Shows a static progress ring when the Reduce Animation setting is on.

SetText updates the copy, SetDisabled updates the disabled state, neither triggers a click callback. Background updates must be placed in core.Update. Maintain focus and tab order when loading, ignore only activations; do not show hover, press appearance, use default cursor. When disabled, it does not participate in Tab navigation; the Disabled of the parent element also limits button interaction, and is processed as disabled when disabled and loaded at the same time. Ghost, Link, Text, and Outline maintain a transparent background when disabled.

Tab / Shift+Tab focuses; Space / Enter activates once when released. Hover, Active and FocusStyle provide hover, press and focus appearance respectively without changing the layout. Buttons draw no persistent border in default, hover, active, selected, disabled, or loading states. Keyboard/programmatic focus alone draws the inset focus ring; pointer focus does not.

The Agent role is button, and the name is text; when loading, the value is loading, and disabled is false; disabled is true when only itself or the parent element is disabled. The icon and spinner are button internal decorations and are not included in the snapshot separately.

Verify: `go run ./examples/components -section button`, add `-theme dark` to check the dark theme. Examples cover nine variants, tinted/compact configurations, custom content/colorways, three sizes, icons, disabled, loading toggles, and narrow containers.

## Selected state

`Selected(true)` / `SetSelected(on)` displays the button as selected, such as the current view in the toolbar and enabled filtering; the Agent sees `selected` as true. Solid buttons darken; secondary, lightweight, text, and tinted buttons use the `Highlight` background with `PrimaryText`. Clicking does not toggle automatically; set it in OnClick. For mutually exclusive options, use [ToggleGroup](toggle_group.md).

Button group see [ButtonGroup](button_group.md).
