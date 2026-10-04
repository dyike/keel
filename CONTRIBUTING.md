# 参与贡献

欢迎提交 Issue 和 Pull Request。

## 授权

Keel 采用 AGPL-3.0 与商业授权双授权（见 [LICENSING.md](LICENSING.md)）。为了让你的贡献也能随商业授权发布，**提交第一个 Pull Request 前需要同意 [贡献者授权协议（CLA）](CLA.md)**：在 PR 描述里写上

> I have read and agree to the Keel Contributor License Agreement (CLA.md).

没有同意 CLA 的 PR 不会被合并。CLA 不转让你的版权，只是允许项目以开源和商业两种授权发布你的贡献。

## 提交前

- 先读 [docs/architecture.md](docs/architecture.md)、[docs/extending.md](docs/extending.md) 和 [docs/testing.md](docs/testing.md)。
- 跑一遍检查，全部通过再提交：

```sh
gofmt -l .
go vet ./...
go test ./...
```

- 新组件要有文档（`docs/kit/<名字>.md`）、组件库示例（`examples/components/<名字>.go`）和窗口级 Agent 测试，`go test` 会检查。
- 一个 PR 只做一件事，提交信息说清楚改了什么、为什么。
