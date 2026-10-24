# Accessibility

English | [简体中文](accessibility.zh-CN.md)

On macOS (VoiceOver) and Linux (Orca and other AT-SPI clients), assistive technology can read a Keel window and press its controls. Applications need no code for this: what a screen reader sees is the same element list an agent reads in [automation](automation.md#element), with the same roles, names, values and states.

Windows (UI Automation), iOS and Android are not connected yet.

## What a screen reader gets on macOS

Every element becomes an `NSAccessibilityElement` under the window's content view:

| Element role | AppKit role |
| --- | --- |
| `button`, `columnheader`, `gridcell` | `AXButton` |
| `link` | `AXLink` |
| `textbox` | `AXTextField` |
| `checkbox` | `AXCheckBox`; value 0, 1, or 2 when mixed |
| `switch`, `toggle` | `AXCheckBox` with subrole `AXSwitch` / `AXToggle` |
| `radio` | `AXRadioButton` |
| `tab`, `tablist` | `AXRadioButton` (`AXTabButton`), `AXTabGroup` |
| `select`, `combobox` | `AXPopUpButton` |
| `slider`, `progressbar` | `AXSlider`, `AXProgressIndicator` |
| `image`, `avatar` | `AXImage` |
| `menu`, `menuitem` | `AXMenu`, `AXMenuItem` |
| `list`, `listbox` | `AXList` |
| `table`, `row`, `tree`, `treeitem` | `AXTable`, `AXRow`, `AXOutline`, `AXRow` |
| `disclosure` | `AXDisclosureTriangle`; value 1 when expanded |
| `toolbar`, `separator`, `tooltip` | `AXToolbar`, `AXSplitter`, `AXHelpTag` |
| scroll containers (`ScrollY`, `ScrollX`) | `AXScrollArea` |
| other containers (`dialog`, `group`, `form`, …) | `AXGroup` |
| `text` and other roles | `AXStaticText`, the text as its value |

Containers keep the elements inside them as children, so VoiceOver moves into a dialog or a group the way it does in native apps. Disabled and selected states carry over. A press (VO-Space) clicks the center of the element's visible part, so it runs the same handlers as a mouse click. Frames are computed when asked, so they stay right while the window moves.

A scroll area holds everything inside it, including what is scrolled out of view, at its real position. VoiceOver can move to such an element and scroll it into view (`AXScrollToVisible`); a scroll area also scrolls by page (`AXScrollUpByPage` and the other directions). Both scroll the way the mouse wheel does, over the area. An element keeps its identity when it scrolls into view, so VoiceOver keeps its place. Agents still see only what shows: out-of-view elements and scroll areas are for accessibility only.

To make a custom component readable, give it a role and a name as automation needs: see [extending Keel](extending.md) and `core.Role`.

## Linux (AT-SPI)

The first window connects to the accessibility bus the session advertises (`org.a11y.Bus.GetAddress`) in the background and registers the application with the registry (`Socket.Embed`). Without an accessibility bus nothing else happens. The application, each window (role frame) and each element are D-Bus objects implementing `Accessible` and `Component`; elements that can be pressed also implement `Action` with one action, "click". Roles follow the same element roles, such as push button, check box, entry, link, page tab, combo box, list item, table row, scroll pane and label (`ui/internal/atspi/roles.go`; the numbers are checked against `atspi-constants.h`). States carry enabled, showing, checked, selected, expanded and editable.

`Component.ScrollTo` scrolls an element into view; `DoAction` clicks it. Changes are announced with `ChildrenChanged`, `StateChanged` (checked, selected, enabled, showing, active, defunct) and `PropertyChange` (name), and window focus with `Window.Activate` and `Deactivate`.

Calls arrive on D-Bus goroutines and read the elements of the last frame under a lock. Actions are queued and run on the window's goroutine after its next event: on X11 and Wayland, Gio's driver runs inside `Window.Event`, so that is where its input state may be touched.

## Cost

Nothing runs until an assistive app asks for elements. From then on, each frame refreshes them: on macOS a refresh queued on the main thread, on Linux on the window's goroutine right after the frame. The refresh rebuilds the element list from the frame's semantic tree, updates only the objects that changed and announces the changes (`AXValueChanged` and `AXLayoutChanged` on macOS, the AT-SPI events above on Linux). An idle window draws no frames, so it does no accessibility work either.

Describing what is scrolled out of view costs every frame while a client is active. On an Apple M4, an el frame of a 500-row list showing about 20 rows took about 856 µs instead of 710 µs (`BenchmarkScrollFrame*` in `ui/el`, +20%), with no extra allocations; without a client the frame is unchanged. Content drawn by a Gio widget (`el.Widget`) or a virtualized list exists only where it is laid out, so only its visible part is described.

## How it works

On macOS, Gio keeps each frame's semantic tree on the main thread, where AppKit asks its accessibility questions. Keel's copy of Gio exposes that tree (`Window.AppendSemantics`) and accepts synthesized input (`Window.ProcessInput`) on that thread; see the patch log in `third_party/README.md`. `ui/window/accessibility_darwin.go` adds `accessibilityChildren`, `accessibilityRole` and `accessibilityHitTest:` to Gio's view class and registers one element class, once per process. No frame lock is taken in these callbacks.

## Limitations

- Keyboard focus is not reported (`accessibilityFocusedUIElement`). VoiceOver follows its own cursor, not Keel's focus.
- Text fields are read-only to assistive apps: their value can be read, but not set, and text ranges and selection are not exposed.
- Sliders do not support increment and decrement; there are no live-region announcements.
- Elements keep their identity through changes of label and state when they take input (a button, a checkbox). Plain text whose words change (a counter) becomes a new element.
- Linux: there are no Text and Value interfaces, so Orca reads an entry's name but not its contents, and a slider's or progress bar's value only as a description. Screen coordinates are window coordinates, since Wayland does not report window positions. CI checks X11 under Xvfb; Wayland runs the same code but is not checked with a client.
- The real-window test (`KEEL_DESKTOP=1 go test -run RealWindowsAccessibility ./ui/window`) asks the view through AppKit's API in the same process. Reading, pressing and scrolling were also checked from another process through `AXUIElement`, the API VoiceOver uses, on `examples/hello` and `examples/components`. On Linux, CI runs `ui/window/testdata/atspi` under Xvfb with at-spi2-core: as a client on the accessibility bus, it finds the application on the registry's desktop, reads roles and names, presses a button, and scrolls a row into view and presses it; `pyatspi_check.py` reads `examples/hello` through libatspi, the library Orca uses, and checks its checkbox. Neither VoiceOver nor Orca has been run.
