# Contributing

English | [简体中文](CONTRIBUTING.zh-CN.md)

Issues and Pull Requests are welcome.

## Licensing

Keel is dual licensed under AGPL-3.0 and a commercial license; see [LICENSING.md](LICENSING.md). To allow your contribution in commercial releases, **agree to the [Contributor License Agreement (CLA)](CLA.md) before your first Pull Request**. Include this sentence in the PR description:

> I have read and agree to the Keel Contributor License Agreement (CLA.md).

PRs without CLA acceptance will not be merged. The CLA preserves your copyright and allows the Project to release your contribution under both open-source and commercial licenses.

## Before submitting

- Read [Architecture](docs/architecture.md), [Extending Keel](docs/extending.md), and [Testing](docs/testing.md).
- Run the checks and submit only after they pass:

```sh
gofmt -l .
go vet ./...
go test ./...
```

- A new component needs documentation (`docs/kit/<name>.md`), a gallery example (`examples/components/<name>.go`), and window-level agent tests. `go test` checks these conventions.
- Keep each PR focused on one change. Explain what changed and why in the commit message.
