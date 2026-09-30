## 代码块横向滚动

代码块左侧显示语言，右侧是换行和复制按钮。将鼠标放到图标上查看提示；复制后图标会显示勾选反馈。

```text
已排队 2 条 · 本轮结束后依次发送 · Backspace 取回
  › 顺便把文档也改了
  › 改完跑一下 go test ./pkg/example/

running · Enter 排队 · Esc 中断
────────────────────────────────────────
›
```

缩窄窗口后，用触控板左右滑动，或拖动底部滚动条，把下面代码块滚到最右侧。普通鼠标可以把指针放在底部滚动条上使用滚轮，也可以点击轨道跳转。第一行应保持一行，并能看到末尾的 END_OF_LONG_LINE；第二行含空格，用来检查是否被自动换行。

```go
const unbroken = "ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz_0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz_0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz_0123456789_END_OF_LONG_LINE"
fmt.Printf("this line has spaces and a 中文注释; it should stay on one visual line inside the code block, even when the window is narrow; the final marker is END_OF_SPACED_LINE")
short := "短行对照"
```

第二个代码块应有独立的横向滚动位置。代码块的语言标签和复制按钮应保持可见。

```text
path=/workspace/projects/markdown/examples/chat/fixtures/deeply/nested/directories/with/a/very/long/file/name/that/must/remain/on/one/line/independent_horizontal_scroll_position_END_OF_SECOND_BLOCK
```

点击代码复制按钮，粘贴后确认长行没有被插入换行。普通正文仍应随窗口宽度换行，消息区的纵向滚动仍应正常。

点击“自动换行”后，长行应在代码块内折行，不再需要横向滚动；再次点击“取消自动换行”恢复原始行。每个代码块的设置应独立，流式追加时已有设置和滚动位置应保留。
