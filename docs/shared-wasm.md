# Shared Go WebAssembly experiment

English | [简体中文](shared-wasm.zh-CN.md)

The build produces a full program and independent category programs. The full Gallery now defaults to the full program to retain cross-category state; embedded component previews still select category programs. This restores state continuity at the cost of a larger initial download, not shared-runtime splitting. This experiment investigates a different path: split a single linked Go/Gio program with [Binaryen wasm-split](https://github.com/WebAssembly/binaryen/tree/version_133/src/tools/wasm-split), keeping one Go runtime, memory, mutable globals and dispatch tables. It is not a new Go dynamic-linking ABI or a production default.

## Reproduce

Use Binaryen version_133's native tools (build-time tools only; no cgo in the application):

```sh
python3 scripts/probe-shared-wasm.py --binaryen-dir /path/to/binaryen/bin --out /tmp/keel-shared --optimize
# Optional for CJK: copy the site font.otf into /tmp/keel-shared.
python3 -m http.server 8767 --bind 127.0.0.1 --directory /tmp/keel-shared
# In another terminal, with Playwright available to Node:
node scripts/verify-shared-gallery.cjs http://127.0.0.1:8767/ work/gallery-shared
```

The browser probe enters text through touch focus and the actual input event path, loads the Dialog child, opens/closes its dialog, returns to Input, and selects/copies the original text through the clipboard. It checks that the memory object is unchanged, the child downloads only on explicit demand and repeated loading does not fetch again. Reports include browser version, requests and screenshots. Use a matching Go compiler and standard library; no compiler-version override is needed.

## Observed scope and limits

A local Chromium 159 run passed with 26 Dialog/gallery functions in a child importing memory, tables, mutable globals and functions from the primary instance. There is no second Go runtime. The experimental host now installs an optional Gallery construction gate: ordinary navigation and deep links wait for the child, show a loading/error page, and offer retry. Independent category and ordinary full builds do not install this hook. A late completion only redraws the current matching page and never changes the navigation destination. Downloads are deduplicated, and failed downloads may be retried without restarting Go. The original run covered only Dialog; the broader experiment below does not establish general category ownership.

This small partition establishes feasibility, not a download-size win. In the measured unoptimized build, the original module was 32,119,816 bytes (gzip-9 7,808,048), while the rewritten primary was 32,143,687 (gzip-9 8,389,066), plus the child (gzip-9 15,231). Binaryen rewriting can change compression materially. It must not replace the existing category build on these results.

Production gates remain: deterministic function ownership for every category and shared dependency; no unloaded function reached by startup, GC, asynchronous events or nested components; async navigation with visible progress/errors and retry; deep links and back/forward while loading; data/state retained across categories; cache/version consistency; smaller measured first transfer and acceptable first-call/frame/memory costs; repeated real browser and device acceptance. Keep the independent/full build fallback until these pass. Never serve mismatched primary/child builds.

The reproducible `--optimize` option applies `wasm-opt -Oz --strip-debug` equally to the unsplit baseline, primary and child and writes raw/gzip-9 measurements to `sizes.json`. In the newer navigation-gated build, the baseline is 29,157,010 bytes / gzip 7,668,807; primary 29,114,017 / gzip 7,652,877; child 46,478 / gzip 13,033. That saves only 15,930 compressed bytes initially and 2,897 after loading Dialog. It is not evidence of a useful category-scale improvement. Startup, first-call, frame and physical-memory comparisons remain unmeasured.

The prototype now content-addresses the primary, child, loader and startup script and records their pairing in `shared-build.json`. Keep referenced older assets available when deploying a new index; content-addressed names alone do not validate production cache headers or prove atomic deployment. A Chromium 159 test held the child request, navigated back to Input, released the request and checked copied input plus the unchanged route; forward navigation then opened the loaded Dialog. The same run verified touch retry after an aborted child request. Reports for this build are local verification artifacts, not evidence of deployment or physical-phone support.

## Multiple lazy partitions

Pass `--partitions dialog,nav_stack,emoji_picker,terminal` to build four children with Binaryen's `--multi-split`. The manifest records explicit function ownership, and every child imports the same primary instance. The browser verifier checks independent lazy downloads, unchanged memory identity and copied input after visiting all four partitions, in addition to retry and history handling.

Dialog's shared kit methods remain in the primary: both Motion and Image use them without visiting the Dialog demo. The original 26-function prototype incorrectly assigned those shared methods to Dialog; the new partition contains only the Dialog demo factory and methods. A Motion visit before any Dialog download is now part of the regression. The other three partitions currently own their view methods; their supporting libraries remain resident. These explicit selections are an experiment, not a general dependency ownership algorithm or a production size improvement.

With the shared Dialog methods kept resident, a fresh Go 1.27.4 four-partition build measured gzip-9 baseline **7,812,742 bytes**, primary **7,780,262**, and children **5,833 / 5,722 / 15,148 / 13,804** bytes (Dialog / NavStack / EmojiPicker / Terminal). Initial savings are only **32,480 bytes**; loading all children costs **8,027 bytes more** than the equally optimized baseline. Chromium 159 passed lazy navigation through all four, retained input via clipboard, shared memory identity, Motion before Dialog, retry, and back/forward during a held download. This is local browser evidence, not a production or physical-device gate pass.

## Category ownership and site integration

The category candidate now compares linker-retained function names from all twelve category builds with the full program and a builder-free baseline. `gallerybuild -analysis` preserves names and uses the same optional imports in every build; partial category sets are rejected. Baseline functions, shared functions, initialization closures and unassigned functions remain in the primary. Only functions present in exactly one category move to its child. The analysis records the primary SHA-256, and splitting rejects a different primary. This is conservative linker evidence, not a complete dynamic-call proof.

```sh
go run ./internal/gallerybuild -analysis -out /tmp/keel-ownership
python3 scripts/analyze-gallery-wasm.py --input /tmp/keel-ownership --binaryen-dir /path/to/binaryen/bin --out /tmp/keel-ownership/ownership.json
python3 scripts/probe-shared-wasm.py --ownership-build /tmp/keel-ownership --binaryen-dir /path/to/binaryen/bin --out /tmp/keel-shared-site-input --optimize
# Supply font.otf here as in the normal site workflow.
go run ./internal/site -demo /tmp/keel-shared-site-input -out /tmp/keel-shared-site
python3 -m http.server 8774 --bind 127.0.0.1 --directory /tmp/keel-shared-site
# In another terminal, with Playwright available:
node scripts/verify-shared-categories.cjs http://127.0.0.1:8774/demo/ work/gallery-shared-categories
node scripts/verify-shared-gallery.cjs http://127.0.0.1:8774/demo/ work/gallery-shared-retry
```

The current 118-section build moves 1,469 functions into twelve children. Equally optimized gzip-9 sizes are **7,812,742 bytes** unsplit versus **6,972,526 bytes** primary: an **840,216-byte (10.75%)** initial reduction before loading a category. Children range from **15,508 to 196,317 bytes**; loading all twelve costs **44,645 bytes more** in total than the unsplit baseline. These are compressed artifact sizes, not observed network timings or deployed content encoding. The earlier four-child measurements remain historical results.

The site publisher accepts `shared-build.json`, publishes only its paired runtime/children/scripts plus the font, and skips analysis artifacts. Its progress loader fetches the primary once and loads the shared host before Go startup. `?bundle=all` selects the separately paired full program without installing lazy hooks. Missing children or mixed independent/shared manifests fail the build. Content-addressing is applied to every referenced asset; old-asset retention remains a deployment responsibility.

Local Chromium 159 passed all 118 section first frames, twelve cold category runs (each requesting only the primary and its own child), unchanged memory identity, copied input retained across categories, and the full-program fallback with actual input. These runs do not exhaust each component's interactive paths. CI remains on the existing build pending startup/frame/physical-memory and real-device gates; site integration is available explicitly, not deployed by this work.

Browser verifiers default to the locally installed Google Chrome channel; set `KEEL_BROWSER_CHANNEL=chromium` to explicitly use Playwright Chromium. Embedded-category testing exposed a Go 1.27.4 JS/WASM trap when an unfocused input queue looked up a nil key in a large handler map. A standalone standard-library reproduction and the actual input regression test fail without the guard; the test passes after `InputHint` returns early when focus is nil. This was the underlying failure behind the secondary `call to released function` console message and Chromium 159 headless-shell crash. See the [Gio patch log](../third_party/README.md). A trial category layout change did not fix the cause and was reverted.

With the input fix, a fresh Chrome 157 run passes all 118 section first frames, twelve cold categories, single-component/category embeds, full-program fallback, exact input retention and zero page/console errors. A separate run passes failed-child retry and back/forward during download. Chromium 159 headless shell also starts the three previously compared embed cases without errors or crashes. These are browser regressions, not physical-phone or performance acceptance. The rebuilt gzip-9 baseline is 7,812,756 bytes, primary 6,972,851; first-primary savings are 839,905 bytes and loading all children adds 44,916 bytes versus the unsplit baseline.
