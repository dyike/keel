# Rendering memory and reuse

English | [简体中文](rendering-memory.zh-CN.md)

Components are layout and state boundaries, not retained pixel surfaces. The current
renderer collects a frame and draws it into a drawable. No general damage-region or
scroll-pixel reuse is implemented by the upload optimization described here.

## What the upload change removes

On the tested Apple M5 Pro, a separate 1×1 Metal clear-only program used about
97 MiB after submission. Adding a 4 KiB buffer copy raised it to about 154 MiB.
The same small allocation cannot account for that difference: submitting the copy
activates additional driver-owned memory. This identifies a workload that triggers
allocation, not private kernel object names or a universal Metal memory constant.

The application trace similarly rose by about 156 MiB when its first coverage
pass completed, even with path geometry draws disabled. That command buffer also
contained uploads. Application textures at that point occupied only about 7 MiB.
It was incorrect to attribute that entire jump to path complexity.

For unified-memory devices, immutable geometry now starts in shared storage, and
a sampled image's first full upload writes shared storage directly. Small mutable
uniform buffers keep their existing path. Mip levels are generated on the CPU with
sRGB decoding/filtering/encoding and linear alpha. There is no retained CPU image
copy. Native comparisons observed at most 2/255 channel differences from Metal's
mipmap generator, including odd dimensions. These are not bit-identical filters.

Later or partial image uploads keep GPU copy ordering and regenerate the entire mip
chain on the GPU. Mutable buffer writes, readback and other GPU copies also retain
the original path. Such workloads can activate the driver allocation again. The
optimization does not establish a process-wide memory cap for every application.
Non-unified devices and failed shared-texture allocations fall back to private
resources. Autorelease pools stay on a locked OS thread; selectors are cached.

The tested CPU mip-chain costs were approximately 0.05 ms at 96², 0.33 ms at 256²,
and 5.25 ms at 1024². Very large images need separate responsiveness profiling;
these figures must not be extrapolated to arbitrary image sizes.

## Damage-region rendering

A useful implementation must track the union of old and new bounds for changed
paint operations, including clipping, transforms, opacity, images and selection.
Layout displacement also damages the previous location. Unknown operations,
resizing, device loss, scale changes and changed clear/background state require a
full redraw. Skipping component layout must not skip input handling.

The renderer currently has no scissor capability in its common driver interface.
An optional driver extension can introduce a damage scissor, with a full redraw
fallback. The window background must be painted inside the damage region rather
than clearing the whole attachment. Offscreen coverage and opacity targets need
independent invalidation; the window scissor cannot be applied to their atlas
coordinates.

Crucially, a drawable is not necessarily the immediately preceding frame. A partial
redraw needs a verified native content-preservation contract and per-drawable damage
history. Otherwise it needs an owned backing surface. At 2920×1898, one RGBA surface
costs about 21.1 MiB; at 5016×2820, about 54.0 MiB before alignment. Allocate that
only after accounting for transient old surfaces and the process peak budget.

## Scroll reuse

Scrolling damages the whole conversation viewport even when message text is
unchanged. A bounded tile cache could reuse unchanged message pixels and redraw
only newly exposed strips. Its key must include stable message identity, content
revision, width, scale, font/theme revisions and rendering effects. Selection and
hover may be drawn as separate overlays only where composition stays equivalent.
Fractional pixel movement, transformed clips and filter boundaries need explicit
pixel tests or a normal-rendering fallback.

Use a fixed byte budget, including resources retained by in-flight commands, rather
than one texture per message. Cache misses and budget exhaustion must draw normally.
Reuse should use the existing render path; introducing a GPU copy solely to shift
pixels can reactivate the memory this patch avoids. Before implementation, compare
cache memory and fill-rate savings against the existing message virtualization.

Acceptance requires native-window comparisons at identical pixel dimensions:
loading, repeated scroll reversals, streaming tool updates, text selection/copy,
resizing and idle recovery. Report physical footprint peaks and presented-frame
intervals separately from CPU submission times. Preserve a full-render reference
for pixel comparison and never use a smaller window to claim a memory improvement.

## Local application validation (2026-10-14)

The installed OneCatch client was compared at the same 2920×1898 pixel size, historical conversation, and three round trips of three-screen scrolling. Builds with presentation timing recorded process peaks of 364 MiB before and 304 MiB after the change. The final build without temporary timing recorded 293 MiB; shrinking and restoring the window subsequently raised its cumulative peak to 333 MiB. One fixed-size run does not establish a stable 300 MiB ceiling across scenarios.

Actual scrolling, resize recovery, and Chinese selection/copy worked. For the subset of presentation intervals below 100 ms, both builds had P50/P95 of 8.33/33.33 ms, but the maximum increased from 41.67 to 91.67 ms. Input timing and sample counts differed; these numbers do not establish unchanged tail latency. New streaming tool output and very large images still need client validation. Detailed samples and evidence paths are recorded in OneCatch's `docs/keel-memory.md`. This patch remains a local workspace version.

## Sharing coverage within a frame (2026-10-14)

Coverage masks with identical path content, absolute offset, and final clip bounds now share an atlas placement and are rasterized once within a frame. Different parents still compose independently in the intersection pass. Paint order, opacity, and final draws are unchanged. The CPU placement index is cleared each frame and released on sustained idle; no GPU texture or persistent scroll-pixel cache is added.

Absolute transform inputs remain in the key to preserve floating-point evaluation order. Tests reject sharing across different offsets, clips, strokes, and paths. A 24-frame native Metal comparison was byte-identical to the previous implementation, including repeated/nested clips, opacity, movement, idle, and trim recovery. For a synthetic headless workload with 100 repeated rounded clips, the median of three runs improved from about 1.68 to 1.12 ms/frame; this is not client frame rate.

An installed-client trace at 5016×2820 reduced cumulative path draws from 27,185 to 15,945. At the same size, loading and three scrolling round trips recorded process peaks of 448 MiB before and 441 MiB after with presentation instrumentation. The final build without temporary tracing recorded 434 MiB, settling to about 143 MiB. The single-run difference is small and does not establish a stable memory reduction or a 300 MiB ceiling. Presentation sample counts differed substantially, so unchanged tail latency is not established.

The earlier part of this round used 2920×1898 and reproduced an approximately 345 MiB resize peak before this change. Desktop dimensions changed after unlocking; later 5016×2820 comparisons are separate and must not be compared with the previous 293 MiB result. Final-build resize peaks at the original size still need verification.

A whole-coverage-atlas cross-frame reuse experiment had zero hits across 114 client frames and was removed. Releasing completed command buffers earlier also failed to establish a resource-lifetime benefit: the old implementation already released the test texture after completion. General damage rendering and scroll-pixel reuse are not implemented; within-frame deduplication must not be described as regional redraw.

Extending the native 1×1 clear experiment showed footprint falling from about 97 to 11 MiB in the second idle second, while Metal resource accounting remained about 0.875 MiB. Both runs, with and without marking the texture volatile, showed this fall. This identifies about 86 MiB of draw-related delayed reclamation on this machine. A high reading immediately after releasing objects does not establish permanent retention; no public control to force this reclamation without affecting responsiveness was identified.

Footprint category snapshots of the minimal process confirmed unmapped graphics falling from about 85 MiB while active to 640 KiB while idle, with IOAccelerator mappings falling from about 3 MiB to 832 KiB. This locates delayed reclamation in driver graphics accounting rather than application texture size.
