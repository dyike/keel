# Unicode emoji catalog

English | [简体中文](README.zh-CN.md)

Source: [Unicode 17.0 emoji-test.txt](https://www.unicode.org/Public/17.0.0/emoji/emoji-test.txt), dated 2025-08-04. Licensed under [Unicode License V3](LICENSE).

Pinned SHA256: 1d8a944f88d7952f7ef7c5167fef3c67995bcae24543949710231b03a201acda.

scripts/generate-emoji.py produces ui/kit/emoji_catalog_generated.go from this exact release. Pass --input /path/to/emoji-test.txt to regenerate offline; add --check to verify without writing. The hash is checked before parsing. Normal component use never downloads data.

All 3,944 fully-qualified sequences are retained. Uniform skin-tone variants are grouped under a named base when available. Mixed-tone combinations and unmatched variants remain independent entries. Standalone components, minimally-qualified and unqualified aliases are excluded. Category order and English names follow the source. The generated constant is parsed lazily on first catalog use; consumers receive copies. No font or emoji image is included.
