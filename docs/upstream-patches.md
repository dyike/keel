# Changes from upstream Gio and go-text

English | [简体中文](upstream-patches.zh-CN.md)

Keel keeps Gio's window, layout, input and GPU architecture, with patches for
resource reuse, text rendering and window lifecycle bugs. Its go-text copy adds
font-storage and shaping optimizations. This page explains the behavior changes
and where to inspect them; the [full patch log](../third_party/README.md) retains
historical measurements and maintenance notes.

## Baselines and module delivery

| Copy | Upstream baseline | Current source and import prefix |
| --- | --- | --- |
| Gio | `gioui.org` v0.10.3 | `gio/`; `github.com/dyike/keel/gio` |
| go-text | `github.com/go-text/typesetting` v0.3.5 | `typesetting/`; `github.com/dyike/keel/typesetting` |
| gogio packager | `gioui.org/cmd/gogio` v0.10.0 | `internal/gogio/`; CLI implementation detail |

The foundation copies are ordinary packages of Keel's root module. Their nested
go.mod/go.sum files are removed, dependencies are maintained in the root, and
internal imports, linker symbols and resource lookups use the new prefixes.
An application requiring Keel therefore receives the patched sources without
copying Keel's former replacements. `gioui.org/shader` remains an external module.

Import relocation changes named Go type identity. An upstream layout.Context
cannot be substituted for Keel's layout.Context simply because their fields
match. Applications and third-party widgets must migrate their imports; see
[One module](single-module.md). Optional interface detection for a rendering
extension is a fallback mechanism, not cross-package type compatibility.

The copies retain their licenses and upstream attribution. Upstream tests and
test data were not copied wholesale; the tests listed here are Keel's patch
regressions. This page was checked against the locally cached baseline sources,
separating import-only changes from implementation changes.

## Gio: GPU rendering and resource lifetime

| Area | Upstream behavior and Keel's change | Source and evidence |
| --- | --- | --- |
| Path caching | Resetting frame ops changes their identity. Keel caches paths by geometry, stroke and transform content, allowing repeated or translated shapes to reuse tessellation and uploads. | [gpu.go](../gio/gpu/gpu.go), [caches.go](../gio/gpu/caches.go); historical same-host component screenshot comparisons in the patch log. |
| Metal buffer pool | Repeated buffer creation/release is replaced by power-of-two size classes. Buffers retired in a frame become reusable after the next BeginFrame waits for prior GPU work. The pool is bounded and cleared on idle. | [pool_darwin.go](../gio/gpu/internal/metal/pool_darwin.go), [metal_darwin.go](../gio/gpu/internal/metal/metal_darwin.go); [idle/resume readback regression](../gio/gpu/headless/trim_darwin_test.go). |
| Quad batching | Instead of drawing every eligible operation separately, Metal batches consecutive solid/image fills with plain rectangular clipping into instanced draws, with up to eight textures per batch. Other backends and ineligible operations keep the per-operation path. | [quads_darwin.go](../gio/gpu/internal/metal/quads_darwin.go), [driver interfaces](../gio/gpu/internal/driver/driver.go); historical visual comparisons and current [GPU lifecycle tests](../gio/gpu/lazy_pipeline_test.go). |
| Tinted images | The optional Metal-build ImageOp.AddTinted extension stores white glyph coverage once and supplies color per draw. Path clipping, opacity or unavailable batching use cached color-baked textures. Coverage quantization can slightly change edge pixels. | [tint_darwin.go](../gio/op/paint/tint_darwin.go), [texture-sharing/fallback tests](../gio/gpu/tint_test.go), [pixel tests](../gio/op/paint/tint_darwin_test.go). |
| Idle resources | Temporary coverage textures become purgeable on Metal when no near-term frame is requested. Idle trimming releases caches and scratch buffers; after one second without drawing, macOS also shrinks spare drawable surfaces to 1×1. The next frame restores them. Compiled pipelines remain cached. | [path.go](../gio/gpu/path.go), [metal_idle_macos.go](../gio/app/metal_idle_macos.go), [trim regression](../gio/gpu/headless/trim_darwin_test.go), [native idle regression](../gio/app/metal_idle_macos_test.go). |
| Pipeline creation | Color, image, gradient and stencil resources initialize on first use, reducing unused shader allocations in simple scenes. | [gpu.go](../gio/gpu/gpu.go), [lazy pipeline lifecycle test](../gio/gpu/lazy_pipeline_test.go). |
| Mutable images | Upstream treats every ImageOp as immutable: changed pixels need a new ImageOp, so the renderer creates and uploads a whole new texture. paint.NewMutableImageOp keeps one texture per image; after Invalidate(rect) each renderer uploads only the areas changed since its last upload, or the whole image when it is more than 32 changes behind. paint.FilterLinearNoMipmap samples linearly without building mipmaps, for images that change often and are not drawn below half size. | [paint.go](../gio/op/paint/paint.go), [dirty.go](../gio/internal/dirty/dirty.go), [gpu.go](../gio/gpu/gpu.go); [partial upload, tint and filter tests](../gio/gpu/mutable_image_test.go). |
| Opaque clear correctness | Clearing the whole window now discards covered opacity layers and their paint operations together. Previously stale layer indexes could replace later foreground operations and hide glyphs in overlays. | [gpu.go](../gio/gpu/gpu.go), [opaque-clear regressions](../gio/gpu/opaque_clear_test.go). |

These patches extend Gio's renderer and resource policy. They do not establish
that every backend implements Metal-specific batching, tinting or purgeable
storage. Releasing scratch resources is also different from releasing live
textures, font faces or all process memory.

## Gio: text, editing and input

Gio still shapes text and computes its layout. Keel adds a raster hook so
`ui/theme` can supply CoreText glyph pixels on macOS; the Gio foundation does
not import Keel's theme or native packages.

| Change | Before and after | Source and regression |
| --- | --- | --- |
| Platform glyph rasterization | Shaper.GlyphFile exposes a file-backed face's path, collection index, size and glyph ID. SetRasterHook/Raster supply an optional solid-color glyph drawing hook; Label, Editor, Selectable and material styles pass color and fractional horizontal placement. A missing/declining hook keeps vector outlines. | [raster.go](../gio/text/raster.go), [shaper.go](../gio/text/shaper.go), [label raster regression](../gio/widget/label_raster_test.go); CoreText integration is in [ui/theme](../ui/theme/coretext_cache_darwin.go). |
| Shared font collections | ParseCollectionShared and built-in Go fonts use immutable backing bytes rather than duplicating font tables. Font.FaceSource carries file/collection identity for platform rasterizers. | [font API](../gio/font/font.go), [OpenType parser](../gio/font/opentype/opentype.go), [shared-table regression](../typesetting/fontscan/openfont_test.go). |
| Typeface parsing cache | Repeated text shaping reuses the most recently parsed typeface and family list instead of parsing them for every run. | [gotext.go](../gio/text/gotext.go); existing text regressions and historical allocation measurements. |
| Editor viewport API | ScrollBounds, ScrollOffset and ScrollTo expose multiline viewport control. Moving the viewport preserves selection; later caret movement/editing resumes caret reveal. | [editor.go](../gio/widget/editor.go), [editor scroll regression](../gio/widget/editor_scroll_test.go). |
| Input clip pruning | Adjacent drawing-only leaf clips no longer accumulate in input collection. Clips carrying handlers, cursor, semantics or window actions remain. | [pointer.go](../gio/io/input/pointer.go), [randomized routing comparisons](../gio/io/input/clip_prune_test.go). |

GlyphFile reports no usable file identity for in-memory or varied faces.
CoreText integration therefore needs fallback behavior; the hook does not
replace shaping, and its presence does not make the macOS text binding cgo-free.

## Gio: native window events

| Platform or area | Change and reason | Source and regression |
| --- | --- | --- |
| Darwin frame scheduling | Display-link Start/Stop only notify their goroutine when the requested state changes, avoiding a synchronous handoff on every animated frame. | [os_darwin.go](../gio/app/os_darwin.go); native idle/resize regressions and historical CPU measurements. |
| macOS visibility and sleep | Covered, off-space or sleeping windows stop animated drawing after their first frame and redraw when visible. Display-link creation can fall back to the main display; destruction ignores an unattached view without a handle. | [os_macos.go](../gio/app/os_macos.go), [os_macos.m](../gio/app/os_macos.m); native window checks, with sleep/space transitions still requiring manual coverage. |
| macOS resize | Three CAMetalLayer drawables remain available while rendering because Core Animation can retain two presented surfaces during resize. Idle trimming still shrinks spare surfaces. | [metal_macos.go](../gio/app/metal_macos.go), [native resize regression](../gio/app/metal_resize_macos_test.go). |
| Wayland teardown | An invalid WaylandViewEvent is returned before native destruction. Destruction waits for the consumer's next Event call, giving it time to release proxies and private queues while the connection is alive. | [deferred_teardown.go](../gio/app/deferred_teardown.go), [os_wayland.go](../gio/app/os_wayland.go), [teardown regression](../gio/app/deferred_teardown_test.go). |
| Wayland Window.Run | Callbacks are queued and drained on the native event thread rather than executed on a background caller. Shutdown stops wakeups and releases waiting callers, avoiding frame-ops resets racing GPU consumption. | [event_funcs.go](../gio/app/event_funcs.go), [queue regression](../gio/app/event_funcs_test.go), [Wayland lifecycle regression](../gio/app/os_wayland_lifecycle_test.go). |
| Native ABI feasibility | Added no-cgo tests exercise Objective-C aggregate callbacks, super calls and copied blocks. These are probes for later binding migration, not a replacement AppKit backend. | [ABI probes](../gio/internal/nativeabi/abi_darwin_test.go). |

There are also keyed-struct-literal vet fixes in
[internal/f32](../gio/internal/f32/f32.go) and
[app/internal/ibus](../gio/app/internal/ibus/ibus.go).

## go-text: font storage, lookup and shaping

| Area | Upstream behavior and Keel's change | Source and regression |
| --- | --- | --- |
| Font table storage | The baseline copies tables on read. Shared resources, including resources with Bytes(), borrow immutable table slices instead. System font files use read-only mmap/MapViewOfFile when available, with ordinary file fallback. | [shared.go](../typesetting/font/opentype/shared.go), [reader.go](../typesetting/font/opentype/reader.go), [mapped-font/shared-table tests](../typesetting/fontscan/openfont_test.go). |
| Font cache size | Glyph-extents storage allocates 128-glyph pages only when measured. CFF parsing retains validated charstring INDEX offsets instead of a slice header per glyph; the expanded parser remains available. Equal immutable coverage sets share storage after full-content comparison. | [cache.go](../typesetting/font/cache.go), [cache-page test](../typesetting/font/cache_pages_test.go), [CFF tests](../typesetting/font/cff/compact_test.go), [coverage test](../typesetting/fontscan/share_coverage_test.go). |
| Font resolution | ASCII faces and family hashes are cached per query. Unchanged query/script selections retain caches, and SetQuery owns the family slice so caller reuse cannot corrupt it. macOS scanning also includes the system asset-font directories used for PingFang. | [fontmap.go](../typesetting/fontscan/fontmap.go), [query/cache tests](../typesetting/fontscan/resolve_cache_test.go), [PingFang regression](../typesetting/fontscan/scan_darwin_test.go). |
| AAT and shaping scratch | Apple-font glyph-class caches stay on font accelerators; AAT and normalization working state reuse Buffer storage. Shaping-plan cache lookup avoids constructing/copying a full plan key for a hit. | [aat_layout.go](../typesetting/harfbuzz/aat_layout.go), [buffer.go](../typesetting/harfbuzz/buffer.go), [shape.go](../typesetting/harfbuzz/shape.go); historical font-equivalence and allocation comparisons. |
| Segmentation and wrapping | Eligible LTR paragraphs bypass unnecessary bidi segmentation. Several runs that fit on one line use a fast wrapping path; mandatory Unicode breaks, bidi ordering and whitespace handling still match the general path. | [input.go](../typesetting/shaping/input.go), [wrapping.go](../typesetting/shaping/wrapping.go), [LTR](../typesetting/shaping/simple_ltr_test.go), [wrapping](../typesetting/shaping/fast_wrap_test.go) and [mandatory-break](../typesetting/shaping/mandatory_break_test.go) regressions. |
| Attachment-chain correctness | A glyph attachment pointing before the buffer returns instead of indexing pos[-1], avoiding a panic observed while shaping Farisi.ttf. | [ot_layout_gpos.go](../typesetting/harfbuzz/ot_layout_gpos.go), [attachment regression](../typesetting/harfbuzz/ot_layout_gpos_keel_test.go). |

Shared bytes must stay unchanged and alive while parsed faces use them.
Successful font mappings are deliberately retained for the faces' process
lifetime. Less Go-heap allocation does not mean no mapped memory or permission
to mutate borrowed font data.

## Packager adaptation and compatibility limits

The copied gogio packager has a serialized Run entrypoint with per-run working
directory, environment, flags and output. GioRoot supplies application-ID and
launch-argument linker symbols, Android permission prefixes and mobile resource
discovery. Default output paths honor the supplied working directory.
See [Run](../internal/gogio/run.go) and
[packaging regressions](../internal/gogio/run_test.go).

Keel owns Windows executable resources separately. Its iOS simulator shader
adapter and SDK/signing orchestration live in the CLI. The
[real-module release fixture](../scripts/check-module-release.py) verifies
external apps and a custom widget, versioned CLI installation, Windows/Web
packaging and macOS packaged IDs/arguments.

The macOS idle helper already uses purego, but AppKit, Metal/GL/headless and
CoreText still have cgo bindings. Linux X11/Wayland/EGL/GL/Vulkan migration is
also pending. The ABI probes currently establish native arm64 feasibility;
native amd64 execution and complete IME protocol parity remain open.
See [current validation evidence](desktop-cgo-free-validation.md), rather than
interpreting the patch list as Desktop CGO-Free acceptance.

## Validation and upstream updates

From the repository root:

```sh
go test ./gio/... ./typesetting/... ./cmd/keel/... ./internal/deps
go test ./internal/site
```

On a macOS desktop, native Metal checks are opt-in:

```sh
KEEL_DESKTOP=1 go test ./gio/app -run 'TestMetalIdleWindow|TestMetalResizeWindow' -count=1
```

Some regressions require system fonts, a GPU or a specific platform and may
skip elsewhere. Historical performance numbers in the patch log describe the
recorded machine and workload, not a fresh benchmark or a portable guarantee.

For an upstream update, compare the recorded baseline with the candidate
version, distinguish import/module relocation from behavior patches, and
reapply the latter with their regressions. Preserve each license, update both
language versions of the patch log, verify external-module delivery, and run
visual/native checks appropriate to the changed paths.
