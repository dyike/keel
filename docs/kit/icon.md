# Icon

`kit.Icon(kit.IconInfo).Size(24).Color(theme.Info)` 创建矢量图标。省略 Color 时每次 Render 读取 theme.Text。支持 Check、Close、Plus、Search、Copy、ChevronDown、ChevronRight、Info、Warning、Error、User、Inbox。

图标遵守父级尺寸约束，本身不提供语义和键盘操作；由包含它的控件提供名称。运行 `go run ./examples/components -section icon -theme dark` 检查深色，省略 theme 检查浅色。
