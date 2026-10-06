# Component gallery

English | [简体中文](README.zh-CN.md)

```sh
go run ./examples/components                    # Gallery: navigation and search on the left, selected component on the right
go run ./examples/components -section select    # One component, matching docs/kit/<name>.md
go run ./examples/components -section inputs    # One category: controls, inputs, overlays, data, shell
go run ./examples/components -theme dark        # Start in dark mode; use the top-right controls to change theme and language
```

The sidebar groups sections into Core capabilities, Basic components, Inputs, Overlays, Data, and App shell. Search filters component names. Each section is created when first opened and retains its state when switching sections. Window-filling sections such as Dock and Settings occupy the right pane; their overlays stay within that pane.

English is the default. Add `-lang zh-CN` for Chinese. Browser examples use `?lang=en` or `?lang=zh-CN`. `demoText(english, chinese)` selects labels, messages, and sample data; framework controls use `ui/locale`. Changing the gallery language rebuilds sections and preserves the selected section and search query.

The "sample code" on the component documentation page directly reads the source code for registering the section and displays it below the online display; just modify the example and regenerate the site. When reused in scaffolding applications, the component construction and Render writing methods are retained, and the registered functions and shared auxiliary functions belong to the component library.

Each kit component corresponds to a `<name>.go`. Use `registerSection` to register a section with the same name to display common states, borders, and light and dark colors; it will automatically appear in the application sidebar after registration. `ui/kit/conventions_test.go` Check out each component with examples.

Generate screenshot:

```sh
go run ./examples/components -section chart -screenshot /tmp/keel-chart.png
```

There are also several sections that verify the basic capabilities of el: `theme` (switching between light and dark colors at runtime), `focus` (Tab / Shift+Tab, subtree disable), `time` (timed closing and cancellation), `overlay` (non-modal click penetration, modal mask, Esc closing and focus restoration).

The interaction test of the component is in `ui/kit/*_test.go`, and the agent snapshot test is in `ui/window/kit_*_test.go`.

`-section scrollable` Validates el horizontal scrolling, wide content cropping, and programmatic positioning.

`-section variable_list` Validates 100,000 row natural height list: positioning, insertion history, expanded content and window width changes.

`-section layout` Validates line wrapping layouts and simple grids, narrowing the window to check wrapping, row and column spacing, and minimum size.
