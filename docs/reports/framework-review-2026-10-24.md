# Framework review · 2026-10-24

English | [简体中文](framework-review-2026-10-24.zh-CN.md)

A review of macOS performance, usability and the scaffold, from a cross-platform
application developer's point of view. Measurements were taken on an Apple M4
with a 60 Hz 4K display. Items marked *unverified* are risks that still need a
check on the platform concerned.

## Measuring

Go's `pprof` misattributes time spent in cgo and the kernel on macOS: it put
~10% of CPU in Metal `newBuffer` and ~23% in `pthread_cond_signal`, while the
Instruments Time Profiler showed 2–3% and under 1%. Use Instruments
(`xcrun xctrace record --template 'Time Profiler' --attach <pid>`) for native
attribution, count over all samples rather than a truncated top list, and
compare runs at the same frame rate. CPU percentages are not linear in frame
count: a lighter load runs on efficiency cores at lower clocks.

## Baseline

| Scenario | Idle footprint | Startup peak | Idle CPU |
| --- | --- | --- | --- |
| `examples/hello` | 37 MB | 96 MB | 0% |
| `examples/chat` | 44 MB | 103 MB | 0% |
| `examples/components`, Button page (3 spinners) | 149 MB | 149 MB | 27–29% of a core |

Idle applications draw nothing and use no CPU. The cost appears as soon as
anything animates: every frame re-runs the whole window (Render, layout,
paint, input and GPU collection, a Core Animation commit). On the Button page,
Keel's own layout and paint were about 45% of process CPU, Gio's frame
processing about 20% and AppKit/Core Animation about 25%.

## Changes made

| Change | Status | Effect on the Button page |
| --- | --- | --- |
| Looping indicators at a shared 30 Hz cadence (`cx.AnimatingAmbient`) | merged, #50 | process CPU 219 → 114 ms/s; energy impact −17% |
| A glyph run as one operation (`paint.ImageOp.PaintQuads`) | merged, #51 | process CPU −10%; Gio `gpu.collect` −33%; pixels byte-identical |
| Pooled immutable Metal buffers | prepared, not committed | removes `newBuffer`/`dealloc`, 2–3% of CPU |

Text was about 4,000 of the page's 5,300 operations per frame before #51. A
Gio event-loop change that removed one of four channel hand-offs per frame
saved about 0.3 µs a frame and was not kept.

## Remaining performance work

Ordered by expected benefit for the effort.

1. **Cache shaped glyphs for unchanged labels.** Keel caches measurements, but
   painting still calls `Shaper.LayoutString` (hashing the string) and rebuilds
   each glyph with `NextGlyph` every frame. About 8% of CPU.
2. **Per-frame bookkeeping in `ui/el`.** Tab navigation walks the tree twice a
   frame (`focusedContainerTrap`, `collectTabTargets`) even without a Tab
   press; each node's state is looked up in `store.states` five or more times
   a frame (assignKeys, prepareKeys, tab walks, blockTree, paint). About 8–10%.
3. **Paint-only frames.** When only an animation advances (no input, updates
   or state changes), reuse the previous element tree and layout and re-run
   paint. About 30% of an animation frame; needs care in the frame flow.
4. **Retained subtrees or damage regions.** The largest gain, including
   scrolling, and an architectural change; see
   [rendering memory](../rendering-memory.md) for the constraints.
5. **Startup peak.** `hello` settles at 37 MB but peaks at 96 MB during the
   first frames; Activity Monitor shows the peak first.
6. **120 Hz displays.** Every per-frame saving doubles on ProMotion; not yet
   measured on a 120 Hz panel.

## Usability and production readiness

### Blocking for production

1. **Cross-compilation.** macOS and Linux UI builds still need cgo (Gio's
   window and GPU bindings); only Windows builds with `CGO_ENABLED=0`. CI needs
   one machine per desktop OS. See the [delivery design](../desktop-cgo-free.md).
2. **Accessibility.** No bridge to NSAccessibility, UIAccessibility, UI
   Automation or AT-SPI exists, so screen readers cannot read Keel windows.
   The semantic tree used by agent automation (roles, names, values, bounds)
   is the natural source for such bridges.
3. **Text outside macOS** (*unverified*). Only macOS renders through the
   platform (CoreText). Windows and Linux rasterize with go-text, so hinting,
   subpixel rendering and font fallback (CJK, emoji) may differ from native
   applications.
4. **Mobile.** `native/*` returns `native.ErrUnsupported` on iOS and Android;
   iOS is verified in the simulator only; examples have no mobile layouts. A
   published support matrix should say so.

### Developer experience

5. **Two kinds of children.** `el.Text(...)` is an `el.Element`, while kit
   components need `.Render(cx)`, because `el.Element` has an unexported
   method. Stateful kit views (`kit.Input`) must be kept in fields, while
   stateless ones (`kit.Button`) are built inline; nothing in the types tells
   them apart.
6. **Threading.** Changing view fields from a goroutine is a data race unless
   it goes through `core.Update`; only `-race` reveals it. The scaffold and
   guides need a "load data, then update the view" example, and possibly a
   helper.
7. **Profiling.** There is no built-in frame overlay (frame time, frames per
   second, operation and glyph counts), and `pprof` misleads on macOS; see
   [Measuring](#measuring).
8. **Reload.** `keel run` rebuilds and restarts on save, losing view state.
9. **Size and toolchain.** `hello` is 17 MB (11.6 MB stripped). Go 1.26.1 is
   the minimum.

### Scaffold and documentation (fixed with this report)

10. The getting-started guide and `cmd/keel` README said `go install …@latest`
    needed an unpublished single-module release; v0.3.8 is one.
11. `keel new` generated an `app.go` with Chinese text only.
12. The generated app told users to re-run `keel run` after saving, although
    `keel run` watches and restarts by default.

### Strengths to keep

A single module requirement, `keel doctor`, development app bundles with icons
from `keel run`, `uitest` interaction tests and `keel-mcp`, which lets agents
drive real windows end to end.
