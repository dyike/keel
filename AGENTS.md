# AGENTS.md

Keel is a desktop UI framework in pure Go: no HTML, CSS, JavaScript or
WebView. `ui/*` draws the interface on Keel's own copy of
[Gio](https://gioui.org) in `third_party/gio`;
`native/*` provides system capabilities Gio lacks (permissions, screen
capture, synthetic input, global hotkeys, notifications, rich clipboard).
Targets: macOS, Windows, Linux, WebAssembly, iOS and Android.

**Read [docs/architecture.md](docs/architecture.md) before changing code.**
It describes the modules, the dependency rules, the frame lock and the
window lifecycle. [docs/testing.md](docs/testing.md) and
[docs/extending.md](docs/extending.md) cover tests and new components.

## Hard rules

- **No new cgo.** The goal is that everything builds with `CGO_ENABLED=0`,
  so any platform cross-compiles from any machine. Never add `import "C"`,
  `.m`, `.c` or `.h` files. Call native code at run time through
  [purego](https://github.com/ebitengine/purego) on macOS and Linux
  (`objc_msgSend`, `dlopen`/`dlsym`) and through `syscall` /
  `golang.org/x/sys/windows` on Windows. When touching an existing cgo file,
  prefer porting it to purego over extending it. The remaining cgo and the
  migration order are tracked in the "cgo inventory" below.
- **purego callbacks are scarce and never freed** (~2000 per process).
  Create them once per signature at startup (package-level, in `init` or a
  `sync.Once`) and route by user data (a token, a window id), never per
  window, call or request. Same for Objective-C classes registered at run
  time: register once, keep per-object state in a Go map keyed by the
  object pointer.
- **Hot paths call by address.** `purego.RegisterFunc` wrappers use
  reflection and allocate; keep them for float/struct signatures off hot
  paths. Per-frame or per-event calls use `purego.SyscallN` with cached
  selectors and function pointers.
- **Objective-C memory is manual.** `alloc`/`init`/`new`/`copy` results are
  owned and must be released; wrap code that creates temporary objects in an
  autorelease pool on a locked OS thread; copy blocks (`_Block_copy`) you
  call later and release them after use. Never pass Go pointers to native
  code that keeps them; use a token (integer handle) instead.
- **Gio and go-text are ours to change.** They live in `third_party/gio`
  and `third_party/typesetting` (import paths
  `github.com/dyike/keel/third_party/...`); never import `gioui.org/...`
  or `github.com/go-text/typesetting` (only `gioui.org/shader` stays
  external). Fix or speed them up in place, keep upstream style there, and
  log every patch with its reason in `third_party/README.md` (both
  languages) so it survives an upstream update. A patch brings its own test.
- **Dependencies only go down.** `internal/deps` enforces the module table
  in the architecture guide: `ui` and `native` never import each other,
  same-layer modules never import each other, heavy dependencies (chroma,
  `net/http`) stay opt-in. Change the table first, then the code.
- **Threading.** All rendering and callbacks run under one frame lock
  (`ui/internal/loop`). Code off the lock goes through `core.Update`.
  Never, while holding the lock, call a window method that waits for the
  main thread (Gio's `Window.Perform`, `Window.Option`, `Window.Run`):
  that deadlocks across windows. Native callbacks (AppKit, Win32, Wayland)
  must not take the frame lock synchronously; post from a goroutine
  (`go core.Update(...)`). Never block the main thread on work that needs it.
- **Performance is a feature.** Nothing polls: work is driven by native
  events or explicit wake-ups; an idle app uses 0% CPU. Avoid per-frame
  allocations, cache what the layout already measured, release scratch GPU
  and glyph resources. Measure before and after with benchmarks
  (`go test -bench`) or screenshots, not by intuition.

## cgo inventory

Gio itself needs cgo on macOS, iOS and Linux (`third_party/gio/app`,
`third_party/gio/internal/gl`), so `CGO_ENABLED=0` is only reachable once Keel
owns its platform layer: windows, surfaces and input through purego and
`syscall`, its own renderers (Metal, Direct3D 11, OpenGL, or the CPU) and
the system's text engines. Keel's own cgo was ported first, in this order,
each step keeping behavior and tests:

1. Done: `native/internal/sys` (darwin: permissions, displays, capture, input,
   hotkeys, clipboard, notifications, process). `native/*` does not depend
   on Gio, so after this step programs that use only `native/*` build with
   `CGO_ENABLED=0` on macOS. `objc_darwin.go` holds the shared helpers
   (`send`, `withPool`, `mainAsync`/`onMain` over libdispatch, blocks).
2. Done: `native/internal/wlclip` (Wayland clipboard; libwayland-client
   1.20+ loaded through purego, one `wl_proxy_add_dispatcher` callback).
3. Done: `ui/window/*_darwin.go` (icon, appearance, position, menu, title
   bar, motion, scroll, glass, native QA) and its real-window fixtures, via
   `ui/internal/appkit` (`ui` cannot import `native`, so it has its own
   helpers), using the `NSView` handle Gio reports in `app.AppKitViewEvent`.
4. Done: `ui/window/*_wayland.go` (scroll sources, xdg-activation) via
   `ui/internal/wayland`, on the connection Gio opened.
5. Done: `ui/window/scene_ios.go`, through purego's objc package.

Keel's own code has no cgo left: no `import "C"`, no `.c`/`.m`/`.h`.
What remains is Gio's (`third_party/gio/app` and its GPU code on macOS,
iOS and Linux). Removing it means owning the platform layer, as described
above, now done inside `third_party/gio`.

## Commands

```sh
gofmt -l .                                       # must print nothing
go vet ./...                                     # third_party too (Windows vet reports upstream unsafe.Pointer uses)
go test ./...                                    # includes internal/deps and convention checks
go test -race ./...
CGO_ENABLED=0 GOOS=windows go build ./...       # Windows is already cgo-free; keep it so
CGO_ENABLED=0 GOOS=linux go vet ./native/... ./cmd/...   # Linux ui/* still needs cgo (Gio)
KEEL_DESKTOP=1 go test -run RealWindows ./ui/window      # real windows; after any ui/window or loop change
go run ./examples/hello -screenshot /tmp/after.png       # pixel comparison for refactors
go run ./examples/components -matrix /tmp/keel-component-matrix
go run ./cmd/keel run ./examples/hello                   # the CLI's dev loop
```

## Conventions

- Docs are bilingual: every `docs/*.md`, README and module README has a
  `.zh-CN.md` twin. Change both in the same commit.
- A new component needs `docs/kit/<name>.md`, a section in
  `examples/components`, `uitest` interaction tests and window-level agent
  snapshot tests (`go test` checks these).
- New roles exposed to agents go into `ui/window/automation.go` and
  `docs/automation.md`.
- `work/`, `TODO.md` and `TODO.zh-CN.md` are local only (gitignored).
- Apps that imported upstream Gio run `keel migrate`; keep it in step when
  more paths move.
- Keep PRs to one change; explain what and why in the commit message.
