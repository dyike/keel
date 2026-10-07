# native/process

English | [简体中文](README.zh-CN.md)

`ForegroundPID(*os.File)` returns a terminal's foreground process group ID; `Directory(pid)` reads a process's working directory. Both return errors, leave file ownership with the caller, and do not invoke a shell. macOS uses libproc and tcgetpgrp; Linux uses /proc and terminal ioctl. Other platforms return `native.ErrUnsupported`. Invalid arguments return `native.ErrInvalidArgument`; disappearing processes and non-terminal descriptors return `native.ErrFailed`.

Dependencies: `native`, `native/internal/sys`. No UI or Gio dependencies. Platform bindings remain inside `native/internal/sys`.
