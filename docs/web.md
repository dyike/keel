# WebAssembly

English | [简体中文](web.zh-CN.md)

`ui/*` compiles to WebAssembly, so the same UI code can run in a browser. `native/*` calls operating-system APIs and is unavailable on the Web.

## Build

In a scaffolded project, run `keel build -target js`. The CLI supplies Go 1.26’s `osusergo` build tag and uses Gio’s packager to generate `index.html`, `wasm.js`, and `main.wasm`. Keep the scaffold’s entry point and `app.go`.

Browsers do not expose system fonts to WebAssembly. To display Chinese text, load a font in your application. Add `font_js.go`, which is compiled only for the js target:

```go
//go:build js

package main

import (
    "log"

    "github.com/dyike/keel/ui/theme"
)

func init() {
    if err := theme.FetchFonts("font.ttf"); err != nil {
        log.Print(err)
    }
}
```

Build, copy the font, and start a static server from your project directory:

```sh
keel build -target js
cp /path/to/NotoSansSC-Regular.ttf dist/web/font.ttf
python3 -m http.server --directory dist/web
```

Prefer a font family listed in `theme.Face`, such as Noto Sans SC. Chinese fonts commonly exceed 10 MB; use a subset where possible. For embedded font files, use `theme.LoadFonts(data)`.

## Differences from desktop applications

- `native/*` is unavailable: no native permissions, screenshots, global hotkeys, synthetic input, system notifications, or rich clipboard.
- There is one browser window; multiple `window.Open` calls draw on the same page.
- Automation mode (`KEEL_AUTOMATION`) and keel-mcp work only on desktop.

## Documentation site

Keel’s site compiles `examples/components` to WebAssembly and embeds it on component pages, as [GPUI Kit](https://gpui-kit.com) does. The gallery is about 43 MB uncompressed and roughly 10 MB over GitHub Pages’ gzip transfer, then cached by the browser.

The browser gallery reads URL parameters: `demo/?section=dock` shows only Dock, and `&theme=dark` selects the dark theme. `&lang=en` or `&lang=zh-CN` selects framework text. The Chinese font is a Noto Sans SC subset of about 1.9 MB.

See [internal/site](../internal/site/README.md) for the generator and local preview. `.github/workflows/site.yml` defines deployment.
