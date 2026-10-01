# 复选框

`SetDisabled(bool)` 禁用鼠标和键盘操作。`SetIndeterminate(bool)` 设置半选并保留原来的布尔值；`Indeterminate()` 查询半选。用户激活半选项后转为选中。`SetValue` 会清除半选且不触发回调；自动化快照通过 `value: mixed` 表示半选，通过 `disabled` 表示禁用。

验证入口：`go run ./examples/components -section checkbox`。
