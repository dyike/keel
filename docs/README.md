# Overview

English | [简体中文](README.zh-CN.md)

Start with Getting started to open your first window. Use Guides to build your application and Components to find UI controls. For framework changes, see Contributing.

## Getting started

- [Getting started](getting-started.md): scaffold a project, edit its UI, run it, and package it.
- [WebAssembly](web.md): build WebAssembly and understand the differences from desktop applications.
- [Migration](migration-kit.md): replace APIs from older versions.

## Mobile

- [Overview](mobile.md): iOS and Android support and system capability limits.
- [iOS](ios.md): builds, simulator runs, scene lifecycle adapter and device signing.
- [Android](android.md): APK packaging, device installation and toolchain checks.

## Guides

- [Elements and views](el.md): layout, state, events, focus, scrolling, and overlays.
- [Windows](app.md): multiple windows, lifecycle, shortcuts, and offscreen screenshots.
- [Theme](theme.md): colors, fonts, dimensions, local themes, and motion.
- [Markdown](markdown.md): streaming, code blocks, caching, and customization.
- [Native APIs](native.md): permissions, screenshots, keyboard and mouse input, global hotkeys, notifications, and the clipboard.
- [Primitives](base.md): reuse keyboard navigation and selection behavior with your own appearance.
- [Accessibility](accessibility.md): what VoiceOver reads and presses in a Keel window.

## Components

- [Overview](kit.md): find components by purpose, with APIs and interactive examples.

## Testing

- [Automation](automation.md): launch applications, operate components, and take screenshots through MCP.
- [Troubleshooting](troubleshooting.md): environment, text layout, redraw, window, and permission issues.
- [Testing](testing.md): interaction regressions, screenshot matrices, and real-window tests.

## Contributing

- [Architecture](architecture.md): module responsibilities, dependency boundaries, and threading rules.
- [Extending Keel](extending.md): add components, native capabilities, and examples.
- [Component guidelines](component-development.md): APIs, state, file organization, and acceptance criteria.
- [Visual guidelines](visual-guidelines.md): typography, dimensions, spacing, colors, and interaction states.
- [Design decisions](decisions.md): framework trade-offs and historical approaches.
- [One module](single-module.md): Gio and go-text as packages of the Keel module, import prefixes, and migration.
- [Changes from upstream](upstream-patches.md): Gio, go-text and packager changes, source locations, regression tests, and compatibility limits.
- [Rendering memory and reuse](rendering-memory.md): renderer allocation changes, native measurements, regression coverage and limitations.
- [Desktop CGO-free delivery design](desktop-cgo-free.md): native binding migration and acceptance gates.
- [Desktop no-cgo implementation evidence](desktop-cgo-free-validation.md): verified build and ABI results, limitations, and pending acceptance.

Module READMEs describe source responsibilities and implementation files. Read the repository’s [contribution guide](../CONTRIBUTING.md) before contributing.
