# 开关

`Size(Small/Medium/Large)` 调整轨道尺寸；默认 Medium。`SetDisabled` 禁用，`SetLoading` / `Loading` 控制加载指示，加载时阻止鼠标和键盘改变状态。应用完成操作后调用 `SetLoading(false)`；组件不发起网络请求。`SetValue` 保持静默。

验证入口：`go run ./examples/components -section switch`。
