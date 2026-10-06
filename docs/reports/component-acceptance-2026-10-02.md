# Component completion acceptance · 2026-10-02

English | [简体中文](component-acceptance-2026-10-02.zh-CN.md)

The original A–E checklist for implementation and automation acceptance has been completed; the visual specification, screenshot matrix, sample process and migration documentation for F have been completed. The following original acceptance records correspond to `0bce878`; subsequent progress is listed separately to `bc54e85`, and the old verification results are not automatically extended to the new code. The complete native scene acceptance F3 is still not completed.

For a comparison of the current 77 components, see [GPUI Kit implementation progress](gpui-progress-2026-10-02.md).

## Follow-up progress (as of `bc54e85`)

| Capability | Implementation | Existing evidence / Verification boundaries |
| --- | --- | --- |
| WebAssembly | `28a0e80` | Build regression, browser hello Chinese, input, checkbox and button; full component browser acceptance is not done |
| Visual Basics | `83c50ed` | Rounded corners/font size/shadow scale, transparency and font style, kit migration; light and dark overlay shadows checked |
| Dock maximization | `bf9f69d` | Menu/double-click to enter, Esc to restore, layout saving and regression testing |
| Multi-theme basics | `bf9f69d` | Registration and JSON parsing test, Nord/Paper example, actual switching |
| CodeEditor | `4f22179` | Line number, highlighting, selection/undo/input method/auto-indent, and diagnostic/completion/hover interface |
| Editor completion repair | `bc54e85` | Real machine input gr pops up, press enter to accept; regression verification Agent completion item, press enter/click to accept and close |
| 200,000 lines of interaction | `bc54e85` | Real machine loading, scrolling, end jump, input; example corrected to exactly 200,000 lines. No frame rate or input lag captured |

The final code of this round passed `go test ./ui/window ./ui/kit`; during the previous work, `go test ./...` passed. The full build/vet/race was not re-executed for this completion fix, so the old check results below only belong to the original acceptance batch.

The component library will be able to open real windows in the future, and the old DisplayLink failure record no longer means that all native windows are inoperable; the title bar, system preferences, and multi-windows are still subject to special review. System screen reading/VoiceOver is suspended and not connected.

## Original batch automation results

The following commands all pass:

```sh
go build ./...
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
```

The above checks were re-run after the last feature modification. `cmd/keel-mcp` covers order, chat, settings and Dock process-level processes; a new keyboard submission test without clicking is added to the order. The tests for kit/window are supplemented with real pointer drag and drop, disabled propagation, focus restoration, overlay boundaries, Agent semantics, themed pixels, and 1× / 2× geometry.

The native Git pre-commit hook is not enabled; verification is done with the above command. Not pushed to the remote end.

## Visual inspection

79 registered display areas, each generating light/dark, 320/680dp, 1×/2×, a total of 632 first frame screenshots. After rebuilding the sample state and rendering them separately, the output directory is `/tmp/keel-component-matrix` and the index is `index.html` therein. The file will not be entered into the repository; it can be rebuilt by pressing [Test document](../testing.md#full-component-screenshot-matrix).

Manually review the light-dark narrow layout overview and review the original image for the question; mixed text, number, and button sizes are covered by examples and existing regressions. This round of centralized inspection found and fixed:

| Objects | Questions and Results | Submit |
| --- | --- | --- |
| OTP | The six-digit verification code is cropped; each grid is shrunk to the same width as the editing area | `4d8629e` |
| Pagination | The last page button is out of bounds; automatic line wrapping, and correction of total overflow | `3301337` |
| Calendar | The outer frame shrinks, but the seven inner columns do not shrink; the seven columns share the available width | `c739e4d` |
| Settings | Fixed sidebar, control column extrusion description; narrow layout changed to top partition, arranged up and down | `ca1ea83` |
| Display container | Fixed wide canvas invalidates narrow window acceptance; limits maximum width, operation line wrapping | `374e199` |
| Stepper | Subsequent steps are not reachable; add horizontal scrolling, disabling and navigation regression | `cfbbc06` |
| Toolbar | The first frame command covers the right slot; retract more first, then expand after measuring | `784d307` |
| Dock | The default docking size of the first frame squeezes out the center; initialize and correct according to the root viewport | `57cf5f1` |
| Order example | The input box after Mod+N has no focus; the next frame is handed over to the customer field | `48e0723` |
| TextArea | Old errors still appear after user correction; cleared simultaneously with single-line input | `0bce878` |

The screenshot only covers the current viewport and does not represent pixel-by-pixel verification of all interaction states. Overlay, content after scrolling and animation interruption are added to correspond to behavioral testing. Dock multi-panels, grids with fixed minimum column widths, etc. still require the application to select an appropriate minimum window size; this does not constitute a commitment to mobile adaptation. Font pixels change with the system font, and PNG golden images are not submitted across machines.

## Performance

Go 1.26.4, darwin/arm64, Apple M4; run three times, median ns/op in the table. This is a native baseline and is not a cross-device performance guarantee.

```sh
go test ./ui/kit -run '^$' -bench 'BenchmarkVariableListFrame|BenchmarkChartDenseSamples' -benchmem -count=3
```

| Benchmark | Median time | Allocation |
| --- | --- | --- |
| Variable List Stable Frame, 1k rows | 27.53μs | 39,019 B, 234 times |
| Variable list stable frame, 100,000 rows | 28.00μs | 39,018 B, 234 times |
| Polyline 100,000 sampling points bucketed | 124.27μs | 32,792 B, 2 times |

List benchmark stable frame, does not include first indexing, batch SetKeys, or rendering of all rows; chart benchmark data sample, does not represent full frame GPU time.

## Unfinished native acceptance

The actual execution of `KEEL_DESKTOP=1 go test -run RealWindows ./ui/window` and `go run ./examples/frameless` in the original batch failed. Gio fires `runtime/cgo: misuse of an invalid Handle` in the cleanup path that created the DisplayLink, before entering the title bar interaction. Building the same frameless example with `2b1648e` before native bridging also fails.

Subsequently, the component library can be opened and the editor can be verified, but the following scenarios still need to be re-verified: multi-window bringing to front/closing, title bar dragging, system double-click preference, window out-of-focus appearance, and real-time notification of macOS reduced animation. Off-screen regression and native code compilation are passed, but do not replace these verifications. Global shortcut keys, system permissions and Chinese input method are still executed according to [Manual verification table](../testing.md#parts-that-require-manual-verification).

## Scope

For migration instructions, see [Migrate to current kit](../migration-kit.md), and for visual rules, see [Component Visual Specification](../visual-guidelines.md). The basic code editor has been implemented in the future, and the advanced editing functions have not yet been completed. HTML rich text, complete TeX, partial theme coverage, and Kbd action binding queries continue to be suspended; cross-window Dock is not included in this round, and full platform screen reading integration is suspended according to the current decision. What is accepted here is the implementation list in the repository, not a one-to-one certification with all GPUI Kit APIs.
