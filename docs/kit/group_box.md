# GroupBox

`kit.GroupBox("通知设置").Description("选择接收方式").Child(views...)` 在内容框外显示标题和说明，内容区带边框、圆角与 Surface 背景。

构造函数只接受标题。Child 接受 el.View 并追加到内容区，SetChildren 批量替换内容。每帧调用子视图的 Render，保留子视图交互。SetTitle 修改标题。Agent 角色 group，标题为名字，说明与内容单独列出。组件本身无键盘操作。

验证：`go run ./examples/components -section group_box -theme dark`，省略 theme 查看浅色。
