# Image

English | [简体中文](image.zh-CN.md)

Display a decoded image, or load an image asynchronously from a Source, scaling by width and maintaining aspect ratio.

```go
logo := kit.Image(img, "Company logo").Width(160).Rounded(8)
photo := kit.Image(nil, "Avatar") // Display the placeholder first; after loading, call photo.SetImage(img) in core.Update
```

- By default, it fills the width of the parent container, but does not exceed the pixel width of the image itself; `Width(dp)` sets the maximum width.
- Displays a placeholder block with alt text when there is no image. It can be loaded with Source, or SetImage can be called after decoding elsewhere.
- `OnClick` makes the image clickable and focusable, `SetDisabled` disables clicks.

Agent: role `image`, name is alt text, `value` is `loaded` or `loading`.

Verify: `go run ./examples/components -section image`, add `-theme dark` to check the dark theme.

`Size(width, height)` specifies the view frame, `Fit(ImageContain)` displays it completely and is centered, `ImageCover` is centered and cropped, and `ImageFill` is stretched and filled. The width is still constrained by the parent container; a height of 0 remains proportional to the actual width. Sizes and rounded corners ignore negative numbers and NaN/Inf, and a width of 0 returns to automatic width. Corner clipping works on both pixels and hit areas.

`Preview()` allows the loaded image to open a modal preview by clicking or keyboard; Esc, close button and background can be closed, and the focus returns to the original image. Removing an image or disabling the associated area will turn off the preview. `OnClick` can be shared with previews.

`SetError(reason)` clears the old image and displays the error; `OnRetry(fn)` adds a retry button. Click to clear the error and return to the loading state before calling the callback. Call `SetImage` on successful loading. Manual background loading uses `core.Update` to submit results; you can also use the built-in Source/Cache below. Every time `SetImage` replaces the draw cache, the old cache reference is released immediately with nil/empty drawing and error status. Do not modify the image pixels after passing them in; call `SetImage` again when changes are needed. The Agent's image status is increased by `error` and the retry button name contains alt text.


## Loading, caching and state content

```go
cache := kit.NewImageCache(32 << 20)
photo := kit.Image(nil, "Product photo").Size(320, 180).
    LoadingContent(kit.Spinner().Label("Loading image")).
    Fallback(kit.Label("Image unavailable")).
    Cache(cache).Source("https://example.com/photo.webp")
```

Source supports local path/file URL, data URL, and HTTP(S) address - the network address must be introduced in the application `_ "github.com/dyike/keel/ui/netimage"` (about 4 MB, if not introduced, `core.ErrNoImageFetcher` will be returned); the formats are PNG, JPEG, WebP, GIF (including animations) and SVG, and the size limit is the same as core.DecodeImage. Requests are executed in the background with a default timeout of 15 seconds; results are submitted through the UI queue. Repeating the same address does not overload; Source("") cancels and clears. SetImage/SetError also cancels and removes the current Source, old results cannot overwrite them. Component uninstallation is not automatically canceled and the app can call Source("") when it is no longer in use.

Loading returns the request status and ImageError returns the loading error. LoadingContent/Fallback accepts custom View, nil restores the default alternative text/error; fixed Size can reserve the loading area, and the custom content should fit in this area. Retain built-in retry button after customizing failure content. In Source mode, Retry clears the current source cache and reloads it; when there is no Source, the original OnRetry callback is used. Disabling the blocking button on itself or its parent, the program Source/SetImage can still be updated.

The default shared ImageCache of 64MiB estimated capacity; Cache(nil) disables caching, and custom caches can be scoped. The cache is distinguished by source string, LRU is eliminated, the source string is included in the budget, and each pixel is conservatively billed at 8 bytes (animated images are accumulated based on the number of frames); images that exceed the budget can still be displayed but will not be retained. Concurrent requests from the same source are merged, and canceling one waiter does not affect others; the request is interrupted after all waiters are cancelled. Failure is not cached. Delete(source)/Clear clears the results and prevents old requests from being refilled, existing waiters still receive their results.

Retry/Delete is called when the content of the source address changes. Cached images are shared as read-only and pixels should not be modified. Network requests are subject to browser CORS, system network, and file permissions.

## SVG and GIF animations

- **SVG**: Draw vectors in real time according to the layout size, and it will be clear at any zoom; the natural size is the viewBox, and all three types of Fit are applicable. The identification is based on the content (`<svg` / `<?xml`) and has nothing to do with the extension. Scripts, external references, and CSS animations are not supported.
- **GIF animation**: Loop playback is delayed for each frame, and the frames are synthesized according to the processing method. The effect is consistent with the browser; frames with a delay of 0 or 1 are played at 100ms. Turn on reduce motion (`theme.ReducedMotion`) or stop at first frame when disabled. Only the first frame is displayed when all frames exceed 96MB after decoding.

## Disk cache

```go
dir, _ := os.UserCacheDir()
cache := kit.NewImageCache(32 << 20).Disk(filepath.Join(dir, "myapp", "images"), 256<<20, 24*time.Hour)
```

`Disk(directory, maxBytes, lifetime)` (also requires `ui/netimage`) saves the raw bytes of the HTTP(S) image to disk, without having to download it again after restarting. Use the local copy directly during the validity period; after expiration, use `If-None-Match` / `If-Modified-Since` to re-verify. If the server returns 304, continue to use the local copy and refresh the validity period. Use expired copy in case of network failure or server 5xx. When the total size exceeds the upper limit, it will be eliminated based on the latest usage time. The response of `Cache-Control: no-store` is not written to disk; local files and data URLs are not copied. The local copy will be deleted if it fails to decode. Directory is empty or upper limit ≤ 0 Turn off disk caching. The memory layer still works according to the original rules, and the disk layer only reads when the memory misses.
