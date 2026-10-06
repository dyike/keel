# Kbd

English | [简体中文](kbd.zh-CN.md)

Kbd displays shortcut key caps, does not register shortcut keys, and does not participate in Tab navigation. The font size is inherited from the parent element, the text color uses theme.Muted, and the border uses theme.Border. Switching the theme at runtime takes effect immediately.

```go
el.Div().Row().Items(el.Center).TextSize(16).Child(
    el.Text("Command palette"),
    kit.Kbd("mod+shift+p").Render(cx),
    kit.Kbd("enter").Plain().Render(cx),
)
```

`Kbd(shortcut string)` returns `*KbdView`. `Plain()` hides borders and preserves spacing. `Size(sp)` sets the font size and adjusts padding proportionally; 0 restores font size inheritance and default padding, negative values, NaN, infinity, and values greater than 128 are ignored. `Style(func(*el.TextEl))` adjusts the background, text/border color, rounded corners, spacing, etc. after the default style of each frame; nil returns to default. The callback does not retain element references, and the Agent name always uses the original shortcut. Default border 1dp, rounded corners 4dp, horizontal padding 6dp, vertical padding 3dp. Keep a single line in narrow containers and truncate.

Formatting is provided by `core.ShortcutLabel(s, goos string) string`, which is shared by kit and window shortcut keys. The syntax is the same as core.ParseShortcut; mod is displayed as Command on macOS and Ctrl on Windows/Linux. Unparsable copy is displayed as is, making it easier to use custom key names. This function only formats and does not query action bindings.

## Show action-bound keys

Shortcut keys can be bound to named actions, see [Elements and Views · Action and Key Table](../el.md#action-and-key-table). `kit.KbdFor("editor.save")` displays the first key currently bound to this action; after the user changes the key (`core.Bind`, `core.LoadKeymap`), the next frame is automatically updated, and nothing is displayed when there is no binding.

```go
kit.KbdFor("editor.save").Render(cx)                   // keycap
kit.Menu().ActionItem("Save", "editor.save", save)    // The same key appears on the right side of the menu
```

The Agent role is text, and the name is the original shortcut passed in; the symbol corresponding to the platform is displayed on the screen, and no additional duplicate snapshot nodes are generated.

Verification entry: `go run ./examples/components -section kbd`, add `-theme dark` to check the dark color. Examples cover 12/16/24sp, Plain, platform symbols, Chinese and English numbers, and narrow containers.

### Show elements in context

`kit.KbdFor("format").At("code")` resolves the binding by the `KeyContext` path of the element with ID `code`, using the same rules it uses for handling keystrokes itself (including the conditional expression for `core.BindIn`). For example, the toolbar button prompts should show the actual shortcut keys in the editor. Display nothing when the element does not exist, is hidden, or is disabled.
