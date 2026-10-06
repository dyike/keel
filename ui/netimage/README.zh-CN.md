# ui/netimage

[English](README.md) | 简体中文

网络图片，可选模块。引入后 Image、Avatar、Attachment 的 `Source`、Markdown 图片和图片磁盘缓存都能加载 `http://`、`https://` 地址。

```go
import _ "github.com/dyike/keel/ui/netimage"
```

- **为什么单独成包**：它用 `net/http`，连带 TLS 等加密库约 4 MB，只显示本地图片的应用不必带上。不引入时本地路径、`file://` 和 data URL 照常工作，网络地址返回 `core.ErrNoImageFetcher`，错误信息里写明要引入的包。
- **依赖**：`ui/core`（实现 `core.ImageFetcher` 并在 `init` 里 `core.SetImageFetcher`）和标准库 `net/http`。
- `netimage.Client` 是使用的 HTTP 客户端，默认 `http.DefaultClient`，可以在第一次请求前换成带代理、超时或鉴权的客户端。

要用别的网络栈，实现 `core.ImageFetcher` 并调用 `core.SetImageFetcher`。
