# Testing

English | [简体中文](testing.zh-CN.md)

Most interface tests do not require opening a window: the `ui/internal/uitest` testing tool sends clicks, keys, and text through Gio's real input routing. The code executed by the component is exactly the same as that of the real window. `go test -race ./...` takes about 10 seconds to run on this machine.

## Linux scaling and window size (local, 2026-11-10)

On Linux amd64 / Hyprland, the 3840×2160 output uses 125% display scaling and
GNOME text-scaling-factor 1.6364. Native protocol traces confirm preferred_scale
150, viewport destinations in logical units and a 2000×1250 buffer for a
1600×1000 window. Before explicit xdg geometry, the same test window could map
at its 560×460 minimum. Programmatic resize also exposed a lost surface-refresh
flag during frame coalescing; a regression test now preserves that flag.

The native scaling fixture exercises 20 resize requests with strict size
acknowledgements, scroll offsets, a saved size and reopening. Both EGL and
Vulkan (`noopengl`) passed on this host. Existing real-window lifecycle tests
also passed. Unit tests cover 100/125/150/200% display densities, independent
font scale, half-pixel rounding, integer fallback, frameless titlebar accounting
and state-file round trips. Focused race tests and vet pass; Linux amd64/arm64
and Windows no-cgo builds pass. A native Workspace screenshot was inspected.

The full suite still fails in existing fontscan, input ink alignment,
hover/status layout and EGL pixel tests. The legacy `giocgo` build cannot be
validated here because Vulkan development headers are absent; disabling Vulkan
then hits a legacy cgo listener pointer-type compiler error. Local commands
use `-gcflags=all=-goversion=go1.27.3` for this host's toolchain-version mismatch.

This is unpublished local work. The first native frame can use integer output
scale before the compositor reports its preferred fractional scale; AutoSize
may adjust once after mapping. Tiling/window-manager policy can override size
requests. Other fractional densities, cross-monitor transitions and real IME
candidate placement have unit coverage or code changes, not native acceptance
on this host. No end-to-end frame-rate or memory improvement is claimed.

## Linux default binding selection (local, 2026-11-10)

Linux desktop now selects purego without build tags, including when CGO is
on. `TestLinuxDefaultBindings` checks amd64/arm64 source selection, the
`giocgo` fallback and precedence of `purego` / disabled CGO. CI runs this
selection check with CGO enabled alongside the no-cgo runtime jobs.

On this Linux amd64 host, whole-module builds, vet, focused binding tests,
real-window lifecycle tests, no-cgo Linux/Windows builds and hello builds with
`purego`, `nox11`, `nowayland` and `novulkan` passed. Local compilation required
`-gcflags=all=-goversion=go1.27.3` because the installed Go toolchain reports
inconsistent version strings; this is not a required Keel build option.
The full test suite did not pass: fontscan, input ink alignment, hover/status
layout and EGL pixel tests failed. Explicit `purego` also reproduces the EGL
failures. This is an unpublished source-selection change, not new evidence of
GPU/IME parity across Linux hardware.


## Motion integration, 2026-11-06 (working branch)

The animation work is **not part of v0.4.2**. Native-window validation now covers
pose, layout-size, shaping-cluster text morph and numeric rolling, including a
resize from 640×400dp to 720×460dp at 2×. Each 600ms animation is observed for
1.2 seconds. Three alternating runs compare the same executable and content
with retained painting enabled versus `-full`, which rebuilds every frame.

Each phase submits **37–38 layout callbacks**. Retained painting runs Render
**once per phase**, versus 37–38 times with full rebuilding. All six runs finish
with **zero additional layout callbacks during three idle seconds**. This proves
retained rendering and settled scheduling, not display presentation FPS or
zero process CPU usage; delayed memory collection can still run without a frame.

| Small native fixture, MiB | Full rebuild median (range) | Retained median (range) |
| --- | ---: | ---: |
| Physical footprint, final two sampled seconds | 46.09 (45.16–46.39) | 45.49 (45.44–46.31) |
| Kernel lifetime peak | 108.50 (108.38–112.52) | 108.14 (107.92–112.30) |

The ranges overlap. **No significant total-memory reduction is established.**
Measurements use `proc_pid_rusage` v4, nominally every 50ms, on Apple M4,
darwin/arm64, Go 1.27.3, `CGO_ENABLED=0`, with identical stripped, trimpath
builds. The display is unlocked, trials are serial, and no tests or builds run
alongside them. No fixture GC is forced; normal idle reclamation remains on.
Three seconds of idle is shorter than the gallery's 20-second recovery period.

The fixture is `ui/window/testdata/pose`; raw samples and phase logs are under
`work/motion-integration/motion-{full,retained}-*.{log,json}`. Text tests also
compare real headless-GPU pixels at entry, interruption, retained/full frames
and rest, including CJK, combining accents, color ranges and emoji. They verify
that an empty string completes its exit and that offscreen/omitted transitions
release old glyphs. These checks do not replace native scroll, drag, presence
interaction or the final validation after shared-element support is added.

### Shared elements and native interaction

The later shared-layout checkpoint adds two native fixture phases: thumbnail
expansion and return, each a 600ms tween observed for 1.2 seconds. The six-phase
`TestRealWindows/./testdata/pose` run passes with retained painting, window resize
and a three-second idle check (`native-shared.log`). This is animation validation;
it does not update the earlier memory measurements.

Native mouse interaction in Components additionally verifies transformed-card
clicks, horizontal drag/snap to 240dp and back to 0dp, scrolling, number/text
changes, icon/image replacement, thumbnail/detail round trips, list replacement
and reordering, width changes, and dialog close/reopen/Esc. A fast release exposed
excessive spring overshoot; settling now absorbs outward velocity at the same
elastic limit as direct dragging. The extreme-velocity regression and a native
mouse rerun pass. These observations are not presentation frame-rate measurements.

Shared-layout tests cover cross-parent interruption, short-path rotation,
spring velocity continuity, destination clips, pointer/anchor geometry, resize,
density, reduced motion and cleanup. Retained/full pixels match byte-for-byte.
Equivalent rounded paths across a change of scale use a two-channel-level edge
quantization tolerance for handoff/interruption comparisons.

The subsequent `Fly` checkpoint passes a seven-phase native fixture, including
an inert ancestor while its shared child is still moving (`native-fly-final.log`).
Real mouse input in the rebuilt Components gallery expands the photo out of its
scroll container and returns it by clicking the displayed image. Pixel/input
tests verify that ancestor clips and opacity do not hide a flight, its own opacity
and local theme survive, anchored overlays follow it, and later modals stay above
it. A flying child outside its dialog's rectangle does not dismiss that dialog;
a genuinely outside click still does. Inert flights continue moving and fading
while input, automation and platform accessibility remain suppressed.


The appearance/text checkpoint adds base color, radius, border-width and opacity
interpolation and cross-owner Text transitions. GPU tests cover different box
proportions, font sizes, interrupted reversal, retained/full parity and exact rest;
they also caught and fixed precision loss from reciprocal affine scales. The
seven-phase native fixture now includes these properties and an inert shared
label (`native-shared-look.log`, passed). A rebuilt native Components window
expands the rounded photo and title together; clicking the displayed title
returns both. These are animation results, not new memory measurements.


Final integration checks pass with Go 1.27.3 on darwin/arm64: `go test ./...`,
`CGO_ENABLED=0 go test ./...`, `go test -race ./...`, `go vet ./...`, the
CGO-free Windows build and Linux native/CLI vet. The component workflow passes
800 matrix images, 100 headless scenes and nine interaction groups. Logs are
`work/motion-integration/final-*.log` and `component-final/summary.json`.
These checks do not establish native runtime support on Windows or Linux.

### Gradient, automatic text and scroll binding, 2026-11-07

The working branch now interpolates shared-element gradients and Elevation shadows,
matches descendant Text by relative paths with `MatchText`, and binds timelines to
scroll containers with `ScrollBinding`. Explicit text identities and manual scroll
sampling already exist in the inspected MyGo implementation; the latter two APIs
are Keel conveniences, not evidence of previously missing MyGo behavior.

GPU tests cover gradient/shadow handoff, interrupted reversal and retained/full
parity. A scaled gradient exposed a Gio paint-transform bug: gradient sampling now
uses the inverse paint transform. Dedicated Gio affine and headless pixel tests
cover that fix. Text tests cover nested owners, cached children, explicit scopes
and cleanup. Vertical and horizontal scroll tests verify content-range changes,
explicit ranges and no repeated wakeups after settling.

Default and CGO-free full tests, race tests, vet, Windows CGO-free builds and Linux
native/CLI vet pass on Go 1.27.3. The updated seven-phase native fixture passes,
including resize, an inert text container and three idle seconds. Evidence is in
`work/motion-integration/followup-*.log`. The component rerun passes 800 matrix
images, 100 headless scenes and nine interaction groups (`component-followup/summary.json`).
The two GIF previews use actual Keel GPU
rendering with a controlled clock; they are not desktop recordings or FPS evidence.
These changes are local and unreleased. Intermediate-size text reflow and arbitrary
multi/inset shadows are still outside this implementation; full MyGo parity is not
claimed. This checkpoint does not update the memory measurements below.

### Matching Components comparison

Three alternating unlocked native-window trials compare v0.4.2 (`d4d26d0`) with
the motion integration checkpoint. Both show the same gallery: the new Motion
example is excluded from this comparison's registry so content stays identical.
The scenario uses 1200×860dp / 2400×1720 pixels, default home animations,
600 acknowledged scroll updates, 20 resizes, eight page visits, 20 seconds idle
and resumed scrolling. All six scroll phases render 601 callbacks. The same
Go 1.27.3/CGO-free build settings and serial physical-footprint sampling apply.

| MiB | v0.4.2 median (range) | Motion checkpoint median (range) |
| --- | ---: | ---: |
| Startup sampled peak | 151.88 (151.77–152.41) | 151.97 (136.02–152.11) |
| Home animation sampled peak | 152.50 (152.35–152.66) | 152.24 (136.53–152.28) |
| Continuous-scroll sampled peak | 161.44 (161.33–161.60) | 161.50 (161.16–161.95) |
| Resize sampled peak | 172.99 (172.06–180.02) | 172.42 (172.03–179.94) |
| Page-switching sampled peak | 174.13 (173.80–175.75) | 173.86 (173.67–174.22) |
| Idle physical footprint | 68.14 (68.11–68.74) | 67.99 (67.97–72.51) |
| Kernel lifetime peak | 184.35 (180.75–184.41) | 184.39 (184.35–184.61) |

There is **no material total-memory reduction**. Current trial 3 has eight idle
layout callbacks versus one in the other five trials, and a higher final idle
reading. It remains in the results. An extra diagnostic rerun has one idle
callback with no pending input and 67.67 MiB idle footprint; it does not reproduce
or explain the earlier event. Do not classify it as a fixed leak or discard it.

Maximum layout callback time over the scenario is 15.78–19.39ms before and
19.84–20.56ms at this checkpoint. Startup-phase maxima are 15.49–16.67ms and
15.32–20.16ms; resumed-layout completion is 13.92–14.45ms and 14.73–16.30ms.
These CPU timings do not measure first-frame presentation latency or display FPS.
The higher page-switching maxima and isolated idle activity remain under review.

The matched, stripped scenario executables are 20,153,666 and 20,325,282 bytes:
+171,616 bytes (+0.85%). This excludes the new gallery demo and is not a packaged
application-size measurement. The native/memory/size samples precede the later
text-exit-clock refinement; final measurements after remaining features are
still required. Source hashes and all raw data are in `work/motion-integration/`.
The first attempted baseline used `/tmp` overlay keys while Go resolved
`/private/tmp`, so it ran the ordinary entry point. It is explicitly excluded
under `excluded-overlay-path/`; corrected builds verify the entry and commit.

## Components inline-font memory, 2026-11-06 (local)

Three alternating before/after native-window trials reduce the full gallery's
post-interaction idle physical footprint from **75.55 to 67.27 MiB (−11.0%)**.
This is a local, unpublished change on top of the text-cache/font-reclamation
work below. It does not establish a broad reduction during animation or scrolling.

Inline input objects used a synthetic font without native font data or a name
table. This made the whole editor fall back to Go shaping, loading a system font
index and parsed font tables even though ordinary text already used CoreText.
The synthetic font now supplies immutable SFNT data, Unicode name records and an
glyph-name-free `post` table. Its PostScript name derives from metrics and rune
mappings, so independent editors with different object widths retain distinct
font identities. The font remains local to its shaper; it is not registered
system-wide. Go shaping remains available on other platforms and when native
shaping is unavailable.

The host is Apple M4, darwin/arm64, Go 1.27.3, `CGO_ENABLED=0`, with identical
stripped builds and unlocked foreground windows. Each trial opens the full
Components gallery at 1200×860dp / 2400×1720 pixels, keeps its default Button
animations running, scrolls VirtualList through 600 acknowledged updates,
resizes 20 times between 1040×760dp and 1200×860dp, and visits Input, Table,
TextView, Button, Table, Input, TextView and VirtualList. It then idles for 20
seconds and resumes scrolling. The baseline renders 601–602 callbacks during the
scroll phase; the modified build renders 601. Measurements are serial, without
builds/tests/profiling or fixture GC; the default idle collector stays enabled. `proc_pid_rusage` v4 provides
physical footprint and the kernel lifetime peak. Samples are nominally 50ms;
idle is the last three seconds before resume.

| Measurement, MiB | Before median (range) | After median (range) |
| --- | ---: | ---: |
| Idle physical footprint | 75.55 (73.89–75.94) | 67.27 (66.97–68.08) |
| Kernel lifetime peak | 187.95 (187.08–189.77) | 184.63 (179.89–184.64) |
| Startup sampled peak | 135.35 (135.17–135.85) | 152.05 (135.45–152.35) |
| Home animation sampled peak | 136.17 (136.17–136.80) | 152.11 (136.28–152.70) |
| Scrolling sampled peak | 161.25 (161.03–161.45) | 161.69 (161.33–161.75) |
| Resizing sampled peak | 172.63 (163.70–180.03) | 171.97 (171.86–184.61) |
| Page-switching sampled peak | 182.56 (181.69–184.38) | 174.17 (173.89–175.03) |

The 8.28 MiB idle reduction is reproducible in this series. The lifetime-peak
median decreases only 1.8%; scrolling is unchanged and resize ranges overlap.
Startup/home ranges also overlap: the change has not run before visiting Input,
so their different medians must not be credited to it. Separate diagnostics show
16 MiB drawable-count steps; unexplained driver-owned allocations remain.
Window-layer experiments without reproducible benefit were discarded.

The maximum layout callback during page switching is 34.79–54.64ms before and
16.07–22.05ms after. Startup maxima are 16.32–20.07ms and 15.56–16.62ms;
request-to-resumed-layout completion is 13.76–15.83ms and 11.45–14.63ms. These
are CPU/layout timings, not presentation latency or end-to-end FPS. In particular,
the first native/font initialization still has a visible-frame latency cost.
Regression tests cover native font identity, BMP/supplementary private-use runes,
1×/2× advances, wrapping and caret selection, while existing input tests cover
Unicode composition, undo and source-preserving reference content. Raw serial
samples and the local scenario overlay are under `work/components-total/`.

Validation passes default full tests and vet, CGO-free full tests, Windows
CGO-free builds, Linux native/cmd vet, and 99 gallery pages plus nine native
interaction groups. The first full race run failed an existing icon allocation
count assertion; five isolated runs and a subsequent full race run passed,
without relaxing the assertion. Go 1.27.3 also passes editorstyle, el and
components tests. English/Chinese input screenshots were visually checked;
byte-identical native/Go text shaping is not claimed.

An early experiment accidentally compared a Go 1.27.3 baseline against a
Go 1.27.1 modified build. That series is excluded from these results. All final
`matched-*` trials use Go 1.27.3, with build metadata checked before startup and
saved per trial. The raw data summary is
`work/components-total/matched-summary.json`.

## Text cache and native font reclamation, 2026-11-06 (local)

The retained changes reduce long-text physical footprint: three alternating
real-window trials give **105.96 → 68.14 MiB idle (−35.7%)** and
**193.60 → 151.31 MiB lifetime peak (−21.8%)**. The ordinary Components list
shows no improvement. These are unpublished working-tree changes, not results
for an existing release.

Paragraph layouts now have an estimated 8 MiB ordinary byte budget, with up to
64 MiB for layouts used within eight explicit shaper frames; the existing 1,000
entry limit still applies. `Shaper.BeginFrame` protects a larger visible working
set from repeated eviction. Without it, the ordinary budget applies. Cached
paragraph keys own their strings, so a short substring does not keep an entire
input document alive. This follows MyGo's byte-budget approach without changing
line wrapping or glyph positions.

On macOS, the theme borrows native fonts synchronously. A 512-object LRU reclaims
unpinned sized fonts while preserving their face indices and exact descriptors.
A descriptor is eligible only after it recreates a `CFEqual` font; other fonts
remain pinned. Existing `GlyphNativeFont` handles retain their shaper-long
lifetime. Descriptors, stable identity records, the existing 256-entry primary
font cache and pinned fonts are outside that LRU budget. This is not an upper
bound on all text or native memory.

Measurements use Apple M4, Go 1.27.3, `CGO_ENABLED=0`, identical stripped builds,
unlocked native windows on a 2× display, and serial trials without concurrent
builds/tests. The baseline retains the preceding lazy-state and idle-collector
changes. `proc_pid_rusage` v4 provides physical footprint and the kernel lifetime
peak; polling is nominally 50 ms, with foreground-activation overhead. Idle is
the mean of the last three seconds of a 12-second quiet period, before resume.
The fixture forces no GC; Keel's existing default idle collector remains enabled.

The long-text window contains 600 mixed-language paragraphs of about 2,000
characters each, at 800×600dp (1600×1200 pixels). It executes 600 scrolling
updates, 20 resizes alternating 900×660dp and 800×600dp, and 480 text-size
updates. Every scroll/text-size update waits for its layout callback; each
accepted trial rendered 600 scrolling and 481 size-change/reset callbacks.
Input is requested at 60 Hz; the framework's frame rate is unchanged. The
ordinary text component rounds sizes to pixels, so these size updates alone
do not exercise the 512-font limit.

| Long-text measurement, MiB | Before median (range) | After median (range) |
| --- | ---: | ---: |
| Idle physical footprint | 105.96 (103.63–107.64) | 68.14 (66.52–69.33) |
| Kernel lifetime peak | 193.60 (190.31–196.67) | 151.31 (151.16–151.67) |
| Startup sampled peak | 105.02 | 105.02 |
| Scrolling sampled peak | 169.19 | 131.70 |
| Resizing sampled peak | 186.25 | 151.16 |
| Text-size changes sampled peak | 193.31 | 140.78 |

The layout-budget-only control gives 68.42 MiB idle and 151.45 MiB kernel peak
(medians, three trials), locating the main benefit in paragraph caching.
Startup maximum layout callbacks range 7.75–8.76 ms before and 8.72–10.83 ms
after; size-change maxima reach 17.72 ms before and 15.76 ms after. The median
request-to-resumed-layout completion is 15.03 → 16.12 ms. These measurements do
not establish improved latency, presentation timing or end-to-end frame rate.

A separate real-window font stress case compares the layout-budget-only control
with the complete change. Four font families render the same short mixed-language
text across 640 fractional pixel sizes, with each update acknowledged by layout,
at 800×600dp. Three alternating trials give:

| Font stress measurement | Layout budget only | With font reclamation |
| --- | ---: | ---: |
| Idle physical footprint, MiB median (range) | 218.78 (218.72–219.92) | 212.22 (212.19–212.37) |
| Kernel lifetime peak, MiB median (range) | 310.30 (302.86–313.50) | 306.20 (304.67–306.55) |
| Resident sized font objects after churn | 3,841 | 1,623 |

The font change independently saves 6.56 MiB (3.0%) at idle in this stress case;
peak ranges overlap, so a reproducible peak benefit is not established. The
1,623 resident objects comprise 512 reclaimable and 1,111 pinned fonts. All
3,841 stable identities remain valid. This is not a claim that ordinary clients
create thousands of fonts or will obtain the same saving.

The existing Components VirtualList regression fixture uses 680×860dp, 600
scroll requests, 40 resizes alternating 820×720dp, then 12 idle seconds. Three
trials per build render 598–599 scrolling callbacks. Idle physical footprint is
53.65 (52.93–55.11) → 55.62 (54.22–58.20) MiB; kernel peak is
148.20 (148.05–148.30) → 148.88 (139.58–149.14) MiB. There is no demonstrated
saving here: the idle median is 1.97 MiB higher, with overlapping ranges. Final
Go heaps stay about 5.4 MiB. The measurements do not identify the cause of the
remaining native/driver/process difference.

Byte/entry limits, oversized paragraphs, visible working-set reuse and substring
ownership have regression tests. Native tests cover old glyph IDs after layout
and font eviction, permanent-handle compatibility and redundant-descriptor
release. Raw masks remain byte-identical for Latin, Chinese, Arabic, emoji and
variable weight/italic text in light and dark ink. Full default/no-cgo tests,
focused race tests, vet, Windows/Linux no-cgo builds and unlocked `RealWindows`
passed. No resolution reduction, frame-rate cap or per-frame GC was introduced.

Raw records, binaries, source overlays, hashes and the aggregate `results.json`
are under local `work/text-memory/`. Locked, occluded, incomplete-frame and
concurrent-check pilots are excluded. Phase sample peaks must not be substituted
for the kernel lifetime peak; layout acknowledgements do not prove GPU
presentation. These results apply to the measured workloads, not every client.


## Idle physical-footprint follow-up, 2026-11-06 (local)

The allocation-only patch above did not meet a total-memory goal. A separate
follow-up allows the existing idle collector to consider a 4 MiB heap after
at least 8 MiB of new allocations, instead of requiring an 8 MiB heap.
The unused-page threshold remains 8 MiB; two seconds of quiet, the five-second
cooldown, frame-lock exclusion and no periodic collection of static windows
remain unchanged. Applications can still disable idle reclamation.

The same native VirtualList workload and sampling method was repeated against
the preceding allocation-only working tree. Three uncontended trials per
version, including one confirmation build from final source, gave:

| Measurement (MiB) | Previous version median (range) | Follow-up median (range) |
| --- | ---: | ---: |
| Idle physical footprint | 55.17 (54.85–55.20) | 53.53 (51.62–54.93) |
| Kernel lifetime physical peak | 146.35 (146.28–146.41) | 146.66 (137.56–146.85) |
| Final Go heap | 7.34 (7.28–7.69) | 5.37 (5.37–5.37) |

Idle physical footprint improves by 1.64 MiB (3.0%) in this series, with some
range overlap. Peak does not improve. This is not a demonstrated improvement
across all client workloads or over every earlier baseline: the earlier
pre-allocation-patch idle median was 52.23 MiB in a different series. The
fixture itself forces no GC; the default collector ran once in each follow-up
trial and zero times in the comparison trials. Resume latency after idle has
not been profiled. A trial overlapping gallery checks was excluded.

An experiment lazily parsing all 74 Chroma styles showed overlapping physical
memory results and was reverted, including copied assets and API changes.
Full tests, focused loop/el/window race tests, vet and the complete unlocked
`RealWindows` gate pass for the retained follow-up. Boundary tests verify
both allocation and heap thresholds and that unchanged state never schedules
periodic collection. This remains an unpublished local change.


## Lazy input state, 2026-11-06 (local)

Element state now allocates text-editing metadata only for inputs. The editor
and its metadata share one allocation with stable addresses; buttons, keyed
virtual rows and scroll containers retain only their common state. Rendering,
font selection, buffering and reclamation policy are unchanged.

Against `9545628`, Apple M4, Go 1.27.1, three benchmark runs:

| Measurement | Before | After |
| --- | ---: | ---: |
| Ordinary element struct | 2,392 B | 1,888 B |
| Ordinary state heap allocation (allocator size class) | 2,688 B | 2,048 B |
| Input state plus editor heap allocation | 4,480 B | 4,352 B |
| 60 keyed rows, replacing three per update | 12,832 B/update | 10,912 B/update |
| Same row benchmark allocation count | 74/update | 74/update |
| Components VirtualList scroll, median | 92,017 B/frame | 90,112 B/frame |
| VirtualList scroll CPU time, median | 148.38 µs/frame | 150.74 µs/frame |

The scene allocation reduction is 2.1%, not the synthetic benchmark's 15%.
CPU time did not improve. These are layout/input-routing benchmarks without
GPU presentation, not physical footprint or end-to-end frame rate.

The original locked-screen native attempt rendered only one frame and is
excluded. After unlocking, three interleaved comparisons used the actual
Components VirtualList window at 680×860dp (2×), 10 seconds of programmatic
scrolling, 10 seconds alternating 820×720dp and 680×860dp, then 12 idle seconds.
Both versions rendered about 598 scroll frames and 60 resize frames. A first
trial overlapping another diagnostic is excluded. Go/GPU reclamation defaults
were retained; the fixture did not force GC. Physical footprint was sampled
every 50ms; idle is the final three seconds. Medians in MiB:

| Native measurement | Before | After |
| --- | ---: | ---: |
| Kernel lifetime physical peak | 146.69 | 146.38 |
| Idle physical footprint | 52.23 | 55.76 |
| Total Go allocation during 10s scrolling | 66.57 | 65.51 |
| Total Go allocation during 10s resizing | 4.79 | 4.72 |

No material peak improvement is established, and observed idle footprint
increases. Default idle reclamation ran once in the median before run and
zero times after; that difference may contribute to the idle result, but is
not proof of its cause. This is an allocation improvement, not a demonstrated
client-memory reduction. Native and headless gallery interaction checks each
pass all 99 scenes. The changes remain local and unpublished.

Regression coverage checks that ordinary states allocate no editor metadata,
input addresses survive rerenders, and removed input states leave the store.
All 792 fixed-time GPU screenshots are byte-identical; Windows/Linux no-cgo
cross-builds pass.
Existing input/document/IME, selection, clipboard, undo and disabled-state tests
pass, as do full cgo/no-cgo tests, vet and focused race tests.


### Headless lifecycle diagnostic

Run while the desktop is locked or in a session without native windows:

```sh
KEEL_MEMORY_REPORT=/tmp/keel-memory.json KEEL_MEMORY_ROUNDS=120 go test ./examples/components -run '^TestComponentMemoryLifecycle$' -count=1
KEEL_MEMORY_REPORT=/tmp/keel-memory-gpu.json KEEL_MEMORY_GPU=1 go test ./examples/components -run '^TestComponentMemoryLifecycle$' -count=1
```

Each round opens and closes 20 windows across VirtualList, Table, Input and
TextView, exercises list scrolling and two sizes, then collects Go memory.
The GPU mode also renders screenshots for each scene at both sizes. The last
sample follows six idle seconds and another collection. `KEEL_MEMORY_PROFILE=1`
adds a heap profile. These forced collections belong only to this opt-in
diagnostic; they do not establish idle physical footprint. A passing run
validates operations, not leak freedom: inspect the retained-memory series.


The headless gallery passed all 99 scenes while the desktop was locked.
The extended lifecycle run then completed 8,000 windows per version. Both
versions stabilize after roughly 4,000 windows: the last 100 rounds retain
19.06–19.37 MiB before and 19.09–19.42 MiB after, with final post-idle heaps
of 19.22 MiB in both and two goroutines. An earlier 2,400-window run was too
short to reveal this plateau; sampled growth was concentrated in the bounded
CoreText glyph cache. This finite run shows no continuing linear growth in
that workload, not proof that all components are leak-free.

A separate GPU lifecycle run per version completed 140 windows and 56
screenshots across both sizes, ending near 4.69 MiB retained Go heap. These
are not GPU/driver-memory measurements. The new diagnostic passes a focused
race run; native gallery checks and the native resize-specific test pass
after unlocking. The broader native gate was not rerun. Detailed local
artifacts are under `work/components-memory/`.


## Components dependency selection, local follow-up

Components keeps all 99 scenes and selects only the optional dependencies its
fixtures use: `ui/highlight/golang` for Go/template/plain-text snippets, and
local/data-URL images without `ui/netimage`. The original full-language
`ui/highlight` and HTTP image provider remain available to applications.
Unknown languages in this gallery now render uncolored; automatic language
detection is limited to the selected registry. HTTP image URLs are not enabled
in this example. This is dependency selection for Components, not an equivalent
reduction for applications that require all languages and remote images.

Go 1.27.1, `CGO_ENABLED=0`, `GOWORK=off`,
`-trimpath -buildvcs=false -ldflags="-s -w"`. Before and after use the same
working tree (including the native decoration split below); the baseline
restores the two previous imports through a Go overlay. Executable sizes:

| Target | Before (MiB) | After (MiB) | Reduction |
| --- | ---: | ---: | ---: |
| macOS arm64 | 24.54 | 19.04 | 22.4% |
| Windows amd64 | 25.76 | 20.01 | 22.3% |
| Linux amd64 | 26.33 | 20.63 | 21.7% |

On macOS, removing the unused HTTP image import alone saves 2.98 MiB;
selecting Go rules then saves another 2.52 MiB. No renderer, font, component,
resolution or frame-rate changes are part of this step. No physical-memory
reduction is claimed. These changes are local and unpublished.

The fixed-time matrix matches all 792 PNGs byte for byte; all 99 visible
window scenes and gallery interaction checks pass. An initial comparison
without a fixed clock differed in 72 animation images; the fixed-time rerun
eliminated those differences. Supported-language parity tests compare the
selected provider against the full provider, including templates, malformed
code, Unicode and multiple styles. Dependency tests reject `net/http` and
the global Chroma lexer registry in Components. Full cgo/no-cgo tests, vet,
focused race tests, desktop cross-builds of Components and its WebAssembly
build pass. This does not resolve the four native-window gate failures
recorded in the preceding optimization; that gate was not rerun here.


## Native window dependency split, local follow-up

Against `19b3db8`, macOS and Windows no longer compile Gio's client-side
fallback title bar into `gio/app`. Native title bars remain; Linux and the
other platforms retain the existing fallback. No font features are removed
from Keel widgets. Go 1.27.1, `CGO_ENABLED=0`, `GOWORK=off`,
`-trimpath -buildvcs=false -ldflags="-s -w"`, executable MiB:

| Client | macOS arm64 before → after | Windows amd64 before → after |
| --- | ---: | ---: |
| Empty Gio window (event loop and frame submission, no text) | 6.718 → 3.364 | 6.499 → 2.998 |
| Empty Keel window | 9.220 → 9.220 | 9.083 → 9.075 |
| Hello | 10.473 → 10.457 | 10.723 → 10.715 |

The roughly 50% reduction belongs to the bare Gio client, not Hello. Keel
still needs its text/theme stack. Linux retains fallback decorations and
these clients grow by 4–8 KiB from the split. The proposed additional 30%
Hello reduction has not been achieved. No GUI memory reduction is claimed.

`noopengl` now removes GPU registration as well as EGL/WebGL headless
contexts. It previously left the GPU backend linked despite excluding app
contexts. On macOS, the updated empty Gio client is 3.139 MiB with the tag;
Hello is 10.312 MiB. Use the tag only with another supported backend (for
example Metal); it does not provide a replacement for WebGL. Normal builds
retain OpenGL. Font removal and disabling development tooling were measured
only with local overlays and are not part of the change.

Full tests with and without cgo, race, vet, Windows/Linux full cross-builds,
WebAssembly build, fallback-decoration tests through a host overlay, and
Metal idle/resize tests pass. The component matrix has 792 byte-identical
PNGs and all 99 gallery scenes pass interaction checks. The workflow's
comparison step initially lacked baseline metadata; a separate complete
byte comparison verified the images without rerendering them.

The complete `RealWindows` gate is **not passing** on the inspected host:
accessibility press, text-cache redraw/resize, glass and the resize probe
fail with both backends. The same four failures reproduce on `19b3db8`
without this change. These failures remain unresolved; the passing Metal
checks do not replace that gate. Windows/Linux runtime behavior is unverified.

## Binary dependency measurements, 2026-11-05

Compared with main `1759621`, this unpublished local change restricts native
X11/DBus implementations to Linux and uses UTF-8 SVG parsing by default.
`-tags keel_legacy_svg` restores upstream legacy charset decoding. Font
fallbacks, renderer selection and automation remain unchanged.

Go 1.27.1 on darwin/arm64; each target uses `CGO_ENABLED=0`, `GOWORK=off`,
`-trimpath -buildvcs=false -ldflags="-s -w"`. Values are executable MiB,
not installers or runtime memory. The native probe imports `native/permission`
and calls `Granted` with a flag-provided kind (default -1, avoiding native
permission requests); its sleep option permits separate memory sampling.

| Target | Hello before → after | Components before → after | Native probe before → after |
| --- | ---: | ---: | ---: |
| macOS arm64 | 11.25 → 10.47 | 25.87 → 24.54 | 3.29 → 2.31 |
| Windows amd64 | 11.50 → 10.72 | 26.78 → 25.77 | 2.34 → 2.16 |
| Linux amd64 | 14.11 → 13.31 | 27.11 → 26.33 | 2.87 → 2.87 |

With macOS's default cgo backend, Hello changes from 11.09 to 10.32 MiB,
Components from 25.48 to 24.18 MiB and the native probe from 3.29 to 2.30 MiB.
Linux retains the native protocols it actually uses. These results do not
establish a GUI physical-memory reduction or Windows/Linux runtime behavior.

Default and no-cgo full tests, vet, race tests and Windows/Linux no-cgo full
builds pass. Parser tests compare pixels with upstream at 1× and 2×;
legacy-tag tests decode Big5, GBK, Shift-JIS and Windows-1252. Native protocol
unit tests also pass on macOS using a test-only build-constraint overlay;
Linux's protocol test binary cross-compiles. Dependency tests check desktop
target selection and reject legacy charset tables in default kit builds.

The first 792-image matrix comparison differed in 12 dark 1× images.
Three interleaved fresh-process comparisons of each affected scene/width all
matched; a complete fresh matrix repeat then matched all 792 PNGs byte for
byte. The initial differences were not reproduced. The real-window gallery
passed all 99 scenes and its interaction checks. No renderer changes were
made in response to the initial screenshot differences.

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

## Testing your app

`ui/window/windowtest` draws a window in memory and drives it the way a person
does: by the visible text and roles of its elements. No native window opens, so
the tests run in CI and with `go test ./...`.

```go
func TestGreeting(t *testing.T) {
    app := windowtest.New(t, window.Options{Content: el.Root(newApp())})
    app.Click("Name")      // the text box labelled Name
    app.Type("Ada")
    app.Click("Say hello")
    app.Want("Hello, Ada!")
}
```

| Method | Effect |
| --- | --- |
| `windowtest.New(t, opts)` | Open the window in memory for this test; 800×600 dp unless `opts` sets a size |
| `app.Click(text)` | Click the element whose name or value best matches `text`: exact over partial, controls over plain text |
| `app.Type(s)`, `app.Press("enter")` | Type into the focused text box; send a key chord (`"tab"`, `"mod+a"`) |
| `app.Scroll(text, dy)` | Turn the wheel over an element |
| `app.Want(text)`, `app.WantNot(text)`, `app.Has(text)` | Check what is visible |
| `app.Find(text)`, `app.Elements()` | Read roles, values, checked and selected states and bounds |
| `app.Resize(w, h)`, `app.Screenshot()` | Change the size; render a PNG with the GPU |

Every query and action renders first, running functions already passed to
`core.Update`, so the result of a goroutine that finished is visible. A failed
lookup lists the visible elements. Elements are what agents see through
`keel-mcp`; see [Agent end-to-end test](automation.md) for roles. Reduced motion
is on, so nothing is caught mid-animation. `window.OpenHeadless` offers the
same actions with errors instead of test failures.

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

On macOS, `TestRealWindowsResize` opens ordinary and frosted-glass windows
and grows/shrinks each through 60 Gio size requests plus 60 native frame
changes with the window center held fixed, then four native zoom/restore
transitions through `ToggleMaximize`. It waits for AppKit's terminal resize
frames before testing fixed-size scrolling. It verifies the NSView uses
`duringViewResize` redraw and `topLeft` content placement, so AppKit redraws
animated size changes instead of scaling cached content. It records the content layer's
`anchorPoint` separately from gravity and checks that its model X/Y transform
scale stays at 1. It also checks non-scaling, top-left layer gravity (accounting for
`contentsAreFlipped`), that drawable size matches layer bounds × backing
scale, and that glass resize presentation runs on the main thread with
`presentsWithTransaction` enabled. Each window also scrolls its text content
through 60 programmatic offsets. Stable-size frames stay asynchronous,
including recovery from idle drawable trimming. A static window must stop
drawing; the probe reports process CPU over two idle seconds after trimming.
The 600ms deadline catches stalls, not display FPS. These property
and scheduling checks do not inspect the compositor's intermediate pixels
or the window snapshots used by system animations. Also drag-resize and
maximize/restore `examples/hello` and `examples/glass`, check text stays
undistorted and correctly anchored, and check idle CPU after GPU completion.
The fix lives in `ui/window` and adds no cgo or Gio patch.

Skipped by default as it will pop up on the screen and requires a graphical interface environment. You must run it once after changing the code related to windows and locks in `ui/window` and `ui/internal/loop`. When adding a new window method, add it to the `testdata/raise` step.

## Parts that require manual verification

The interfaceless test cannot cover these, so you need to change the relevant code and run it manually:

| Scenario | How to verify |
| --- | --- |
| Open, close, and bring to front the real window | `go run ./examples/components -section multiwindow -lang en`, click "Open settings" twice; there should be only one settings window |
| Exit after closing the last window | Close all windows and the process in the terminal should end |
| Window shortcut keys | Press ⌘+, | in the main window
| Global shortcut keys | `go run ./examples/components -section hotkey -lang en`, open the preview and enable the shortcut, then switch applications and press Cmd+Shift+K on macOS or Ctrl+Shift+K elsewhere; the count increases |
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
