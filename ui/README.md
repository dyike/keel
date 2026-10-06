# ui

English | [简体中文](README.zh-CN.md)

Collection of interface modules. Like `native/`, each subdirectory is a module with a single responsibility.

| Module | What it does | Dependencies |
| --- | --- | --- |
| [core](core/README.md) | Foundation: `Widget` interface, callback, thread rules (`Update`) | None |
| [theme](theme/README.md) | Color, size, font | None |
| [locale](locale/README.md) | Frame text, switch between Chinese and English | None |
| [base](base/README.md) | Component behavior, excluding appearance: keyboard navigation, initial jump, multi-select, open state | None |
| [el](el/README.md) | GPUI style: view + chain style element + flexbox | core, theme, locale |
| [kit](kit/README.md) | Components: buttons, forms, tables, overlays, application shells, charts | base, el, core, theme, locale |
| [window](window/README.md) | Window: `Open`, `Main`, shortcut keys, screenshots | core, theme |
| [markdown](markdown/README.md) | Markdown rendering, optimized for AI streaming output | el, core, theme, locale |

See [Architecture](../docs/architecture.md#modules) for module boundaries and dependency graphs. For the minimal form of application, see [Quick Start](../docs/getting-started.md#write-your-first-window).

`internal/` saves frame lock, editing status, text drawing, image loading and interfaceless testing tools, and cannot be directly referenced by external applications.
