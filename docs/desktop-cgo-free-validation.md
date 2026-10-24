# Desktop no-cgo implementation evidence

English | [简体中文](desktop-cgo-free-validation.zh-CN.md)

Validated on 2026-10-13 with Go 1.27.1 on macOS 27.0.1, darwin/arm64.
The single-module step D1 is implemented and validated locally. Desktop
CGO-Free is still pending: macOS/Linux UI bindings continue to require cgo.
The visual baseline for this run is `8d59576`.

## Root-level relocation follow-up

The foundation directories have physically moved to gio/ and typesetting/, with
imports under `github.com/dyike/keel/gio` and `github.com/dyike/keel/typesetting`.
Both `go list -json ./gio/layout ./typesetting/shaping` results report
`github.com/dyike/keel` as Module.Path. Only the root go.mod/go.sum remain.
Old third_party foundation paths occur only in the bilingual migration table,
explicitly marked as unpublished intermediate paths. Upstream core references
are historical attribution, migration explanations or rejection checks; shader
remains external. The site excludes the relocated upstream documentation.

Formatting, vet, full tests, full race tests, site tests, Windows amd64/arm64
no-cgo gates and the real-source external release fixture were rerun successfully
after relocation. The fixture also rejects obsolete foundation sources/modules
and verifies the new linker symbols through a packaged native macOS identity
probe. Native arm64 ABI probes and a byte-identical hello screenshot passed.

The first real-window run timed out in reopen/update/now while full checks ran
concurrently. An isolated rerun passed all real-window fixtures in 16.443 seconds.
The timeout's cause was not established; no lifecycle implementation was changed.
The 792-image matrix comparison below belongs to the earlier single-module
validation, before this root-level relocation.

## Single-module implementation (D1)

Keel, Gio and typesetting share the root go.mod/go.sum. Both foundation trees
use Keel-prefixed imports and ship inside its module archive; shader stays
external. Source checks cover imports in all platform files, including files
excluded by the host's build tags, and enforce foundation dependency boundaries.

The scaffold writes one Keel requirement, or one Keel replacement for a local
checkout. Exact offline versions need no release metadata or network access;
uncached version queries fail explicitly. Local paths may contain spaces and
must identify the module root. Staging, tidy rollback and concurrent-file
protection remain covered. The earlier unpublished separate-fork lock manifest,
resolver, backend mode and deps command are superseded by this implementation.

The maintained gogio v0.10.0 copy uses the relocated Gio root for app IDs,
arguments, permissions and resource discovery. Run options supply the directory,
environment and output; calls serialize upstream package state. Default outputs
are relative to the supplied directory. Windows resources and the iOS simulator
shader adapter remain in the CLI. Architecture, migration instructions, patch
logs and CI now describe the single module.

## Commands and results

| Check | Result |
| --- | --- |
| `gofmt -l .`, `git diff --check`, `go mod tidy -diff` | Passed; no formatting, whitespace or module changes needed. |
| `go vet ./...` | Passed, including both foundation trees. |
| `go test ./...`, `go test -race ./...` | Passed, including the MCP tests and foundation regressions. |
| Targeted CLI, packager, boundary and site tests/race checks after follow-up edits | Passed. |
| `python3 scripts/check-module-release.py` | Passed the real-source release fixture described below. |
| `check-desktop-cgo-free.sh native`, six desktop OS/architecture pairs | Passed cross-builds with nonexistent CC/CXX paths. |
| `check-desktop-cgo-free.sh all`, Windows amd64/arm64 | Passed complete root-module cross-builds with cgo disabled. |
| `CGO_ENABLED=0 go build ./examples/hello`, darwin/arm64 | Still fails in Gio GL with `undefined: Functions`; this remains a migration gate. |
| `CGO_ENABLED=0 go test ./gio/internal/nativeabi -count=1` | Passed on native arm64. |
| Native ABI test binary, darwin/amd64, cgo disabled | Cross-compiled; not executed on a native Intel Mac. |
| `KEEL_DESKTOP=1 go test -run RealWindows ./ui/window -count=1` | Passed on the native macOS host. |

These are local command results. The updated CI jobs have not been executed by
a remote runner in this session. Cross-builds do not establish native runtime
support, and nonexistent compiler paths do not replace compiler-free isolation.

## Release and packaging fixture

The file proxy serves the real working-tree archive as `v0.0.0-fixture`.
With a fresh module cache and `GOWORK=off`, the fixture verifies:

- An external UI app requires only Keel and has no replacement. Both foundation
  package trees report Keel as their Module.Path, with real source files present
  and no nested manifests or upstream core modules in the graph.
- An external widget implements core.Widget using the relocated Gio layout,
  unit, clip and paint packages and compiles for Windows.
- Ordinary versioned `go install` installs the CLI from that archive. Its
  default scaffold selects its own release. Exact offline generation works
  before any release is cached; local generation uses one replacement. App,
  checkout and module-cache paths containing spaces work.
- The installed CLI packages Windows amd64/arm64 and WebAssembly. Web outputs
  include main.wasm, wasm.js and index.html.
- On native macOS, the maintained packager creates a real app bundle using the
  existing cgo backend. Its executable reports `dev.keel.releaseprobe` and
  both the native command-line argument and linker-supplied launch argument.

Focused tests also verify WebAssembly app ID/argument linking and HTML metadata,
and discover Android/iOS resources and Android permission packages at the new
paths. Full APK/iOS packaging was not run: this host has Command Line Tools,
not the full mobile SDK toolchains. Production tags were not published; remote
installation of the eventual release remains a publication gate. Only the
synthetic fixture disables checksum-server verification.

## Visual and window acceptance

Hello's before/after PNGs are byte-identical. Both complete component matrices
contain 792 PNGs (99 sections, light/dark, two widths and two scales).

Animated components differ when sampled at different wall-clock times. A
temporary Go overlay fixes only the screenshot frame context to
`time.Unix(0, 0)` in both checkouts; it does not change production sources.
With that same frame time, all 792 PNGs and the matrix index are byte-identical.
The baseline overlay uses the resolved `/private/tmp` source path on macOS.
Real-window lifecycle/multiwindow fixtures also pass.

## M1 progress: Keel's CoreText binding (2026-10-24)

`ui/theme` rasterizes glyphs through CoreText in `coretext_cgo_darwin.go`
(`cgo && !purego`) or, without cgo, in `coretext_purego_darwin.go`
(`!cgo || purego`); font caching, padding and the linear-light conversion
are shared. The bitmap context draws into a Go buffer whose lifetime is
bounded by one function, as in the cgo code; a context owning its pixels
measured three times slower. Integer and pointer calls go by address
through `SyscallN`; the text colors are CGColors created once, and only the
bounds query keeps a `RegisterFunc` wrapper on amd64, where its rect result
uses a hidden pointer.

| Check | Result |
| --- | --- |
| `go test ./ui/theme`, with and without `-tags purego`; `-race -tags purego` | Passed. |
| Component screenshots, 11 text-heavy sections × light/dark × 1x/2x, plus `hello` | 44 and 1 PNGs byte-identical between bindings. |
| 12 emoji rasterizations (Apple Color Emoji, three sizes) and text masks, light and dark | Identical SHA-256 between bindings. |
| `BenchmarkCoreTextMask` (one glyph-cache miss), Apple M4 | cgo 2.7 µs, 4 allocs; purego 3.3 µs, 19 allocs. Steady-state frames do not rasterize. |
| `CGO_ENABLED=0 go vet ./ui/theme` | Still fails in Gio's `internal/gl`, which `ui/theme` reaches through the headless renderer: the next M1 steps. |

The cgo binding remains the default until macOS is promoted, then is removed.
Native amd64 execution was not run.

## M1 progress: Gio's GL binding (2026-10-24)

`gio/internal/gl` has a purego binding for macOS and Linux desktop,
selected by `CGO_ENABLED=0` or `-tags purego`; iOS, Android and the BSDs
keep `gl_unix.go`, as `go list` confirms for each configuration. With it,
`CGO_ENABLED=0 go vet ./gio/internal/gl` passes for darwin and linux, amd64
and arm64. On macOS the headless and paint tests pass with
`-tags "nometal purego"`, rendering through OpenGL; `gio/gpu` fails to build
under `nometal` with either binding, as on `main` (its tint test uses the
Metal-only `AddTinted`). The no-cgo build of `hello` now stops in `gio/app`.
Linux GL performance and runtime remain to be measured with L1.

## M1 progress: Gio's Metal backend (2026-10-24)

`gio/gpu/internal/metal` has a purego binding for macOS, selected by
`CGO_ENABLED=0` or `-tags purego`; iOS keeps the cgo file, as `go list`
confirms for iOS, macOS with cgo and macOS without it. On an Apple M4
(darwin/arm64), comparing the two bindings:

- Pixels: hello is byte-identical. Of 793 component-matrix screenshots, 721
  are byte-identical; the other 72 are animated pages (spinner, progress,
  skeleton, status marker, and pages containing them), and the same 72
  differ between two runs of the cgo binding.
- Tests: `gio/...` and `ui/...` pass with `-tags purego`, including the
  real-window tests (`KEEL_DESKTOP=1`).
- Allocations per frame: 6 instead of 7 (rotating path) and 4 instead of 5
  (repeated clip). Of the purego binding's own calls, only the blocking wait
  allocates.
- CPU: 6,000 headless frames took 1.97 to 2.13 s of user and system time
  with purego and 2.01 to 2.07 s with cgo over four alternating runs; the
  difference is within noise. These frames have few draws; a call costs
  about 55 ns, so frames with thousands of unbatched draws may cost more.
- Leaks: Objective-C object counts by class stay constant over about 15,000
  frames with both bindings; encoder calls run without an autorelease pool.

amd64 was only compiled and vetted (Rosetta is not installed), so its
by-value struct calls through `RegisterFunc` have not run. The no-cgo build
of `hello` still stops in `gio/app`; `gio/gpu/headless` remains cgo.

## M1 progress: Gio's headless context (2026-10-24)

`gio/gpu/headless` creates its Metal device and queue through purego on
macOS without cgo; iOS keeps the cgo file. With it, Gio's GPU stack
(`gpu`, `gpu/headless`, `gpu/internal/metal`, `internal/gl`,
`internal/cocoainit`) passes its tests with `CGO_ENABLED=0` on an Apple M4.
With `-tags purego`, hello is byte-identical to the cgo build and the same
721 of 793 component screenshots match; the other 72 are the animated pages
that also differ between cgo runs. `gio/app` is now the only cgo package in
the no-cgo build of `hello`.

## M2 progress: Gio's macOS app binding (2026-10-24)

`gio/app` has a purego binding on macOS, so with `CGO_ENABLED=0` no package of
`examples/hello` uses cgo and the binary links only libSystem. On an Apple M4:

- `CGO_ENABLED=0 go test ./...` passes; `gio/app` real-window tests
  (`KEEL_DESKTOP=1`: idle trim, live resize, wrapped input view, activation
  policy, text input) and `ui/window` real windows pass with cgo, with
  `-tags purego` and without cgo. `-race -tags purego` passes for `gio/app`,
  `gio/gpu/...`, `ui/window` and `ui/el`. `gio/app` still cross-compiles for
  iOS, and its amd64 build vets.
- Text input: a new test sends the `NSTextInputClient` messages through
  `objc_msgSend` to a focused editor: marked text with its range and caret,
  commit, cancel, a surrogate pair counted as two UTF-16 units, the
  substring with its actual range, and the caret rectangle inside the
  window frame. It passes under both bindings and fails without the
  protocol or with a skewed range. No real input method was driven.
- Pixels: the no-cgo `hello` screenshot is byte-identical to cgo's. A
  visible no-cgo window took typing of Chinese and emoji, a click and a
  redraw through the automation socket.
- OpenGL (`nometal`): a new test opens and resizes a window under cgo,
  purego and no cgo; its pixels were not compared. `hello` does not build
  with `nometal` with either binding (`ui/theme` uses the Metal-only
  `PaintTintedQuads`), as before.
- Cost, real windows, 10 s after a 3 s warmup, four alternating runs:
  the spinner page used 0.94–1.07 s of CPU without cgo (mean 1.00) and
  0.89–0.98 s with cgo (mean 0.94), about 6% more; phys_footprint was
  108–121 MB against 108–118 MB. Idle `hello` used no measurable CPU with
  either. Native callbacks and float or struct messages go through purego's
  reflective wrappers, which allocate.

Not done: amd64 execution, a real Chinese input method with candidate
windows across scale changes, display sleep/wake and full screen by hand,
and the pixel check of the OpenGL window. The macOS default stays cgo.

## L1 progress: EGL, xkbcommon and X11 on Linux (2026-10-24)

The "Linux without cgo" workflow runs on GitHub's ubuntu-24.04 (amd64) and
ubuntu-24.04-arm runners with no C compiler and runtime libraries only, on
Mesa's software renderers. With `CGO_ENABLED=0`:

- EGL and OpenGL ES: a new headless pixel test renders fills, a clip and a
  blended ellipse through surfaceless EGL and reads them back correctly on
  both architectures; the GL binding's symbol and string tests pass.
- xkbcommon: a new test compiles the default keymap and checks KEY_A's
  keysym, text, Shift and repeat; the cgo job runs it too and its header
  checks pass.
- X11: `examples/hello` builds and its window opens under Xvfb. The cgo job
  checks the Go X structure layouts and constants against the headers.
  The ui/window real-window tests first timed out in resident and reopen
  with both bindings: Perform and Configure ran outside the event loop and
  Xlib never flushed their requests. With XFlush added, the whole set
  passes under Xvfb with both bindings on amd64 and arm64. CI also found two
  cgo regressions of the refactor (structs holding Go pointers passed to
  C), fixed before this note.

Wayland and Vulkan are still cgo-only on Linux; the no-cgo build uses X11
and EGL. IBus input, keyboard layouts beyond the default and real desktops
were not exercised.

## L2 progress: Wayland on Linux (2026-10-24)

`gio/app` opens Wayland windows without cgo. The no-cgo workflow runs the
Wayland lifecycle tests and the six ui/window real-window tests under
weston's headless backend with `DISPLAY` unset, so no X11 fallback exists;
all pass on amd64 and arm64. Requests are marshalled by shared Go code
through `wl_proxy_marshal_array_flags` under both bindings; the cgo job
checks their 61 opcodes, 30 protocol values and the wl_cursor layouts
against the headers, which caught wrong text-input content purpose values
before this note. Events reach shared handlers through Gio's listener
tables (cgo) or one `wl_proxy_add_dispatcher` callback (purego).

Not exercised: a real compositor and desktop, text-input with an input
method, clipboard between applications, touch, HiDPI output scales and
client-side decorations. Vulkan remains cgo-only on Linux.

## Preserved ABI work and remaining gates

D0's compile gate and M0's no-cgo Objective-C probes remain in the root module.
Native arm64 probes cover ranges, register exhaustion, nested rects, nullable
native out buffers, super calls and copied blocks; scalar floating callback
returns remain explicitly unsupported. Native amd64 execution and exact
NSTextInputClient protocol behavior are still required.

Proceed with M0/M1 after single-module review: migrate AppKit, cocoainit,
Metal/GL/headless and Keel CoreText bindings, then validate IME, lifecycle and
GPU/performance parity on both Mac architectures. Linux X11, Wayland and Vulkan
migration, native GPU/IME acceptance, compiler-free isolation, mobile packaging
and production release publication remain pending. This structural change
does not promote a no-cgo macOS/Linux backend.
