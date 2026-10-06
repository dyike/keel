# internal/site

English | [简体中文](README.zh-CN.md)

Builds Keel’s documentation site. Guides under `docs/` and module/example READMEs become static HTML. Component pages embed the WebAssembly build of `examples/components`. `.github/workflows/site.yml` publishes the site to GitHub Pages on pushes to main and release tags.

Local preview:

```sh
go run gioui.org/cmd/gogio@v0.10.0 -target js -tags osusergo -o /tmp/demo ./examples/components
python3 internal/site/subset_font.py . NotoSansSC-Regular.otf /tmp/demo/font.otf
go run ./internal/site -out _site -demo /tmp/demo
python3 -m http.server --directory _site
```

Without `-demo`, the builder generates documentation but the live examples cannot load. `-version v0.1.1` displays a release link in the header. CI uses the latest `v*` tag.

Progress reports under `docs/reports/` stay in the repository. Links to them are removed from published tables and lists because they are working records.

The Markdown indexes own the navigation: H2 sections in `docs/README.md` define guide groups and reading order; H2 sections in `docs/kit.md` define component categories. Module and example READMEs appear under Packages. Missing index entries or duplicate component categories fail the build. The sidebar opens Getting started and the current page’s category.

Getting started covers scaffold creation, development, running, and packaging. The former `docs/cli.html` redirects there while preserving query parameters and section anchors. Component pages display collapsible, copyable source below the live example. The builder reads the file that registers the gallery section, so the displayed source stays current.

Previous/Next links follow the sidebar order across guides, components, and source READMEs. The first document has only Next; the last has only Previous. The home page is excluded. Narrow screens stack the links vertically.

## Languages

English is the default. Canonical `.md` files contain English; `.zh-CN.md` counterparts contain reviewed Simplified Chinese. Every published page requires both files. Keep heading levels and order aligned so the language switch can preserve the current section. Translate prose and sample labels as appropriate; retain API names and executable commands.

The English site uses `/`; Chinese uses `/zh-CN/`. Each language has its own navigation, reading sequence, search index, UI labels, and source links. The header switches to the same document and matching section. Existing Chinese heading anchors remain valid on default English URLs. Both languages share `/demo/`, with `lang=en` or `lang=zh-CN` selecting framework text; application-owned example data retains its original language.

Repository language links connect the paired Markdown files and are omitted from rendered content in favor of the header switch. Chinese repository links point to Chinese files; English links use English heading anchors.

| File | Responsibility |
| --- | --- |
| `main.go` | Collects pages, builds both languages, writes search indexes and assets |
| `navigation.go` | Builds groups and reading order from indexes; checks category coverage |
| `i18n.go` | Resolves language counterparts, localized UI, and section mappings |
| `render.go` | Renders Markdown, rewrites links, and highlights code with Chroma |
| `demo.go` | Publishes the gallery with localized download progress |
| `subset_font.py` | Subsets Chinese fonts using repository characters plus GB2312 |
| `assets/` | Templates, styles, search, copy, and language-switch scripts |

`go test ./internal/site` builds both languages and checks local links, anchors, translation coverage, category expansion, reading order, search isolation, and gallery source integrity.
