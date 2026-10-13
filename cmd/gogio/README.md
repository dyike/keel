# gogio command

English | [简体中文](README.zh-CN.md)

Packages Keel's Gio applications without a `keel.json`, including the site's
component gallery:

```sh
go run ./cmd/gogio -target js -tags osusergo -o /tmp/demo ./examples/components
```

This command passes arguments to [`internal/gogio`](../../internal/gogio/),
which is also used in-process by `cmd/keel`. It imports no GUI packages.
Use `go run ./cmd/gogio -h` for packaging flags and supported targets.
