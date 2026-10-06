package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var languageLineRE = regexp.MustCompile(`(?m)^(?:English \| \[简体中文\]\([^\n]+\)|\[English\]\([^\n]+\) \| 简体中文)\n`)

func canonicalSource(src string) string {
	if strings.HasSuffix(src, ".zh-CN.md") {
		return strings.TrimSuffix(src, ".zh-CN.md") + ".md"
	}
	return src
}
func sourceLanguage(src string) string {
	if strings.HasSuffix(src, ".zh-CN.md") {
		return "zh-CN"
	}
	return "en"
}
func localizedSource(src, lang string) string {
	src = canonicalSource(src)
	if lang == "zh-CN" {
		return strings.TrimSuffix(src, ".md") + ".zh-CN.md"
	}
	return src
}
func otherLanguage(lang string) string {
	if lang == "zh-CN" {
		return "en"
	}
	return "zh-CN"
}
func translateUI(p *page, zh, en string) string {
	if p.Lang == "zh-CN" {
		return zh
	}
	return en
}
func localizedNavTitle(lang, title string) string {
	if lang == "zh-CN" {
		return title
	}
	return map[string]string{"示例": "Examples", "界面模块": "UI", "系统能力": "Native", "命令行工具": "CLI", "工具与内部实现": "Internals", "源码导览": "Packages"}[title]
}
func (s *site) switchURL(p *page) string {
	root := relURL(p.Out, ".")
	if s.lang == "zh-CN" {
		return root + "../" + p.Out
	}
	return root + "zh-CN/" + p.Out
}
func (s *site) demoURL(p *page) string {
	url := relURL(p.Out, "demo/index.html")
	if s.lang == "zh-CN" {
		url = relURL(p.Out, ".") + "../demo/index.html"
	}
	return url
}
func localizedRedirect(lang string) string {
	if lang == "zh-CN" {
		return scaffoldRedirect
	}
	return strings.NewReplacer(`lang="zh-CN"`, `lang="en"`, "快速开始 · Keel", "Getting started · Keel", "脚手架与打包已并入", "Scaffolding and packaging are covered in ", "快速开始", "Getting started", "。</p>", ".</p>").Replace(scaffoldRedirect)
}
func (s *site) headingIDs(src string) []string {
	if ids, ok := s.headingCache[src]; ok {
		return ids
	}
	if s.headingCache == nil {
		s.headingCache = map[string][]string{}
	}
	data, err := os.ReadFile(filepath.Join(s.root, src))
	if err != nil {
		return nil
	}
	ctx := parser.NewContext(parser.WithIDs(&githubIDs{seen: map[string]int{}}))
	doc := md.Parser().Parse(text.NewReader(data), parser.WithContext(ctx))
	var ids []string
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if h, ok := n.(*ast.Heading); ok && entering {
			id, _ := h.AttributeString("id")
			ids = append(ids, string(id.([]byte)))
		}
		return ast.WalkContinue, nil
	})
	s.headingCache[src] = ids
	return ids
}
func (s *site) anchorPairs(src string, from, to string) map[string]string {
	canonical := canonicalSource(src)
	a := s.headingIDs(localizedSource(canonical, from))
	b := s.headingIDs(localizedSource(canonical, to))
	pairs := map[string]string{}
	if len(a) != len(b) {
		return pairs
	}
	for i, id := range a {
		pairs[id] = b[i]
		if from == "en" && id != b[i] {
			pairs[b[i]] = b[i]
		}
	}
	return pairs
}
func (s *site) switchAnchors(p *page) string {
	data, _ := json.Marshal(s.anchorPairs(p.Src, s.lang, otherLanguage(s.lang)))
	return string(data)
}
func (s *site) translatedFragment(src, fragment string) string {
	if s.lang == "en" {
		if target := s.anchorPairs(src, "zh-CN", "en")[fragment]; target != "" {
			return target
		}
	}
	return fragment
}
func (s *site) legacyAnchors(p *page, html string) string {
	if s.lang != "en" {
		return html
	}
	for old, id := range s.anchorPairs(p.Src, "zh-CN", "en") {
		if old == id {
			continue
		}
		// Keep existing bookmarks to Chinese section names working on the default URL.
		needle := `id="` + id + `"`
		start := strings.Index(html, needle)
		if start < 0 {
			continue
		}
		begin := strings.LastIndex(html[:start], "<h")
		if begin >= 0 {
			html = html[:begin] + `<span class="anchor-alias" id="` + old + `"></span>` + html[begin:]
		}
	}
	return html
}
