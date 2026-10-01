# Settings

设置页：左侧是分区列表，右侧是当前分区的设置项，顶部的搜索框可以跨分区查找。

```go
s := kit.Settings().
    Section("通用", kit.IconUser,
        kit.SettingItem{Label: "语言", Description: "界面显示的语言", Control: lang},
        kit.SettingItem{Label: "深色模式", Control: kit.Switch("", false)}).
    Section("通知", kit.IconInbox, kit.SettingItem{Label: "邮件提醒", Control: mail})
```

- 每个设置项：左边是标签和说明，右边是控件。控件区宽 240dp，输入框、下拉框会撑满这个宽度，开关、复选框靠右对齐。
- 控件不需要再传标签，设置项的标签会作为它的无障碍名称，和 Form 一样。
- 搜索按标签和说明匹配，不区分大小写，结果按分区分组显示。
- `Value()` / `SetValue(title)` 读取或切换当前显示的分区。
- 会撑满父容器给的空间，通常直接作为窗口内容：`el.Root(settings)`。

Agent：左侧是 `navigation`，分区是 `link`；每个分区和每个设置项都是 `group`，名字分别是分区标题和设置项标签。

验证：`go run ./examples/components -section settings`，加 `-theme dark` 检查深色。
