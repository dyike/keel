# ui/netimage

English | [简体中文](README.zh-CN.md)

Network pictures, optional modules. After the introduction, the `Source`, Markdown images and image disk cache of Image, Avatar, Attachment can load the `http://` and `https://` addresses.

```go
import _ "github.com/dyike/keel/ui/netimage"
```

- **Why it is packaged separately**: It uses `net/http`, and together with encryption libraries such as TLS, it is about 4 MB. Applications that only display local images do not need to bring it. When not imported, the local path, `file://` and data URL work as usual, the network address returns `core.ErrNoImageFetcher`, and the error message indicates the package to be imported.
- **Dependencies**: `ui/core` (implements `core.ImageFetcher` and `core.SetImageFetcher` in `init`) and the standard library `net/http`.
- `netimage.Client` is the HTTP client used. The default is `http.DefaultClient`. It can be changed to a client with proxy, timeout or authentication before the first request.

To use another network stack, implement `core.ImageFetcher` and call `core.SetImageFetcher`.
