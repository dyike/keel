# GroupBox

`kit.GroupBox("通知设置").Description("选择接收方式").Child(elements...)` 在内容框外显示标题和说明，内容区带边框、圆角与 Surface 背景。

Child 接受 el.Element，保留子元素交互；兼容构造函数中的 el.View 和 SetChildren。SetTitle 修改标题。Agent 角色 group，标题为名字，说明与内容单独列出。组件本身无键盘操作。

验证：`go run ./examples/components -section group-box -theme dark`，省略 theme 查看浅色。
