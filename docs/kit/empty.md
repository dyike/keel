# Empty

`kit.Empty("暂无订单").Description("新建订单开始").Icon(kit.IconInbox).Action(action)` 垂直居中显示图标、标题、说明与操作区。Action 接受 el.Element，暂可用 el.Div().OnClick 组合按钮。

SetTitle、SetDescription 更新文案。Empty 不新增语义角色，标题和说明分别作为 text，操作元素保留自身语义与键盘行为。窄容器文字换行。验证：`go run ./examples/components -section empty`，加 `-theme dark` 检查深色。
