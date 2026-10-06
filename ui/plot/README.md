# ui/plot

English | [简体中文](README.zh-CN.md)

Exposed custom chart basics: continuous/categorical scales, positive and negative stacking, pie chart layout, and columns, lines, areas, arcs, points, guides, and axes on the clipped canvas.

Directly depends on `ui/core` (drawing context), `ui/theme` (axis font), and indirectly depends on `ui/internal/loop`. Does not reference kit, el, window, or native; currently used by the app and component library examples, existing kit diagrams are not migrated to this package.

The numerical layout has no mutable global state, and the current frame Canvas is used for drawing. The application is responsible for colors, size conversions, interactions, prompts, and accessible data descriptions. See [Public Drawing Basics](../../docs/kit/plot_primitives.md) for complete instructions.
