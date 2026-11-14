# Component gallery

English | [简体中文](README.zh-CN.md)

```sh
go run ./examples/components                    # Gallery: navigation and search on the left, selected component on the right
go run ./examples/components -section select    # One component, matching docs/kit/<name>.md
go run ./examples/components -section inputs    # One category, e.g. inputs, charts, navigation or motion
go run ./examples/components -theme dark        # Start in dark mode; use the top-right controls to change theme and language
```

The sidebar groups sections into General, Form & input, Data display, Lists & tables, Charts, Chat & messages, Feedback, Navigation, Overlay, Layout & shell, Motion, and Core capabilities. Category headings use stronger text color and semibold weight. Drag the divider to resize the sidebar (180–420dp), or focus it and use the arrow keys, Home and End. The chosen width survives theme/language changes during this session. Search filters component names. Each section is created when first opened and retains its state when switching sections. Window-filling sections such as Dock and Settings occupy the right pane; their overlays stay within that pane.

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

Run `go run ./examples/components -section motion` for pose retargeting, moving hit regions and anchors, keyed entry/exit, ordered replacement, layout resizing and shared component transitions.

Screenshot and matrix commands enable reduced motion so entry transitions do not leave content transparent in first-frame captures. Interactive runs retain animations.

The Motion page also demonstrates text morphing, rolling numbers, icon/image replacement and a shared image moving between thumbnail and detail parents.

Motion also demonstrates gradient/shadow handoffs and container-level MatchText without separate identities on each label.


All demonstrations previously in examples/motion now live here. Use the Motion sidebar category or select a topic:

- motion_values: springs, easing and keyframes.
- motion_gestures: hover, press and drag/inertia.
- motion_layout: entry/exit and reordering.
- motion_scroll: scroll-linked progress.
- motion_presets: Smooth / Snappy / Bouncy.
- motion_pages: page fades and retained input.
- motion_pose: moving hit regions and anchored overlays.
- motion_timeline: replay, pause, reverse, seek and scroll scrubbing.
- motion_effects: pointer glow, ripples and order confetti.
- rising_glow and shimmer_text: decorative particle and text effects.

For example: go run ./examples/components -section motion_effects.

TitleBar includes the former frameless example. Open Layout & shell → TitleBar and click **Open frameless window** to launch the native preview; repeated clicks raise that window, and closing it leaves the gallery open. Search, commands and sidebar toggling remain available in the title bar.

The former hotkey example is under Core capabilities → Global shortcuts & displays (section hotkey). Its preview explicitly enables/disables the global shortcut and refreshes display information on demand. macOS uses Cmd+Shift+K; other platforms use Ctrl+Shift+K. Closing the preview unregisters the shortcut. Platform errors are shown; no background clock or polling is started. Native calls run off the frame lock, and callbacks update the UI through core.Update.

Multiple windows is under Layout & shell (section multiwindow). Open settings repeatedly to reuse its window; close and reopen it to create fresh settings while the main page keeps its input. Ctrl/Cmd+, opens settings while the page is active. The former orders application is no longer a public example; its MCP regression fixture lives in cmd/keel-mcp/testdata/orders. Table, Form, Dialog and related components remain in the gallery.

## Narrow screens and browser bundles

Below 720 dp (scaled with accessibility text size), the gallery opens a searchable category index. Tap a component to open a full-width page; the 44 dp back button returns to the index. Search and built section instances survive navigation and desktop/compact layout changes within one runtime. The browser records gallery pages in history and accepts `#component=button` deep links. The native gallery remains a single complete application.

Web publication can use independently linked category programs (`go run ./internal/gallerybuild -out demo`). The complete gallery, including `#component=input` deep links, loads one full program by default so category changes retain demo state. Embedded `?section=button` previews select the smaller category program. Explicit `?bundle=general` opts into category programs; moving between these programs reloads and resets runtime-local state. `?bundle=all` explicitly selects the default full program. These are independent Go runtimes, not shared/dynamically linked WASM modules.

The browser entry excludes screenshot/matrix export and filesystem theme watching. Categories without code/Markdown examples omit the optional Go syntax-highlighting provider. Chinese font resources are shared by URL. Browser safe-area padding follows CSS environment insets on resize; Gio already reports visual-viewport keyboard insets. Physical-phone keyboard/IME and gesture acceptance still requires real-device testing.

The Layout demo uses one column below 480 dp, two below 720 dp and three on desktop; thresholds follow text scaling. Dialog and Sheet demo launch buttons wrap and use 44 dp heights on compact screens. Mobile interaction tests cover opening and closing Dialog, Sheet, Popover and Select at 320/390 dp, plus editing, saving and revisiting a dialog after a shortened viewport. These headless checks do not replace real-device keyboard/IME testing.

Component deep links take precedence over an outdated `bundle` parameter; explicit `bundle=all` still keeps the full runtime. Hash changes and browser history restore the selected component without pushing another history entry. If the current runtime lacks that component, the route is replaced with its category URL.

### Remaining comparison work

The NSH comparison has five open workstreams; these are not five isolated bugs:

- Audit every component at narrow widths, enlarged text and touch input. Current coverage is the gallery shell and selected layout/input/overlay examples.
- Preserve state across category runtime reloads. The full bundle currently provides state continuity within one runtime.
- Investigate a shared WASM runtime. Current category executables each include their own Go runtime.
- Validate real browsers/devices and deployment: keyboard/IME, gestures, safe areas, history, cache hits and negotiated compression. Headless/Node tests and generated gzip files do not establish these outcomes.
- Build a component-by-component capability gap list and prioritize additions. A larger catalog has not been delivered by the gallery/loading work.

Run the full mobile first-frame census with `KEEL_MOBILE_AUDIT=/tmp/keel-mobile KEEL_MOBILE_AUDIT_IMAGES=1 go test ./examples/components -run '^TestGalleryMobileAudit$'`. It emits semantic candidates and 320 dp images at normal/enlarged text sizes; clipped semantic bounds are not proof of layout correctness. For a served category build, run `node scripts/verify-gallery-web.cjs http://127.0.0.1:8766/demo/` with Playwright available to Node (for example through `NODE_PATH`). It uses Chromium touch emulation and saves screenshots and a request report under `work/gallery-web`; it does not validate physical keyboards/IME or deployment headers.

Narrow layouts also stack Resizable panes with separate saved sizes per orientation and use Dock compact tabs without rewriting its desktop layout. ToggleGroup scrolls horizontally, Radio cards and action rows wrap, and StaticTable scrolls horizontally to keep order identifiers and amounts readable. These adaptations preserve the existing component APIs; responsive presentation is selected by the examples.
