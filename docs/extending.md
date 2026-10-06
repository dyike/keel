# Extending Keel

English | [简体中文](extending.zh-CN.md)

Before adding something, determine which existing module it belongs to. In most cases, a file is added to an existing module; for the standards for creating new modules, see [Architecture · When to create a new module](architecture.md#when-to-create-a-new-module).

| What you want to add | Where to put | Examples |
| --- | --- | --- |
| Interfaces and business components in applications | Not included in the library: Use `ui/el` to write functions or views, see [Elements and Views](el.md#make-reusable-components) | Order cards, toolbars |
| Generic element capabilities | `ui/el` | New styling methods, layout properties, element types |
| Common components | `ui/kit/` New files | `select.go`, `chart.go`, `dock.go` |
| Adjustable parameters such as color, font size, etc. | `ui/theme/theme.go` | Value used by more than two components |
| Related to the window itself | `ui/window/` | Window position, pinned |
| System capabilities, useful even without opening a window | New module under `native/` | `native/clipboard`, `native/notify`, `native/tray` |
| There is only one interface fragment used on the page | Not included in the library, the business code is written as a function that returns `el.Element` | Consider moving it in after using it for the second time |

## Add new components

Common components are placed in `ui/kit`, one component is a file, and views are written in `ui/el`. The complete specification and acceptance list are in [Component Development Specification](component-development.md), here we only go through the skeleton. Take a counter as an example (the kit already has `NumberInput`, this is just to illustrate the structure):

```go
// CounterView shows a number between − and + buttons.
type CounterView struct {
    value    int
    disabled bool
    onChange func(int)
}

func Counter(value int) *CounterView { return &CounterView{value: value} }

func (v *CounterView) Value() int         { return v.value }
func (v *CounterView) SetValue(n int)     { v.value = n } // Program assignment does not trigger callback
func (v *CounterView) SetDisabled(d bool) { v.disabled = d }
func (v *CounterView) OnChange(fn func(int)) *CounterView {
    v.onChange = fn
    return v
}

func (v *CounterView) set(n int) {
    v.value = n
    if v.onChange != nil {
        v.onChange(n) // Only triggered by user actions
    }
}

func (v *CounterView) Render(cx *el.Context) el.Element {
    minus := Button("−", func() { v.set(v.value - 1) }).Variant(ButtonSecondary)
    plus := Button("+", func() { v.set(v.value + 1) }).Variant(ButtonSecondary)
    minus.SetDisabled(v.disabled)
    plus.SetDisabled(v.disabled)
    return el.Div().Row().Gap(8).Items(el.Center).Child(
        minus.Render(cx),
        el.Text(strconv.Itoa(v.value)).TextColor(theme.Text), // Color is read at Render time
        plus.Render(cx),
    )
}
```

Key points:

1. **Constructor `Xxx(...)` returns `*XxxView`. ** Business status is stored in the structure; interaction status such as hover, press, and focus are saved by el according to the element position without declaration.
2. **Value components are provided `Value`, `SetValue`, `OnChange`, `SetDisabled`.** Programmatic assignment does not trigger callbacks.
3. **The color and framework text are read from `theme` and `locale` during Render**, and are not saved during construction, nor hard-coded Chinese.
4. **The capabilities that el lacks are added to el** first (focus, timing, overlay, dragging), and the Gio input routing is not written directly in the component.

Then complete the supporting files, `ui/kit/conventions_test.go` will check which one is missing:

- `ui/kit/counter_test.go`: Use `page(v)`, `click(t, h, "name")` and other auxiliary functions (in `kit_test.go`) to take the real input route;
- `ui/window/kit_*_test.go`: The role, name, value, and status in the Agent snapshot are correct; the container role that needs to list the sub-elements separately is added to `containerRoles`, and the table of [Agent end-to-end test](automation.md#element) is added;
- `docs/kit/counter.md`, and registered in the corresponding category of [component reference](kit.md);
- `examples/components/counter.go`, sign up for `-section counter`.

## Add native capabilities

Take "Read Clipboard Text" as an example and walk through it. This example was compiled and tested when writing the document, and it was not merged into the repository.

**Step one: Write a C function.**Append to `native/internal/sys/sys_darwin.m`:

```objc
int keel_clipboard_text(char **out){
 @autoreleasepool {
 NSString *s=[[NSPasteboard generalPasteboard] stringForType:NSPasteboardTypeString];
 if(!s){*out=NULL;return 0;}
 *out=strdup(s.UTF8String);return *out?0:100;
 }
}
```

Return value convention: `0` is successful, `1` has no permission, `2` is not supported, `3` parameters are wrong, `4` has wrong thread, `5` times out, `6` conflicts, other values are failures. `status()` in `sys_darwin.go` to convert them to `native.Err*`. When a new error category is needed, add both sides together.

Declare in `sys_darwin.h`:

```c
int keel_clipboard_text(char **out);
```

**Step 2: Package it into a Go function.** `sys_darwin.go`:

```go
func ClipboardText() (string, error) {
    var p *C.char
    if err := status(C.keel_clipboard_text(&p)); err != nil {
        return "", err
    }
    if p == nil {
        return "", nil
    }
    defer C.free(unsafe.Pointer(p))
    return C.GoString(p), nil
}
```

Memory allocated by C is freed by Go side `C.free`. Don't give Go pointers to C for long-term storage.

**Step 3: Other platforms.** `sys_windows.go` (Win32) and `sys_linux.go` (X11) must also have functions with the same name. If it cannot be done temporarily, it will return to `native.ErrUnsupported` first. `sys_other.go` covers the rest of the platforms:

```go
func ClipboardText() (string, error) { return "", native.ErrUnsupported }
```

If any file is missing, the platform will not be able to compile it. Check one by one:

```sh
GOOS=windows go vet ./native/...
CGO_ENABLED=0 GOOS=linux go vet ./native/...
CGO_ENABLED=0 GOOS=freebsd go vet ./native/...
```

**Step 4: Make the package public.** Create a new `native/clipboard/clipboard.go`. Parameter verification is placed in this layer. The `sys` layer only does translation:

```go
// Package clipboard reads the system clipboard.
package clipboard

import "github.com/dyike/keel/native/internal/sys"

// Text returns the clipboard's plain text, or "" when it holds none.
func Text() (string, error) { return sys.ClipboardText() }
```

**Step 5: Register module boundaries.** Add a row to the `allowed` table of `internal/deps/deps_test.go`:

```go
"native/clipboard": {"native", "native/internal/sys"},
```

Without registration, `TestEveryModuleIsListed` will fail. This step forces you to think clearly: who the new module depends on, and whether it is really independent.

**Step 6: Write the document.** Write clearly in `native/clipboard/README.md` what it does, what it depends on, and how to use it alone (copy the README format of other modules); add a section each in the table of [native/README.md](../native/README.md) and [native capability](native.md) to clearly indicate what permissions are required, whether it will block, and which thread it can be called from.

For system APIs involving the main thread (most UI classes in AppKit), please note: Gio's event loop occupies the main thread, and `dispatch_sync(dispatch_get_main_queue(), ...)` is used to cut through it in C code (refer to the existing `onMain`). But if the caller itself is on the main thread, `dispatch_sync` will deadlock, so `[NSThread isMainThread]` is judged first.

## New example

`examples/<name>/main.go`, an example demonstrates one thing. To be able to produce screenshots, please refer to `examples/hello` to support the `-screenshot` parameter.

## Check before submission

```sh
gofmt -l .                          # 应无输出
go vet ./...
CGO_ENABLED=0 GOOS=linux go vet ./native/... ./cmd/...   # 桩函数齐全
go test -race ./...                 # 包括模块边界检查
go run ./examples/hello -screenshot /tmp/after.png   # 改了样式时对比截图
```
