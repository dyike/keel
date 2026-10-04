# Settings

设置页提供页面导航、页面内多组设置、搜索和整页重置。窗口窄于 600dp 时，导航移到顶部，控件排到说明下方。

```go
name := kit.Input("")
s := kit.Settings().GroupVariant(kit.GroupBoxFill).
    Page(kit.SettingPage{
        Title: "通用", Icon: kit.IconSettings, Resettable: true,
        Groups: []kit.SettingGroup{
            {Title: "账户", Items: []kit.SettingItem{
                {Label: "显示名称", Description: "公开显示的名称",
                 Keywords: []string{"name", "profile"}, Control: name,
                 Reset: func() { name.SetValue("") }},
            }},
            {Title: "外观", Items: []kit.SettingItem{
                {Label: "主题", Description: "立即生效，无需重启",
                 DescriptionContent: markdown.New("**立即生效**，无需重启。"),
                 Control: themeSelect},
            }},
        },
    })
```

`Section(title, icon, items...)` 仍适用于只有一组的页面，可与 `Page` 混用。页面标题须唯一；`Value` / `SetValue` 读取和切换页面，首次添加的页面默认打开。`TitleSuffix` 接收页面标题后的视图。

`SettingGroup.Footer` 位于组的表面之外，跟随组一起滚动、过滤；它不是搜索内容，也不成为导航项。`GroupVariant` 设置统一外观，组的 `Variant` 指针可覆盖它，支持 GroupBox 的 Surface、Normal、Fill、Outline。未命中任何行的组不显示。

搜索不区分大小写，匹配 `Label`、`Description` 和独立的 `Keywords`。关键词不会显示。当前页面仍有匹配时保持选择，否则切到第一个匹配页；导航只显示匹配页。零结果保留选择但显示空状态，清空搜索保留当前选择。`Query` / `SetQuery` 可从应用操作搜索。

`DescriptionContent` 可接现有 `markdown.New(...)` 或任意 `el.View`，替代纯文本说明的显示；搜索内容仍由 `Description` 提供。富文本视图应在构建设置项时创建，以保留解析和交互状态。`Content` 替换整个行体，完全自定义的行也可用关键词搜索。`Vertical` 强制上下排列，`Disabled` 禁用整行，包括自定义内容内的交互。标准控件列宽 240dp；行标签会补给没有名称的控件。`RowSpacing` 调整行的上下留白，0 恢复主题默认；控件尺寸由传入的控件配置。

页面声明 `Resettable: true` 后显示本地化重置按钮。点击会调用该页所有未禁用行的 `Reset`，包括搜索隐藏的行；没有回调的行跳过。`ResetPage(title)` 提供程序入口，未开启重置的页面不执行。默认值、应用状态和持久化由回调维护，不通过反射修改控件；例如还原主题时，需要同时更新开关和 `theme.Apply`。

添加页面时会复制组、行、关键词和 Variant 值；视图和回调仍由应用共享。各行身份不随搜索结果的位置变化，控件状态在切页和搜索后保留。

通常直接用 `el.Root(settings)` 填满窗口。运行 `go run ./examples/components -section settings` 查看多组、关键词、Markdown 和重置示例；加 `-theme dark` 检查深色。自动测试覆盖原有 1×/2× 窄布局、键盘和状态保持，以及分组过滤、页尾、数据副本、整页重置、自定义行与禁用；本批未做真机视觉验收。

## 尺寸与分组导航

`Size(SettingsSizeXSmall/Small/Medium/Large)` 统一调整行标签/说明字号、内边距、横向间距和控件列宽，默认 Medium 保持原布局。`RowSpacing` 的非零值覆盖尺寸预设的行内边距；自定义 Content 和控件自身尺寸仍由应用配置。

`GroupNavigation(true)` 将有标题且命中搜索的分组加入导航；窄窗口显示当前页的分组按钮。`ShowGroup(page, groupTitle)` 切换页面并定位分组，可独立于导航开关使用。缺失、重名或被搜索过滤的组返回 false，不清空查询。目标位置由布局计算，支持离屏分组。`Value()` 始终返回页面标题，`SetValue` 取消待处理的分组定位。

自动测试覆盖双倍率宽窄窗口、离屏定位、分组点击、搜索过滤、重名拒绝与尺寸切换保留控件值；组件库默认展示 Small 和分组导航。原生视觉未验收。
