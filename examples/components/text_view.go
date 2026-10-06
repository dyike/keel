package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/markdown"
	"github.com/dyike/keel/ui/theme"
	"github.com/yuin/goldmark/ast"
)

func init() {
	registerSection("text_view", "data", func() core.Widget {
		status := ""
		expanded := false
		fadeDoc := markdown.New("").StreamFade(true)
		chunks := []string{demoText("Text", "文字"), demoText("Click to add", "按新增"), demoText("Snippet", "片段"), demoText(" appears gradually.", "逐渐显示。"), demoText("\n\n**Bold", "\n\n**加"), demoText("Rich content", "粗内容"), demoText("** remains continuous.", "**也保持连续。"), "\n\n```go\n", "fmt.", "Println", "(\"hello\")", "\n```"}
		streamAt, streamRun := 0, false
		doc := markdown.New(demoText("# Code block extensions\n\n```go\nfmt.Println(\"hello\")\n```\n\n```notice\nThe application renders this fenced block as a notice card.\n```\n\n> ```text\n> Nested code blocks also have action buttons.\n> ```\n", "# 代码块扩展\n\n```go\nfmt.Println(\"hello\")\n```\n\n```notice\n这段代码围栏由应用渲染为提示卡片。\n```\n\n> ```text\n> 嵌套代码块也有操作按钮。\n> ```\n"))
		doc.CodeBlockActions(func(cx *el.Context, block markdown.CodeBlockContext) el.Element {
			return kit.Button(demoText("View information", "查看信息"), func() {
				status = fmt.Sprintf(demoText("Language %q, %d lines", "语言 %q，%d 行"), block.Language, strings.Count(strings.TrimSuffix(block.Text, "\n"), "\n")+1)
			}).ID(block.ID + "/inspect").Variant(kit.ButtonGhost).Size(28).Render(cx)
		})
		doc.CodeBlockRenderer("notice", func(cx *el.Context, block markdown.CodeBlockContext) el.Element {
			return el.Div().P(16).Gap(8).Rounded(theme.RadiusMd).Bg(theme.Subtle).Child(
				el.Text(demoText("App-defined tooltip", "应用自定义提示")).Bold(), el.Text(block.Text),
				kit.Button(demoText("Confirm", "确认"), func() { status = demoText("Alert acknowledged", "已确认提示") }).ID(block.ID+"/confirm").Size(28).Render(cx))
		})
		pluginDoc := markdown.New(demoText("The plugin renders `@Keel` as an inline tag; other `code` remains unchanged.\n\n> NOTE: A block plugin renders this quotation.\n", "插件将 `@Keel` 渲染成行内标签，其他 `code` 保持原样。\n\n> NOTE: 这条引用由块插件渲染。\n")).Plugins(markdown.Plugin{
			Blocks: map[ast.NodeKind]func(ast.Node, []byte) el.View{
				ast.KindBlockquote: func(n ast.Node, source []byte) el.View {
					label := strings.TrimSpace(string(n.Text(source)))
					if !strings.HasPrefix(label, "NOTE:") {
						return nil
					}
					return el.ViewFunc(func(cx *el.Context) el.Element {
						return el.Div().P(12).Rounded(theme.RadiusMd).Bg(theme.Subtle).Child(el.Text(strings.TrimPrefix(label, "NOTE:")).Bold())
					})
				},
			},
			Inlines: map[ast.NodeKind]func(ast.Node, []byte) markdown.InlineObject{
				ast.KindCodeSpan: func(n ast.Node, source []byte) markdown.InlineObject {
					label := string(n.Text(source))
					if !strings.HasPrefix(label, "@") {
						return markdown.InlineObject{}
					}
					return markdown.InlineObject{Text: label, Widget: el.Embed(kit.Tag(label).Size(20))}
				},
			},
		})
		fmDoc := markdown.New(demoText("---\ntitle: Release notes v0.0.2\nauthor: yike\ndate: 2026-10-04\n---\n## Content starts here\n\nYAML metadata between the opening `---` delimiters is hidden by default. Read it with `Meta()`.\n", "---\ntitle: 发布说明 v0.0.2\nauthor: yike\ndate: 2026-10-04\n---\n## 正文从这里开始\n\n文档开头 `---` 之间的 YAML 元数据默认不显示，可以用 `Meta()` 读取。\n"))
		showMeta := false
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			if streamRun {
				cx.After(struct {
					Doc  *markdown.Doc
					Part int
				}{fadeDoc, streamAt}, 120*time.Millisecond, func() {
					fadeDoc.Append(chunks[streamAt])
					streamAt++
					if streamAt == len(chunks) {
						streamRun = false
						fadeDoc.SetStreaming(false)
					}
				})
			}
			if expanded {
				doc.MaxLines(0)
			} else {
				doc.MaxLines(6)
			}
			label := demoText("Expand full text", "展开全文")
			if expanded {
				label = demoText("Collapse preview", "收起预览")
			}
			return el.Div().P(24).Gap(12).ScrollY().Child(
				kit.Button(demoText("Locate and highlight nested code", "定位并高亮嵌套代码"), func() {
					expanded = true
					doc.MaxLines(0)
					word := demoText("Nested code block", "嵌套代码块")
					start := strings.Index(doc.RenderedText(), word)
					if start < 0 {
						return
					}
					r := markdown.TextRange{Start: start, End: start + len(word)}
					doc.SetRangeHighlights([]markdown.RangeHighlight{{Range: r, Background: theme.Highlight}})
					doc.RevealRange(r)
				}).Render(cx),
				kit.Button(demoText("Clear highlights", "清除高亮"), func() { doc.SetRangeHighlights(nil) }).Render(cx),
				doc.Render(cx),
				kit.Button(label, func() { expanded = !expanded }).Render(cx),
				el.Text(fmt.Sprintf(demoText("Clipped in the previous frame: %v", "上一帧发生截断：%v"), doc.IsClamped())).TextColor(theme.Muted),
				el.Text(status).TextColor(theme.Muted),
				kit.Button(demoText("Demo streaming fade-in", "演示流式淡入"), func() {
					fadeDoc.SetSource("")
					fadeDoc.SetStreaming(true)
					streamAt, streamRun = 0, true
				}).Render(cx), fadeDoc.Render(cx), pluginDoc.Render(cx),
				kit.Button(demoText("Show / hide YAML metadata", "显示 / 隐藏 YAML 元数据"), func() { showMeta = !showMeta; fmDoc.ShowFrontMatter(showMeta) }).Variant(kit.ButtonSecondary).Render(cx),
				el.Text("Meta()[\"title\"] = "+fmDoc.Meta()["title"]).TextColor(theme.Muted),
				fmDoc.Render(cx))
		}))
	})
}
