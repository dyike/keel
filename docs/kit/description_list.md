# DescriptionList

`kit.DescriptionList().Item("订单号", "SO-1001").ItemElement("状态", element).LabelWidth(96)` 按行显示标签和值，默认标签列 96dp，值列占剩余宽度并换行，顶部对齐。窄容器允许标签列收缩。

纯文本条目向 Agent 暴露一个 text，名字为“标签：值”。ItemElement 保留传入元素的语义和交互，标签单独可读。SetItems 替换全部文本条目；构造函数的 Description 参数是兼容入口。

组件本身无键盘操作。验证：`go run ./examples/components -section description-list -theme dark`，省略 theme 查看浅色。
