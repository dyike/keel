# Architecture

English | [简体中文](architecture.zh-CN.md)

Keel's code is divided into two groups: `ui/*` is responsible for the interface, and `native/*` is responsible for system capabilities that Gio cannot do. Both groups are divided into small modules according to their responsibilities, with each module having a directory and a README. The interface is built on a convention: the rendering of all windows and all callbacks are executed serially under the same lock.

## Modules

```
github.com/dyike/keel
├── ui/
│   ├── core/             Foundation: Widget interface, callbacks, thread rules
│   ├── theme/            Color, size, font
│   ├── locale/           The text displayed by the frame itself: OK, copy, close...
│   ├── base/             Unstyled component behavior: keyboard navigation, initial jump, multiple selection
│   ├── el/               GPUI Styles: views, chained style elements, flexbox
│   ├── plot/             Public scale, graph layout and instant drawing
│   ├── kit/              Components: Button, Input, Table, Dialog, Chart... (one component, one file)
│   ├── window/           Window: Open, Main, shortcut keys, screenshot
│   ├── markdown/         Markdown Rendering, for AI streaming output
│   ├── highlight/        Optional: code highlighting (chroma), kit and markdown are colored after introduction
│   ├── netimage/         Optional: http(s) image loading (net/http), only after importing can the image use the network address
│   └── internal/         loop（Frame lock), editorstyle (input drawing), inputcontent (atomic reference editing), imageload (image loading), uitest (test tool)
├── native/
│   ├── permission/       Permission check and application
│   ├── screen/           Monitor list, screenshots
│   ├── input/            Synthesize mouse and keyboard events
│   ├── hotkey/           Global shortcut keys
│   ├── notification/     System notification (not dependent on UI)
│   ├── clipboard/        Asynchronously read text, encoded images and file paths (not dependent on UI)
│   ├── internal/sys/     cgo Bindings, all Objective-C code only goes here
│   ├── internal/wlclip/  Linux Wayland cgo binding for clipboard (only clipboard reference)
│   └── native.go         shared error value
├── cmd/
│   ├── keel/             Scaffolding: Create a new project, run it, generate icons for each platform and package it
│   └── keel-mcp/         MCP server：Agent Use it for end-to-end testing of your application
├── internal/
│   ├── deps/             Module bounds checking
│   ├── svgicon/          Draw SVG icons to PNG (common to site and scaffolding)
│   ├── appicon/          Cut out application icons according to macOS, Windows, and GNOME specifications (shared by scaffolding and ui/window)
│   └── site/             Documentation site generator
├── examples/
└── docs/
```

There is a README in each module directory, which describes what it does, who it depends on, and who it depends on.

### Depends on direction

```
ui:
  kit ───────► el ──┐
  kit ───────► base（Component behavior, does not depend on other modules)
  markdown ──► el   ├──► theme、locale
  plot ─────────────┤
  window ───────────┤
                    └──► core

  theme、locale ──► internal/loop（Notify all windows to redraw when switching themes or languages)
  el ──► internal/editorstyle（input box drawing)
  el ──► internal/inputcontent（Atomic reference content and editing transactions)
  markdown ──► internal/imageload（Picture decoding and placeholder)

native:
  permission ─┐
  screen     ─┼──► internal/sys ──► native（error value)
  input      ─┤
  hotkey     ─┘
```

There are only three rules:

1. **Dependence only goes down.** The lower layer does not know the existence of the upper layer: `core` only relies on the internal frame lock, `theme` only refers to the internal frame loop to notify the theme to redraw; `el` does not know that there is `kit`; `window` only recognizes the `core.Widget` interface, and does not know what specific components there are.
2. **Same layers do not reference each other.** `kit`, `markdown` and `window` do not reference each other; the four `native` modules do not reference each other.
3. **`ui` and `native` do not reference each other.** Programs that do not require windows (background screenshots, global shortcut keys) only reference the required `native/*` and will not be brought into Gio.

`markdown` uses `el` for typesetting, and pictures use `internal/imageload` (loading, decoding restrictions, placeholders). The input box of `el` uses `internal/editorstyle` to draw the cursor and selection; this internal package is only responsible for Gio input drawing and glyph measurement and does not rely on other Keel modules.

`cmd/keel-mcp` does not reference any Keel package, nor does it reference Gio: it only drives the automation mode of `ui/window` through the JSON protocol on the socket, see [Automation](automation.md#principle).

`cmd/keel` does not reference the interface package: it only uses `internal/svgicon` to draw placeholder icons, uses `golang.org/x/image` and `tc-hib/winres` to generate icons and Windows resources for each platform, and calls `go`, Gio's gogio and macOS's `codesign` when packaging, see [Quick Start·Packaging](getting-started.md#package).

**Large dependencies are introduced on demand.** chroma (code highlighting) and `net/http` (network pictures) are about 4 MB each, and are not placed in the dependencies of kit, markdown, el, core, and window: core defines two interfaces, `Highlighter` and `ImageFetcher`. `ui/highlight` and `ui/netimage` are registered and implemented in `init`, and are introduced when the application uses them. `internal/deps`'s `TestHeavyDependenciesAreOptIn` prevents them from being pulled back in.

These rules are enforced by the test of `internal/deps`: it writes the packages that each module is allowed to depend on into a table. If it goes out of bounds or the new directory is not registered, `go test ./...` will fail. When changing the architecture, change the table first, and then change the code.

### What to put on each layer

| Module | What to put | What not to put |
| --- | --- | --- |
| `ui/core` | Interfaces and functions that all interface modules must comply with | Any specific components, colors |
| `ui/theme` | Visual parameters, global palette switching, local theme scope and redraw notifications | Components |
| `ui/locale` | The framework itself displays or reports text to the Agent, switching languages at runtime | Apply its own text and translation system |
| `ui/base` | Component behavior: keyboard navigation, initial jump, multi-select, open state | Any drawing, color, Keel dependency other than Gio |
| `ui/plot` | Scale bar, stacked/pie chart layout, instant drawing basics | Finished charts, input processing, window management |
| `ui/kit` | el and base based components | Gio input and overlay infrastructure, window management |
| `ui/window` | Things bound to the window: life cycle, shortcut keys, root view, screenshots | Specific components |
| `ui/el` | Elements, styles, layout engines, element states, views | Business components (they are written as functions or views in the application) |
| `ui/internal/loop` | Mutable state shared across windows: frame locks, update queues | any Gio type |
| `ui/internal/inputcontent` | Atomic input references, grapheme boundaries, coordinate mapping and undo transactions | Gio, other Keel modules and platform events |
| `ui/internal/editorstyle` | Input box cursor, selection drawing and glyph measurement | Specific Keel components, windows and themes |
| `ui/internal/imageload` | Image source analysis, asynchronous loading, size restrictions, loading and failed placeholder | Component appearance, click and other interactions |

### When to create a new module

First check if it is the responsibility of an existing module. Yes, just add files in that module: the new drop-down box is the kit component, which is `ui/kit/select.go`; the new layout capabilities (such as line breaks) belong to el, which is the change in `ui/el/layout.go`.

Only create a new directory if the existing modules cannot be installed and it has its own clear responsibilities. For example, the clipboard: it does not belong to any of the four existing capabilities, and there are scenarios where it is only used, so it is a new module `native/clipboard`.

When creating a new module, write README and register it in the module diagram and `internal/deps` table above.

## Frame lifecycle

Gio is an immediate mode framework: every frame calls `Layout` of all components from the beginning. The components process input events since the previous frame in `Layout` and output drawing instructions at the same time. Keel's component object only saves the state (text, check, input box content), but does not save the drawing result.

Each window has its own goroutine, which is executed after receiving `FrameEvent`:

```
lockFrame()
drainUpdates()               // Execute core.Update queued functions
window.handleShortcuts(gtx)  // Window shortcut keys
root.Layout(gtx, content)    // background + scroll + 24dp padding + component tree
    └─ Layout of each component: handle events → core.Call (callback) when necessary → draw yourself
unlockFrame()
e.Frame(ops)                 // Submitted to GPU, executed outside lock
```

The callback occurs in the middle of the layout. The component in front of the button has been drawn this frame, and the old value is seen. Therefore, after `core.Call` executes the callback, it will request all windows to draw another frame. The actual effect is that the callback modification is displayed on the screen one frame later, about 16ms on a 60Hz screen, which is imperceptible to the human eye.

## Threading

**One sentence: Modify components directly in the callback; components modified by other goroutines must be included in `core.Update`.**

There is a global frame lock in `ui/internal/loop`. Each window holds it when drawing a frame. Component callbacks occur during the frame drawing process, so the lock must be held when the callback is executed. The components themselves are not locked, and their concurrency security depends entirely on this lock.

| Where does the code run | Can the component be modified directly | How to do it |
| --- | --- | --- |
| Callbacks for buttons, input boxes, and check boxes | Can | directly change fields, call `SetValue`, etc. |
| Callback of `window.Options.Shortcuts` | Can | be changed directly |
| `window.Options.OnClose` | Can | Change directly |
| The goroutine you started, `time.AfterFunc` | cannot | `core.Update(func() { ... })` |
| `hotkey.Register`'s callback | cannot | `core.Update(func() { ... })` |
| `main` in `window.Main()` before | Can | The window has not started to be drawn, there is no competition |

`core.Update(fn)` Put `fn` into the queue and wake up all windows; the next window that starts drawing frames first executes the queue and then layouts it. It is safe to call from anywhere, including inside callbacks.

How to write background tasks:

```go
kit.Button("刷新", func() {
    v.status = "加载中…"
    go func() {
        data, err := fetch()           // Time-consuming operations outside the lock
        core.Update(func() {             // The result returns to the interface
            if err != nil {
                v.status = err.Error()
                return
            }
            v.status = data
        })
    }()
})
```

The cost of this model:

- **All windows are rendered serially.** One callback is stuck for 2 seconds, and all windows are stuck for 2 seconds. In the callback, only microsecond-level things such as changing the state are done, and I/O and calculation are done in goroutine.
- **`core.Update` is asynchronous.** `fn` has not been executed when the call returns. Waiting for the result of `core.Update` in the callback (for example, using a channel to wait for it to finish executing) will cause a deadlock: the callback holds the lock, and `fn` has to wait for the lock.
- **If there is no window drawing frames, the queue will not be executed.** The function `core.Update` will be executed in the first frame of the first window before the program opens the window.

Why choose one big lock instead of locking every component? See [Design Decisions](decisions.md#a-global-frame-lock).

## Avoid waiting for the main thread while holding the lock

Here are the rules for those who maintain `ui/window` and write components: **While holding a frame lock, you cannot call any Gio window method that will wait for the main thread to finish executing before returning**. In Gio they are `Window.Perform`, `Window.Option`, `Window.Run` (when called after the window is created).

The reason is a three-way waiting loop. Take "The settings window is open, return to the main window and press ⌘+," as an example:

```
主窗口 goroutine   持有帧锁，在快捷键回调里调 settings.Raise()
                   └─ Gio Perform Wait for the main thread to execute it
主线程             正在给设置窗口派发事件（焦点变了），等设置窗口画完这一帧
设置窗口 goroutine  要画帧，等帧锁  ← 被主窗口 goroutine 持有
```

The three parties are waiting for each other and the interface is stuck. This bug has actually appeared. `Raise` and `Close` are now executed outside the lock using `go w.win.Perform(...)`. `win.Invalidate()` does not wait for the main thread and can be called within the lock.

When adding new window methods (setting position, sticking to top, changing title), handle them in the same way, and cover them in real window tests such as `ui/window/testdata/raise`, see [Testing](testing.md#real-window-test).

## Window lifecycle

- `window.Open` returns immediately; the native window is created asynchronously in its own goroutine.
- `Close` and `Raise` will wait until the first frame of the window is drawn before they are actually executed. Gio v0.10.3 has a bug on macOS: the native window will be closed before it is built, and the process will crash. Humans can't click so fast, Agent can, so Keel waits here. The regression test is `ui/window/testdata/reopen`.
- After closing the window (either the user clicks Close, or calls `w.Close()`), `OnClose` is executed within the lock, and `w.Closed()` becomes `true`. Closed windows cannot be reopened, reopen if necessary `window.Open`.
- After the last window is closed, the process calls `os.Exit(0)` to exit. The code after `window.Main()` in `main` will not be executed. It needs to be cleaned up and put into `OnClose`.

`native/notification` relies only on `native` and `native/internal/sys`, returns permission/delivery results via asynchronous completion callbacks, and does not reference the UI or Gio. macOS system calls are initiated from the main queue, and Go completion callbacks are executed in independent goroutines; UI writeback uses `core.Update`. kit.Notifier is accessed through the public NoticeSystemBackend interface, and the application layer adapts native/notification, and there is no direct reference between modules.

`ui/plot` is designed for custom charts, directly relying on core and theme (theme font), without relying on kit, el or window; applications can embed drawings into Widgets. The finished Chart/Plot remains in the kit, the existing plotting implementation has not yet been migrated to the public package.

`native/clipboard` relies on native/internal/sys, native/internal/wlclip and native; wlclip is the cgo binding for Linux Wayland to read the clipboard. It is packaged separately and is a program that only uses modules such as screenshots and shortcut keys, so it does not link to libwayland. The Wayland connection is taken from `window.Window.WaylandDisplay()` by the application and handed over to `clipboard.UseWaylandDisplay`. The two sets of modules still do not reference each other. macOS reads the clipboard snapshot from the main queue and then delivers it to the background goroutine. The application adapts the result to core.ClipboardData and accesses it through Input/TextArea.PasteReader; el is processed after core.Update and does not wait for the main thread in the frame lock. The UI module does not have a new native dependency.

The SVG icon is parsed by `oksvg` and `rasterx` in `ui/kit` and rasterized by physical pixels, and is still drawn by Gio; no new Keel module dependency is added. See [Icon](kit/icon.md) for entries, format subsets, and cache limits.
