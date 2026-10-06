# Avatar

English | [简体中文](avatar.zh-CN.md)

`kit.Avatar("Alex Chen").Size(40).Status(kit.AvatarOnline)` displays a circular avatar. Size accepts float32 dp. The recommended small, medium and large sizes are 24, 40 and 56. Image accepts a decoded image.Image, nil restore name fallback. Use core.Update to update after background loading is completed.

In Chinese, the first character is taken, in English, the first letters of the first two words are taken, and a question mark is displayed for an empty name. The background is picked from the current theme based on the name hash. Status supports AvatarOnline, AvatarBusy, and AvatarOffline, and empty values are hidden; the status point is located inside the avatar and does not change the layout.

Agent role avatar, the name is the person's name, the value is the state, and it is empty if there is no state. Pure display, no keyboard handling. Verification: `go run ./examples/components -section avatar -theme dark`, omit theme to see the light theme.

`Source(url)` loads avatars in the background, supporting data URL, local path, file URL, and HTTP(S) after introducing `ui/netimage`; PNG, JPEG, GIF first frame, WebP and SVG. The name display falls back during loading or on failure, and the diameter remains unchanged. The default request timeout is 15 seconds, the encoded data limit is 16 MiB, and the decoded image limit is 32 million pixels. HTTP requests in a browser environment are still subject to CORS restrictions.

```go
avatar := kit.Avatar("Ada Lovelace").Source("https://example.com/ada.png")
```

If the same source is set repeatedly for the same instance, the request will not be re-requested. `Loading()` queries the loading state, `ImageError()` returns the latest loading error, and `Retry()` re-requests the current source. Switching sources, calling `Image` or `Source("")` will cancel the old request, and late results will not overwrite the new state. `Source("")` restores name fallback; loading completed is applied in the next frame via `core.Update`. When uninstalling the avatar, you can call `Source("")` to cancel the unfinished request.

No cross-instance image caching is performed; when authentication requests or unified caching are required, the application provides pixels through `Image`. Calling `core.DecodeImage(ctx, source)` directly reuses the same decoder, which should be used in a background thread and provide its own timeout.

Avatar group see [AvatarGroup](avatar_group.md).

## Appearance

| Method | Function |
| --- | --- |
| `Rounded(dp)` | Rounded corners, the default is circle; rounded square corners are commonly used for team and application avatars, such as `Rounded(theme.RadiusLg)` |
| `Colors(bg, fg)` | The background color and text color of the initial letter and placeholder icon; the zero value remains the default (select the background color by name, and use `theme.Text` for text) |
| `Border(dp, c)` | Outer ring, such as stacked avatars separated by `theme.Surface` colored rings; 0 removed |
| `Placeholder(icon)` | The icon displayed when there is no picture or name, the default is `IconUser`; `IconNone` displays nothing |
| `Style(fn)` | Adjust the avatar frame after the default style in each frame to handle situations not covered by the above methods |

The placeholder icon is displayed when there is no name, and the question mark is no longer displayed.
