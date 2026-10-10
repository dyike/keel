# ui/core

English | [简体中文](README.zh-CN.md)

The foundation for all interface modules:

| Name | Function |
| --- | --- |
| `Widget` | Interface: Something that can `Layout`. All components and containers implement it |
| `Func` | Use a Gio layout function as a component |
| `ViewportWidget`, `ViewportFunc` | Keep the parent's visible rectangle when embedding or wrapping a widget; skip offscreen paint while retaining full content dimensions |
| `DecodeImage(ctx, source)` | Read local/HTTP/data images, limit encoding size and number of pixels; must be called in the background and manage timeout. `ReadImageSource` / `DecodeImageBytes` are two steps of disassembly. The Image of the kit is used to decode SVG and GIF animations at the same time |
| `Update(fn)` | Modify the interface from any goroutine: `fn` is executed in the next frame |
| `Call(gtx, fn)` | For those who write components: execute user callbacks and let all windows redraw |
| `Semantic(gtx, w, ops...)`, `Role(...)` | For those who write components: declare the role, name, and status of the component so that the Agent can see it |
| `Bind`, `BindIn`, `Bindings`, `BindingsIn` | Key table: action name to shortcut key, the context of `BindIn` can write conditional expressions (`Editor && !ReadOnly`, `Pane > Editor`), see [Menu · Context Conditional Expression](../../docs/kit/menu.md#context-conditional-expression) |
| `CurrentScrollGesture`, `ReportScrollGesture` | Scrolling device (wheel/trackpad) and gesture stage; `ui/window` reads reports from macOS, Wayland, Carousel and other components |
| `WindowControls`, `CurrentWindow()` | The component gets the activation status of the window where it is located, title area registration, minimize, maximize, close capabilities, and is used to customize the title bar. `ui/window` Registers the current window during layout |

- **Dependencies**: Only depends on Gio, and the internal `ui/internal/loop`.
- **Used by**: `el`, `kit`, `window`, `markdown`.

Threading rules: Modify components directly in callbacks; modify components in other goroutines and package them into `core.Update`. See [Architecture · Threading Rules](../../docs/architecture.md#threading) for the reason.

`ClipboardData`, `ClipboardImage`, `ClipboardReader` define rich paste exchange data and asynchronous reading interface; core does not read the system clipboard and does not reference native. The application adaptation platform reads the results and is dispatched back to the UI thread by the input component.
