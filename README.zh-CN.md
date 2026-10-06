<p><img src="docs/images/keel.svg" alt="Keel" width="80" height="80"></p>

# Keel

[English](README.md) | 简体中文

用纯 Go 写桌面界面，不需要 HTML、CSS、JavaScript 或 WebView。界面由 [Gio](https://gioui.org) 绘制；Gio 没有的原生能力（权限、截图、合成输入、全局快捷键、系统通知、富剪贴板）在 `native/` 下。支持 macOS、Windows、Linux，也能编译成 WebAssembly 在浏览器里运行，并支持 [iOS 和 Android 移动端](docs/mobile.zh-CN.md)。

新建一个应用：

```sh
go install github.com/dyike/keel/cmd/keel@latest
keel new my-app && cd my-app
keel run            # 运行
keel build          # 打包当前平台（带图标的 .app / .exe / Linux 安装包）
```

应用开发从生成的 `app.go` 开始。查看仓库中的示例和组件库：

```sh
go run ./examples/hello
go run ./examples/components   # 组件库：侧栏选择、搜索
go run ./examples/chat -sample=all
go test -race ./...
```

需要 Go 1.26+；macOS 上需要 Xcode Command Line Tools，Linux 上需要 Wayland/X11 的开发包（`keel doctor` 会检查）。在线文档和组件库：https://keel.dyike.com

## 文档

- [文档导览](docs/README.zh-CN.md)：按开发任务选择文档。
- [快速开始](docs/getting-started.zh-CN.md)：用脚手架新建、开发、运行和打包应用。
- [Mobile 支持](docs/mobile.zh-CN.md)：iOS、Android 的构建运行与系统能力边界。
- [组件参考](docs/kit.zh-CN.md)：按用途分类，查 API 和在线交互示例。
- [Agent 端到端测试](docs/automation.zh-CN.md)：让 Agent 点击、输入、滚动和截图。
- [架构](docs/architecture.zh-CN.md)：源码模块、依赖边界与线程规则。

## 授权

Keel 采用 [AGPL-3.0](LICENSE) 与商业授权双授权。开源项目和个人自用按 AGPL 免费使用；在闭源软件里使用 Keel 需要购买商业授权。详见 [LICENSING.md](LICENSING.zh-CN.md)，参与贡献见 [CONTRIBUTING.md](CONTRIBUTING.zh-CN.md)。
