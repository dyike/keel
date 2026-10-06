# Settings

English | [简体中文](settings.zh-CN.md)

The settings page provides page navigation, multiple groups of settings within the page, search, and full page reset. When the window is narrower than 600dp, navigation is moved to the top and controls are arranged below the description.

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

`Section(title, icon, items...)` still works for pages with only one set, and can be mixed with `Page`. The page title must be unique; `Value` / `SetValue` reads and switches pages, and the page added for the first time is opened by default. `TitleSuffix` The view after receiving the page title.

`SettingGroup.Footer` is outside the group's surface and scrolls and filters with the group; it is not search content and does not become a navigation item. `GroupVariant` sets a unified appearance. The `Variant` pointer of the group can overwrite it. It supports Surface, Normal, Fill and Outline of GroupBox. Groups that do not hit any rows are not displayed.

The search is case-insensitive and matches `Label`, `Description`, and `Keywords` alone. Keywords will not be displayed. If there is still a match on the current page, keep it selected, otherwise it will switch to the first matching page; the navigation only displays the matching page. Zero results retain the selection but display an empty state, and a clear search retains the current selection. `Query` / `SetQuery` can be searched from the app actions.

`DescriptionContent` can be connected to the existing `markdown.New(...)` or any `el.View`, replacing the display of plain text description; the search content is still provided by `Description`. Rich text views should be created when building settings to preserve parsing and interaction state. `Content` replaces the entire line body, and fully customized lines can also be searched by keywords. `Vertical` forces top-bottom arrangement, `Disabled` disables interaction within the entire row, including custom content. Standard control column width is 240dp; row labels are supplied for controls without names. `RowSpacing` adjusts the upper and lower margins of the row, 0 restores the theme default; the control size is configured by the incoming control.

Display the localization reset button after the page declares `Resettable: true`. Clicking will call `Reset` for all non-disabled rows on the page, including searching for hidden rows; rows without a callback are skipped. `ResetPage(title)` provides program entry. Pages that are not reset will not be executed. Default values, application state and persistence are maintained by callbacks and controls are not modified through reflection; for example, when restoring a theme, switches and `theme.Apply` need to be updated at the same time.

Groups, rows, keywords, and Variant values are copied when adding pages; views and callbacks are still shared by the app. The identity of each row does not change with the position of the search results, and the control state is retained after page cutting and searching.

Usually the window is filled directly with `el.Root(settings)`. Run `go run ./examples/components -section settings` to see multi-group, keyword, Markdown, and reset examples; add `-theme dark` to check dark colors. The automatic test covers the original 1×/2× narrow layout, keyboard and status retention, as well as group filtering, page footer, data copy, full page reset, custom rows and disabling; this batch does not undergo real device visual acceptance.

## Size and group navigation

`Size(SettingsSizeXSmall/Small/Medium/Large)` uniformly adjusts row label/caption font size, padding, horizontal spacing, and control column width. The default Medium maintains the original layout. A non-zero value for `RowSpacing` overrides the size's default inline margins; custom Content and the control's own size are still configured by the application.

`GroupNavigation(true)` Add the group with title and hit search to the navigation; the narrow window displays the group button of the current page. `ShowGroup(page, groupTitle)` switches pages and locates groups, and can be used independently of the navigation switch. If the group is missing, has the same name, or is searched and filtered, false will be returned and the query will not be cleared. The target position is calculated by the layout, and off-screen grouping is supported. `Value()` always returns the page title, `SetValue` cancels pending group targeting.

Automatic testing covers double-ratio wide and narrow windows, off-screen positioning, group clicks, search filtering, duplicate name rejection and size switching to retain control values; the component library displays Small and group navigation by default. Native vision is not accepted.
