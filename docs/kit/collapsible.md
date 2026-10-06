# Collapsible

English | [简体中文](collapsible.zh-CN.md)

The expanded/collapsed state of a single section of content can be used as a whole, or the trigger and content can be placed separately in the layout.

```go
section := kit.Collapsible("Advanced filters", form)
section.SetValue(true)
// Default bordered panel: section.Render(cx)
row.Child(section.Trigger().Render(cx))
body.Child(section.Content().Render(cx))
```

`Value` / `SetValue` reads and sets the status, and the program settings do not call back; `OnChange` is called when the user switches. `SetDisabled` disables both independent triggers and content. The Trigger/Content of each instance is rendered once. Do not call the overall Render at the same time.

`Heading(view)` Customize the display content of the title. The label during construction is still used as the accessibility name. Place text, icons and other display content inside the title to avoid nesting another interactive control.

The trigger supports Tab, Enter / Space, the role is disclosure, and the value is expanded / collapsed. When the content being edited is closed, the focus returns to the trigger and the input state is retained. The content being folded cannot be operated.

Expand and collapse use 180ms height animation; repeated switching reverses from the current height, and the text maintains a natural layout without squeezing the text. Switch immediately when reduced motion, and the initial expanded state is also displayed directly.

Verify: `go run ./examples/components -section collapsible`, add `-theme dark` to check the dark theme.
