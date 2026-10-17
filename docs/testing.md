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

### Coverage decode and icon callback reuse (local follow-up)

The disk-index decoder interns identical rune coverage before allocating a
second slice. Its table belongs to one read; full page/bitmap comparisons
check hash hits, and decoded data never aliases the reused record buffer.
Named-instance metadata stays independent and version-7 caches remain valid.
IconView reuses one immutable renderer/color snapshot. Replacing the cached
renderer never changes an older element's captured color; copied views get
their own renderer. Element handles and Context.Cache lifetimes are unchanged.
Local mygo's retained icon raster callback is the reference, not borrowed
frame scratch or a change to Keel's pointer API.

```sh
go test ./typesetting/fontscan -run 'Test(IndexDecode|RuneSetDecode)'
go test ./ui/kit -run 'TestIcon(RenderReuses|CachedElement)'
go test ./typesetting/fontscan -run '^$' -bench BenchmarkFontIndexDeserialize -benchmem -count=3
go test ./ui/kit -run '^$' -bench 'Benchmark(IconRender|RetainedIconFrame)' -benchmem -count=3
```

Apple M4, Go 1.27.1, medians of three 500ms samples, against main `afe213c`:
the synthetic 64-file/four-face/512-page shared index changes from
4,959,613 B / 888 allocations to 259,418 B / 633; CPU time is 2.01ms → 1.48ms.
With every coverage distinct, the lookup table costs 4,959,578 → 4,996,580 B,
886 → 899 allocations and 2.05ms → 2.29ms. Keep this worst-case cost visible.
Warm retained IconView.Render changes from 944 B / 3 allocations to 920 B / 2;
fresh views change from 992 B / 4 to 984 B / 4. These include independent
element/node storage, not just callbacks; whole frames are not zero-allocation.
All 792 fixed-time GPU screenshots are byte-identical. Three short-run gallery
byte regressions did not persist in five interleaved fresh-process rechecks.
Microbenchmarks do not establish native footprint, peak memory or FPS.

Three fresh interleaved native card clients per version, 800×600dp at native
2× (1600×1200px), compare main `afe213c`, this unpublished local follow-up,
and current local mygo `fefd36a5`. The existing v7 font index is retained.
Scroll/resize last 10s each, idle 12s; footprint sampling is 100ms (50ms at
startup), with idle values from the last 3s. Only the fixture disables idle
active reclamation to match mygo; GOGC/GOMEMLIMIT and library defaults are
unchanged. Forced GC is zero; profiles/screenshots run separately.

| Native three-run median | Before | Local follow-up | Mygo |
| --- | ---: | ---: | ---: |
| Startup TotalAlloc / MiB | 26.74 | 21.62 | 20.03 |
| Startup HeapSys / MiB | 27.22 | 19.25 | 23.22 |
| Scroll allocation / MiB/s | 4.88 | 4.07 | 45.33 |
| Scroll allocations / draw call | 693 | 93 | 2,568 |
| Scroll GC cycles | 7 | 6 | 38 |
| Resize allocation / MiB/s | 2.33 | 1.93 | 15.30 |
| Resize GC cycles | 3 | 3 | 12 |
| Idle footprint, final 3s / MiB | 58.80 | 57.24 | 73.30 |
| Kernel lifetime peak / MiB | 159.10 | 164.06 | 163.67 |

Peak does not improve: before 157.86–162.45 MiB, after 158.10–165.38 MiB;
all processes peak during resizing. The observed median increases about 5 MiB.
Three-run footprint ranges overlap, and small idle differences are not an
established benefit. Startup GC can increase in an individual run despite
less allocation. Mygo's accessibility work materially affects its allocation
rate; this is not a framework ranking or an end-to-end FPS result. No controls
category peak improvement is established. A separate profile reports no
IconView.Render self-allocation in scrolling; pointer frameHandles remain
about 57% of sampled bytes. Startup RuneSet decoding is about 1 MiB sampled,
not the old approximately 6 MiB. Sampling/GC lag prevents treating profile
totals as exact phase allocation or live physical memory.

### Paint-allocation follow-up (2026-10-15)

The local follow-up keeps the pointer-color API and element lifetimes. Material
labels use an additive value-color layout method; selectable labels own their
color. Anchors are cleared at paint start (after Render can read LastSize), with
maps above 4,096 entries discarded on the following paint. Stroke admission
keeps only hashes until a key occurs in consecutive frames. Cache entries never
borrow mutable Ops or stroke scratch. No drawable, resolution, GC or idle-reclaim
default changes are included. Local mygo's bounded scratch and explicit reset
ownership are the reference, not its public handle API.

Isolated 99-scene benchmarks, three samples of 50 warmed frames per scene:
1,402,961 → 1,086,720 B/frame and 11,144 → 9,876 allocations/frame summed over
the scenarios; none regressed. These measure layout/op generation, not GPU
execution. Warm ordinary material labels change from 128 B/1 allocation to
zero; selectable labels change from 133 B/2 to 5 B/1. Reusing 512 anchors changes
111,288 B/17 allocations to zero. The 792 fixed-time GPU screenshots are identical.

The native controls comparison uses five fresh interleaved processes per build:
main `3b88f26`, font+stroke (`af8d7dd` + `b45e325`, now main `6ab9dae`) and the
local follow-up. At 800×600 native 1×, scroll and resize last 10s each, followed
by 12s without input; controls animations continue in that last phase.

| Controls, five-run median | Main before | Font+stroke | Local follow-up |
| --- | ---: | ---: | ---: |
| Startup TotalAlloc / MiB | 245.54 | 131.75 | 123.01 |
| Scroll allocation / MiB/s | 22.88 | 14.40 | 10.51 |
| Scroll GC cycles | 9 | 6 | 4 |
| Resize allocation / MiB/s | 86.80 | 39.53 | 28.41 |
| Resize GC cycles | 32 | 14 | 10 |
| No-input footprint, final 3s / MiB | 162.38 | 162.33 | 162.47 |
| Kernel lifetime peak / MiB | 187.39 | 175.72 | 188.36 |

Peak ranges are 180.60–193.42, 170.08–187.38 and 175.70–209.03 MiB.
The follow-up reduces garbage but does not reduce physical peak: its median
is higher than font+stroke alone. All five follow-up processes first reach
their lifetime peak during initial scrolling, not resizing. In contrast, all
five old-main processes reach it during startup. Phase timing depends on the
client and natural GC; a previous run's resize peak is not a universal cause.

Existing category snapshots show approximately 51 MiB unmapped graphics,
8.3 MiB IOAccelerator graphics and 5.7 MiB IOSurface in all three controls
builds. Startup Untagged medians are about 70 → 87 → 98 MiB, while no-input
values return to about 73–74 MiB. Untagged is not synonymous with Go heap.
However, the follow-up's run 1 trace shows the heap reaching 101 MB before
natural GC at 3.501s (107 MB goal), then 29 MB live. This is evidence against
attributing the peak variation solely to GPU resources or calling it a leak.
Category tools round values and snapshots need not coincide with a brief peak.

Physical footprint is sampled every 100ms (50ms startup), separately from the
kernel lifetime peak. Every formal phase has zero forced GC. Only the fixture
disables Keel's active idle reclamation to match mygo; GOGC/GOMEMLIMIT are unset.
Builds, tests, profiles and extra category sampling are excluded from formal
runs. A warmed font index is used. The latest merged font/stroke changes are
not attributed to an existing release; this follow-up remains local.

The matching 600-card client uses three interleaved runs per build, including
local mygo `7e868828`; all forced-GC counts remain zero:

| Cards, three-run median | Font+stroke | Local follow-up | Local mygo |
| --- | ---: | ---: | ---: |
| Startup TotalAlloc / MiB | 26.65 | 26.58 | 19.20 |
| Startup footprint / MiB | 47.13 | 47.66 | 54.52 |
| Scroll footprint / MiB | 105.17 | 103.31 | 124.91 |
| Resize footprint / MiB | 107.47 | 107.63 | 126.25 |
| Idle footprint, final 3s / MiB | 50.10 | 50.19 | 63.88 |
| Kernel lifetime peak / MiB | 115.97 | 115.83 | 131.33 |
| Scroll allocation / MiB/s | 4.91 | 4.88 | 45.21 |
| Resize allocation / MiB/s | 2.29 | 2.42 | 15.42 |
| Scroll / resize GC cycles | 7 / 3 | 7 / 3 | 38 / 12 |

All card processes peak during resizing. Card peak ranges overlap:
114.42–116.08 versus 115.56–116.74 MiB for the two Keel builds. The card
client barely exercises these controls hotspots; its resize allocation rate
does not improve. Resize frame-event counts vary despite the same 20 size
requests per second. These are not end-to-end FPS measurements. Mygo's local
macOS accessibility work materially affects its allocation rate; do not
generalize this single client into a framework-wide ranking.

Separate 64 KiB-sampled allocs profiles show no LabelStyle self-allocation or
stroke-cache put hotspot after the change. Text shaping plus toGioGlyphs account
for about 36% of the controls resize sample; frameHandles account for about
57% of card scrolling. Controls scroll writeVertCache accounts for 15.18 MiB
but only four sampled growth allocations, not a per-frame leak. Profile
accounting can lag natural GC by about two cycles; use MemStats differences
for phase totals and treat these shares as approximate.

A separate single-pair frame-statistics check (not the formal memory series)
records first-frame layout+submission 124.67 → 141.49ms. After excluding the
first 20 frames, P95 is 8.56 → 7.57ms. A first-frame stall remains and this
pair does not establish presentation FPS or a repeatable first-frame change.
Two-frame admission may recompute newly stable geometry twice; hash collisions
can prevent admission. These are cache-hit/CPU tradeoffs, not safety shortcuts.

### Font startup allocation checks (2026-10-15)

Run the font initialization and index microbenchmarks independently:

```sh
go test ./typesetting/font -run '^$' -bench BenchmarkGlyfInitialization -benchmem -benchtime=500ms -count=3
go test ./typesetting/fontscan -run '^$' -bench 'BenchmarkFontIndex(RecordRead|Serialize)' -benchmem -benchtime=200ms -count=3
```

Apple M4, darwin/arm64, Go 1.27.1, median of three isolated samples:

| Operation | Before bytes | After bytes | Before allocations | After allocations |
| --- | ---: | ---: | ---: | ---: |
| Go Regular glyf initialization | 191,033 | 30,720 | 2,125 | 2 |
| Read a 3,143,610-byte index record | 8,388,283 | 3,178,546 | 17 | 3 |
| Encode 64 synthetic files, four 512-page coverages each | 28,489,518 | 6,009,804 | 293 | 278 |

Unchanged font files also skip index serialization entirely. Cold/updated
indexes retain the version-7 format and stream one file at a time. Headers,
variation point counts, outlines and metrics are compared with the eager
parser; malformed-record fuzzing checks matching validation results.
All 792 fixed-time component screenshots matched main byte-for-byte.

The unlocked native follow-up used 600 matching nonvirtualized cards at
800×600 pixels, native 1×, three interleaved fresh processes per version.
The baseline is main `3b88f26`; the local changed build combines font patch
`af8d7dd` with stroke patch `b45e325` through a Go overlay. Those measurements
preceded publication; the patches are now merged as PRs #35 and #36, not
attributed to a released version. The font index already existed. Local mygo
is `7e868828`.

| Native metric, three-run median | Keel before | Local changed build | Local mygo |
| --- | ---: | ---: | ---: |
| Startup TotalAlloc / MiB | 95.38 | 26.65 | 19.17 |
| Startup GC cycles | 10 | 5 | 4 |
| Startup HeapSys / MiB | 83.22 | 27.22 | 23.22 |
| Whole-process kernel peak / MiB | 115.36 | 115.55 | 132.41 |

Startup allocation falls 72.1%, but physical peak reduction is not established.
Keel's three-run kernel-peak ranges overlap: 114.75–124.24 before,
112.70–115.58 after. Startup kernel-peak medians are 106.86 → 106.00 MiB;
later resizing raises the lifetime peak. HeapSys includes released pages and
is not physical footprint. Card scrolling remains about 4.91 MiB/s, with
GC 6 → 7; fewer startup allocations do not imply fewer cycles in every phase.

Each process ran startup/loading, 10s scroll, 10s continuous resizing and 12s
idle. Physical footprint was sampled at 100ms (50ms startup), separately from
the kernel lifetime peak, with idle values taken over the final three seconds.
Forced GC was zero throughout. Only the fixture disables Keel's idle active
reclamation to match mygo; library defaults and GOGC/GOMEMLIMIT are unchanged.
Locked-window pilots, profiles and screenshots are excluded from these medians.
The native scroll/resize workflows completed without unresponsive timeouts.

A separate single-pair frame-statistics diagnostic reports first-frame
layout+submission 104.00 → 60.89ms: a first-frame stall remains. These are
not presentation latency, three-run timing medians or end-to-end FPS.

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

The microbenchmark is CPU vertex construction, not a whole frame or FPS.
An unlocked native follow-up compared the actual controls category at 800×600
pixels, native 1×, three interleaved fresh processes per version. Main
`3b88f26` is the baseline; the changed binary combines local `af8d7dd` font
and `b45e325` stroke patches through an overlay. Those measurements preceded
publication; both patches are now merged as PRs #35 and #36, not attributed
to a released version.

| Native controls phase | Before MiB/s | After MiB/s | GC before → after |
| --- | ---: | ---: | ---: |
| 10s scrolling plus component animation | 22.91 | 14.41 | 9 → 5 |
| 10s continuous resizing | 86.17 | 39.55 | 32 → 14 |
| 12s without input, animation continues | 23.11 | 14.50 | 10 → 6 |

This is not a static-idle CPU check or a mygo gallery comparison. Native
scroll/resize workflows completed without unresponsive timeouts, but the
fixture has no frame counter and does not establish FPS or input latency.
Forced GC was zero throughout; only the fixture disables idle active
reclamation, matching mygo. Shaders, drawable count and library GC/reclamation
defaults are unchanged.

Allocation improves, but physical memory improvement is not established.
Scrolling footprint medians are 157.45 → 159.86 MiB, resizing
160.99 → 161.22 MiB. Kernel
lifetime-peak medians rise 183.58 → 190.85 MiB; the three-run ranges overlap,
176.45–194.60 → 173.08–200.31 MiB. Keep this observed cost in the report.

Separate no-forced-GC profiles put LabelStyle at about 13.3% and anchors
insertion at 9.3% of the remaining scrolling samples. During resizing, shaping
plus Gio glyph conversion account for about 28%; the new cache's owned copies
account for 15.8%. Changing-size paths need admission/reuse work, not unsafe
scratch borrowing. Profile accounting can lag by roughly two GC cycles;
these shares are approximate and sampled totals are not exact phase bytes.

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

Native glass needs a compositor check as well as transparent GPU pixels. On macOS 14.4+, run `KEEL_DESKTOP=1 go test -tags keelnativeqa -run TestNativeMaterialComposition ./ui/window`. This fixture captures its own red/blue reference and frosted window together, then asserts that both colors reach the glass. It uses current-process ScreenCaptureKit content without requesting screen recording. The filter retains system desktop layers because excluding them flattens behind-window materials; other applications' content windows are excluded, and the saved image is cropped to the test window.
