# cmd/keel

English | [简体中文](README.zh-CN.md)

Keel's scaffolding: `keel new` creates a new project, `keel run` runs it, `keel build` packages it by platform (with icon, name, version), and `keel doctor` checks the environment. For usage, see [Quick Start](../../docs/getting-started.md).

```sh
go install github.com/dyike/keel/cmd/keel@latest
```

- **Dependencies**: `internal/svgicon` (draw the SVG of the original image into PNG), `internal/appicon` (the icon shape of each platform, `ui/window` also uses it when running), `golang.org/x/image` (scaling, vector rasterization), `github.com/tc-hib/winres` (Windows resources: icons, manifests, version information). Keel's interface package is not referenced; macOS and browser packaging call Gio's gogio (fixed in v0.10.0), and macOS signature calls the system's `codesign`.
- **Test**: `go test ./cmd/keel` generates the project, checks the packaging commands for each platform (`-n`), and compiles the generated project with Keel in this repository (`-short` skips).

| File | Responsibility |
| --- | --- |
| `main.go` | Subcommand distribution, running external commands |
| `config.go` | `keel.json` Reading, writing and verification |
| `new.go` | New project, template in `template/` |
| `run.go` | `keel run` |
| `build.go` | Packaging for each platform, Windows resources, Info.plist, Linux desktop files |
| `icons.go` | Generate icons, `keel icon`, .ico by platform and size |
| `doctor.go` | Environmental Check |
