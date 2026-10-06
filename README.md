<p><img src="docs/images/keel.svg" alt="Keel" width="80" height="80"></p>

# Keel

English | [简体中文](README.zh-CN.md)

Build desktop UIs in pure Go, without HTML, CSS, JavaScript, or a WebView. [Gio](https://gioui.org) draws the interface; `native/` provides permissions, screenshots, synthetic input, global hotkeys, system notifications, and a rich clipboard. Keel runs on macOS, Windows, and Linux, compiles to WebAssembly for the browser, and offers [experimental iOS support](docs/ios.md).

Create an application:

```sh
go install github.com/dyike/keel/cmd/keel@latest
keel new my-app && cd my-app
keel run            # Run
keel build          # Package for the current platform (.app / .exe / Linux package with icons)
```

Start developing in the generated `app.go`. To explore the repository examples and component gallery:

```sh
go run ./examples/hello
go run ./examples/components   # Component gallery: sidebar navigation and search
go run ./examples/chat -sample=all
go test -race ./...
```

Requires Go 1.26+. On macOS, install Xcode Command Line Tools. On Linux, install the Wayland/X11 development packages; `keel doctor` checks the environment. Browse the [documentation and live gallery](https://keel.dyike.com).

## Documentation

- [Overview](docs/README.md): find guides for your development task.
- [Getting started](docs/getting-started.md): create, develop, run, and package an application with the scaffold.
- [iOS (experimental)](docs/ios.md): scaffold build/run and current verification scope.
- [Components](docs/kit.md): browse components by purpose, with APIs and interactive examples.
- [Automation](docs/automation.md): let an agent click, type, scroll, and take screenshots.
- [Architecture](docs/architecture.md): modules, dependency boundaries, and threading rules.

## Licensing

Keel is dual licensed under [AGPL-3.0](LICENSE) and a commercial license. Open-source projects and personal use are free under the AGPL. Using Keel in closed-source software requires a commercial license. See [LICENSING.md](LICENSING.md) and the [contribution guide](CONTRIBUTING.md).
