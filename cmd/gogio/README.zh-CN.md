# gogio 命令

[English](README.md) | 简体中文

无需 `keel.json` 即可打包使用 Keel 内置 Gio 的应用，例如站点组件库：

```sh
go run ./cmd/gogio -target js -tags osusergo -o /tmp/demo ./examples/components
```

命令将参数交给 [`internal/gogio`](../../internal/gogio/)，`cmd/keel` 也在进程内
调用这个包。命令不导入 GUI 包。执行 `go run ./cmd/gogio -h` 查看打包参数和支持的目标。
