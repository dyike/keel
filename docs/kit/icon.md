# Icon

`kit.Icon(kit.IconInfo).Size(24).Color(theme.Info)` 创建矢量图标。省略 Color 时每次 Render 读取 theme.Text。Size 接受 float32 dp。零值 `IconNone` 不画任何东西、不占空间，可选图标的字段不设就是没有图标。内置图标来自 Material Design：Check、Done、Close、Plus、Minus、Search、Copy、ChevronLeft、ChevronRight、ChevronDown、Info、Warning、Error、User、Inbox、Star、StarOutline、Calendar、Clock、Settings、Bell、Lock、Folder、File、Archive、Receipt、Home、Trash、Edit。按含义选图标，不要借用形状相近的图标（例如用 Info 表示设置）。解析结果会缓存，每帧调用 `Icon` 不会重复解码。

图标遵守父级尺寸约束，本身不提供语义和键盘操作；由包含它的控件提供名称。运行 `go run ./examples/components -section icon -theme dark` 检查深色，省略 theme 检查浅色。
