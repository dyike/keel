# gogio

[English](README.md) | 简体中文

Keel 维护的 Gio v0.10.0 打包器通过 `Run(args, Options)` 在 CLI 进程内运行。
独立入口 `cmd/` 也用于打包站点示例。上游许可证保留在 `LICENSE`，补丁记录在
[third_party/README.zh-CN.md](../../../../third_party/README.zh-CN.md)。

`GioRoot` 指定应用 ID、启动参数、Android 权限和移动端资源使用的包前缀。
Options 提供工作目录、环境变量和输出流。上游实现保留包级状态，因此调用串行执行。
打包器不导入 Gio 或其他 Keel 包。Windows 资源和 Keel 的 iOS 模拟器适配仍由 CLI 处理。

在仓库根目录执行 `go test ./cmd/keel/internal/gogio/...`。
