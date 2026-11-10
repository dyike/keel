# ui/el

English | [简体中文](README.zh-CN.md)

GPUI-style elements and views: Views are ordinary structs, and each frame `Render` returns an element tree built with chained styles; flexbox layout; element status (hover, scroll, input box content) is automatically saved according to element path or ID; Agent semantics are automatically generated.

- **Dependencies**: `core`, `theme`, `locale`, and internally `ui/internal/loop`, `ui/internal/editorstyle`, `ui/internal/inputcontent`, `ui/internal/animation`. Not dependent on `kit`, `window`.
- **Used by**: Application code. `window` recognizes `el.Root` through the `FillsWindow` interface and does not reference this package.

| File | Responsibility |
| --- | --- |
| `element.go` | All chain methods of `Element`, `Node`, `Styled[T]`, `Div`, `Text`, `Widget`, `Map`; `Decorate` package drawing, `VisitWidgets` reads the component coordinates after layout |
| `arena.go` | Per-root node chunks, reusable node buffers, independent pointer handles and build-generation checks; Cache and standalone elements keep persistent storage |
| `reload.go` | `core.ReloadState`: scroll offsets and input text across `keel run` reloads, matched by tree position, role and name; passwords are never saved |
| `time.go` | Frame times, declarative timers, explicit keys and reduced animations |
| `overlay.go` | Declarative overlay, anchor positioning, modal input isolation, focus constraints and hover query |
| `mount.go` | `cx.Mount`: Hang the view at the root of the window (`Dialog.Show`, `Sheet.Show`, `WindowNotifier` of the kit use it) |
| `key_context.go` | `KeyContext`, `cx.ActionAt`, `cx.Perform` (menus and command panels are executed by action name) |
| `scrollbar.go`, `scrollbar_mode.go` | Scroll bar drawing, display strategy (including following system) and fade in and out |
| `focus.go` | Native focus order, program focus, key bubbling and default activation, `FocusVisible` |
| `input.go` | `Input`、`TextArea` |
| `style.go` | `Style`, length (`Dp`, `Frac`, `Full`), alignment constants |
| `layout.go`, `flow.go` | flexbox, line wrapping and simple grid layout |
| `paint.go` | Drawing, click area, scrolling, input box, semantic information |
| `text_measure.go` | Dimensions-only text measurement with a frame-local cache; shares Gio shaping without generating drawing operations |
| `text_paint.go` | Reuses vector glyph fragments for short, single-line numeric labels; falls back to Gio labels for complex or overlapping text |
| `text_atlas.go` | Optional numeric label image cache; prepares visible labels before painting, bounds frame storage, and validates paint-time inputs |
| `viewport.go` | Procedural scrolling that draws coordinates, visible area, and nearest scroll container |
| `state.go` | Element state storage and recycling; long press on the touch screen to open the right-click menu |
| `root.go` | `View`, `ViewFunc`, `Context`, `Root`, `Embed`, the execution order of each frame |

Usage Guide: [Elements and Views](../../docs/el.md).

`input_paste.go` handles OnPaste before default text insertion; can be connected to core.ClipboardReader, and falls back to the Gio plain text path if native reading fails. Completed asynchronously via core.Update, old results are discarded after editing content or selection changes.

Atomic input references are connected to `ui/internal/inputcontent` by `el.InputDocument`, which only saves text, reference range, selection and editing transactions, and does not rely on Gio or other Keel modules. kit and markdown only depend on it indirectly via el.

Draggable boxes support parent-axis movement, bounds, rubber-band overshoot, release velocity and snap points. See [Dragging and release motion](../../docs/el.md#dragging-and-release-motion).

Text content transitions crossfade, roll numeric changes or morph whole shaping clusters. See [Changing text](../../docs/el.md#changing-text).

SharedLayout connects successive elements across parents, with matching input and anchor geometry. See [Shared elements](../../docs/el.md#shared-elements-across-parents).

Add `.Fly()` to a shared layout for handoffs beyond ancestor clips, scoped below later overlays; inert exit state still suppresses input and semantics.

Shared handoffs interpolate base solid colors, corners, border width and opacity. Combine SharedLayout with ContentTransition on Text for cross-owner glyph continuity, including interrupted moves and font changes.

MatchText on a shared container pairs descendant labels by relative ID/path. Shared handoffs also interpolate linear gradients and Elevation shadows.
