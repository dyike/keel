# GooeyDock

English | [简体中文](gooey_dock.zh-CN.md)

A native proximity dock based on [Ruixen Gooey Dock](https://ruixen.com/docs/components/gooey-dock). Keep the view between renders.

```go
dock := kit.GooeyDock(
    kit.GooeyDockItem{ID: "home", Label: "Home", Icon: kit.IconHome, Active: true, Action: openHome},
    kit.GooeyDockItem{ID: "search", Label: "Search", Icon: kit.IconSearch, Action: openSearch},
)
// Inside Render:
return dock.Render(cx)
```

The reference's geometry is fixed: 40→58 dp cells, 14 dp lift, 150 dp cosine influence, 20 dp icons scaling to 1.2, and spring mass/stiffness/damping 0.1/200/14. The bar uses 2 dp gaps, 8/12/10 dp padding, an 18 dp radius and a 1 dp border. Hover labels enter and exit over 150 ms. Ink follows the light/dark theme. An extra 44 dp above the bar reserves room for lifted cells and labels. Allow enough horizontal space for the expanded dock; it does not shrink icons to fit.

`IconView` accepts a retained SVG icon in place of `Icon`. The Components demo uses the seven Lucide icons from the reference, with Home active. `SetItems` replaces actions and resets springs; IDs must be nonempty and unique. `SetActive` changes dots without invoking actions. `SetDisabled` and item `Disabled` suppress interaction. Arrow keys, Home and End move focus; Enter/Space activate. Reduced motion applies state changes immediately, and settled springs request no frames.

`OnHover(func(id string))` observes pointer entry into a different item, throttled to 80 ms, for application audio integration. This implementation does **not** play the reference's Web Audio tick by itself. The visual parameters follow the source; browser-versus-native pixel and timing parity has not been established. Font rasterization and platform text scaling can differ.

```sh
go run ./examples/components -section gooey_dock
```

Implemented entirely in Go without new cgo, JavaScript or WebView dependencies.

`Surface(color)` supplies the opaque color behind the dock (default `theme.Bg`). Set it when placing the dock on a different background: CSS-style sRGB colors are precomposited against this surface because Gio blends translucent colors in linear light.
