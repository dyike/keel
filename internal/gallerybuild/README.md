# internal/gallerybuild

English | [简体中文](README.zh-CN.md)

Builds the complete Components WebAssembly application and independently linked category applications. Uses only the Go standard library and invokes Keel's gogio CLI. Run from the repository root:

```sh
go run ./internal/gallerybuild -out demo
# Faster development build: complete application and these categories only.
go run ./internal/gallerybuild -out /tmp/demo -categories general,inputs
```

The AST scanner finds literal `registerSection(name, category, builder)` calls. Temporary `-overlay` sources retain selected builders and replace other registrations with metadata-only entries. The original calls remain in constant-false branches to preserve import validity without retaining linker roots. No checkout files or module manifests are rewritten. Failed builds stop before publishing a new `groups.json` manifest.

Each category has its own runtime. This reduces first-category download size, but duplicates common code across categories and resets application state when changing programs. This is not shared-runtime dynamic linking. The site publisher consumes `groups.json`, hashes asset names, shares the font URL and exposes `?bundle=all` for the complete program. CI builds all categories; the subset flag is for development.

Browser categories without Markdown/code demos use the internal `keel_gallery_small` tag to omit syntax highlighting. Desktop builds keep their existing provider. Update that category allowlist if a demo begins using highlighting.

Tests cover registration parsing and pruning. Build all categories to verify actual compiler/linker compatibility. Measure raw/compressed sizes and real cold/warm startup separately; code splitting alone does not establish lower total downloads or higher frame rates.
