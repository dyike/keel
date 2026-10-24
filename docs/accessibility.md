# Accessibility

English | [简体中文](accessibility.zh-CN.md)

On macOS, VoiceOver and other assistive apps can read a Keel window and press its controls. Applications need no code for this: what a screen reader sees is the same element list an agent reads in [automation](automation.md#element), with the same roles, names, values and states.

Linux (AT-SPI), Windows (UI Automation), iOS and Android are not connected yet.

## What a screen reader gets

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

## Cost

Nothing runs until an assistive app asks a window for its elements. From then on, each frame of that window queues one refresh on the main thread. The refresh rebuilds the element list from the frame's semantic tree, updates only the objects that changed and posts `AXValueChanged` and `AXLayoutChanged` notifications. An idle window draws no frames, so it does no accessibility work either.

Describing what is scrolled out of view costs every frame while a client is active. On an Apple M4, an el frame of a 500-row list showing about 20 rows took about 856 µs instead of 710 µs (`BenchmarkScrollFrame*` in `ui/el`, +20%), with no extra allocations; without a client the frame is unchanged. Content drawn by a Gio widget (`el.Widget`) or a virtualized list exists only where it is laid out, so only its visible part is described.

## How it works

Gio keeps each frame's semantic tree on the main thread, where AppKit asks its accessibility questions. Keel's copy of Gio exposes that tree (`Window.AppendSemantics`) and accepts synthesized input (`Window.ProcessInput`) on that thread; see the patch log in `third_party/README.md`. `ui/window/accessibility_darwin.go` adds `accessibilityChildren`, `accessibilityRole` and `accessibilityHitTest:` to Gio's view class and registers one element class, once per process. No frame lock is taken in these callbacks.

## Limitations

- Keyboard focus is not reported (`accessibilityFocusedUIElement`). VoiceOver follows its own cursor, not Keel's focus.
- Text fields are read-only to assistive apps: their value can be read, but not set, and text ranges and selection are not exposed.
- Sliders do not support increment and decrement; there are no live-region announcements.
- An element whose content shows a changing label (a counter, "Save" becoming "Saved") gets a new identity, so VoiceOver may move to the start of the window.
- The real-window test (`KEEL_DESKTOP=1 go test -run RealWindowsAccessibility ./ui/window`) asks the view through AppKit's API in the same process. Reading, pressing and scrolling were also checked from another process through `AXUIElement`, the API VoiceOver uses, on `examples/hello` and `examples/components`. It has not been checked with VoiceOver running.
