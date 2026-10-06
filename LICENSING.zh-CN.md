# 授权

[English](LICENSING.md) | 简体中文

Copyright © 2026 dyike

Keel 采用**双授权**：你可以在下面两种授权中任选一种使用。

## 1. 开源授权：AGPL-3.0

完整条款见 [LICENSE](LICENSE)（GNU Affero General Public License v3.0，仅限第 3 版，SPDX：`AGPL-3.0-only`）。

按 AGPL 使用时可以免费使用、修改、分发，包括商业用途，但必须遵守它的条件：

- **分发**用到 Keel 的程序（无论是否修改过 Keel）时，要以 AGPL-3.0 公开**整个程序**的源代码，并附上授权声明。
- 程序经由**网络**提供给用户使用时（比如把基于 Keel 的界面跑在服务器上让用户远程操作），同样要向这些用户提供完整源代码。
- 不能附加更严格的限制，也不能把 Keel 的代码并入闭源程序。

适合：开源项目、个人学习和自用工具、愿意以 AGPL 公开源码的产品。

## 2. 商业授权

不想公开自己程序的源代码，比如开发闭源的桌面软件、在公司内部产品里使用但不打算以 AGPL 发布，就需要购买商业授权。商业授权：

- 不要求公开你的源代码，可以把 Keel 用在闭源、收费的软件里；
- 按约定提供更新和技术支持；
- 具体价格、范围（按产品、按开发者人数或按公司）和期限，以双方签订的商业授权合同为准。

联系：在 GitHub 上联系 [@dyike](https://github.com/dyike)。

## 怎么判断该用哪一种

| 你的情况 | 需要 |
| --- | --- |
| 个人学习、写自用小工具，不分发 | AGPL 即可（不分发就不触发公开源码的义务） |
| 开源项目，愿意以 AGPL 或兼容协议公开全部源码 | AGPL 即可 |
| 公司或个人开发**闭源**软件并分发给用户 | 商业授权 |
| 基于 Keel 提供在线服务，不愿向用户公开源码 | 商业授权 |
| 拿不准 | 联系我们 |

这里的说明只是帮助理解，权利义务以 [LICENSE](LICENSE) 原文和商业授权合同为准。

## 第三方组件

Keel 依赖的第三方库都是宽松协议（MIT、BSD、Apache-2.0、Unlicense），各自的版权和协议见它们的源码。它们与 AGPL-3.0 及商业授权都兼容：

- Gio（`gioui.org`）：Unlicense 或 MIT
- goldmark、chroma、go-text/typesetting、jezek/xgb 等：MIT 或 BSD
- `golang.org/x/*`：BSD-3-Clause
- modelcontextprotocol/go-sdk：MIT

文档站在线示例使用的 Noto Sans SC 字体（构建时下载，不在仓库中）采用 SIL Open Font License 1.1。

## 贡献

为了能同时以 AGPL 和商业授权发布，外部贡献需要同意 [贡献者授权协议（CLA）](CLA.zh-CN.md)，见 [CONTRIBUTING.md](CONTRIBUTING.zh-CN.md)。
