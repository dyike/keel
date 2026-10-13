# gogio

English | [简体中文](README.zh-CN.md)

Keel's copy of Gio's v0.10.0 packager runs through `Run(args, Options)` inside
the `keel` CLI. The standalone [`cmd/gogio`](../../cmd/gogio/) entrypoint also
packages examples for the site.
The upstream license is retained in `LICENSE`; patches are recorded in
[third_party/README.md](../../third_party/README.md).

`GioRoot` selects the import prefix used for application IDs, launch arguments,
Android permissions and mobile resources. Options supply the working directory,
environment and output writers. Calls are serialized because the upstream
implementation keeps package-level state. The packager does not import Gio or
other Keel packages. Windows resources and Keel's iOS simulator adapter stay in
the CLI.

Run `go test ./internal/gogio/...` from the repository root.
