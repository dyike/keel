#!/usr/bin/env python3
"""Subset a CJK font for the WebAssembly demo.

Keeps every character in the repository's Go and Markdown files (all the
text the gallery shows) plus the 6,763 hanzi of GB2312, so what visitors
type renders too. Usage: subset_font.py <repo> <font.otf> <out.otf>
Needs fontTools (pip install fonttools).
"""
import pathlib
import sys

from fontTools import subset


def main():
    repo, src, out = pathlib.Path(sys.argv[1]), sys.argv[2], sys.argv[3]
    chars = set(chr(c) for c in range(0x20, 0x7F))
    for p in list(repo.rglob("*.go")) + list(repo.rglob("*.md")):
        if any(part in (".git", "work", "node_modules", "_site") for part in p.parts):
            continue
        chars.update(p.read_text(encoding="utf-8", errors="ignore"))
    for hi in range(0xB0, 0xF8):
        for lo in range(0xA1, 0xFF):
            try:
                chars.add(bytes([hi, lo]).decode("gb2312"))
            except UnicodeDecodeError:
                pass
    chars.update("，。、；：？！“”‘’（）《》【】—…·￥")
    chars = {c for c in chars if c.isprintable()}
    options = subset.Options()
    options.layout_features = ["*"]
    options.name_IDs = ["*"]
    options.notdef_outline = True
    font = subset.load_font(src, options)
    sub = subset.Subsetter(options)
    sub.populate(text="".join(sorted(chars)))
    sub.subset(font)
    subset.save_font(font, out, options)
    print(f"{len(chars)} characters -> {out}")


if __name__ == "__main__":
    main()
