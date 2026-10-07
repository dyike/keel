# 参与贡献

[English](CONTRIBUTING.md) | 简体中文

欢迎提交 Issue 和 Pull Request。

## 授权

Keel 采用 MIT 协议（见 [授权说明](LICENSING.zh-CN.md)），贡献也以 MIT 协议发布。**提交第一个 Pull Request 前需要同意 [贡献者授权协议（CLA）](CLA.zh-CN.md)**：在 PR 描述里写上

> I have read and agree to the Keel Contributor License Agreement (CLA.md).

没有同意 CLA 的 PR 不会被合并。CLA 不转让你的版权，只是授予项目使用和分发你的贡献的权利。

## 提交前

- 先读 [docs/architecture.md](docs/architecture.zh-CN.md)、[docs/extending.md](docs/extending.zh-CN.md) 和 [docs/testing.md](docs/testing.zh-CN.md)。
- 跑一遍检查，全部通过再提交：

```sh
gofmt -l .
go vet ./...
go test ./...
```

- 新组件要有文档（`docs/kit/<名字>.md`）、组件库示例（`examples/components/<名字>.go`）和窗口级 Agent 测试，`go test` 会检查。
- 一个 PR 只做一件事，提交信息说清楚改了什么、为什么。
