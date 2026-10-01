# Tag

`kit.Tag("紧急").Color(kit.TagDanger)` 显示染色胶囊标签。Color 支持 TagDefault、TagPrimary、TagSuccess、TagWarning、TagDanger，背景由当前主题 Surface 与等级色混合。Tone 是旧版兼容入口，SetText 更新文案。

Agent 角色 tag，名字是文字，value 为颜色名。纯展示版本不响应键盘。多个标签的自动换行依赖后续 wrap，M1 示例只展示单行组合。验证：`go run ./examples/components -section tag -theme dark`，省略 theme 查看浅色。
