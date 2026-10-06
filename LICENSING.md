# Licensing

English | [简体中文](LICENSING.zh-CN.md)

Copyright © 2026 dyike

Keel is **dual licensed**. Choose either of the following licenses.

## 1. Open-source license: AGPL-3.0

The full terms are in [LICENSE](LICENSE): GNU Affero General Public License v3.0, version 3 only, SPDX `AGPL-3.0-only`.

Under the AGPL, you may use, modify, and distribute Keel for free, including commercially, subject to its conditions:

- When **distributing** a program that uses Keel, whether or not you modified Keel, disclose the source code of the **entire program** under AGPL-3.0 and include the license notice.
- When users access the program over a **network**, such as a Keel interface running on a server for remote operation, provide those users with the complete source code as well.
- Do not impose additional restrictions or incorporate Keel’s code into a closed-source program.

Suitable for open-source projects, personal learning and tools, and products willing to publish their source under the AGPL.

## 2. Commercial license

If you do not want to disclose your application’s source code—for example, for closed-source desktop software or internal company products that will not be released under the AGPL—purchase a commercial license. It:

- Allows Keel in closed-source and paid software without requiring disclosure of your source code.
- Includes updates and technical support as agreed.
- Defines pricing, scope (per product, developer count, or company), and duration in the commercial license agreement signed by both parties.

Contact [@dyike](https://github.com/dyike) on GitHub.

## Which license should I use?

| Your situation | License |
| --- | --- |
| Personal learning or tools for your own use, without distribution | AGPL; without distribution, the source-disclosure obligation is not triggered |
| An open-source project willing to publish all source under the AGPL or a compatible license | AGPL |
| A company or individual distributing **closed-source** software to users | Commercial license |
| A Keel-based online service that will not disclose source to users | Commercial license |
| Unsure | Contact us |

This explanation helps you understand the options. Rights and obligations are governed by [LICENSE](LICENSE) and the commercial license agreement.

## Third-party components

Keel’s dependencies use permissive licenses: MIT, BSD, Apache-2.0, and Unlicense. See their source repositories for copyright and license terms. They are compatible with both AGPL-3.0 and commercial licensing:

- Gio (`gioui.org`): Unlicense or MIT
- goldmark, chroma, go-text/typesetting, jezek/xgb, and others: MIT or BSD
- `golang.org/x/*`: BSD-3-Clause
- modelcontextprotocol/go-sdk: MIT

The documentation gallery’s Noto Sans SC font is licensed under SIL Open Font License 1.1. It is downloaded during the build and is not stored in this repository.

## Contributions

To support both AGPL and commercial releases, external contributors must agree to the [Contributor License Agreement (CLA)](CLA.md). See [CONTRIBUTING.md](CONTRIBUTING.md).
