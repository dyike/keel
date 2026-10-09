# third_party

English | [简体中文](README.zh-CN.md)

Keel carries its own copies of the libraries it draws with, so it can fix
and speed them up without waiting for upstream releases, and so every app
gets the fixes through its Keel requirement alone (a `replace` in Keel's
`go.mod` would not reach apps).

| Directory | Upstream | Base | License |
| --- | --- | --- | --- |
| `gio/` | [Gio](https://gioui.org) `gioui.org` | v0.10.3 | Unlicense OR MIT (`gio/LICENSE`) |
| `gio/cmd/gogio/` | Gio's packager, `gioui.org/cmd/gogio` | v0.10.0 | Unlicense OR MIT |
| `typesetting/` | [go-text](https://github.com/go-text/typesetting) | v0.3.5 | Unlicense OR BSD-3-Clause (`typesetting/LICENSE`) |

`gioui.org/shader` stays an ordinary dependency: precompiled shaders, no
types in any API.

Import paths are `github.com/dyike/keel/third_party/gio/...` and
`github.com/dyike/keel/third_party/typesetting/...`. Apps that imported the
upstream packages run `keel migrate` once.

## Rules

- Keep upstream style in these directories; change what Keel needs, nothing
  cosmetic. Each change is listed below with the reason, so a rebase onto a
  newer upstream can redo it.
- Upstream tests and test data are not copied. A patch comes with its own
  test next to it.
- Updating to a newer upstream: copy the release without `*_test.go` and
  `testdata`, rewrite the import paths (`gioui.org/` except
  `gioui.org/shader`, `github.com/go-text/typesetting/`), then reapply the
  patches below.

## Patches

- Import paths rewritten to this module; `go vet` fixes for unkeyed struct
  literals in `gio/internal/f32` and `gio/app/internal/ibus`.
- `gio/cmd/gogio`: looks for `github.com/dyike/keel/third_party/gio/app`
  instead of `gioui.org/app`. `keel build` compiles it from the Keel module
  the app requires.
- `typesetting/font/opentype`: `Shared` resources (`NewShared`) hand out
  table slices instead of copies and never write into a caller's buffer.
  `typesetting/fontscan` maps system font files read-only (`mmap`,
  `MapViewOfFile`) and parses them as `Shared`, so a face's tables are clean
  file-backed pages, not Go heap: hello's idle footprint went from 153 to
  103 MB on macOS. `gio/font/opentype.ParseCollectionShared` does the same
  for bytes that never change; `gio/font/gofont` and Keel's theme use it.
  Tested by parsing, describing, shaping and outlining every system font
  from read-only mappings (`fontscan/openfont_test.go`).
- `typesetting/harfbuzz`: an attachment chain pointing before the buffer
  returns as in HarfBuzz instead of indexing `pos[-1]`, which panicked
  shaping macOS's Farisi.ttf.
