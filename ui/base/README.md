# ui/base

English | [简体中文](README.zh-CN.md)

Component behavior, without appearance: keyboard navigation of lists, jump by first letter (typeahead), multi-selection with shift range, open/closed state. Doesn't draw anything, just normal Go state and functions.

- **Dependencies**: Only relies on Gio's key name (`third_party/gio/io/key`) and does not rely on any Keel module.
- **Used by it**: `kit`'s List, Tree, Menu, Select, Command, Sidebar, and Table use it to handle keyboard and selection; applications can also use it with `el` to write components with completely customized appearance.

| File | Responsibility |
| --- | --- |
| `list.go` | `List`: ↑ ↓ Home End PageUp PageDown, skip disabled items, loopable |
| `typeahead.go` | `Typeahead`: Enter letters continuously to jump to the item starting with it, and loop when the same letter is pressed repeatedly |
| `selection.go` | `Selection[K]`: Multi-selection saved by key, click, Cmd/Ctrl switching, Shift range |
| `disclosure.go` | `Disclosure`: Disableable on/off state, distinguishing user operations and program settings |

Documentation: [Unstyled Base Layer](../../docs/base.md)
