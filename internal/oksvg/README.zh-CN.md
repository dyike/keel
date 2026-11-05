# internal/oksvg

[English](README.md) | 简体中文

Keel 内部的 oksvg 副本，供 ui/kit 解析 SVG 图标和图片。依赖 rasterx 和 x/image，不依赖其他 Keel 包。默认支持 UTF-8；`-tags keel_legacy_svg` 通过 x/net 和 x/text 加入上游旧字符集解码。见[来源与补丁](../../third_party/README.zh-CN.md#svg-解析器)。
