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
