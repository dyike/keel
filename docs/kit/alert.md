# Alert

English | [简体中文](alert.zh-CN.md)

`kit.Alert("保存失败").Tone(kit.ToneDanger).Description("网络不可用").OnClose(fn)` Displays inline hints. The default is ToneInfo; Tone supports ToneNeutral, ToneInfo, ToneSuccess, ToneWarning, and ToneDanger. SetTone, SetTitle, SetDescription can be updated programmatically.

The left level bar and icons are colored according to the theme, the title is bold, and the default description is 13sp Muted.

- `Size(AlertSizeXSmall/AlertSizeSmall/AlertSizeMedium/AlertSizeLarge)` sets four levels of font size, icon and padding. The default is Medium to retain the original layout; illegal values are ignored.
- `Icon(IconName)` replaces the level icon, `IconNone` hides the icon and its spacing.
- `Content(el.View)` replaces the description, and can pass Markdown document, rich text or action button; pass nil to restore Description. The text control retains its own keyboard and click behavior, and inherits Alert's disabled state.
- `Banner(true)` uses a full-width, right-angled, borderless dyed banner, without displaying a separate title line; displays Content or Description, and uses the title as the message when there is no body text. The Close button and Agent name still use the title. Pass false to restore inline hints.

```go
notice := kit.Alert("维护通知").Banner(true).
    Description("今晚进行例行维护").Icon(kit.IconCalendar).
    Size(kit.AlertSizeSmall)
// Rich text is composed by the application: notice.Content(markdown.New("**NOTE**: Please save your work first"))
```

The close button is displayed only when OnClose is set; Space/Enter is closed after click or Tab focus, and is hidden first and then called back. Visible queries the status, SetVisible does not call the callback when restoring or hiding.

Agent container role alert, the name is the title, and the value is the level. The description and close button are individually readable. Narrow container text wraps. Run `go run ./examples/components -section alert -theme dark`; omit theme to see the light themes.

SetDisabled(true) disables closing and passes disabled semantics to the button; it can be refocused and closed after being restored, and the visible and hidden state will not be lost due to measurement or theme switching.
