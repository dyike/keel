package main

import (
	"embed"
	"strings"
)

//go:embed samples/*.md
var sampleFiles embed.FS

type demoSample struct {
	name, title, description string
	keywords                 []string
}

var demoSamples = []demoSample{
	{"selection", "跨段落选择", "跨标题、段落、列表、代码和表格拖选；全选与复制。", []string{"selection", "跨段", "拖选", "选择", "选中"}},
	{"math", "数学公式", "行内公式、独立公式、分式、求和与矩阵。", []string{"math", "latex", "公式", "数学"}},
	{"code-scroll", "代码横向滚动", "含空格的长行、连续长字符串和短行对照。", []string{"code-scroll", "横向", "长代码", "长行"}},
	{"click-selection", "双击与三击", "英文词、中文句子、标点和跨样式文字。", []string{"click-selection", "双击", "三击", "选词", "选段"}},
	{"references", "脚注与引用", "重复脚注、多段脚注、跨块引用和同块链接对照。", []string{"references", "脚注", "引用", "链接"}},
	{"images", "图片", "本地 PNG、带链接的图片、空替代文字和缺失图片。", []string{"images", "image", "图片"}},
}

func sourceFor(name string) (string, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "all":
		var sources []string
		for _, s := range demoSamples {
			src, _ := sourceFor(s.name)
			sources = append(sources, src)
		}
		return strings.Join(sources, "\n\n---\n\n"), true
	case "table":
		return tableAnswer, true
	case "downloader":
		return downloaderAnswer, true
	}
	for _, s := range demoSamples {
		if s.name == name {
			src, err := sampleFiles.ReadFile("samples/" + name + ".md")
			if err != nil {
				panic(err) // an embedded fixture named in the catalogue is missing
			}
			return strings.TrimSpace(string(src)), true
		}
	}
	return "", false
}

func answerFor(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	if src, ok := sourceFor(q); ok {
		return src
	}
	if strings.Contains(q, "全部") || strings.Contains(q, "todo") {
		src, _ := sourceFor("all")
		return src
	}
	// Specific interactions win over broad words such as "选择" or "表".
	for _, name := range []string{"click-selection", "math", "code-scroll", "references", "images", "selection"} {
		for _, s := range demoSamples {
			if s.name != name {
				continue
			}
			for _, keyword := range s.keywords {
				if strings.Contains(q, keyword) {
					src, _ := sourceFor(s.name)
					return src
				}
			}
		}
	}
	if strings.Contains(q, "表") {
		return tableAnswer
	}
	if strings.Contains(q, "样例") || strings.Contains(q, "示例") || strings.Contains(q, "验证") || strings.Contains(q, "帮助") {
		src, _ := sourceFor("all")
		return src
	}
	return downloaderAnswer
}
