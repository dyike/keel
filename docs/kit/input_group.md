# Input Group

English | [简体中文](input_group.zh-CN.md)

Group text boxes, icons, units, or action buttons into the same border.

```go
query := kit.Input("").Placeholder("客户或订单号").Clearable()
search := kit.InputGroup("查询订单", query).
    Prefix(kit.Icon(kit.IconSearch)).
    Suffix(kit.Button("查询", func() { lookup(query.Value()) }))
```

The Input instance is responsible for text, filtering, read-only and callbacks, and the Input Group is responsible for borders, labels and additional content before and after. The incoming Input is only rendered through this group and cannot be rendered elsewhere at the same time. Additional content width is determined by itself; leave enough space for long buttons and narrow containers.

- `Prefix(el.View)` / `Suffix(el.View)` places additional content, pass nil to remove. Dynamic addition and deletion of additional content retains input focus and text.
- `Value` / `SetValue` / `OnChange` uses the state of the internal Input; programmatic assignment does not trigger a callback. Format, password, clear and commit configuration are still placed on the Input.
- `SetDisabled` disables the input and all attached operations, preserving the contents when restored; disabling the input itself also disables the entire group. Read-only input still allows the use of additional buttons.
- `SetError` / `Error` / `FocusID` can access Form. The group label also serves as the accessible name of the input; click on a visible label to focus the editor. Accepts Form field names when empty label.
- Tab accesses the editor and focusable additional operations in sequence, and the text editing shortcut keys are the same as Input.

Agent: The outer layer is `group`, the inner editor is `textbox` with field names; additional buttons retain their own names, roles, and disabled states.

Verify: `go run ./examples/components -section input_group`; add `-theme dark` to verify dark color.

`Addon(id, alignment, view)` can add any number of additional content in four locations:

| alignment | location |
| --- | --- |
| `InputGroupInlineStart` | Left side of editor, after Prefix |
| `InputGroupInlineEnd` | Right side of editor, before Suffix |
| `InputGroupBlockStart` | Above the entire input line |
| `InputGroupBlockEnd` | Below the entire input line |

The same position is arranged in the order of addition. The ID must be non-empty and stable; replace the content and position with the same ID, and pass nil to delete. The upper and lower areas each occupy one row, and Row/Wrap can be used to combine counts, toolbars, buttons or menus within the View. Click to focus input on text, icons, and blank spaces; interactive controls retain their own click and focus behaviors.

```go
message := kit.TextArea("").Rows(3)
composer := kit.InputGroup("备注", message).
    Addon("heading", kit.InputGroupBlockStart,
        el.ViewFunc(func(*el.Context) el.Element { return el.Text("填写订单备注") })).
    Addon("send", kit.InputGroupBlockEnd,
        kit.Button("保存", func() { save(message.Value()) }).Size(28))
```

Additional buttons directly use `kit.Button`, which can be configured with Variant, Size, Icon, Name, Loading and disabled state; compact buttons can be set to `.Variant(kit.ButtonGhost).Size(24)`. Icon button requires `Name`. Menus, Popovers and Tooltips can also be combined as additional Views, inheriting their respective interfaces.

The TextArea shares these locations with single-line input; the top and bottom panels grow with content, and the number of lines of text is controlled by the internal TextArea. `Rows` sets the minimum height; to increase the height with the content and set an upper limit, use `AutoGrow(minRows, maxRows)` of TextArea, see [Input](input.md). There is only one layer of field borders outside the input and additional content, and the focus box covers the entire group. Read-only allows additional actions, input-disabled, group-disabled, and ancestor-disabled disable all additional actions.
