# native/clipboard

English | [简体中文](README.zh-CN.md)

Reads a snapshot of the clipboard independently, without referencing the UI or Gio. Implemented for macOS 14+/cgo, Windows, Linux X11, and (when the app provides the connection) Linux Wayland; other platforms and iOS return `native.ErrUnsupported`.

```go
clipboard.Read(func(data clipboard.Data, err error) {
    // Background goroutine; modifying the interface requires core.Update.
    // data.Text is text, Images is encoded pictures, and Files is file path.
})
```

- `Read` returns immediately and completes the callback by calling the background goroutine after reading from the main queue; nil callback does not read.
- Pictures are preferred PNG, otherwise TIFF, keep the encoded data, do not decode or convert pixels. The file URL is converted into a path; the image preview that comes with the file is not repeatedly counted as an image.
- Maximum of 128 pasteboard items, the total text/path UTF-8 bytes and image encoded data does not exceed 16MiB. `native.ErrFailed` is returned if the limit is exceeded, serialization fails, or the clipboard version changes during reading, and partial results are not returned.
- Does not write to the clipboard, does not open the referenced file, and does not read the file content; an empty clipboard returns empty Data successfully.
- Linux Wayland: Only the client with the keyboard focus can see the clipboard, so the connection of the focused window: `clipboard.UseWaylandDisplay(w.WaylandDisplay())` is passed from the application and is re-passed when the focus changes to the window. Fallback to X11/XWayland when no connection is available or not supported by the synthesizer. This path loads libwayland-client at run time (no cgo); `-tags nowayland` leaves it out.
- **Dependencies**: native, native/internal/sys, native/internal/wlclip. The application can adapt the result to core.ClipboardData and connect it to Input/TextArea.PasteReader.

macOS has actually read the picture snapshot; the reading of Windows and Linux has passed the format and protocol test, but has not yet been accepted on the real device.
