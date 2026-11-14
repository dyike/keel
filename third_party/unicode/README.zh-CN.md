# Unicode 表情目录

[English](README.md) | 简体中文

来源：[Unicode 17.0 emoji-test.txt](https://www.unicode.org/Public/17.0.0/emoji/emoji-test.txt)，日期为 2025-08-04。许可证为 [Unicode License V3](LICENSE)。

固定 SHA256：1d8a944f88d7952f7ef7c5167fef3c67995bcae24543949710231b03a201acda。

scripts/generate-emoji.py 使用此固定版本生成 ui/kit/emoji_catalog_generated.go。传 --input /path/to/emoji-test.txt 可离线生成，加 --check 可只检查而不写入。解析前验证哈希。正常使用组件不会下载数据。

保留全部 3,944 个 fully-qualified 序列。存在对应名称的基础表情时，将统一肤色变体归入该基础条目；混合肤色组合及未匹配变体保留为独立条目。不包含独立组件、minimally-qualified 和 unqualified 别名。分类顺序及英文名称沿用源文件。生成常量在首次使用目录时才解析，提供给使用者的是副本。不包含字体或表情图片。
