# ui/locale

English | [简体中文](README.zh-CN.md)

Text displayed by the Keel framework itself or reported to the Agent: OK, Cancel, Copy, Close, Please Select, "Line 36", etc. The usage is the same as `theme`. After switching, all windows are redrawn and `cx.Cache` automatically becomes invalid.

```go
locale.Apply(locale.English())                      // Before opening the window or during callback
core.Update(func() { locale.Apply(locale.English()) }) // background goroutine

s := locale.Chinese() // Just change a few things: copy from the default and change it
s.OK = "好的"
locale.Apply(s)
```

- Default: `Chinese()` (default), `English()`. `Apply` replaces all text, so you need to start from the default when customizing.
- Counting function: `Rows(n)`, English will distinguish between 1 row and 2 rows.
- `Name(action, target)` Splice accessible names, such as "Close Saved Successfully", "Close Saved".
- **Just framework text.** The application is responsible for its own copy: you can press `Current().Lang` to select your own copy table, read it in Render, and it will be automatically redrawn after switching languages; the key of the self-built cache must bring `Revision()`.
- The component reads `locale.Current()` during Render or Layout and cannot save text during construction.

- Dependencies: `ui/internal/loop` (only used to notify redraws).
- Used by: `el` (cache invalidation), `kit`, `markdown`.
