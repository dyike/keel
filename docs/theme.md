# Theme

English | [简体中文](theme.zh-CN.md)

`ui/theme` tube color, size scale and font. The component reads from the theme every time it renders, so after switching the theme, all windows will be changed to a new look in the next frame without rebuilding the view.

## Switch themes

```go
theme.Use("nord")          // Apply registered themes by name and remember names
theme.CurrentName()        // "nord"
theme.Names()              // All registered names can be used directly in the theme selection box
theme.Apply(myPalette)     // Apply an unnamed palette
```

`Use` and `Apply` should be called before opening the window, or called in the interface callback (holding the frame lock). Other goroutines use `core.Update(func() { theme.Use("dark") })`.

Built-in themes: `light`, `dark`, `nord`, `paper`, `solarized-dark`, `high-contrast`, and `aurora` with gradient.

## Palette and caching

`Light()`, `Dark()`, `Current()` returns a copy of the palette. Modify the copy when customizing and replace all colors with `Apply`; zero value fields are also applied.

```go
p := theme.Light()
p.Primary = theme.RGB(0x15803d)
theme.Apply(p)
```

`Apply` increments `theme.Revision()`, invalidates `cx.Cache` and requests a redraw. Your own rendering cache should also include the version number in the key. Colors should be read during Render; already constructed static elements, custom fixed colors, and Markdown independent colors need to be updated by the application itself. Modifying package-level color variables directly does not synchronize the Material, refresh the cache, or request a redraw.

| color | light default value | where to use |
| --- | --- | --- |
| `Bg` | `#f5f6f8` | Window background |
| `Surface` | `#ffffff` | Card and input box background color |
| `Border` | `#e3e5e8` | Borders, dividers |
| `Text` / `Muted` | `#1f2328` / `#6b7280` | Main text / secondary text, placeholder text |
| `Primary` / `PrimaryHover` / `PrimaryText` | `#2563eb` / `#1d4ed8` / `#1d4ed8` | Main button, focus border/hover/link and other blue text |
| `Danger` / `DangerHover` / `DangerText` | `#dc2626` / `#b91c1c` / `#b91c1c` | Danger Button / Hover / Wrong Text |
| `Success` / `Warning` / `Info` | `#15803d` / `#a16207` / `#0369a1` | Status text, icon |
| `Subtle` / `SubtleHover` | `#eceef1` / `#e2e5e9` | Secondary button, hover background color |
| `OnColor` | `#ffffff` | Text on solid primary color, dangerous background |
| `Highlight` | `#dbeafe` | Selected rows, selected items |
| `Scrim` | 40% black | Mask behind modal overlay |
| `CodeBg` / `CodeText` | `#f0f1f3` / `#1f2328` | CODE BLOCK |
| `Chart` | 8 classified colors | Chart series colors, used in order; one set of light colors and one set of dark colors, all passed color vision deficiency verification |

`OnColor` is for use on Primary / Danger solid backgrounds and is not guaranteed to fit all status colored backgrounds. Success, warning and prompt colors can be used for text and icons.

## Theme files

The theme file is JSON: choose a basic theme and only write the colors you want to change. The color name is the field name of `Palette`, which is not case-sensitive. The value can be written as `#rgb`, `#rgba`, `#rrggbb` or `#rrggbbaa`.

```json
{
  "name": "forest",
  "base": "light",
  "colors": {
    "bg": "#e8f3ea",
    "primary": "#2f7d4a",
    "primaryHover": "#26693d",
    "chart": ["#2f7d4a", "#c2410c"]
  }
}
```

```go
name, palette, err := theme.ParseTheme(data)
theme.Register(name, palette)
```

`chart` has up to 8 colors, and the basic theme will be used if not specified. `ParseTheme` returns an error when the wrong color name or color value is written.

## Gradient

`bgGradient` gives the window background, and `primaryGradient` gives the main button, progress bar, and user's chat bubble:

```json
"bgGradient":      {"from": "#10142b", "to": "#1f1843", "angle": 120},
"primaryGradient": {"from": "#7c5cff", "to": "#2fb8ff", "angle": 0}
```

`angle` is measured in degrees: 0 from left to right, 90 from top to bottom. The main button is dimmed when hovered and pressed, and the gradient does not disappear.

The gradient is just the appearance. `Bg` and `Primary` are still solid colors in these two places. The selected border, focus box, link and other places use solid colors. Therefore, you need to set `bg` and `primary` to a certain color in the gradient to look harmonious. When there is no gradient (zero value), the original solid color is not affected at all.

Elements drawn by yourself can also be used: `el.Div().BgGradient(theme.Gradient{From: a, To: b, Angle: 90})`. Passing a zero value will not change anything, so you can directly pass `theme.PrimaryGradient`, which will be a solid color without a gradient theme. Adjusting `Bg` later will replace the gradient.

## Watch theme files

When the designer changes the theme file, the effect can be seen without restarting the program:

```go
w, err := theme.WatchThemes("./themes", time.Second, func(names []string, err error) {
    if err != nil {
        log.Print(err) // A file was written incorrectly; other files are loaded as usual
    }
})
defer w.Stop()
theme.Use("forest")
```

- All `*.json`s in the directory are immediately registered on startup, so you can `Use` them immediately.
- Then check the modification time of the file at intervals. Changed files will be re-registered in the next frame (when the frame lock is held); if the theme being used (`CurrentName`) changes, it will be re-applied immediately.
- Using polling instead of file system events requires no additional dependencies and the behavior is consistent across platforms.
- Deleting the file will not unregister the registered theme.
- The component library example can be tried like this: `go run ./examples/components -themes ./mythemes -theme forest`.

## Spacing and sizing

The component takes its value from the scale and does not write numbers by itself. The same purpose looks the same everywhere, and if you want to change the design, you only need to change one place.

| Scale | Constant |
| --- | --- |
| Spacing (dp) | `SpaceXxs` 2, `SpaceXs` 4, `SpaceSm` 6, `SpaceMd` 8, `SpaceLg` 12, `SpaceXl` 16, `Space2xl` 24, `Space3xl` 32 |
| Rounded corners (dp) | `RadiusSm` 4, `RadiusMd` 6, `RadiusLg` 8, `RadiusXl` 12, `RadiusFull` |
| Font size (sp) | `TextXs` 11, `TextSm` 12, `TextMd` 13, `TextControl` 14, `TextBody` 15, `TextLg` 17, `TextXl` 20, `TextHeading` 22 |
| Shades | `ElevationSm`, `ElevationMd`, `ElevationLg`, color `theme.Shadow` |

Usage of spacing: `SpaceSm` between icons and text, `SpaceMd` between controls in a row, `SpaceLg` between control inner margins and form fields, `SpaceXl` for card inner margins, and `Space2xl` for dialog box and page inner margins.

```go
el.Div().Row().Gap(theme.SpaceMd).P(theme.SpaceXl).Rounded(theme.RadiusLg)
```

The spacing of the kit has been changed to this set of scales; the remaining 10, 14, and 20dp are deliberate visual fine-tuning, such as aligning the input box text and button text.

## Local themes

If you want to change the color of a certain area in a window (such as a dark sidebar in a light-colored window), use `cx.Themed`, see [Elements and Views · Local Theme](el.md#local-topic).

When writing your own Gio drawing code, you can use `theme.Scope(p)` to temporarily replace the palette and call the returned function to restore it; Scope does not trigger redrawing or change the version number. Partial themes do not automatically follow the system appearance.

## Fonts

`theme.Face` specifies the font priority of the main text, with glyph fallback; `theme.MonoFace` specifies the same-width font priority. The desktop version gives priority to using system fonts, and fonts are only used when the system does not have corresponding fonts.

`theme.LoadFonts(data...)` receives the byte content of TTF, OTF, and TTC files, and redraws all windows after loading. The browser downloads and loads through `theme.FetchFonts("font.ttf")`; for Chinese font preparation, see [Run](web.md#build) in the browser.

`BodySize` / `SmallSize` / `HeadingSize` are 15 / 13 / 22sp respectively, and the standard single-line field height is `ControlHeight` (36dp). `theme.Material`'s glyph formatter can be reused when drawing directly with Gio.

## Reduced motion

macOS defaults to following the system "reduce dynamic effects" after running `window.Main()`, including runtime changes. Apps can override or restore this preference:

```go
theme.SetReducedMotion(true) // Explicitly turn off animation
theme.FollowSystemMotion()   // Restore follows the most recent system value
```

Called within a UI callback or `core.Update`. Automation mode uses explicit overrides to keep screenshots stable; other platforms allow animation by default and the app can still be closed.
