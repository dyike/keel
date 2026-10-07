# native/process

[English](README.md) | 简体中文

`ForegroundPID(*os.File)` 查询终端的前台进程组 ID，Unix 上通常是组长 PID；`Directory(pid)` 查询进程的当前工作目录。接口返回错误，不执行 shell，也不接管文件所有权。macOS 使用 libproc 和 tcgetpgrp，Linux 使用 /proc 和终端 ioctl。其他平台返回 `native.ErrUnsupported`，非法参数返回 `native.ErrInvalidArgument`，进程已退出或文件不是终端时返回 `native.ErrFailed`。

依赖 `native`、`native/internal/sys`，不依赖 UI 或 Gio。平台绑定留在 `native/internal/sys`，业务项目只使用 Go API。
