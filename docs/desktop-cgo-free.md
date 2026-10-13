# Desktop CGO-free delivery design

English | [简体中文](desktop-cgo-free.zh-CN.md)

Status: implementation handoff, not a statement of current platform support.
The single-module revision below supersedes the earlier separate-fork delivery
approach. Complete [module unification](single-module.md) first.
Baseline: `b2696a6`, inspected on 2026-10-12 with Go 1.26.4 on darwin/arm64.

## Goal and boundaries

Deliver reproducible applications using Keel's Gio and go-text patches, and
build desktop executables with `CGO_ENABLED=0` without a C compiler or native
development headers. Keep native windows and GPU rendering, component APIs,
rendering algorithms and existing performance patches. Module unification changes
Gio/go-text import paths as documented in the single-module migration guide.
System frameworks, shared libraries and GPU drivers remain runtime dependencies.

Targets are macOS arm64/amd64, Windows amd64/arm64, and Linux amd64/arm64 with
X11 and Wayland. Cross-compilation and execution are separate acceptance gates.
iOS/Android migration is deferred; their existing builds must keep working.
WebAssembly remains on its current path. Signing, notarization, icon conversion
and mobile packaging tools are separate from desktop executable compilation.

## Historical starting point

| Area | Evidence and consequence |
| --- | --- |
| Distribution | Root `go.mod` replaces two nested modules. `cmd/keel/new.go` adds their replacements only for `new -replace`. Ordinary generated applications resolve upstream. Nested modules are excluded from the root module zip. |
| macOS dependency scan | `hello` reaches cgo in `internal/gl`, `gpu/internal/metal`, `internal/cocoainit`, `gpu/headless`, `app` inside Gio, and Keel's `ui/theme/coretext_darwin.go`, plus `runtime/cgo`. |
| macOS build | `CGO_ENABLED=0 go build ./examples/hello` fails with `undefined: Functions` in Gio's `internal/gl/util.go`. GPU backend registration makes OpenGL part of the graph even with default Metal rendering. |
| Native APIs | `CGO_ENABLED=0 go build ./native/...` succeeds on this macOS host. Static search finds no `import "C"` in `native`. This is not cross-platform runtime acceptance. |
| Linux source selection | `GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go list -e -deps` selects cgo in Gio's X11/Wayland app, xkbcommon, GL, EGL and Vulkan. Standard-library net/os/user and purego's optional cgo loader also appear; their no-cgo alternatives must be checked with CGO disabled. This was a dependency scan, not a Linux build or runtime test. |
| Existing CI | Windows native chrome and Linux Wayland lifecycle jobs exist; the Linux job installs development headers. No desktop no-cgo build matrix was found. |

The original inventory has since been corrected in `AGENTS.md`. These baseline
observations are historical; see the [validation report](desktop-cgo-free-validation.md)
for subsequent work. Its separate-fork D1 results do not validate module unification.

## Decision 1: one Keel module

Follow [the single-module design](single-module.md): Gio and typesetting become
ordinary packages beneath Keel, use Keel-prefixed imports and share root go.mod,
go.sum and release version. This replaces the former fork manifests, separate
releases, application replacements and upstream compatibility mode. Keep shader
as its separate external dependency. Adapt package-sensitive tooling, including
gogio, and preserve the useful transactional scaffold tests from previous work.

The user explicitly chose this structure, superseding the old module-path rule.
Component behavior remains stable; applications importing Gio types must migrate
imports. Removing nested manifests makes both libraries part of the Keel zip;
removing root replacements allows ordinary versioned CLI installation, subject
to the external-module acceptance fixture. Module unification does not remove CGO.

## Decision 2: migrate native bindings with a bounded transition

Use an explicit backend-selection truth table per migrated desktop platform:

| CGO_ENABLED | `purego` tag | Transitional backend |
| --- | --- | --- |
| 1 | absent | Existing cgo |
| 1 | present | Pure Go |
| 0 | absent or present | Pure Go |

For macOS-specific files, legacy constraints are
`darwin && !ios && cgo && !purego`; replacements use
`darwin && !ios && (!cgo || purego)`. Apply consistent selection to existing
`.m`/`.c` files, assembly, common definitions and tests as well as Go files.
Shared Darwin code needs explicit iOS separation. Preserve mobile constraints;
do not accidentally select desktop code for iOS. Add no C/Objective-C sources
or C helpers. Keep old files solely for comparison during migration.

Each completed platform switches to Pure Go by default after the acceptance
gates below; then retire its legacy bindings in a separate change. Until then,
`CGO_ENABLED=1 -tags purego` allows race tests of the new implementation.
Do not run `-race` with CGO disabled or equate the race runtime's cgo use with
application binding requirements.

### macOS work packages

1. **ABI feasibility gate.** Exercise both amd64 and arm64 native calls and
   native-to-Go callbacks for NSRect/NSPoint/NSSize/NSRange, floating arguments,
   out parameters, objc super calls and blocks. Specifically test the signatures
   used by NSTextInputClient. Check purego 0.11.1's actual support; do not assume
   RegisterClass implies every aggregate callback ABI works. Use Go/assembly
   adapters where needed, with focused native tests. If signatures cannot be
   supported without new cgo, return the limitation to design before rollout.
2. **Shared runtime and complete first frame.** Add minimal helpers within
   Gio's own `internal` boundary; Gio cannot import Keel's `ui/internal/appkit`
   or `native/internal/sys`. Port cocoainit, AppKit bootstrap, macOS window/view,
   display link, Metal layer/context, GPU Metal bindings including `pool` and
   `quads`, and headless context. Port Keel's CoreText/CoreGraphics binding in
   `ui/theme`. Resolve GL's compile dependency with real bindings and preserve
   the existing `nometal` path; empty stubs are not a renderer implementation.
   Keep shader code, command scheduling, atlas behavior and resource policies.
3. **Interaction parity.** Preserve all current window operations, keys,
   modifiers, mouse/scroll, focus, clipboard, existing drag/drop behavior,
   Retina/display changes, full screen, occlusion and display sleep/wake.
   Chinese IME must cover marked text, replacement/selection ranges, commit,
   cancel, UTF-16 conversion for surrogate pairs, and candidate coordinates
   across scale changes. Preserve AppKitViewEvent and invalid-view/teardown order.
4. **Default promotion.** Only after both architectures pass ABI, real-window,
   text and GPU acceptance. First-frame success alone is not platform support.

Register Objective-C classes and callbacks once; route through integer tokens
or object maps, and invalidate routing on teardown before releasing objects.
No retained Go pointers. Native-owned buffers or an explicitly valid bounded
call lifetime are required for bitmap/texture storage. Pair every owned object
and copied block with release, including error paths. Use autorelease pools on
locked threads. Do not catch a failed Metal initialization by silently changing
the renderer; return a useful error through the existing lifecycle.

Integer/pointer hot calls use cached addresses/selectors and `SyscallN`.
Float/struct calls require ABI-correct typed adapters; never force them through
integer calling conventions for speed. Benchmark their cost. Native callbacks
must not synchronously acquire Keel's frame lock. Preserve Gio event ordering;
UI mutations are posted via the existing update mechanism. No polling replaces
display link, window events or explicit wakeups.

### Linux work packages

Port the existing Xlib/X11 binding via purego first, including Xcursor/Xfixes,
xkbcommon and EGL/GL. Keeping the existing protocol/event model bounds the
change; adopting an unrelated pure Go X11 stack is unnecessary here. Validate
XIM/IBus behavior actually supported by the current backend, keyboard layouts,
clipboard, DPI, resize, close and wakeups.

Next port Wayland protocol metadata, proxy construction and marshalling,
listeners/dispatcher, xdg-shell/decoration/text-input, cursor and EGL integration.
Reuse the single-dispatcher pattern conceptually, not Keel's private package.
Preserve display sharing and private queues used by `native/clipboard`, thread
affinity, configure acknowledgment and view invalidation before native teardown.
Wayland objects/signatures retained by libwayland require native-safe lifetime.
Use a bounded callback registry, never one FFI callback per proxy or request.

Vulkan is also in the default Linux dependency graph. Port its loader/surface
bindings without changing the renderer. Intermediate X11/EGL validation may
use `nowayland,novulkan`, but that is only a milestone. Final Linux acceptance
requires both window systems and the normal default build with no extra tags,
plus existing `nox11`, `nowayland` and `novulkan` configurations. Exclude Android
and BSD from desktop-specific constraints rather than breaking their old paths.
Load required libraries with clear missing-library/symbol errors and use the
existing backend fallback policy. No compiler, pkg-config or headers at runtime.

## Implementation sequence and ownership

Use separate reviewable changes in this dependency order:

| Step | Main locations | Deliverable / exit gate |
| --- | --- | --- |
| D0 | `AGENTS.md`, test scripts, `.github/workflows/` | Correct inventory, reproducible baseline and no-cgo matrix. |
| D1 (revised) | Root manifest, fork imports, dependency rules, CLI/packager and release fixture | Complete the single-module design; an external app gets both libraries with only a Keel requirement. |
| M0 | Gio internal native helpers and ABI tests | Native aggregate callbacks pass on both Mac architectures. |
| M1 | Gio app/Metal/GL/headless/cocoainit, Keel theme | No-cgo hello, offscreen screenshots and native first frame; cgo comparison still available. |
| M2 | Gio app, existing `ui/window` fixtures, theme tests | IME, lifecycle, GPU/performance parity; promote macOS default. |
| L1 | Gio X11/xkb/EGL/GL | X11 runtime acceptance, initially with documented temporary tags. |
| L2 | Gio Wayland/Vulkan and native clipboard tests | Default Linux no-cgo build and X11/Wayland runtime acceptance; promote Linux default. |
| R | CLI build/doctor/run, release fixtures, bilingual docs | New app through packaged artifact works; remove legacy desktop bindings only after parity. |

Complete revised D1 first, preserving existing D0/M0 work. Then resume the
binding migration. A design handoff does not establish implementation completion.
Do not report Desktop CGO-Free complete while L2 or either Mac architecture
remains unverified. Every fork patch has its own test and bilingual entry in
`third_party/README*`. Update architecture/dependency tables before introducing
any new Keel package; keep ui/native isolated. Update README, getting-started,
testing, troubleshooting and CLI docs in both languages as support changes.

## Acceptance evidence

1. **Distribution:** use the single-module external archive fixture, a fresh
   module cache and `GOWORK=off`. The app requires only Keel; both library
   package trees belong to that root module. Validate ordinary CLI installation,
   local/offline generation and migrated packager paths. No fork lock manifest
   or separate nested-module release is part of acceptance.
2. **Compilation:** no-cgo builds of `./...` and `./native/...` for each target
   above; root checks now include Gio and typesetting. Collect
   `go list -deps -json` with CGO disabled: no selected CgoFiles or package
   errors. Enable cgo only for diagnostic inventory and the race lane.
   Build in a compiler/header-free environment; Go alone must link the binary.
   These are later CGO gates, not prerequisites for accepting module unification.
3. **Compatibility:** test each transition selection and renderer tag; Windows
   keeps its existing implementation. Verify migrated custom-widget imports and
   the absence of old upstream core packages, with shader explicitly allowed.
   Preserve mobile/web builds and packaging. The old test that removed fork
   replacements to compile against upstream is retired with module unification.
4. **Runtime:** run on native macOS arm64 and amd64, Windows, Linux X11 and
   Wayland. Repeat `KEEL_DESKTOP=1 go test -run RealWindows ./ui/window` and
   Gio app lifecycle tests. Exercise immediate close/reopen, raise/close from
   callbacks, full screen, focus, teardown and clipboard. Real Chinese IME
   composition/candidate-window checks are required, not only synthesized text.
   Record desktop OS, GPU, scale and input method. Software-rendered CI tests
   cannot replace a native GPU acceptance run; untested hardware stays pending.
5. **Visual/performance:** before changing bindings, capture hello and component
   matrix screenshots on the same host. Require identical output for unchanged
   rasterization; investigate any difference. Run CoreText/emoji, Metal idle,
   resize, readback and resource-recreation regressions. Compare at least five
   alternating before/after benchmark runs with `-benchmem` and benchstat,
   including `BenchmarkGlyphRendererScrolling`; record frame latency p50/p95,
   CPU, allocation and GPU/process memory for idle, scrolling, resize and text.
   Treat a repeatable >5% CPU/latency/allocation regression outside observed
   noise as a rollout blocker needing explanation and resolution. Idle windows
   must introduce no periodic wakeups. No assumption that FFI itself is faster.
6. **CLI and release:** `keel run` and desktop build honor the selected backend
   and CGO setting. Remove Linux's host-only restriction after setting GOOS
   explicitly and proving packaging works. Audit pinned gogio for forced CGO;
   adapt desktop invocation or compile the executable directly if necessary.
   Report build tools separately from `.app`/signing tools in doctor. Validate
   `go build` independently of packaging, then test actual `keel new → run →
   build` artifacts. Run the repository's formatting, vet, test and race checks
   including the now-integrated library trees and updated `AGENTS.md` rules.

Implementation reports must distinguish commands run, failures, skipped desktop
checks and pending release publication. Cross-build success, simulated events,
and a design document must never be reported as native runtime acceptance.
