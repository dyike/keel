# ui/kit

English | [简体中文](README.zh-CN.md)

Ready-made component layer, based on el to organize appearance and interaction, base provides keyboard navigation, first letter jump and multi-selection behavior.

- Usage and classification index: [Components](../../docs/kit.md).
- Implementation requirements: [Component development specification](../../docs/component-development.md).
- Supporting example: [Component library](../../examples/components/README.md).

Directly dependent on `core`, `theme`, `locale`, `el`, `base`, not `window`. SVG parsing and rasterization use `oksvg`, `rasterx`. For the support range, see [Icon](../../docs/kit/icon.md) and [Image](../../docs/kit/image.md).

One component corresponds to one `<name>.go`, the shared field frame is in `field.go`, and the shared surface and overlay styles are provided by `surface()` and `floating()`. `conventions_test.go` Check out supporting documentation, examples, Agent tests, and public API conventions.

Atomic input references access `ui/internal/inputcontent` via `el.InputDocument`; the kit does not directly depend on this internal package.
