# Component gallery

English | [简体中文](README.zh-CN.md)

```sh
go run ./examples/components                    # 组件库应用：左侧导航和搜索，右侧是选中的组件
go run ./examples/components -section select    # 只显示一个组件，名字与 docs/kit/<名字>.md 一致
go run ./examples/components -section inputs    # 一类：controls、inputs、overlays、data、shell
go run ./examples/components -theme dark        # 深色启动；应用右上角也能切换深浅色和中英文
```

The sidebar of the application is grouped by "basic capabilities, basic components, input, overlay, data, application shell", and the search box at the top filters component names. Each component is created when it is first opened, and the state is retained when it is removed and returned. The components of the entire window (Dock, Settings, etc. el.Root) occupy the content area on the right, and the built-in overlay is also limited to the content area.

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
