# Testing

English | [简体中文](testing.zh-CN.md)

Most interface tests do not require opening a window: the `ui/internal/uitest` testing tool sends clicks, keys, and text through Gio's real input routing. The code executed by the component is exactly the same as that of the real window. `go test -race ./...` takes about 10 seconds to run on this machine.

## Desktop no-cgo compilation gate

```sh
GOOS=windows GOARCH=amd64 bash scripts/check-desktop-cgo-free.sh
GOOS=linux GOARCH=arm64 bash scripts/check-desktop-cgo-free.sh native
```

The default `all` mode checks native capabilities and then the complete
module, which includes Keel's Gio and go-text. It disables cgo and workspaces,
rejects selected dependency errors/CgoFiles, and fails on any build error.
`native` mode checks only `native/*`; it does not accept desktop UI support.
The six-target CI matrix checks native capabilities on every push/PR and full
Windows builds. Full macOS/Linux gates are pending migration and can be run
with the workflow's `full` input; they are not silently allowed to fail.
Setting CC/CXX to nonexistent paths prevents accidental C compiler invocation,
but does not substitute for an isolated compiler/header-free release build.
Compilation is not runtime, IME, native GPU, packaging or performance acceptance;
see [the desktop delivery design](desktop-cgo-free.md).

The Ubuntu `distribution` job installs Linux graphics development libraries
before the scaffold tests. `TestNewProjectCompiles` vets a generated GUI project
on the host; on Linux it enables cgo and still needs Gio's Wayland, X11,
xkbcommon, EGL/GLES and Vulkan headers. This is separate from the `compile`
matrix, which disables cgo and uses nonexistent C/C++ compiler paths.

`python3 scripts/check-module-release.py` archives the working tree as a
fixture Keel release in a local module proxy, creates an app with `keel new`
that requires only Keel, checks that its Gio and go-text packages come from the
Keel module (never upstream `gioui.org` or go-text), compiles an external custom
widget and runs `go install` of the CLI from that release. It checks default,
offline and local scaffolds with space-containing paths, packages Windows
amd64/arm64 without cgo and WebAssembly, and on macOS runs a packaged identity
probe using the existing cgo backend. Dependencies use the configured Go proxy;
the synthetic release is served locally, with a fresh module cache and no workspace.

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
# Edit the code
go run ./examples/hello -screenshot /tmp/after.png
cmp /tmp/before.png /tmp/after.png && echo Identical
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
# One component: narrow viewport, 1x, dark theme
go run ./examples/components -section button -width 320 -scale 1 -theme dark -screenshot /tmp/button.png
```

The matrix generates a total of eight first-frame screenshots in light/dark, 320/680dp width, 1×/2× for each registered component, and outputs a browsable `index.html`. Reconstruct the component for each case to avoid the layout cache from the previous case affecting the results. The large image is opened through the image link in the index; the screenshot only contains the current viewport, and the scrolled content and overlay still need to be interactively tested.

`window.Screenshot` Keep the default of 2×; use `ScreenshotAtScale` when additional scaling is required. Dimensions are expressed in dp, PNG dimensions are scaled and rounded to physical pixels. Illegal size or scaling returns an error.

The fixed display width in the component example must also be set to `MaxW(el.Full)`; otherwise, a cropped wide canvas will be tested instead of the narrow layout of the component. Use Wrap for parallel operations to retain the content width needed to verify horizontal scrolling. The Sidebar example is initially collapsed below 600dp and can still be expanded manually; the multi-column workspace of the Dock needs to be wide enough, and the 320dp screenshot is not regarded as a mobile layout commitment.

Example-level regressions of `cmd/keel-mcp` include: keyboard-only process for order filtering, details and deletion, new order; chat streaming output, code/form and interruption; settings page theme, search and language switching; Dock mobile panel, retaining input after switching, menu Esc and closing panel. `TestOrdersKeyboardOnly` Without clicking on the control, start with Mod+N and use Tab, arrow keys, Space, and Enter to complete the save.

## Tray and residence regressions

`go test -race ./native/tray ./ui/internal/loop` checks menu snapshots, stale events, callback serialization, discarding callbacks after close and updates without windows. SNI/DBusMenu protocol tests exercise pure Go serialization and menu models without a desktop session.

```sh
KEEL_DESKTOP=1 go test -run RealWindows ./ui/window
go run ./examples/tray -smoke
```

Real-window fixtures start without windows, reopen after closing the last window and check exit callbacks. Tray smoke also creates, updates and removes the native icon. Actual menu clicks, appearance, high DPI, Explorer restart and Linux host restart require verification on the respective desktop; cross-compilation does not establish runtime acceptance.

Menu regressions cover frozen icon snapshots, atomic invalid updates, enabled/disabled template caches with zero warmed icon-layout allocations, fixed-time animation progress, reduced motion and stopped frame scheduling. Agent tests check menu-item labels, disabled/check states, action routing and controller cancellation. Windows target tests verify bitmap attachment, premultiplied pixels, aspect-ratio padding and cleanup on replacement/failure; cross-compiling these tests does not execute them. macOS fixtures create real AppKit images and popup tracking loops.

```sh
python3 scripts/check-components.py --offline --desktop --only checks interactive
```

## One-command component validation

```sh
python3 scripts/check-components.py
# Offline: requires dependencies and the Go toolchain already cached.
python3 scripts/check-components.py --offline
# CI without visible windows (GPU backend/development libraries still required).
python3 scripts/check-components.py --offline --headless
```

The standard workflow runs sequentially: runner regression tests, formatting, vet, all Go tests, race,
Windows no-cgo build and Linux native/CLI vet; then isolated component and
large-data benchmarks; the complete fixed-time GPU screenshot matrix; finally
`go run ./examples/components`, navigation to every registered scene and the
interaction smoke checks described below. `--desktop` also runs real-window
lifecycle fixtures during checks, including macOS menu image replacement and both native cancellation modes; it also runs the macOS native tray fixture and `examples/tray -smoke`. Desktop tray smoke requires a working notification area/SNI host. Python 3 and Go are required; Python uses only
the standard library. macOS needs Xcode command-line tools; Linux still needs
Gio's graphics development libraries and a suitable graphics/display backend.

Default results go to a new `work/component-checks/<timestamp>/` directory:
`summary.json`, the actual Go registry's `manifest.json`, per-command logs,
raw benchmarks and median `benchmarks.csv`, `matrix/index.html`, and gallery
interaction reports/screenshots. Failures stop the workflow and return a
nonzero exit code. Existing nonempty output directories are rejected. All
benchmark scenes run in fresh processes; checks, GPU rendering and interactions
never run concurrently with measurements.

```sh
# Run selected phases; default benchmarks use 3 × 200ms per gallery scene.
python3 scripts/check-components.py --only benchmarks --count 3 --benchtime 1s
python3 scripts/check-components.py --only matrix interactive --offline
# Save a baseline, then repeat after editing on the same host.
python3 scripts/check-components.py --output work/component-baseline
python3 scripts/check-components.py --baseline work/component-baseline
# Fast workflow smoke: 1x samples validate the runner, not performance gains.
python3 scripts/check-components.py --offline --count 1 --benchtime 1x --large-benchtime 1x
```

`--baseline` requires matching scene sets, fails on changed matrix PNGs, and
writes `comparison.csv` (positive percentages mean regression). Benchmark
percentages are review data, without automatic timing/allocation thresholds.
Compare on the same machine, fonts, Go version and benchmark settings. The
matrix clock is fixed through a generated Go overlay; production code is not
modified. Intentional visual changes require a reviewed new baseline.

The gallery smoke covers checkbox/disabled state, switch, Unicode/multiline
input, dialog save/Escape/reopening, searchable and 10,000-option selects,
100,000-item virtual lists, table sorting/scrolling, calendar keyboard navigation
past disabled Sundays, date-picker selection after reopening, and application-menu icons, disabled items, action routing and Escape. Calendar checks
read displayed dates rather than hard-coding a month. The automation operates
Gio's shadow input router while the default app has a visible native window;
`--headless` suppresses that native window. Neither mode verifies an OS input
method candidate panel or native mouse delivery. GPU presentation latency and
animation performance require separate profiling.

`--offline` disables Go module/VCS/toolchain downloads; an uncached dependency
fails explicitly. It does not turn screenshots into a no-cgo build or disable
optional network-image requests made by the example.

## Component frame allocation benchmarks

The gallery benchmark discovers every registered scene, including controls,
inputs, overlays, navigation, charts and virtualized content:

```sh
go test ./examples/components -run '^$' -bench '^BenchmarkComponentFrame$' -benchmem -benchtime=200ms
# Isolate one scene; the second anchored expression matches the whole name.
go test ./examples/components -run '^$' -bench '^BenchmarkComponentFrame$/^calendar$' -benchmem -benchtime=1s -count=3
go test ./ui/kit -run '^$' -bench '^BenchmarkTableFrame300k$' -benchmem -benchtime=1s -count=3
go test ./ui/kit -run '^$' -bench '^BenchmarkCodeEditorPlainLargeFileFrame$' -benchmem -benchtime=1s -count=3
```

`BenchmarkComponentFrame` warms each scene at 680×1040dp, 1× scale, uses a
real Gio input router and fixes animation time. It measures CPU layout and
input registration, without GPU submission or background update delivery.
Closed-overlay first frames do not measure open menus/dialogs, and a static
scene does not measure scrolling, typing or animation latency. Run interactive
regressions and the GPU screenshot matrix alongside it.

Steady-frame allocation count is a regression metric. The ordinary test suite
keeps ceilings for the label (35), list (250), button (260), calendar (1,050),
table (550) and input (500) gallery scenes after three warm-up frames. These ceilings are guards, not performance
targets; benchmark bytes and counts before and after every allocation change.

Set `KEEL_FRAME_STATS=all` to log every native-window frame with layout and
submission time, allocation count and bytes observed for the process, and GC
cycles completed. `KEEL_FRAME_STATS=1` logs frames over 8ms; another
non-negative number is the threshold in milliseconds. Leave it unset in
normal runs: reading runtime metrics is diagnostic work. The allocation
counters cover the whole process during the frame, so concurrent background
work can contribute.

For comparisons, build baseline and changed test binaries with the same
benchmark code, then run each scene in a fresh process. Optional highlighting
and image fetching may otherwise spill into the following scene's counters.
A Go overlay can select baseline production files without reverting changes.
The plain-editor benchmark leaves its language unset; the existing highlighted
large-file benchmark still includes whole-document background highlighting.

### Paint-path allocation checks, 2026-10-15

The path-space stroke cache has separate renderer microbenchmarks:

```sh
go test ./gio/gpu -run '^$' -bench 'Benchmark(Rotating|Changing)StrokeVertices' -benchmem -benchtime=500ms -count=3
```

Apple M4, darwin/arm64, Go 1.27.1, medians of three isolated samples:

| Cubic stroke vertex build | Before time | After time | Before → after bytes | Before → after allocations |
| --- | ---: | ---: | ---: | ---: |
| Same path, rotating transform | 17.88 µs | 5.29 µs | 42,648 → 0 | 103 → 0 |
| Changing path content | 18.53 µs | 16.44 µs | 42,939 → 8,232 | 103 → 2 |

The first case warms storage; the second changes a control coordinate every
frame and includes cache misses and ownership copies. Scratch buffers do not
alias cached geometry or recorded Ops. The cache retains at most 128 outlines
and 1 MiB of path/quad data, expires unused entries and bypasses oversized paths.
It is released on renderer trim/release. Affine vertices and bounds match the
original stroker exactly, including shear and reflection. Warm/fresh GPU pixels
match under rotation, rounded ancestor clips and opacity; all 792 fixed-time
component screenshots match main byte-for-byte.

This is CPU vertex construction, not a whole frame, end-to-end FPS or native
memory result. Native scrolling/resizing and four-phase footprint/GC comparison
remain pending an unlocked desktop. Shaders, drawable count, GC and idle
reclamation are unchanged.

`BenchmarkCategoryFrame` uses the actual `-section` compositions at 1360×1720
pixels (680×860dp, 2×), a real input router and animation time advancing at
60 Hz. Run each category in a fresh process, without other benchmarks running:

```sh
go test ./examples/components -run '^$' -bench '^BenchmarkCategoryFrame$/^controls$' -benchmem -benchtime=1s -count=3
```

On Apple M4, darwin/arm64, Go 1.27.1, three one-second samples per category
gave these medians. The baseline already contains the node arena. Both binaries
use the same benchmark; the changed binary adds bounded icon variants, borrowed
glyph scratch and retained views in the status-marker/theme demos.

| Category | Before B/frame | After B/frame | Before allocs/frame | After allocs/frame |
| --- | ---: | ---: | ---: | ---: |
| controls | 334,888 | 152,289 | 2,169 | 2,050 |
| inputs | 484,933 | 183,947 | 2,176 | 2,021 |
| overlays | 175,744 | 38,832 | 558 | 487 |
| data | 726,864 | 599,584 | 3,262 | 3,191 |
| shell | 384,267 | 94,172 | 1,339 | 1,182 |

This measures CPU layout and paint-op generation, not GPU execution, native
footprint, scrolling latency or end-to-end FPS. No GC or drawable policy changed.
The icon cache trades bounded retention (eight variants and 1 MiB of pixels per
icon, except one oversized current image) for fewer rasterizations and uploads.

Set `KEEL_STATE_STATS=1` to log state creation/removal per root and frame, with
the first eight new element IDs, roles and names. Stable root IDs distinguish
recreated roots from unstable element keys. Logging is opt-in and never requests
a redraw; new visible rows legitimately create state. This identified 21 and four
state replacements per warmed controls frame in the status-marker and theme
demos respectively: both constructed pointer-identified views during rendering.
Retaining those views removed the churn in a follow-up trace. The theme inputs
now also retain their editing state across renders.

### Further allocation reductions, 2026-10-15

On Apple M4, darwin/arm64, Go 1.27.1, the next pass used the completed node
arena as its baseline. All 99 gallery scenes reduced both allocation bytes
and counts: median per-scene reductions were 42.2% and 44.2%. The unweighted
mean fell from 57,243 to 35,420 B/frame. Each scene ran in its own process,
three 200ms samples; neither version tuned GC settings. These archived local
measurements used the temporary three-drawable and idle-reclaim-disabled
workspace experiments. Neither policy change is part of the arena PR: Metal
keeps main's policy and idle heap reclamation remains enabled by default.

| Scene | Bytes/frame, before → after | Allocations/frame, before → after |
| --- | ---: | ---: |
| Label | 9,808 → 3,336 | 51 → 26 |
| Button | 106,583 → 64,989 | 463 → 202 |
| Calendar | 413,218 → 190,688 | 2,344 → 842 |
| Table | 149,992 → 79,336 | 817 → 424 |
| List | 58,464 → 29,848 | 374 → 200 |
| Input | 178,293 → 122,612 | 806 → 400 |

The changes inline optional style values, allocate a private visual style
only when a callback runs, right-size independent handle batches, reuse
frame-local measurement entries and editor filters, and cache immutable
document presentations. Button visual callbacks share a bounded value-only
cache; application callbacks and frame contexts are never cached there.
Calendar grids share one disabled-style callback rather than one per day.
Styles still distinguish absent values from transparent colors, normal font
weight and zero offsets. Callback copies and pointer handles are never
overwritten in a later frame.

The 300,000-row scrolling table, three 300ms samples, fell from 368,397 to
297,361 B/frame and 1,622 to 1,197 allocations/frame; median layout time was
525µs → 506µs. Gallery layout time improved by a per-scene median of 5.0%,
but small timing changes remain host-load sensitive. Initial short-run
regressions in Bubble and the large plain editor did not persist in five 1s
samples per version: 57.2µs → 55.0µs and 2.50ms → 2.33ms, respectively.
These CPU-frame measurements do not establish end-to-end FPS, physical
footprint, GPU memory or startup latency. The pointer API still needs fresh
checked handles, and shaping, widget construction and application callbacks
still allocate; this is not a zero-allocation result.

The first screenshot pass caught a default text-color regression in Themed
subtrees: an inline value captured the global color too early. Default colors
now resolve against the paint-time palette; explicit transparent colors remain
explicit. A root-level regression test covers both cases. After that fix,
all 792 matrix PNGs matched byte-for-byte and all 99 scene navigations and
interaction groups passed. Native-window lifecycle, idle resume and 60-resize
checks also passed. Tests additionally cover zero-valued style inheritance,
callback-copy identity, stale pointer handles, hash collisions, released text
references, and editor content/object-mode changes. The arena does not add
forced collections or alter idle heap reclamation thresholds.

### Node arena measurements, 2026-10-15

On Apple M4, darwin/arm64, Go 1.27.1, all 99 gallery scenes reduced both
bytes and allocation counts per warmed frame. Each scene ran in its own
process, three 200ms samples. The baseline already includes direct semantic
ops, cached focus filters, cell IDs and visible columns. The per-scene median
reductions from the arena were 42.4% in bytes and 37.7% in counts; the
unweighted mean bytes per scene fell from 106,662 to 57,243.

| Scene | Bytes/frame, before → after | Allocations/frame, before → after |
| --- | ---: | ---: |
| Button | 181,059 → 106,583 | 699 → 463 |
| Calendar | 918,307 → 413,218 | 3,861 → 2,344 |
| Table | 558,217 → 149,992 | 2,056 → 817 |
| List | 162,097 → 58,464 | 679 → 374 |
| Input | 218,070 → 178,293 | 935 → 806 |

The 300,000-row scrolling table benchmark, three 300ms samples per version,
fell from 1,121,028 to 368,397 B/frame and 3,930 to 1,622 allocations/frame.
Its median layout time was 590µs → 525µs. Short timing regressions in Radio
and Button did not persist in five 1s samples per version: medians were
30.7µs → 29.8µs and 152.6µs → 147.6µs, respectively. Small timing changes
remain sensitive to host load.

All 792 GPU matrix PNGs matched byte-for-byte. The gallery smoke passed
all 99 scene navigations and its input, overlay, list, table and calendar
interactions. Arena tests cover stale handles before and after slot reuse,
same-frame rerenders, root ownership, cache escape, dropped-node references,
one-shot roots and nested roots. A warmed 101-node construction allocates at
most six small handle batches; complete component frames still allocate for
styles, callbacks and text work. These CPU-frame results do not measure
native-window physical footprint, GPU memory or startup latency.

Native-window checks also passed (`RealWindows`, `TestMetalIdleWindow` and
`TestMetalResizeWindow`). Three idle resumes took 28.8–31.8ms; the slowest
of 60 native resizes took 27.4ms. These checks do not measure end-to-end FPS.

### Earlier allocation measurements, 2026-10-13

On 2026-10-13, darwin/arm64, Apple M5 Pro, all 99 gallery scenes reduced both
allocated bytes and allocation counts per frame. The median per-scene reductions
were 8.7% and 16.2%, respectively (three runs per scene, 200ms each, separate
processes). The five apparent timing regressions were rechecked with five 1s
runs; none persisted. Small timing differences remain sensitive to host load.
These are same-host observations, not cross-machine performance thresholds.

| Scene | Bytes/op, before → after | Allocations/op, before → after |
| --- | ---: | ---: |
| Calendar | 1,139,345 → 959,473 | 7,085 → 5,184 |
| Form | 92,141 → 82,192 | 467 → 348 |
| Chart | 292,385 → 267,760 | 1,445 → 1,229 |
| Tree | 118,003 → 104,320 | 644 → 520 |
| VirtualList | 142,285 → 111,931 | 645 → 452 |
| MessageScroller | 1,057,741 → 682,160 | 2,385 → 1,590 |

Separate large-data benchmarks use three 1s runs. Both table versions used
identical CPU/memory profiling flags; neither plain-editor version used profiles:

| Scenario | Time/op, before → after | Bytes/op, before → after | Allocations/op, before → after |
| --- | ---: | ---: | ---: |
| 300,000-row table scroll | 702µs → 608µs | 1,515,575 → 1,138,768 | 6,524 → 4,522 |
| 200,000-line plain editor cursor move | 2,594µs → 2,319µs | 540,302 → 250,920 | 1,745 → 1,077 |

The changes allocate layout memos only for cached nodes, create Gio editors only
for inputs, skip unused keymap/layer indexes, share immutable keymap ancestry,
use stack scratch for small layout child lists, reserve child capacity once,
and hash state IDs without concatenating strings. State hashing remains
compatible with existing identities. Component and calendar IDs write directly
into byte scratch without changing their format. Microbenchmarks reduce ordinary
ID generation from 43ns to 19ns and calendar cell IDs from 94ns/3 allocations to
43ns/1 allocation. Editor glyph scratch retains at most
4,096 glyphs; larger buffers are discarded.

Regressions cover cache width/DPI/inherited font changes, independent action
scopes, hints without handlers, stable editor focus tags and Unicode input.
The 792-image GPU matrix remains byte-identical to the baseline; a test-only
overlay fixed screenshot time in both runs. Running `go run ./examples/components`
also covered navigation to all 99 scenes, input, checkbox/switch state, dialog
focus/save/reopening, 10,000-option search, 100,000-item virtual scrolling and
table sorting, calendar keyboard navigation past disabled dates and retained
date-picker selection after reopening. Gallery category checks prevent registered scenes from being
omitted from the sidebar.
