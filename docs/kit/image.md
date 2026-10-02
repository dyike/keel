# Image

显示已解码的图片，按宽度缩放并保持宽高比。

```go
logo := kit.Image(img, "公司标志").Width(160).Rounded(8)
photo := kit.Image(nil, "头像") // 先显示占位；加载完后在 core.Update 中调用 photo.SetImage(img)
```

- 默认撑满父容器的宽度，但不超过图片自身的像素宽度；`Width(dp)` 设置最大宽度。
- 没有图片时显示带替代文字的占位块。kit 不负责加载图片：在别处解码好，再通过 `SetImage` 交给它。
- `OnClick` 让图片可以点击和聚焦，`SetDisabled` 禁用点击。

Agent：角色 `image`，名字是替代文字，`value` 为 `loaded` 或 `loading`。

验证：`go run ./examples/components -section image`，加 `-theme dark` 检查深色。

`Size(width, height)` 指定视框，`Fit(ImageContain)` 完整显示并居中，`ImageCover` 居中裁剪，`ImageFill` 拉伸填满。宽度仍受父容器约束；高度为 0 时随实际宽度保持比例。尺寸与圆角忽略负数及 NaN/Inf，宽度 0 恢复自动宽度。圆角裁剪同时作用于像素和命中区域。

`Preview()` 让已加载图片通过点击或键盘打开模态预览；Esc、关闭按钮和背景可关闭，焦点返回原图。移除图片或禁用所属区域会关闭预览。`OnClick` 可与预览共用。

`SetError(reason)` 清除旧图并显示错误；`OnRetry(fn)` 添加重试按钮，点击先清除错误并回到加载状态，再调用回调。加载成功调用 `SetImage`。后台加载和 URL 缓存由应用管理，使用 `core.Update` 提交结果；组件不跨实例缓存图片。每次 `SetImage` 替换绘制缓存，nil/空图和错误状态立即释放旧缓存引用。传入后不要修改图片像素；需要变化时重新调用 `SetImage`。Agent 的图片状态增加 `error`，重试按钮名称包含替代文字。
