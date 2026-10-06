# cmd/keel

[English](README.md) | 简体中文

Keel 的脚手架：`keel new` 新建项目，`keel run` 运行，`keel build` 按平台打包（带图标、名称、版本），`keel doctor` 检查环境。用法见 [快速开始](../../docs/getting-started.zh-CN.md)。

```sh
go install github.com/dyike/keel/cmd/keel@latest
```

- **依赖**：`internal/svgicon`（把占位原图的 SVG 画成 PNG）、`internal/appicon`（各平台图标形状，`ui/window` 运行时也用它）、`golang.org/x/image`（缩放、矢量光栅化）、`github.com/tc-hib/winres`（Windows 资源：图标、清单、版本信息）。不引用 Keel 的界面包；macOS 和浏览器打包调用 Gio 的 gogio（固定在 v0.10.0），macOS 签名调用系统的 `codesign`。
- **测试**：`go test ./cmd/keel` 生成项目、检查各平台的打包命令（`-n`），并用本仓库的 Keel 编译生成的项目（`-short` 跳过）。

| 文件 | 内容 |
| --- | --- |
| `main.go` | 子命令分发，运行外部命令 |
| `config.go` | `keel.json` 读写与校验 |
| `new.go` | 新建项目，模板在 `template/` |
| `run.go` | `keel run` |
| `build.go` | 各平台打包、Windows 资源、Info.plist、Linux 桌面文件 |
| `icons.go` | 按平台和尺寸生成图标、`keel icon`、.ico |
| `doctor.go` | 环境检查 |
