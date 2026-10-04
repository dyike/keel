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
		chunks := []string{"文字", "按新增", "片段", "逐渐显示。", "\n\n**加", "粗内容", "**也保持连续。", "\n\n```go\n", "fmt.", "Println", "(\"hello\")", "\n```"}
		streamAt, streamRun := 0, false
		doc := markdown.New("# 代码块扩展\n\n```go\nfmt.Println(\"hello\")\n```\n\n```notice\n这段代码围栏由应用渲染为提示卡片。\n```\n\n> ```text\n> 嵌套代码块也有操作按钮。\n> ```\n")
		doc.CodeBlockActions(func(cx *el.Context, block markdown.CodeBlockContext) el.Element {
			return kit.Button("查看信息", func() {
				status = fmt.Sprintf("语言 %q，%d 行", block.Language, strings.Count(strings.TrimSuffix(block.Text, "\n"), "\n")+1)
			}).ID(block.ID + "/inspect").Variant(kit.ButtonGhost).Size(28).Render(cx)
		})
		doc.CodeBlockRenderer("notice", func(cx *el.Context, block markdown.CodeBlockContext) el.Element {
			return el.Div().P(16).Gap(8).Rounded(theme.RadiusMd).Bg(theme.Subtle).Child(
				el.Text("应用自定义提示").Bold(), el.Text(block.Text),
				kit.Button("确认", func() { status = "已确认提示" }).ID(block.ID+"/confirm").Size(28).Render(cx))
		})
		pluginDoc := markdown.New("插件将 `@Keel` 渲染成行内标签，其他 `code` 保持原样。\n\n> NOTE: 这条引用由块插件渲染。\n").Plugins(markdown.Plugin{
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
			label := "展开全文"
			if expanded {
				label = "收起预览"
			}
			return el.Div().P(24).Gap(12).ScrollY().Child(
				kit.Button("定位并高亮嵌套代码", func() {
					expanded = true
					doc.MaxLines(0)
					word := "嵌套代码块"
					start := strings.Index(doc.RenderedText(), word)
					if start < 0 {
						return
					}
					r := markdown.TextRange{Start: start, End: start + len(word)}
					doc.SetRangeHighlights([]markdown.RangeHighlight{{Range: r, Background: theme.Highlight}})
					doc.RevealRange(r)
				}).Render(cx),
				kit.Button("清除高亮", func() { doc.SetRangeHighlights(nil) }).Render(cx),
				doc.Render(cx),
				kit.Button(label, func() { expanded = !expanded }).Render(cx),
				el.Text(fmt.Sprintf("上一帧发生截断：%v", doc.IsClamped())).TextColor(theme.Muted),
				el.Text(status).TextColor(theme.Muted),
				kit.Button("演示流式淡入", func() {
					fadeDoc.SetSource("")
					fadeDoc.SetStreaming(true)
					streamAt, streamRun = 0, true
				}).Render(cx), fadeDoc.Render(cx), pluginDoc.Render(cx))
		}))
	})
}
