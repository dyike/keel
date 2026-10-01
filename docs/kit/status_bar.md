# StatusBar

`kit.StatusBar().Left(el.Text("就绪")).Right(el.Text("第 12 行"))` 创建固定 24dp 状态栏，顶部 1dp 边框，默认 12sp Muted 文本。左右内容分组，中间空间由 Grow 撑开，长文本限制单行。父视图决定状态栏位置。

Left/Right 替换各自的元素组。Agent 角色 status，子元素单独列出；组件本身无键盘操作，传入控件保留交互。溢出菜单留到 M5。

验证：`go run ./examples/components -section status-bar -theme dark`，省略 theme 查看浅色。
