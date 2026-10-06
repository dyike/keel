# Kbd

[English](kbd.md) | 简体中文

Kbd 展示快捷键键帽，不注册快捷键，也不参与 Tab 导航。字号继承父元素，文字颜色使用 theme.Muted，边框使用 theme.Border，运行时切换主题立即生效。

```go
el.Div().Row().Items(el.Center).TextSize(16).Child(
    el.Text("命令面板"),
    kit.Kbd("mod+shift+p").Render(cx),
    kit.Kbd("enter").Plain().Render(cx),
)
```

`Kbd(shortcut string)` 返回 `*KbdView`。`Plain()` 隐藏边框并保留间距。`Size(sp)` 设置字号并按比例调整内边距；0 恢复字号继承和默认内边距，负值、NaN、无穷和大于 128 的值忽略。`Style(func(*el.TextEl))` 在每帧默认样式之后调整背景、文字/边框颜色、圆角、间距等；nil 恢复默认。回调不要保留元素引用，Agent 名称始终使用原 shortcut。默认边框 1dp、圆角 4dp、水平内边距 6dp、垂直内边距 3dp。窄容器中保持单行并截断。

格式化由 `core.ShortcutLabel(s, goos string) string` 提供，kit 和窗口快捷键共用。语法同 core.ParseShortcut；mod 在 macOS 显示为 Command，在 Windows / Linux 显示为 Ctrl。无法解析的文案原样显示，便于使用自定义键名。该函数只格式化，不查询动作绑定。

## 显示动作绑定的键

快捷键可以绑定到命名动作上，见 [元素与视图 · 动作与键位表](../el.zh-CN.md#动作与键位表)。`kit.KbdFor("editor.save")` 显示这个动作当前绑定的第一个按键；用户改键（`core.Bind`、`core.LoadKeymap`）后下一帧自动更新，没有绑定时什么都不显示。

```go
kit.KbdFor("editor.save").Render(cx)                   // 键帽
kit.Menu().ActionItem("保存", "editor.save", save)    // 菜单右侧显示同一个键
```

Agent 角色为 text，name 为传入的原始 shortcut；屏幕上显示平台对应的符号，不额外生成重复快照节点。

验证入口：`go run ./examples/components -section kbd`，加 `-theme dark` 检查深色。示例覆盖 12 / 16 / 24sp、Plain、平台符号、中英文数字和窄容器。

### 按元素所在的上下文显示

`kit.KbdFor("format").At("code")` 按 ID 为 `code` 的元素所在的 `KeyContext` 路径解析绑定，和它自己处理按键时用的规则一样（包括 `core.BindIn` 的条件表达式）。比如工具栏按钮上的提示要显示编辑器里的实际快捷键。元素不存在、隐藏或禁用时什么都不显示。
