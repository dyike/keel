# Design decisions

English | [简体中文](decisions.zh-CN.md)

Document the trade-offs that impact the big picture: the options at the time, what was chosen, and what the cost was. When overturning a decision, do not delete it. Add a new one at the end and indicate the replacement relationship.

## Choose Gio

**Decision**: The interface is drawn with Gio, replacing the original go-gui implementation.

**Goal**: The application code is only written in Go, without writing HTML, CSS, JS, or introducing WebView.

| Candidate | Why not selected |
| --- | --- |
| Wails, WebView class solutions | Direct conflict with the goal: the interface is still HTML/CSS/JS |
| go-gui (original implementation) | It works, but for multi-window, life cycle, show-hide, and shortcut keys, Keel maintains about 940 lines of window management and platform startup code (excluding tests), part of which is to bypass upstream behavior (such as asynchronous window creation without failure callback). After switching to Gio, the corresponding code is the 235 lines related to the window in `ui/` |
| giu (Dear ImGui) | ImGui relies on C++. Its appearance is tool and debugging panel style. For ordinary applications, a lot of style changes are required |
| Fyne | Mature, complete components. But it comes with a complete application life cycle, theme system and component system. Keel can only cover it with a thin shell. After modularization, each layer will repeat the corresponding concept of Fyne |

The benefits of Gio: The interface is completely drawn by Go, and the components are just ordinary Go structures. The immediate mode is easily wrapped into a simple stateful API; it comes with off-screen rendering (`window.Screenshot` uses it) and programmable input routing (`ui/internal/uitest` uses it).

The **price** is explicitly accepted when choosing Gio:

- There is no native control appearance, and the appearance of all components is drawn by Keel himself.
- The window cannot be hidden and then shown, it can only be closed and reopened.
- Chinese relies on system font fallback. It is necessary to fix the font priority (`theme.Face`) and fine-tune the vertical position of the text: el moves the text to the middle of the line box according to the measured glyph position of the font, and ensure that the descending parts (g, y) do not fit in the line box. See `textShift` of `ui/el/paint.go`.
- You have to write the components yourself. When making this decision, there are only text, buttons, links, input boxes, and check boxes; now `ui/kit` has more than 80 components, covering the entire component directory of GPUI Kit, see [kit component](kit.md).

## Stateful components wrap outside immediate mode

**Decision**: Gio is in real-time mode and re-describes the interface every frame. Keel provides `widget.Button(...)`, a stateful component that returns a pointer. The user builds the component tree once, and then only changes the state.

**Reason**: To write Gio directly, you need to manage state objects such as `widget.Clickable` yourself and call layout functions for each frame. There is a lot of boilerplate code for each page. Stateful components bring business code close to "declaration interface + writing callbacks".

**Price**: Dynamic structure (adding and deleting rows in the list, displaying based on conditions) needs to be supported by the component itself; when you want a completely flexible layout, use `core.Func` to return to Gio writing. (Later solved by `ui/el`: the view rebuilds the element tree every frame, and the dynamic structure is ordinary Go conditions and loops, see "New ui/el" below.)

## A global frame lock

**Decision**: The rendering of all windows and all callbacks are executed serially under a lock on `ui/internal/loop`, and the component itself is not locked. Other goroutines use `core.Update` to queue modifications.

| Candidate | Why not selected |
| --- | --- |
| Each component has its own lock | The component reads its own status during the `Layout` process, and the callback will change other components. The lock sequence is difficult to guarantee and deadlock is easy; the author of each new component must handle concurrency |
| A lock for each window | Callbacks often change components across windows (the main window button changes the text of the setting window), but global coordination is still required |
| All operations go through `core.Update` | Changing components in the callback also requires a layer of wrapping, which is tedious to write |

**cost**:

- Windows are rendered serially; a slow callback will freeze all windows together.
- You cannot wait for the main thread while holding the lock, otherwise it will form a waiting loop with other windows. This problem actually occurred (⌘+ under multiple windows, stuck), so `Raise` and `Close` were changed to asynchronous. See [Architecture · Cannot wait for main thread](architecture.md#avoid-waiting-for-the-main-thread-while-holding-the-lock) within lock.
- The text formatter (`text.Shaper`) is not concurrency-safe, and the global lock incidentally ensures that multiple windows sharing a formatter are safe. It's not a price, but remember to deal with it when removing the lock.

## Native capabilities are separated from the interface, and cgo is concentrated in one place

**Decision**: `native/` does not depend on any interface module; all Objective-C and cgo code are placed in `native/internal/sys`, and the public package only performs parameter verification.

**Reason**: Tools that do not require a window (background screenshots, global shortcut keys) can only use `native`. The cgo code is concentrated in one package, and the platform stub functions are also concentrated in `sys_other.go`. If a function is omitted, the compilation will fail directly and no error will occur during runtime.

**Cost**: To add a native capability, you need to change 4 files (`.m`, `.h`, `sys_darwin.go`, `sys_other.go`), plus a public package.

## Divide modules hierarchically according to responsibilities

**Decision**: The interface is split into five modules: `ui/core`, `theme`, `layout`, `widget`, and `window`. The dependencies only go down; the system capabilities are split into four modules: `native/permission`, `screen`, `input`, and `hotkey`, which do not reference each other. One directory, one README per module, with boundaries enforced by tests for `internal/deps`.

This structure was finalized after three attempts. The process is recorded here to avoid having to go through it again in the future:

| Version | Structure | Problem |
| --- | --- | --- |
| First edition | Top tiles `ui`, `theme`, `widget`, `box`, `app` | The names do not show any relationship (what is `box`?), and are juxtaposed on the top layer with `native`. They look like five unrelated things; I am worried that there will be more and more top layers in the future |
| Second version | All interfaces are merged into one `ui` package, divided by files | 15 files in one directory, windows, components, themes, and locks are mixed together. It is difficult to understand which ones are the foundation and which are the upper layers |
| Third edition (current) | `ui/` is divided into five sub-modules, symmetrical with `native/` | The business code needs to introduce 2–3 packages |

Reasons for choosing the third edition:

- **There are only two groups of `ui` and `native` on the top level**. You can see the "interface" and "system capabilities" at a glance.
- **Each sub-module has a single responsibility, and the name directly describes the responsibility**: `core`, `theme`, `layout`, `widget`, `window`.
- **Layering makes dependencies readable**: Look at `core` and `theme` first to understand the foundation, `window` does not recognize any specific components. To understand a module, you only need to look at it and the modules below it.
- **Growth method fixed**: New components are new files under `widget/`, new containers are new files under `layout/`, and new directories will not pop up.

**Cost**: The business code introduces three packages `layout`, `widget`, and `window`, which are two more lines of import than the second version. `layout` and Gio's `layout` (`third_party/gio/layout`) have the same name, and they must be aliased when used in the same file.

## Agent tests are rendered in memory and do not drive the real window

**Decision**: In `KEEL_AUTOMATION` mode, `ui/window` assigns a shadow window to each window: the same set of components, independent Gio input routing, Agent operations are only sent to the shadow window; `cmd/keel-mcp` drives it through sockets. In visible mode, the real window is displayed as usual, the component state is shared on both sides, and the user can watch the Agent operate; in `KEEL_HEADLESS=1`, there is only a shadow window.

| Candidate | Why not selected |
| --- | --- |
| Use `native/input` to move the real mouse, send real keys, and then take screenshots for recognition | Requires accessibility and screen recording permissions; will take away the user's mouse and keyboard; it will fail if the window is blocked or the monitor is changed; "What is on the page" can only rely on image recognition, and the name and status of the elements cannot be obtained |
| Read window contents through macOS Accessibility API | Gio does not expose controls to Accessibility API on macOS and cannot be read |
| Injecting events into the real window | Gio's `app.Window` does not open the interface for injecting input events |

The advantages of the memory window: strong certainty, no permissions required, no interruption to the user, and a complete process in about 1.5 seconds; element information comes from Gio’s semantic tree, and the name, status, and position are all precise values.

**Cost**: Unable to detect system window-level issues, such as window position, system menu, input method, and main thread-related deadlocks between real windows. The latter is complemented by `KEEL_DESKTOP=1`'s real window testing. Components must declare semantic information themselves, otherwise Agent cannot see it.

**Why is it a shadow window instead of forwarding the input of the real window to Keel's own router**: The latter needs to take over all the input of the real window, including the candidates and group words of the Chinese input method, the clipboard, the cursor shape, and the system Tab focus, which is equivalent to rewriting the input layer of Gio. The most prone to problems happens to be the Chinese input. The shadow window allows the user's side of the code path to remain unchanged at the expense of separate recording of focus on both sides.

**Another option**: Let the application act as the MCP server. No choice, because testing often requires restarting the application and changing the application to test. This can only be done by placing the launcher in an independent process; and the standard output of the application cannot be occupied by the MCP protocol.

## Make Go version of GPUI (ui/el) on Gio

**Decision**: Add `ui/el`, adding a layer on top of Gio based on GPUI ideas: chained style builder (`Styled[T]`), flexbox layout engine, and element status automatically managed according to element paths. Gio continues to be responsible for windows, GPU rendering, input methods, and clipboard. The old `ui/layout` + `ui/widget` are retained and migrated one by one.

**Why**: After writing a batch of components such as tables and drop-down boxes, the cost of Gio is very clear: each interactive component must declare its own state variables (`widget.Clickable`, etc.), keyboard focus requires three steps of registering filtering, registering a handler, and executing focus commands. The layout must be nested `layout.Flex{}.Layout(gtx, layout.Rigid(...))` layer by layer, and the Agent semantics must be handwritten one by one; if one step is missed, problems such as "the first frame cannot be clicked" and "the node disappears" will appear. These should all be handled uniformly by the framework.

| Candidate | Why not selected |
| --- | --- |
| Continue to write components on Gio | The above costs must be paid again for each new component |
| Implement GPUI (window, GPU, text typesetting, input method) from scratch | The biggest workload is exactly these platform layers, Gio has already done it |
| Copies GPUI's Entity/Context/cx.listener | This mechanism is largely for Rust's borrow checking. Go has closures and GC. It is enough to make the view a normal struct and callback to capture the pointer directly; only keep the parts that are still needed in Go (`cx.Shortcut`, use `core.Update` for background updates) |

**Practice trade-offs**:

- Distribute events first and then render: the click effect is drawn in the same frame, unlike old components that have to wait for the next frame.
- The key of the element state is "path on the tree", which takes `ID` for each level. If not, it takes the sequence number. There is no need to write an ID for a static structure; it is required for a changing list.
- Only elements that are interactive, have semantics or require state are pushed into Gio's clipping area. Other `Div` are not pushed, and child elements can overflow it.
- Semantics are automatically inferred, `Role`/`Name`/`Value`/`Selected` are only used to supplement or cover.

**Price**: The layout is a subset of flexbox (no wrap, grid, horizontal scrolling, min-content), no virtual lists and animations; these are supplemented as needed. During the transition period, two sets of writing methods coexist. Newcomers should know to read `el` first.

**Migration order**: The page structure of the order example has been written with `el`, and tables, drop-down boxes, radio selections, switches, and dialog boxes are temporarily embedded with `el.Widget`. Then migrate according to frequency of use: buttons and form controls → drop-down boxes, tabs → tables (requires virtual lists) → dialog boxes (requires el overlay). With each migration, `examples/orders`'s end-to-end tests must continue to pass.

## Markdown streaming rendering

**Decision**: `ui/markdown` uses goldmark parsing, chroma highlighting, and the richtext of the Gio extension package to shuffle in-line drawings. The source text is cut into pieces according to "blank lines outside the code block", and the parsing results are cached for each piece; during streaming output, only the last piece is re-parsed, and its unclosed inline syntax is temporarily completed. The written block reuses elements and layout via `el.Context.Cache`.

**Why it's diced like this**: AI output is only appended at the end. According to the top-level block cache, the cost of appending is only related to the size of the last block and has nothing to do with the length of the entire article. The cost is that cross-block reference links do not take effect, and the same list separated by blank lines will become two lists (the starting number of the ordered list will be retained). Both situations are rare in AI output.

**Why complete unclosed syntax**: If not completed, `**bold` will be displayed as an asterisk before closing. When closed, the entire paragraph will be suddenly rearranged, and it will keep flashing during streaming output. The completion only works on the last paragraph, and it will be re-analyzed according to the original text after the answer is completed, without affecting the final result.

**Three changes in performance** are all analyzed:

1. Richtext typesetting is expensive, and el will measure the same block several times per frame. The rich text block caches the results of the last 8 measurements (differentiated by constraints).
2. el's layout engine originally arranged the stretched and stretched child elements twice (first measure the natural size and then arrange the final size). When the width of the container is known, it is directly arranged according to the stretched width. `Grow` is changed to the semantics of `flex: 1` (the initial size is calculated as 0). Each sub-element is arranged only once per frame. This also allows `cx.Cache`'s layout reuse to achieve stable hits.
3. Elements outside the scroll container skip drawing.

Result: Dropped from 5.0ms to 0.5ms per frame when streaming output (11KB document).

| Candidate | Why not selected |
| --- | --- |
| Each append is analyzed and formatted in its entirety | The cost is directly proportional to the length of the document, and long answers will become increasingly stuck |
| Rendering Markdown with WebView | Against project goals |
| Write your own Markdown parser | goldmark is CommonMark standard implementation, GFM extension is complete, no need |

## The program calls SetXxx without triggering the callback

**Decision**: `Field.SetValue`, `Check.SetValue` do not trigger `OnChange`, only user actions trigger.

**Reason**: The common way of writing is "update B when A changes, update A when B changes". If the program call also triggers a callback, it will loop endlessly. Gio's `Editor.SetText` itself generates change events, and `Field` filters it out by recording the last notified content.

## Exit when the last window is closed

**Decision**: Call `os.Exit(0)` after the last window is destroyed.

**Reason**: Gio's `window.Main()` will not return, and a desktop program without a window will appear to the user to have exited.

**Price**: The code after `window.Main()` in `main` and `defer` will not be executed, and the cleanup work must be put into `OnClose`. When the tray is resident in the future, it needs to be changed to a configurable exit strategy.

## Things not to do

- **Does not do HTML/CSS rendering and does not embed WebView.** This is the raison d’être of the project.
- **Does not imitate the appearance of native controls on each platform.** All platforms have one look and feel, with minimal maintenance costs.
- ~~The `native` capability of Windows and Linux is not implemented yet. ~~ Implemented: The interface was designed in a platform-independent manner from the beginning. Later, `sys_windows.go`, `sys_linux.go` and other files were added. For the platform status of each capability, see [Native capability](native.md).


## New component based on el, ui/widget frozen

**Decision**: Starting from 2026-10-01, `ui/widget` will only fix bugs and no longer add components or capabilities. The new component is placed in `ui/kit` and is implemented based on `ui/el`. kit directly depends on `core`, `theme`, `el`, but does not depend on `widget`, `layout`, `window`. This decision replaces the previous "new components in widgets" and the old migration order.

**Why**: Focus, keyboard, disable, timing and overlays need to be provided by el. Continuing to implement new components in the widget will re-implement these mechanisms during migrations.

**Cost**: The corresponding interactive components cannot be delivered before the infrastructure is completed. During the transition period, the two sets of components coexist, the old code continues to work, and the new code uses the kit. M0 only establishes rules, module boundaries and global topics, and components start from M1.

| Stages | Work and Completion Standards |
| --- | --- |
| M0 | Component specification, dependency registration, success/warning/prompt semantic color, light and dark switching during runtime; review after completion |
| M1 | Alert, Empty, Avatar, Tag and other display components; add focus and buttons, disable, and time, and then create Spinner and Skeleton. After the focus and button interfaces are completed, review them first, and then continue to rely on their components after confirmation |
| M2 | Popover, Tooltip, Menu, DropdownButton, Dialog, Sheet, Notification; the dialog box of the order example is changed to kit |
| M3 | Migrate the old form controls, and then do NumberInput, Combobox, Calendar, DatePicker, and form verification; the "New Order" form of the order example all uses kit |
| M4 | Virtual list, Tree, kit table, Command, extract chat message components; the example no longer references ui/widget |
| M5 | Application Shell |
| M6 | Visualization; delete ui/widget after migration is completed, and clean up all remaining dependencies |

Each component is implemented, verified, and submitted individually. Execute `go build ./... && go vet ./ui/... && go test ./... -count=1` before submission; do not use the test result cache to prevent the sample source code changes started by the end-to-end test from not being detected by the cache. For specific requirements, see [Component Development Specification](component-development.md).

The code editor, HTML rich text, complete TeX, partial topic coverage, and Kbd action binding query are not currently available. These capabilities require their own models and interfaces, which can be decided separately when there are actual needs.

## Delete ui/widget and ui/layout

**Decision**: Remove `ui/widget` and `ui/layout` after M6 is completed, leaving Keel with only one set of components, `ui/kit`. This decision replaces the previous "stateful components + containers" and "old ui/layout + ui/widget retention".

**Why**: The kit has covered all functionality of the old component, and examples and tests have been migrated. Two sets of components coexist, and two copies of focus, disable, overlay, and semantics must be maintained, and the document must also be written in two ways.

**How to do it**: The image loading, decoding restrictions and occupancy used by Markdown are moved to `ui/internal/imageload`, and Markdown exposes `ImageLoader` and `DecodeImage`. `window.Options.Content` and `Overlay` still accept any `core.Widget`, and your own Gio code can still be used.

**Price**: No compatibility layer. Code using the old component is rewritten as [kit component](kit.md): `widget.Xxx` generally corresponds to `kit.Xxx`, `layout.Column` / `Row` / `Card` uses `el.Div` instead.

## Frame text is concentrated in ui/locale

**Decision**: All text displayed by Keel or reported to Agent (OK, Cancel, Copy, Close, Please Select, "Line 36", etc.) are put into `ui/locale`. The default is Chinese, and English preset is provided. `locale.Apply` is switched at runtime, all windows are redrawn, and `cx.Cache` automatically expires. The method is consistent with `theme`.

**Why**: A total of more than 20 Chinese characters were written in the kit, widgets and markdown at that time. When the application switches to the English interface, these framework text cannot be changed accordingly. The more components there are, the higher the cost of modification later, so focus on processing while the kit is still in its early stages.

**Scope**: Only framework text. The application is responsible for its own text, and Keel does not operate a translation system; `Current().Lang` tells the application what language it is currently in, and the interface will be re-rendered after switching.

**Restraint**: The test of `internal/deps` prohibits writing Chinese string literals in the framework code, except for the glyph probe "国" and "国Ag" used to measure CJK line height.

**Price**: Accessibility names are unified with "action + space + object", "clear search" becomes "clear search"; the alternative text of Markdown images changes from "[image: x]" to "[image x]".

## Custom title bar is not implemented yet

**Decision**: M5 does not do TitleBar (custom window title bar).

**Why**: Customizing the title bar requires a borderless window, the title bar area must be able to drag the window and double-click to maximize it, and the traffic light button on macOS must also remain in the correct position. All of these must be implemented in conjunction with `native` and `ui/window` (macOS full-size content view, drag area registration). The window cannot be dragged by just relying on kit to draw a component that "looks like a title bar". Before there is a specific demand, this part of the original work requires high investment and uncertain returns.

**Price**: Apps can only use the system title bar. If necessary, create another one: first add a borderless window and drag area to `native`, and then provide a TitleBar in the kit.

## Custom title bar: implemented using Gio’s borderless window

**Replaces**: The previous article "Customized title bar is not implemented yet".

**Decision**: `window.Options.Frameless` Turn on Gio's borderless mode (`app.Decorated(false)`), `kit.TitleBar` draw the title bar and window buttons yourself. Drag to register the area with Gio's `system.ActionInputOp(system.ActionMove)`, and the window button calls the window through the `WindowControls` interface of `ui/core`.

**Why the judgment was changed**: The previous judgment requires a new borderless window and drag area to be written in `native`. After actually checking Gio v0.10.3, I found that it already provides both: the borderless mode on macOS will extend the content to the title bar and make the title bar transparent; when pressing the area where ActionMove is registered, the system will complete the window dragging. So there is no need to add native code.

**trade-off**:

- Gio will hide the traffic light button of macOS in borderless mode, so the button is drawn by kit itself, and the color follows the system convention.
- kit cannot reference `ui/window`, so put a minimal interface (Frameless, Minimize, ToggleMaximize, Maximized, Close) in `ui/core`. The window registers itself as the "current window" during layout, and the component gets it during Render for callback use. All renderings are done serially under the same frame lock, so this is safe.
- When Gio determines whether a location is a dragging area, it only looks at whether dragging is registered there, regardless of whether there are buttons on it. Therefore, the dragging area can only be placed next to the button, not around the button. The layout of the TitleBar is designed according to this constraint: window buttons, application content, and drag areas are arranged side by side.

**Cost**: Double-clicking the title bar on macOS will not zoom the window, because the press event is directly handed over to the system for dragging, and the application cannot receive the double-click; when the window loses focus, the button will not turn gray.


## Custom title bar: follow window focus and access macOS double-click

**Decision**: `core.WindowControls` Add window focus query and title drag area registration. `window` updates the activation status from Gio ConfigEvent; each time TitleBar is drawn, it only registers the middle area without buttons and slots, and disables and clears undrawn frames.

**Reason**: Gio's macOS ActionMove directly calls AppKit native drag when pressed, and ordinary component double-click callbacks cannot receive the second complete click. `window` Use the current NSView's local event monitor to intercept double clicks only within the registered area, handling them according to the system's no-op/minimize/zoom preferences. The header control continues to receive normal input.

**Threads and life cycle**: retain first when passing the NSView handle, asynchronously submit it to the main thread and then release it; the native registration table is removed when the view is changed or the window is closed. The native callback only submits actions to the Go update queue and does not wait for the frame lock on the main thread. UI does not reference `native`, `kit` does not reference `window`.
