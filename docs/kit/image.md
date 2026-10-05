# Image

显示已解码图片，或通过 Source 异步加载图片，按宽度缩放并保持宽高比。

```go
logo := kit.Image(img, "公司标志").Width(160).Rounded(8)
photo := kit.Image(nil, "头像") // 先显示占位；加载完后在 core.Update 中调用 photo.SetImage(img)
```

- 默认撑满父容器的宽度，但不超过图片自身的像素宽度；`Width(dp)` 设置最大宽度。
- 没有图片时显示带替代文字的占位块。可用 Source 加载，也可在别处解码后调用 SetImage。
- `OnClick` 让图片可以点击和聚焦，`SetDisabled` 禁用点击。

Agent：角色 `image`，名字是替代文字，`value` 为 `loaded` 或 `loading`。

验证：`go run ./examples/components -section image`，加 `-theme dark` 检查深色。

`Size(width, height)` 指定视框，`Fit(ImageContain)` 完整显示并居中，`ImageCover` 居中裁剪，`ImageFill` 拉伸填满。宽度仍受父容器约束；高度为 0 时随实际宽度保持比例。尺寸与圆角忽略负数及 NaN/Inf，宽度 0 恢复自动宽度。圆角裁剪同时作用于像素和命中区域。

`Preview()` 让已加载图片通过点击或键盘打开模态预览；Esc、关闭按钮和背景可关闭，焦点返回原图。移除图片或禁用所属区域会关闭预览。`OnClick` 可与预览共用。

`SetError(reason)` 清除旧图并显示错误；`OnRetry(fn)` 添加重试按钮，点击先清除错误并回到加载状态，再调用回调。加载成功调用 `SetImage`。手动后台加载使用 `core.Update` 提交结果；也可使用下方内置 Source/Cache。每次 `SetImage` 替换绘制缓存，nil/空图和错误状态立即释放旧缓存引用。传入后不要修改图片像素；需要变化时重新调用 `SetImage`。Agent 的图片状态增加 `error`，重试按钮名称包含替代文字。


## 加载、缓存和状态内容

```go
cache := kit.NewImageCache(32 << 20)
photo := kit.Image(nil, "商品照片").Size(320, 180).
    LoadingContent(kit.Spinner().Label("正在加载图片")).
    Fallback(kit.Label("图片暂时不可用")).
    Cache(cache).Source("https://example.com/photo.webp")
```

Source 支持本地路径/file URL、data URL，以及 HTTP(S) 地址——网络地址要在应用里引入 `_ "github.com/dyike/keel/ui/netimage"`（约 4 MB，不引入时返回 `core.ErrNoImageFetcher`）；格式为 PNG、JPEG、WebP、GIF（含动图）和 SVG，大小限制与 core.DecodeImage 相同。请求在后台执行，默认 15 秒超时；结果通过 UI 队列提交。重复同一地址不重载；Source("") 取消并清空。SetImage/SetError 也会取消并移除当前 Source，旧结果不能覆盖它们。组件卸载不会自动取消，应用可在不再使用时调用 Source("")。

Loading 返回请求状态，ImageError 返回加载错误。LoadingContent/Fallback 接受自定义 View，nil 恢复默认替代文字/错误；固定 Size 可预留加载区域，自定义内容应适配该区域。自定义失败内容后仍保留内置重试按钮。Source 模式下 Retry 清除当前源缓存并重新加载；没有 Source 时使用原 OnRetry 回调。自身或父级禁用阻止按钮，程序 Source/SetImage 仍可更新。

默认共享 64MiB 估算容量的 ImageCache；Cache(nil) 禁用缓存，自定义缓存可限定作用域。缓存按源字符串区分、LRU 淘汰，源字符串计入预算，每像素按 8 字节保守计费（动图按帧数累计）；超过预算的图仍可显示但不保留。并发同源请求合并，取消一个等待者不影响其他人；所有等待者取消后中断请求。失败不缓存。Delete(source)/Clear 清除结果并阻止旧请求重新填入，现有等待者仍收到自己的结果。

源地址内容变化时调用 Retry/Delete。缓存图片按只读共享，不应修改像素。网络请求受浏览器 CORS、系统网络和文件权限约束。

## SVG 与 GIF 动图

- **SVG**：按布局尺寸实时绘制矢量，任何缩放都清晰；自然尺寸取 viewBox，三种 Fit 都适用。识别依据是内容（`<svg` / `<?xml`），与扩展名无关。不支持脚本、外部引用和 CSS 动画。
- **GIF 动图**：按每帧延迟循环播放，按处置方式合成帧，效果与浏览器一致；延迟为 0 或 1 的帧按 100ms 播放。开启减少动态效果（`theme.ReducedMotion`）或禁用时停在第一帧。所有帧解码后超过 96MB 时只显示第一帧。

## 磁盘缓存

```go
dir, _ := os.UserCacheDir()
cache := kit.NewImageCache(32 << 20).Disk(filepath.Join(dir, "myapp", "images"), 256<<20, 24*time.Hour)
```

`Disk(目录, 上限字节, 有效期)`（同样需要 `ui/netimage`）把 HTTP(S) 图片的原始字节存到磁盘，重启后不必重新下载。有效期内直接使用本地副本；过期后带 `If-None-Match` / `If-Modified-Since` 重新验证，服务器返回 304 就继续用本地副本并刷新有效期。网络失败或服务器 5xx 时使用过期副本。总大小超过上限时按最近使用时间淘汰。`Cache-Control: no-store` 的响应不落盘；本地文件和 data URL 不复制。本地副本解码失败会被删除。目录为空或上限 ≤ 0 关闭磁盘缓存。内存层仍按原规则工作，磁盘层只在内存未命中时读取。
