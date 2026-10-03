# Avatar

`kit.Avatar("张三").Size(40).Status(kit.AvatarOnline)` 显示圆形头像，Size 接受 float32 dp，推荐小、中、大尺寸为 24、40、56。Image 接受已解码的 image.Image，nil 恢复姓名回退。后台加载完成后用 core.Update 更新。

中文取第一个字，英文取前两个单词首字母，空姓名显示问号。背景根据姓名哈希从当前主题选取。Status 支持 AvatarOnline、AvatarBusy、AvatarOffline，空值隐藏；状态点位于头像内部，不改变布局。

Agent 角色 avatar，名字为人名，value 为状态，无状态为空。纯展示，不处理键盘。验证：`go run ./examples/components -section avatar -theme dark`，省略 theme 查看浅色。

`Source(url)` 在后台加载头像，支持 HTTP(S)、data URL、本地路径和 file URL；PNG、JPEG、GIF 首帧和 WebP。加载期间或失败时显示姓名回退，直径不变。默认请求超时 15 秒，编码数据上限 16 MiB，解码图片上限 3200 万像素。浏览器环境中的 HTTP 请求仍受 CORS 限制。

```go
avatar := kit.Avatar("Ada Lovelace").Source("https://example.com/ada.png")
```

同一实例重复设置相同来源不会重新请求。`Loading()` 查询加载状态，`ImageError()` 返回最近的加载错误，`Retry()` 重新请求当前来源。切换来源、调用 `Image` 或 `Source("")` 会取消旧请求，迟到结果不会覆盖新状态。`Source("")` 恢复姓名回退；加载完成通过 `core.Update` 在下一帧应用。头像卸载时可调用 `Source("")` 取消尚未结束的请求。

不做跨实例图片缓存；需要认证请求或统一缓存时，由应用通过 `Image` 提供像素。直接调用 `core.DecodeImage(ctx, source)` 可复用同一解码器，应在后台线程使用并自行提供超时。

头像组见 [AvatarGroup](avatar_group.md)。
