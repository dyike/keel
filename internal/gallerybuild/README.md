# internal/gallerybuild

English | [简体中文](README.zh-CN.md)

Builds the Components Gallery with **one Go runtime and fine-grained lazy leaf and shared-dependency modules by default**. Run from the repository root with Python 3 and Binaryen **version_133** installed (build-time tools; application builds set `CGO_ENABLED=0`):

```sh
BINARYEN_DIR=/path/to/binaryen/bin go run ./internal/gallerybuild -out /tmp/demo-new
# Equivalent: -binaryen-dir /path/to/binaryen/bin
# Go-only fallback: full application and independent category applications.
go run ./internal/gallerybuild -mode independent -out /tmp/demo-independent
# Development subset (implicitly independent).
go run ./internal/gallerybuild -categories general,inputs -out /tmp/demo-subset
```

Shared output must be empty or absent. Builds run in temporary directories, analyze linker ownership, split, and optimize both the full fallback and all shared modules with the same `-Oz --strip-debug` options. Only content-addressed runtime assets, `sizes.json` and the final `shared-build.json` are copied out; analysis artifacts and independent manifests are excluded. Missing/wrong Binaryen versions and failed stages fail the command; there is no silent change of mode. Supply `font.otf` afterwards, then publish with `go run ./internal/site -demo /tmp/demo-new -out /tmp/site-new`. The site exposes `?bundle=all` for the equally optimized full fallback.

The AST scanner finds literal `registerSection(name, category, builder)` calls. Temporary `-overlay` sources retain selected builders and replace other registrations with metadata-only entries. Constant-false branches preserve import validity without retaining linker roots. Source files and module manifests are never rewritten. `-analysis` preserves Go symbols, uses identical optional imports and emits a builder-free baseline; the default pipeline consumes this internally. Partial analysis is rejected.

The independent mode uses one runtime per category. Its internal `keel_gallery_small` allowlist omits syntax highlighting from categories without Markdown/code demos; update that list if those demos change. Ordinary desktop builds are unaffected.

CI checks ownership/loader tests, all section first frames, retained input, fallback, retry/history, and controlled Chrome loading performance before Pages publication. See [shared WASM](../../docs/shared-wasm.md) for thresholds, reproduction and deployment limits. Compressed file sizes alone do not establish actual startup or memory improvements.
