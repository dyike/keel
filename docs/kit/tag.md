# Tag

`kit.Tag("紧急").Color(kit.TagDanger)` 显示染色胶囊标签。Color 支持 TagDefault、TagPrimary、TagSuccess、TagWarning、TagDanger，背景由当前主题 Surface 与等级色混合。Tone 是旧版兼容入口，SetText 更新文案。

Agent 角色 tag，名字是文字，value 为颜色名。纯展示版本不响应键盘。Selectable().OnChange(fn) 启用选择，Value/SetValue 查询和程序赋值，程序赋值不触发回调。OnRemove(fn) 添加移除按钮，Tab 聚焦后 Space/Enter 或 Backspace/Delete 触发；调用方负责从数据中删除，组件不会自行隐藏。SetDisabled 禁止交互，Agent 报告 disabled 与 selected。多个标签的自动换行依赖后续 wrap，M1 示例只展示单行组合。验证：`go run ./examples/components -section tag -theme dark`，省略 theme 查看浅色。
