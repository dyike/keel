# third_party

English | [简体中文](README.zh-CN.md)

Keel carries its own copies of the libraries it draws with, so it can fix
and speed them up without waiting for upstream releases. They keep their
module paths and are wired in with `replace` in Keel's `go.mod`:

```
replace (
	gioui.org => ./third_party/gio
	github.com/go-text/typesetting => ./third_party/typesetting
)
```

| Directory | Upstream | Base | License |
| --- | --- | --- | --- |
| `gio/` | [Gio](https://gioui.org) `gioui.org` | v0.10.3 | Unlicense OR MIT (`gio/LICENSE`) |
| `typesetting/` | [go-text](https://github.com/go-text/typesetting) `github.com/go-text/typesetting` | v0.3.5 | Unlicense OR BSD-3-Clause (`typesetting/LICENSE`) |

`gioui.org/shader` and the packager `gioui.org/cmd/gogio` stay upstream.

A `replace` only applies to the module that writes it, and module zips
leave nested modules out. So:

- Keel's own builds, tests and examples use the copies.
- Projects made with `keel new -replace <Keel checkout>` get the same
  `replace` lines. Other apps add them, pointing at a Keel checkout, to use
  the copies; without them they build against upstream Gio and go-text.
- Keel's code outside `third_party` must therefore build against upstream
  too: a patch adds no API Keel calls, or only one upstream accepts
  unchanged (as the `Bytes` method below). Check with the `replace` lines
  removed: `go mod edit -dropreplace gioui.org -dropreplace
  github.com/go-text/typesetting && go build ./...`, then restore `go.mod`.

## Rules

- Keep upstream style in these directories; change what Keel needs, nothing
  cosmetic. Each change is listed below with the reason, so a rebase onto a
  newer upstream can redo it.
- Upstream tests and test data are not copied. A patch comes with its own
  test next to it. Each copy is its own module: test it from its directory
  (`cd third_party/typesetting && go test ./...`).
- Updating to a newer upstream: copy the release without `*_test.go` and
  `testdata`, keep `go.mod` (gio's also replaces go-text with `../typesetting`),
  then reapply the patches below.

## Patches

- `go vet` fixes for unkeyed struct literals in `gio/internal/f32` and
  `gio/app/internal/ibus`.
- `typesetting/font/opentype`: resources that expose their bytes (`Shared`,
  or any with a `Bytes() []byte` method, so callers build against upstream
  too) hand out table slices instead of copies and never write into a
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
- `gio/app` (macOS): the window's `CAMetalLayer` keeps at most two
  drawables. The renderer waits for the previous frame before the next, so
  a third only held a window-sized surface (5 MB at 640x512 pt, 33 MB on
  4K); hello's idle footprint varied with it by 5 MB between runs.
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
