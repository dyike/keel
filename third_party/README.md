# third_party

English | [简体中文](README.zh-CN.md)

Keel carries its own copies of the libraries it draws with, so it can fix
and speed them up without waiting for upstream releases. They are ordinary
packages of Keel's module, with no `go.mod` of their own, imported under
Keel's path. Sources live in the root-level gio/ and typesetting/ directories;
this directory keeps their attribution and patch logs:

| Upstream import prefix | Keel's import prefix |
| --- | --- |
| `gioui.org` (Gio's root module) | `github.com/dyike/keel/gio` |
| `github.com/go-text/typesetting` | `github.com/dyike/keel/typesetting` |

| Directory (repository-relative) | Upstream | Base | License |
| --- | --- | --- | --- |
| `gio/` | [Gio](https://gioui.org) `gioui.org` | v0.10.3 | Unlicense OR MIT (`gio/LICENSE`) |
| `typesetting/` | [go-text](https://github.com/go-text/typesetting) `github.com/go-text/typesetting` | v0.3.5 | Unlicense OR BSD-3-Clause (`typesetting/LICENSE`) |
| `internal/gogio/` | Gio's packager `gioui.org/cmd/gogio` | v0.10.0 | Unlicense OR MIT (`internal/gogio/LICENSE`) |

`gioui.org/shader` (including `gioui.org/shader/gio`) is a separate module
and stays upstream.

Both copies ship in Keel's module zip and share its version, so any app that
requires `github.com/dyike/keel` builds with them; no `replace` is needed.
Changing the import path changes Go type identity: an app or widget that
imports upstream `gioui.org/...` or go-text types next to Keel must switch to
the prefixes above (see [docs/single-module.md](../docs/single-module.md)).

## Rules

- Keep upstream style in these directories; change what Keel needs, nothing
  cosmetic. Each change is listed below with the reason, so a rebase onto a
  newer upstream can redo it.
- The copies are foundations, not Keel modules: Gio may import itself and
  typesetting, typesetting only itself, and neither imports other Keel
  packages or upstream Gio/go-text (`internal/deps` checks this).
- Upstream tests and test data are not copied. A patch comes with its own
  test next to it; run them from the root (`go test ./gio/... ./typesetting/...`).
- Updating to a newer upstream: copy the release without `go.mod`, `go.sum`,
  `*_test.go` and `testdata`, rewrite the two import prefixes (leaving
  `gioui.org/shader` alone), then reapply the patches below.

## Patches

- `gio/gpu`, `gio/internal/stroke`: cache stroke outlines in path coordinates,
  before the existing affine transform and vertex splitting. Rotation no longer
  rebuilds the same outline; width/content changes remain distinct. Exact input
  comparison prevents hash-collision reuse. Owned entries expire after an unused
  frame and are bounded to 128 entries/1 MiB; oversized paths bypass caching.
  Renderer trim/release drops the cache. Reusable, independently owned stroking
  scratch also reduces changing-path misses and discards oversized arrays.
  Tests compare the old algorithm's curves and affine vertices byte-for-byte,
  check ownership/eviction/allocation bounds, and compare warmed/fresh GPU pixels
  under rotation, nested rounded clipping and opacity. No shader or GC changes.

- `gio/widget.Icon`: retain an LRU of up to eight immutable size/color images,
  with a 1 MiB pixel budget per icon. A larger image retains only itself, as
  before. Alternating variants no longer rerasterize or change texture identity;
  tests cover warm zero-allocation hits, eviction and reference release.
- `gio/text`, `gio/widget`: labels and editor painting borrow one 32-glyph
  scratch buffer from their shaper. CoreText opts into a synchronous, read-only
  raster-hook contract; existing `SetRasterHook` callers keep independent
  buffers. Overlapping borrows and hook changes cannot reuse a live lease.
  Tests cover legacy retained slices, lease invalidation and allocation-free reuse.
- `typesetting/fontscan` tests isolate the variable PingFang UI collection
  when checking named instances and immutable-table sharing. Installed static
  MobileAsset faces can legitimately win system fallback. CoreText tests cover
  both the explicit variable fixture and actual installed weight resolution,
  retaining exact PostScript-identity and nonblank-glyph assertions.

- Share identical stencil masks within each frame without adding GPU textures. Preserve absolute transform inputs and parent intersection behavior; tested with native pixel comparisons. See [rendering memory](../docs/rendering-memory.md).

- `gio/gpu/internal/metal`: unified-memory devices initialize immutable geometry and
  the first complete sampled-image upload directly in shared resources. This avoids
  activating GPU copy work for ordinary UI uploads, which added about 57 MiB in a
  minimal local Metal experiment. CPU-generated mip levels retain linear-light sRGB
  filtering; native comparisons cover odd sizes, strides and all levels (maximum
  observed channel difference 2/255). Partial/repeated uploads, mutable buffers,
  render targets, non-unified devices and allocation failures retain the ordered
  GPU path. No per-frame GC, frame-rate cap or drawable-count change. See
  [rendering memory and reuse](../docs/rendering-memory.md) for scope and limitations.

For a behavior-by-behavior summary with source/test links, see
[Changes from upstream](../docs/upstream-patches.md). Numbers below are historical
local measurements, not fresh benchmarks or portable application guarantees.

- Root-level relocation: the foundation trees now live in gio/ and typesetting/
  with matching import prefixes. Link symbols, resource discovery, CI and release
  fixtures follow those paths; source checks reject obsolete foundation imports.
- Module merge: the copies' `go.mod`/`go.sum` were removed and their imports
  rewritten to Keel's prefixes, so they no longer depend on application-level
  `replace` lines (which Go does not propagate) to reach apps.
- `internal/gogio`: the packager runs inside `keel` instead of `go run
  gioui.org/cmd/gogio`, because upstream gogio hard-codes Gio's import path. It
  takes that path from `GioRoot` (for `app.ID`, the log app ID, extra args,
  permissions and the `app/` package lookup); flags, working directory,
  environment and output are per-run options (`Run`) rather than process
  globals; upstream package-level state is serialized. Default outputs resolve
  against the supplied working directory. A WebAssembly packaging regression
  checks app ID/launch-argument linking, HTML title and output files; the release
  fixture builds a custom widget from the real root archive. The Windows target
  and upstream end-to-end tests are not copied: Keel writes
  its own Windows resources. `cmd/gogio` runs it directly for
  builds without `keel.json`, such as the site's gallery.

- Packager relocation: the standalone entrypoint lives at `cmd/gogio` and the
  shared implementation at `internal/gogio`, so both commands can use it without
  nesting a command inside `cmd/keel`. Resource-discovery tests use the new
  directory depth; CI and release fixtures invoke the standalone command.

- `gio/internal/nativeabi`: no-cgo macOS ABI probes call Go Objective-C IMPs
  through the real system `objc_msgSend`. They cover range pairs/register
  exhaustion, nested rect arguments/returns, floating point arguments, nullable
  native-allocated out buffers and canaries. Classes/callbacks register once.
  Scalar floating callback results are explicitly documented as unsupported by
  purego 0.11.1. This is an M0 feasibility test, not a migrated window backend;
  Expanded probes cover aggregate super calls and native invocation of copied
  blocks. Native arm64 tests pass; amd64 execution and exact protocol signatures
  remain required.

- `gio/widget.Editor`: exposes optional `ScrollBounds`, `ScrollOffset` and `ScrollTo` methods for multiline scrollbar drawing and dragging. Moving the viewport preserves the selection; the next edit or caret movement resumes caret scrolling. Capability detection retains native wheel scrolling when the extension is absent; import-path compatibility still follows the migration guide. A 1×/2× regression checks clamping, selection, caret reveal and content replacement.

- `gio/gpu/internal/metal`: idle trimming waits for the last submitted GPU work, then releases its command buffer, staging buffer and quad instance buffer. Live textures and pipelines remain cached; temporary buffers are recreated on the next draw or readback. This lowers retained idle memory without requesting frames. A headless regression test checks repeated trim, readback and resumed rendering.

- `go vet` fixes for unkeyed struct literals in `gio/internal/f32` and
  `gio/app/internal/ibus`.
- `typesetting/font/opentype`: resources that expose their bytes (`Shared`,
  or any with a `Bytes() []byte` method) hand out table slices instead of
  copies and never write into a
  caller's buffer.
  `typesetting/fontscan` maps system font files read-only (`mmap`,
  `MapViewOfFile`) and parses them as `Shared`, so a face's tables are clean
  file-backed pages, not Go heap: hello's idle footprint went from 153 to
  103 MB on macOS. `gio/font/opentype.ParseCollectionShared` does the same
  for bytes that never change, used by `gio/font/gofont`; Keel's theme
  passes its font files with a `Bytes` method.
  Tested by parsing, describing, shaping and outlining every system font
  from read-only mappings (`fontscan/openfont_test.go`).
- `typesetting/harfbuzz`: an attachment chain pointing before the buffer
  returns as in HarfBuzz instead of indexing `pos[-1]`, which panicked
  shaping macOS's Farisi.ttf.
- `gio/gpu`: the path renderer's coverage textures (stencil and
  intersection FBOs, often larger than the window since they pack every
  rounded shape) are redrawn every frame, so after a committed frame they
  are marked volatile (`driver.Volatile`, Metal's purgeable state) and made
  non-volatile again when a frame resizes them for use. An idle window's
  largest GPU allocations leave its footprint: hello 72 -> 64 MB on macOS.
  Other backends do not implement it yet.
  Since an animating window would pay a kernel call per frame for it, the
  marking happens only when a frame asks for no frame within 100 ms
  (`app.Window` calls `gpu.Idle`, using `input.Router.PeekWakeup`, which
  does not consume the wakeup). Each texture remembers its state, so
  nothing is called while it is unchanged.
- `gio/gpu/internal/metal`: GPU buffers are pooled by power-of-two size
  class instead of created and released every frame (`newBuffer` and
  `CFRelease` took a fifth of an animating window's CPU). A buffer released
  during a frame is reused only after the next `BeginFrame`, which waits
  for the previous command buffer; the pool is bounded (32 MB, 64 per
  class) and emptied by `gpu.Idle`.
- `gio/gpu`: clip paths are cached by content (a hash of the path data,
  with stroke width, outline and transform) instead of by the ops that
  recorded them, whose key changes whenever the `Ops` are reset. Keel
  records each frame anew, so every border and rounded shape was
  re-tessellated and re-uploaded every frame. The offset is not part of the
  key, so moved content reuses its paths too. An animating gallery
  allocated 881 -> 470 MB per 10 s (51 -> 29 GCs). All static component
  screenshots are byte-identical.
- `gio/app` (macOS, iOS): `displayLink.Start` and `Stop` only send to the
  display link's goroutine when the requested state changes. Windows call
  `Start` on every animated frame, and the unbuffered send blocked the main
  thread on a goroutine handoff each time.
- `gio/app` (macOS): keep three `CAMetalLayer` drawables during rendering.
  Core Animation can retain two presented surfaces during resize; limiting
  the pool to two intermittently blocked the main thread in `nextDrawable`
  for about a second. Idle trimming still shrinks spare surfaces. The native
  `TestMetalResizeWindow` regression drives 60 growing/shrinking frames and
  rejects the compositor stall (`KEEL_DESKTOP=1 go test ./gio/app`). The
  native Metal tests keep their windows on top and skip, rather than hang,
  when the window is occluded anyway (covered, display asleep or locked),
  since an occluded window draws nothing.
- `gio/gpu`, `gio/gpu/internal/metal`: consecutive ops clipped to a plain
  rectangle and filled with a solid color or a texture are drawn as one
  instanced draw (`driver.QuadBatcher`, up to 8 textures a batch) instead of
  a draw call each: a terminal-sized text grid issued 3,300 draws a frame.
  The Metal shader is compiled at run time and computes what Gio's blit
  shaders compute; other backends keep the per-op path. Every static
  component screenshot is byte-identical.
- `typesetting/harfbuzz`: AAT shaping (`morx`, `kerx`, used by Apple fonts
  such as Menlo) keeps each subtable's glyph class cache on the font's
  accelerator, as HarfBuzz does; upstream copied it into a per-call context
  and threw the filled copy away. The context and the subtable drivers'
  state live in the reused `Buffer` instead of 1.3 KB and more allocated per
  run. Shaping output is identical for 239 macOS faces with `morx`; text
  allocation in the benchmark grid scene halved.
- `typesetting/fontscan`: `FontMap.ResolveFace` resolves ASCII runes from a
  per-query table and hashes the query's families once per `SetQuery`
  instead of once per rune. `SetQuery` and `SetScript` keep the caches when
  the query or script is unchanged; Gio calls both for every run of text it
  shapes. A test checks that cached and fresh maps resolve the same faces.
- `gio/app` (macOS): a window nobody can see (covered, on another space, or
  with the screen locked or asleep) draws nothing after its first frame and
  stops its display link; it redraws when it shows again
  (`windowDidChangeOcclusionState:`). AppKit reports changes only, so the
  state is also read when the view attaches. Gio drew every animation frame
  of hidden windows: the benchmark grid scene used 0.49 s of CPU per second
  with the screen locked, now 0.001 s.
- `gio/app` (macOS): with every display asleep, creating the display link
  failed and window creation then panicked in `gio_onDestroy` (a view
  without a handle). The display link falls back to the main display, and a
  view without a handle skips `gio_onDestroy`.

- `gio/op/paint`, `gio/gpu`, `gio/gpu/internal/metal`: the optional Metal-build
  `ImageOp.AddTinted` extension multiplies an immutable image by a per-draw
  linear color. Keel's glyph atlas stores white coverage once instead of
  duplicating it for every text color (the benchmark grid uses one 1 MiB page
  instead of six). Metal reserves the frame's quad storage before texture
  upload, so shared tinted textures cannot enter the untinted fallback during
  a render pass. Path clips, opacity layers and unavailable batching bake the
  tint into separately cached textures. Builds without the extension retain
  Keel's colored atlas pages through interface detection. Tests cover texture
  sharing, cache eviction, fallback colors, transparency, transforms, clipping
  and opacity; 8-bit coverage quantization can change edge pixels slightly.
  On an Apple M4, three alternating runs reduced median process footprint
  from 205 to 145 MB for the grid and 191 to 148 MB for widgets. A 240-frame
  offscreen grid test reduced median time from 2.776 to 2.452 ms/frame and
  allocation from 1.18 to 1.04 MB/frame. These are local measurements with
  both atlases enabled, not memory limits or promises for other applications.

- `gio/app` (macOS): after one second without drawing, a cancellable one-shot
  callback trims GPU textures, paths and scratch buffers, then shrinks the
  Metal drawable pool to 1×1. Core Animation retains the displayed frame;
  the next render restores the drawable size. Compiled pipelines stay alive.
  Native tests cover repeated idle/resume and closure with a pending callback.
  Native idle helpers use purego v0.11.1; the combined module's Go version
  is specified by the root go.mod.
- `gio/gpu`: color/texture/gradient fallback pipelines and the stencil renderer
  initialize on first use, avoiding unused shader resources in quad-only scenes.
  Headless pixel tests exercise rendering and readback before and after trimming.
- `gio/io/input`: discard adjacent drawing-only leaf clips when collecting input.
  Handler, cursor, semantic and window-action clips remain. Randomized tests
  compare hit testing against the original collector.
- `typesetting/font`: allocate glyph-extents cache pages only for measured
  glyphs. CFF fonts retain validated charstring INDEX offsets instead of a Go
  slice header per glyph; the public expanded CFF parser remains available.
  Tests cover cache boundaries, malformed INDEX data and equivalent outlines.
- `typesetting/fontscan`: share equal immutable system-font coverage sets,
  checking full contents after hashing; retain ownership of query family slices
  so caller reuse cannot corrupt resolution caches.
- `gio/text`, `typesetting/shaping`, `typesetting/harfbuzz`: cache parsed
  typefaces, bypass unnecessary wrapping and segmentation for eligible text,
  reuse normalization scratch and avoid copying shaping-plan keys.
  Fast paths preserve mandatory Unicode breaks and have regression tests.
- `typesetting/fontscan` (macOS): also scan `/System/Library/AssetsV2/com_apple_MobileAsset_Font*`, where current macOS keeps system fonts such as PingFang. Without it, "PingFang SC" never resolved and Chinese fell back to Hiragino Sans GB W3, with no Medium or Semibold. `fontscan/scan_darwin_test.go` resolves all three weights.
- `gio/text`, `gio/font`, `gio/widget`, `gio/widget/material`: platform glyph rasterizers. `Shaper.GlyphFile(id)` reports a glyph's font file, face index, size and glyph id (faces from `font.FaceSource` keep their path; built-in, variable-instance and varied faces report false). `Shaper.SetRasterHook` registers a function that draws a solid-color line of glyphs; `Label`, `Editor` and `Selectable` take an optional `Color` (set by `material.LabelStyle` and `EditorStyle`) and, with a hook, translate each line by whole pixels and pass the fraction to it, falling back to outlines when it reports false. Keel's theme sets the hook on macOS to draw with CoreText and detects optional hooks through interface assertions. This does not allow mixing upstream and Keel import paths.
- `gio/gpu`: an opaque full-window clear discards completed opacity layers together with their covered paint operations. Stale layer indexes otherwise replace unrelated foreground operations, causing missing glyphs in fullscreen overlays. Regression tests cover discarded layers, subsequent layers, partial covers and fills inside active opacity groups.

## Wayland teardown acknowledgment

The Wayland backend defers surface/display destruction until the next `Event`
call after returning the invalid `WaylandViewEvent`. Queuing that event is not
an acknowledgment: Keel must first release its scroll-watch pointer, seat and
private queue while the display is still alive. Previously `close` disconnected
immediately, causing a use-after-free in `wlScrollWatch.close` on Linux.
`deferred_teardown_test.go` covers notification, acknowledgment and single cleanup.

- `gio/app` (Wayland): queue `Run` callbacks for the native event thread instead of executing them on a background caller. Asynchronous maximize/configure otherwise races frame collection and can reset recorded operations while the GPU reads them. Shutdown stops queue wakeups before closing the native pipe and releases waiting `Window.Run` callers. A regression decodes a frame before draining a queued reset.

- `gio/gpu`: track opacity membership per image operation. An unrelated opacity layer no longer disables shared tinted textures for the whole frame; path clips and images inside layers retain baked colors. Texture-count and native pixel tests cover unrelated and nested opacity. CoreText uses 256×256 pages with the same 32 MiB pixel budget to reduce whole-page color copies in the fallback path; cache tests cover eviction and pages used in the current frame.

- `typesetting/fontscan` (macOS): scan the system FontServices `Resources/Reserved` directory. Current macOS stores PingFangUI.ttc there; omitting it silently falls back to another CJK family. A native font-resolution test covers Helvetica Neue and PingFang SC.

- `ui/theme` native glyph rendering: resolve the selected TTC face by its PostScript identity through purego. CoreText descriptor indexes differ from TTC face indexes after named-instance expansion. Native tests inspect the created PingFang SC Medium and SF Regular objects; this prevents the former from being drawn as HK Thin. No new cgo bindings.

- `typesetting/font/opentype/tables`: fix `fvar` named-instance record strides and bound each record so an absent PostScript name cannot consume the next record. `fontscan` indexes named weights with default non-weight axes, loads their coordinates, and bumps the disk cache to version 7. `gio/text.Shaper.GlyphFileInstance` exposes the selected instance to CoreText; native tests verify Regular and Semibold identities from shaping through rasterization.

### 2026-10-14: retained framebuffer and variable-font tables

- `gio/gpu/path.go`: calculate framebuffer waste as retained area / requested
  area. The inverse ratio never shrank an oversized framebuffer after a smaller
  frame. `TestFramebuffersReleaseOversizedStorage` covers shrinking, stable reuse
  and release.
- `typesetting/fontscan`: named instances of the same file and TTC index share
  immutable parsed font tables. Each instance retains a distinct `Font` identity
  and `Face`, preserving location metadata and independent variation coordinates.
  The Darwin PingFang instance test checks shared GSUB storage and distinct weights.
- `typesetting/font`: retain glyph headers and variation point counts at load
  time, then cache full TrueType contours only when an outline is requested.
  This avoids retaining every CJK contour merely for native text rendering.
  Shared caches are synchronized; the public Face scratch state remains separate.
  Tests compare every Go Regular outline and representative PingFang glyphs at
  four weights against eager decoding, including concurrent cache access.

### 2026-10-14: covered quad batching and conservative clip interiors

- Metal batches path/intersection-covered color and image quads with coverage
  textures in the existing eight-slot table. Coverage uses the original
  `min(abs(value), 1)` operation. Clipped tinted glyphs can share atlas textures;
  opacity layers and unsupported devices keep their existing fallback. Tests
  cover texture-slot exhaustion, ordering, tint, allocation fallback, and native
  pixels for rounded/nested clips.
- Rounded rectangle path operations carry a conservative interior rectangle.
  A paint can skip coverage only when every path ancestor contains its complete
  pixel bounds; an extra pixel margin protects antialiased edges. Unknown paths,
  strokes, rotations and unsupported transformed rectangles retain coverage.
  Collection tests cover these cases. The internal clip operation encoding grew
  by 16 bytes; all readers use the updated length.
- OneCatch's 5016×2820 comparison differed by at most 1/255 per channel. Actual
  window loading/scrolling still exceeds 300 MiB; neither pixel tests nor lower
  frame submission time establishes that memory target or end-to-end frame rate.

- `typesetting/shaping`: trailing whitespace trimming now verifies the source cluster contains only Unicode whitespace. Missing outline metrics (including the installed PingFang UI face) also yield zero ink width and must not erase the final visible glyph advance. Both the fitting multi-run path and wrapped-line path use the check. Synthetic LTR/RTL/vertical tests and mixed-font label width/pixel regressions cover it.
- `typesetting/shaping`: line breaking measures candidate lines with the same check. `Output.advanceSpaceAware` left a zero-ink final glyph out of the width as if it were a space, so with PingFang (no outline metrics) every wrapped CJK line held one glyph more than its width and the last glyph was clipped. It now drops the glyph only when its source text is whitespace (`whitespaceCluster`, shared with trimming). Synthetic tests cover the wrapped line count and the width check.

### 2026-10-14: presentation device-loss recovery

- `gio/app`: port upstream Gio commit [bf9f995](https://github.com/gioui/gio/commit/bf9f995d86c22d555d7ce2db154395bca94541b7). When `Present` reports `gpu.ErrDeviceLost` (including wrapped errors), release the GPU context and schedule an immediate new frame. The acknowledged frame cannot be retried. Keel only runs idle GPU reclamation after successful presentation, while the context remains locked.
- Regression tests cover success, direct/wrapped device loss, unrelated errors, acknowledgment count, and context lock/release ordering. The Windows cross-build passes; Windows client behavior has not been verified on hardware. Metal currently returns no presentation error, so this patch does not reduce macOS memory peaks.
