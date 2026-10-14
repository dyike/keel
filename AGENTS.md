# AGENTS.md

Keel is a desktop UI framework in pure Go: no HTML, CSS, JavaScript or
WebView. `ui/*` draws the interface on [Gio](https://gioui.org), with
Keel's patched copy in `gio`;
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
- **Gio and go-text are ours to change.** Keel's copies live in
  `gio` and `typesetting` as ordinary packages of
  Keel's single module, imported as `github.com/dyike/keel/gio/...`
  and `github.com/dyike/keel/typesetting/...` (never upstream
  `gioui.org/...` or go-text; `gioui.org/shader` stays an external module).
  Apps get them by requiring Keel; no `replace`, no nested `go.mod`. Fix or
  speed them up in place, keep upstream style there, and log every patch
  with its reason in `third_party/README.md` (both languages). A patch
  brings its own test. The copies never import other Keel packages
  (`internal/deps` checks this). The packager is Keel's copy of gogio in
  `internal/gogio`, which knows Gio's new path.
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

## Memory and rendering hard rules

- **No redraw for unchanged resident checks.** Use `core.UpdateIf` to compare
  visible state under the frame lock. Return true only for a visible change.
  An unchanged check must not restart idle reclamation. Do not animate a static
  status indicator just because a background task is running.
- **Measure the actual client.** Compare the same content, window pixel size
  and interaction. Record startup/loading, continuous scrolling, resizing and
  idle physical footprint and process peaks separately. RSS, Go heap size,
  headless tests and forced-GC snapshots do not establish total client memory.
  Frame submission timings do not establish end-to-end frame rate.
- **Preserve rendering and responsiveness.** Do not meet a memory target by
  reducing resolution, capping scroll frame rate or forcing GC every frame.
  Changes to drawable count, synchronization or reclamation require real-window
  scroll and resize checks; reducing triple buffering previously caused a
  resize wait. Preserve antialiasing, clipping, opacity, text selection and copy.
- **Shared tables, distinct font identities.** Cache immutable tables by file
  and TTC index, but keep named-instance metadata, variation coordinates and
  Face scratch state independent. Synchronize shared lazy caches. Match native
  fonts by exact face/instance identity, never by CoreText descriptor position.
- **Keep GPU fallback correct.** Reserve resources before selecting a shared
  texture path; unsupported devices, allocation failures and opacity layers
  must retain correct tint and coverage. Clip shortcuts require containment by
  every ancestor, including antialiased boundaries. Add pixel and fallback tests.
- **Report evidence without overstating it.** Separate observed allocations
  from unexplained driver memory, and idle success from peak success. Document
  first-frame stalls even when P95 improves. Record whether a fix is published
  or supplied only by a local workspace; never attribute local patches to an
  existing release. Revert experiments without reproducible benefit.

## cgo inventory

Desktop migration is not complete. Preserve Gio's UI, layout, GPU algorithms
and performance patches; replace native bindings, not the renderers. The
2026-10-12 darwin/arm64 dependency scan of `examples/hello` selects cgo in:

- Gio: `app`, `internal/gl`, `internal/cocoainit`, `gpu/internal/metal`
  (including pool/quads), and `gpu/headless`.
- Keel: `ui/theme/coretext_darwin.go` (CoreText/CoreGraphics).
- `runtime/cgo`, as a consequence of the selected bindings.

Linux also needs migration of X11/Wayland, xkbcommon, EGL/GL and Vulkan
bindings. Windows already builds without cgo. iOS/Android migration is deferred.
`CGO_ENABLED=0 go build ./examples/hello` currently fails on macOS in Gio GL;
`CGO_ENABLED=0 go build ./native/...` passes on the inspected macOS host.
Neither observation establishes cross-platform runtime support.
See [the delivery design](docs/desktop-cgo-free.md) for gates and sequencing.
Completed Keel binding migrations, each retaining behavior and tests:

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

These completed migrations do not include Keel's CoreText binding or Gio's
platform/GPU bindings. Since the single-module merge
([docs/single-module.md](docs/single-module.md)) applications receive the
patched Gio and go-text by requiring Keel alone.

## Commands

```sh
python3 scripts/check-components.py --offline   # standard component workflow; cached dependencies required
gofmt -l .                                       # must print nothing
go vet ./...
go test ./gio/... ./typesetting/...                        # the Gio and go-text copies' own patch tests
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
- Keep PRs to one change; explain what and why in the commit message.
