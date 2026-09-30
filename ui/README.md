# Go UI

`keel/ui` 用 Go 组件树描述界面，直接生成 Go-Gui 控件。按钮、输入框和状态都在 Go 中，无网页资源、JS 桥接或 npm 构建。

```go
name := ui.Input(ui.InputOptions{Label: "名字", Placeholder: "请输入"})
result := ui.Text("等待操作")
page := ui.NewPage("Hello", ui.Card(
    name,
    ui.Button("提交", func() {
        result.SetText("你好，" + name.Value())
    }),
    result,
))
kit := keel.New()
_, err := kit.Window.New(keel.WindowOptions{UI: page})
if err != nil { log.Fatal(err) }
if err := kit.Run(); err != nil { log.Fatal(err) }
```

完整可运行程序见 [单窗口示例](../examples/ui/main.go)。

## 组件

| API | 用途 |
| --- | --- |
| `Text`、`Heading` | 普通文本、标题，`SetText` 更新 |
| `Button(text, callback)` | 默认主按钮，Go 回调 |
| `NewButton(ButtonOptions)` | 主、次、危险按钮以及禁用状态 |
| `Input(InputOptions)` | 单行输入，`Value` / `SetValue` |
| `TextArea(InputOptions)` | 多行输入 |
| `Checkbox(CheckboxOptions)` | 布尔状态，`Value` / `SetValue` |
| `Column`、`Row`、`Card` | 纵向、横向、卡片布局 |
| `Divider` | 分隔线 |
| `NewPage(title, components...)` | 独立窗口的页面 |

输入框的 `OnChange` 在每次接受文本编辑后调用，按钮回调直接读取 `Value()`。`ReadOnly` 允许选择和复制但禁止编辑；`Disabled` 禁止交互；`MaxLength` 按 Unicode 字符计数。`Required` 会在点击页面按钮前检查，失败显示错误并调用 `page.SetOnError` 注册的处理器。

`PasswordInput` 遮蔽文本；`EmailInput` 和 `NumberInput` 提供软键盘布局提示，不自动校验邮件和数字。`SearchInput` 目前采用普通单行输入外观，没有额外清空按钮。需要业务校验时，在 Go 回调中处理。

## 组合成自己的组件

复用的是构建函数，每次调用创建新的控件对象：

```go
func NameForm(onSubmit func(string)) ui.Component {
    input := ui.Input(ui.InputOptions{Label: "名字"})
    return ui.Card(input, ui.Button("提交", func() {
        onSubmit(input.Value())
    }))
}
```

为每个窗口创建新的 Page 和组件。一个 Page 只能绑定一个窗口；复用组件对象会共享其 Go 状态。组件树创建后固定，文本、值和禁用状态可以更新。动态添加或删除组件、路由和自定义样式尚未封装，复杂界面可以通过 `WindowOptions.View` 使用 Go-Gui 原始组件。

## 后台任务

事件回调在 UI 线程运行，更新后自动刷新。耗时工作放到后台，完成后显式 `Refresh`：

```go
status := ui.Text("等待")
page := ui.NewPage("任务", status)
// 创建窗口之后：
go func() {
    result := doWork()
    status.SetText(result)
    page.Refresh()
}()
```

组件状态的读取和设置有锁；Go-Gui 的窗口和渲染操作应在 UI 线程执行。裸 `*gui.Window` 的后台操作通过 `QueueCommand` 调度。工作任务可以监听 `win.Host().Ctx().Done()`，在窗口关闭时停止。

## 测试

```sh
go test -race ./ui
go run ./examples/smoke
```

单元测试使用 Go-Gui 软件渲染和 `TestType` / `TestClick`，验证实际输入与按钮分发、禁用和只读、必填和长度限制、回调 panic 报告，以及页面状态隔离。原生 smoke 再验证真正的多窗口生命周期。

设置页快捷键示例见 [多窗口程序](../examples/multiwindow/main.go)。使用 `win.RegisterShortcut("super+comma", openSettings)` 绑定 `⌘ + ,`，回调在 UI 线程运行，无需手动调度。
