# Primitives

English | [简体中文](base.zh-CN.md)

The component appearance of `kit` is fixed. When you want a component that behaves the same as kit but looks completely different, you don't have to write the keyboard logic from scratch: `ui/base` provides the behavior of the component, and `el` takes care of the appearance. The combination of the two is equivalent to styleless component libraries such as Radix and Headless UI.

`base` does not draw anything and does not rely on any Keel modules. It is just ordinary states and functions that can be tested independently. Kit's own List, Tree, Menu, Select, Command, Sidebar, and Table are all built on it, so the keyboard feel of custom components and kit components is the same.

## List: keyboard navigation

```go
nav := base.List{Count: len(items), Disabled: func(i int) bool { return items[i].Off }}
next, ok := nav.Key(e.Name, current) // ↑ ↓ Home End PageUp PageDown
```

- Skips disabled items; stays put when there is nothing to go to.
- When there is currently no selected item (`-1`), ↓ goes to the first item, and ↑ goes to the last item.
- `Wrap: true` is connected end to end, and this is used for menus. `Page` is the page turning step size, the default is 10.
- `First`, `Last`, `Next(i, dir)` are available individually, such as focusing on the first available item when opening.

## Typeahead: jump by first letter

```go
var find base.Typeahead

if s, ok := base.Text(e.Name, e.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt) != 0); ok {
    if i, found := find.Find(time.Now(), s, current, nav, func(i int) string { return items[i].Label }); found {
        current = i
    }
    return true
}
```

- Continuously input `r`, `e` to jump to "Red"; pause for more than 1 second (`TypeaheadPause`) to restart.
- Press the same letter repeatedly to cycle through items starting with it.
- Case insensitive, skipping disabled items.
- `Text` Converts key names into input text, excluding function keys named with a single symbol such as ↑ and ⏎, as well as shortcut keys with Ctrl, Cmd, and Alt.
- `Reset` is called every time the overlay is opened, so that the letters entered last time will not affect this time.

Kit's List, Tree, Menu, and Select already support jumping by first letter.

## Selection: Multiple selection

```go
var sel base.Selection[string] // Save by ID, sort, filter and select without losing

sel.Click(ids, i, mods.Contain(key.ModShift), mods.Contain(key.ModShortcut), disabled)
sel.Has(id)
sel.In(ids)      // List selected IDs in ids order
sel.Indexes(ids) // The subscript of the selected item
sel.Keep(func(id string) bool { return exists[id] }) // Clear the non-existent data after changing it
```

`Click` is processed according to system convention:

| Operation | Result |
| --- | --- |
| Click | to select just this item and set it as the start of the range |
| Cmd click (Ctrl on other platforms) | Add or remove this item |
| Shift click | Select this item from the starting point, skipping disabled items |
| Cmd + Shift click | Add this range to the existing selection |

## Disclosure: open/close

```go
d := base.Disclosure{OnChange: func(open bool) { ... }}
d.Toggle()     // User operation: OnChange only when the status changes
d.Set(false)   // Program settings: Do not adjust OnChange
d.SetDisabled(true) // When disabled, it will be closed at the same time and cannot be opened later.
```

The distinction between "user-modified" and "program-modified" is consistent with the agreement of the kit component `OnChange`: the program calls `SetValue` and does not trigger a callback.

## Example: Custom appearance of a radio-select list

```go
type swatches struct {
    colors []string
    nav    base.List
    find   base.Typeahead
    at     int
}

func (s *swatches) Render(cx *el.Context) el.Element {
    s.nav.Count = len(s.colors)
    box := el.Div().ID("swatches").Role("listbox").Name("颜色").Focusable(true).Row().Gap(8).
        OnKey(func(e el.KeyEvent) bool {
            if e.State != el.KeyPress {
                return true
            }
            if t, ok := base.Text(e.Name, false); ok {
                s.at, _ = s.find.Find(time.Now(), t, s.at, s.nav, func(i int) string { return s.colors[i] })
                return true
            }
            name := e.Name
            switch key.Name(name) { // Horizontal arrangement: ← → treated as ↑ ↓
            case key.NameLeftArrow:
                name = string(key.NameUpArrow)
            case key.NameRightArrow:
                name = string(key.NameDownArrow)
            }
            var ok bool
            s.at, ok = s.nav.Key(name, s.at)
            return ok
        })
    for i, c := range s.colors {
        box.Child(el.Div().Role("option").Name(c).Selected(i == s.at).Size(el.Dp(28)).Rounded(theme.RadiusFull).
            Border(2, theme.Border).When(i == s.at, func(d *el.DivEl) { d.Border(2, theme.Primary) }).
            OnClick(func() { s.at = i }))
    }
    return box
}
```

The appearance is all determined by `el`; the keyboard and jump rules are consistent with the kit list. The headless page of the component library example (`examples/components/headless.go`) has a fully runnable version and also includes a multi-select list made with `Selection`.
