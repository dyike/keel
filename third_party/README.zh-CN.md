# third_party

[English](README.md) | 简体中文

Keel 自带它绘制界面所用的库的副本，这样可以直接修复和优化，不必等上游发版；应用只要依赖 Keel 就能拿到这些修改（Keel 的 `go.mod` 里写 `replace` 传不到应用）。

| 目录 | 上游 | 基线 | 许可 |
| --- | --- | --- | --- |
| `gio/` | [Gio](https://gioui.org) `gioui.org` | v0.10.3 | Unlicense 或 MIT（`gio/LICENSE`） |
| `gio/cmd/gogio/` | Gio 的打包器 `gioui.org/cmd/gogio` | v0.10.0 | Unlicense 或 MIT |
| `typesetting/` | [go-text](https://github.com/go-text/typesetting) | v0.3.5 | Unlicense 或 BSD-3-Clause（`typesetting/LICENSE`） |

`gioui.org/shader` 仍是普通依赖：只有预编译的着色器，没有出现在任何 API 里的类型。

导入路径是 `github.com/dyike/keel/third_party/gio/...` 和 `github.com/dyike/keel/third_party/typesetting/...`。之前直接导入上游包的应用运行一次 `keel migrate`。

## 规则

- 这两个目录保持上游风格；只改 Keel 需要的，不做纯格式修改。每处修改都记在下面并写明原因，换到更新的上游时照着重做。
- 不复制上游的测试和测试数据。补丁自带测试，放在补丁旁边。
- 升级上游：复制发布版本，去掉 `*_test.go` 和 `testdata`，改写导入路径（`gioui.org/`，`gioui.org/shader` 除外；`github.com/go-text/typesetting/`），再重做下面的补丁。

## 补丁

- 导入路径改为本模块；修复 `gio/internal/f32` 和 `gio/app/internal/ibus` 里 `go vet` 报的无字段名结构体字面量。
- `gio/cmd/gogio`：查找 `github.com/dyike/keel/third_party/gio/app` 而不是 `gioui.org/app`。`keel build` 从应用所依赖的 Keel 模块编译它。
