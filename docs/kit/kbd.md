# Kbd

Kbd 展示快捷键键帽，不注册快捷键，也不参与 Tab 导航。字号继承父元素，文字颜色使用 theme.Muted，边框使用 theme.Border，运行时切换主题立即生效。

```go
el.Div().Row().Items(el.Center).TextSize(16).Child(
    el.Text("命令面板"),
    kit.Kbd("mod+shift+p").Render(cx),
    kit.Kbd("enter").Plain().Render(cx),
)
```

`Kbd(shortcut string)` 返回 `*KbdView`。唯一配置项 `Plain()` 隐藏边框并保留间距；不提供 Size，调用方通过父元素 TextSize 设置字号。默认边框 1dp、圆角 4dp、水平内边距 6dp、垂直内边距 3dp。窄容器中保持单行并截断。

格式化由 `core.ShortcutLabel(s, goos string) string` 提供，kit 和窗口快捷键共用。语法同 core.ParseShortcut；mod 在 macOS 显示为 Command，在 Windows / Linux 显示为 Ctrl。无法解析的文案原样显示，便于使用自定义键名。该函数只格式化，不查询动作绑定。

## 显示动作绑定的键

快捷键可以绑定到命名动作上，见 [元素与视图 · 动作与键位表](../el.md#动作与键位表)。`kit.KbdFor("editor.save")` 显示这个动作当前绑定的第一个按键；用户改键（`core.Bind`、`core.LoadKeymap`）后下一帧自动更新，没有绑定时什么都不显示。

```go
kit.KbdFor("editor.save").Render(cx)                   // 键帽
kit.Menu().ActionItem("保存", "editor.save", save)    // 菜单右侧显示同一个键
```

Agent 角色为 text，name 为传入的原始 shortcut；屏幕上显示平台对应的符号，不额外生成重复快照节点。

验证入口：`go run ./examples/components -section kbd`，加 `-theme dark` 检查深色。示例覆盖 12 / 16 / 24sp、Plain、平台符号、中英文数字和窄容器。
