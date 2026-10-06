# Testing

English | [简体中文](testing.zh-CN.md)

Most interface tests do not require opening a window: the `ui/internal/uitest` testing tool sends clicks, keys, and text through Gio's real input routing. The code executed by the component is exactly the same as that of the real window. `go test -race ./...` takes about 10 seconds to run on this machine.

## Interactive testing

Modules under `ui/` can reference it (it is in `internal` and cannot be referenced by external projects):

```go
func TestInputSubmit(t *testing.T) {
    var got string
    f := Input("").OnSubmit(func(s string) { got = s })   // in package kit
    h := uitest.New(el.Root(f)) // layout a frame
    h.Click(20, 10)      // Click to get focus
    h.Type("你好")        // Enter text
    h.Key("⏎", 0)        // Enter
    if got != "你好" {
        t.Fatalf("OnSubmit saw %q", got)
    }
}
```

| Function | Effect |
| --- | --- |
| `uitest.New(w)` | Place the component in the upper left corner (0,0) and lay out one frame |
| `uitest.NewFunc(fn)` | Use any frame function, such as `window` for package testing, use `w.layout` with root view and shortcut keys |
| `h.Click(x, y)` | Press and release the left button at (x, y), then draw a frame |
| `h.Key(name, mods)` | Press and release a key, then draw a frame. `name` is the key name of Gio: capital letters `"A"`, press Enter `"⏎"`, you can also use constants such as `key.NameReturn` |
| `h.Type(s)` | Insert text into the focused input box and then draw a frame |
| `h.Frame()` | Execute the `core.Update` queue and draw a frame |

Coordinate rules: The test viewport is 400×300, 1dp = 1 pixel. `uitest.New` Without root view, the component starts from (0,0); when tested with window layout, the content starts from (24,24).

Each operation such as `h.Click` is followed by a frame, but when the callback modifies the state and the redraw result is to be seen, call `h.Frame()` one more time.

## Module boundary testing

`internal/deps` checks that each module only references allowed packages (see [Architecture · Modules](architecture.md#modules)):

- The lower module refers to the upper layer (for example, `el` refers to `kit`), the same layer refers to each other, or `native/*` refers to `ui`, Gio, and the test fails;
- The module directory was added but not registered in the permission list, and the test failed.

It runs with `go test ./...`.

## What should be tested

- Number of callback triggers: trigger once per click, not trigger when disabled.
- Callback parameter: `OnChange` received the new value.
- The difference between program calls and user operations: `SetValue` does not trigger `OnChange`, typing does.
- Parsing function (shortcut key string): legal input and every illegal input.
- `native` package: parameter verification path (no permissions required, and you won’t actually move the mouse).

## Agent end-to-end testing

`cmd/keel-mcp` allows Agent to drive applications through MCP: read elements, click, input, scroll, and take screenshots. The application is rendered in memory without pop-up windows. For usage, see [Agent end-to-end test](automation.md).

There are two layers of related tests in the repository, both run with `go test ./...`:

- `ui/window/automation_test.go`: The in-process test automation mode itself, including scrolling, Tab moving focus, closing windows in callbacks, and reporting of disabled and checked statuses.
- `cmd/keel-mcp/main_test.go`: Start `keel-mcp` as MCP client, go through the multiwindow example completely, and test the connection to the application started by the user (`attach`). It will be the first to fail after changing the semantic information of a protocol, tool or component.

## Screenshot comparison

After changing the theme, spacing, and fonts, take screenshots before and after rendering to confirm that only the expected changes have occurred:

```sh
go run ./examples/hello -screenshot /tmp/before.png
# 修改代码
go run ./examples/hello -screenshot /tmp/after.png
cmp /tmp/before.png /tmp/after.png && echo 完全一致
```

Pure refactoring should output "exactly consistent". Screenshots are rendered off-screen by the GPU, and the results are stable on the same machine; pixels may be different on different machines and different system fonts. Do not submit screenshots as cross-machine benchmark files.

## Real window test

Some bugs only appear in real windows, such as deadlocks between multiple windows. `ui/window/desktop_test.go` runs in the child process `ui/window/testdata/raise`: open two windows and call `Raise` and `Close` in the interface code (holding the frame lock). If the process is not completed within 15 seconds, it is determined to be a deadlock.

```sh
KEEL_DESKTOP=1 go test -run RealWindows ./ui/window
```

This set of tests also covers the window being centered for the first time, immediately closed and reopened. Multi-monitor switching, window managers for each platform, and system preference changes still require manual verification.

Skipped by default as it will pop up on the screen and requires a graphical interface environment. You must run it once after changing the code related to windows and locks in `ui/window` and `ui/internal/loop`. When adding a new window method, add it to the `testdata/raise` step.

## Parts that require manual verification

The interfaceless test cannot cover these, so you need to change the relevant code and run it manually:

| Scenario | How to verify |
| --- | --- |
| Open, close, and bring to front the real window | `go run ./examples/multiwindow`, click "Open Settings Window" twice, there should be only one setting window |
| Exit after closing the last window | Close all windows and the process in the terminal should end |
| Window shortcut keys | Press ⌘+, | in the main window
| Global shortcut keys | `go run ./examples/hotkey`, switch to another application and press ⌘⇧K, the count will increase |
| Backend `core.Update` | The clock in the hotkey example ticks every second |
| Permissions, screenshots, synthetic input | Authorization required, manual test according to the instructions of [native capability](native.md) |
| Chinese input method | Use Pinyin to input in the input box, the candidate box is in the correct position, and the content is correct after it is displayed on the screen |

## Full component screenshot matrix

```sh
go run ./examples/components -matrix /tmp/keel-component-matrix
# 单个组件：窄窗口、1×、深色
go run ./examples/components -section button -width 320 -scale 1 -theme dark -screenshot /tmp/button.png
```

The matrix generates a total of eight first-frame screenshots in light/dark, 320/680dp width, 1×/2× for each registered component, and outputs a browsable `index.html`. Reconstruct the component for each case to avoid the layout cache from the previous case affecting the results. The large image is opened through the image link in the index; the screenshot only contains the current viewport, and the scrolled content and overlay still need to be interactively tested.

`window.Screenshot` Keep the default of 2×; use `ScreenshotAtScale` when additional scaling is required. Dimensions are expressed in dp, PNG dimensions are scaled and rounded to physical pixels. Illegal size or scaling returns an error.

The fixed display width in the component example must also be set to `MaxW(el.Full)`; otherwise, a cropped wide canvas will be tested instead of the narrow layout of the component. Use Wrap for parallel operations to retain the content width needed to verify horizontal scrolling. The Sidebar example is initially collapsed below 600dp and can still be expanded manually; the multi-column workspace of the Dock needs to be wide enough, and the 320dp screenshot is not regarded as a mobile layout commitment.

Example-level regressions of `cmd/keel-mcp` include: keyboard-only process for order filtering, details and deletion, new order; chat streaming output, code/form and interruption; settings page theme, search and language switching; Dock mobile panel, retaining input after switching, menu Esc and closing panel. `TestOrdersKeyboardOnly` Without clicking on the control, start with Mod+N and use Tab, arrow keys, Space, and Enter to complete the save.
