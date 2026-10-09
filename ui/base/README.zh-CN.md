# ui/base

[English](README.md) | 简体中文

组件的行为，不含外观：列表的键盘导航、按首字母跳转（typeahead）、带 Shift 范围的多选、打开/关闭状态。不绘制任何东西，只是普通的 Go 状态和函数。

- **依赖**：只依赖 Gio 的按键名（`third_party/gio/io/key`），不依赖任何 Keel 模块。
- **被谁依赖**：`kit` 的 List、Tree、Menu、Select、Command、Sidebar、Table 用它处理键盘和选择；应用也可以拿它配合 `el` 写外观完全自定的组件。

| 文件 | 内容 |
| --- | --- |
| `list.go` | `List`：↑ ↓ Home End PageUp PageDown，跳过禁用项，可循环 |
| `typeahead.go` | `Typeahead`：连续输入字母跳到以它开头的项，同一字母重复按下时循环 |
| `selection.go` | `Selection[K]`：按 key 保存的多选，单击、Cmd/Ctrl 切换、Shift 范围 |
| `disclosure.go` | `Disclosure`：可禁用的打开/关闭状态，区分用户操作和程序设置 |

文档：[无样式基础层](../../docs/base.zh-CN.md)
