# native

系统能力模块集合。每个子目录是一个独立模块，可以只用其中一个，也可以完全不用 `ui` 界面。

| 模块 | 能力 | 依赖 |
| --- | --- | --- |
| [permission](permission/) | 检查、申请系统权限 | `native`、`native/internal/sys` |
| [screen](screen/) | 显示器列表、截图 | 同上 |
| [input](input/) | 合成鼠标、键盘事件 | 同上 |
| [hotkey](hotkey/) | 全局快捷键 | 同上 |

模块之间互不引用，也不引用 `ui` 和 Gio，这一点由 `internal/deps` 里的测试保证。

两个共享部分不是独立模块：

- `native`（本目录的 `native.go`）：所有模块共用的错误值 `native.Err*`。
- `native/internal/sys`：cgo 绑定，所有 Objective-C 代码只在这里。外部不能引用。

只实现了 macOS 14+，其他平台返回 `native.ErrUnsupported`。

文档：[原生能力](../docs/native.md) · [新增原生能力](../docs/extending.md#新增原生能力)
