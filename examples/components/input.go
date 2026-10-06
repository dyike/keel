package main

import (
	"fmt"

	"github.com/dyike/keel/native/clipboard"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func init() {
	registerSection("input", "inputs", func() core.Widget {
		search := kit.Input(demoText("Search", "搜索")).Placeholder(demoText("Customer name or order number 123", "客户名称或单号 123")).Clearable().Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Icon(kit.IconSearch).Size(16).Color(theme.Muted).Render(cx)
		}))
		search.ContextMenu(kit.Menu().Item(demoText("Example: search customers", "示例：搜索客户"), "", func() { search.SetValue(demoText("Customer", "客户")) }).Separator().Item(demoText("Clear search", "清空搜索"), "", func() { search.SetValue("") }))
		price := kit.Input(demoText("Unit price · Digit grouping", "单价 · 数字分组")).Placeholder("0.00").NumberMask(',', 2).Suffix(el.ViewFunc(func(*el.Context) el.Element { return el.Text(demoText("CNY", "元")) }))
		phone := kit.Input(demoText("Phone · Template mask", "电话 · 模板掩码")).Mask("(###)-###-####").Placeholder("(123)-456-7890").Clearable()
		pass := kit.Input(demoText("Password", "密码")).Password()
		bad := kit.Input(demoText("Email", "邮箱 Email"))
		bad.SetValue("not-an-email")
		bad.SetError(demoText("Invalid email address", "邮箱格式不正确"))
		note := kit.TextArea(demoText("Notes", "备注")).Placeholder(demoText("Auto height: 2–6 lines; scroll internally beyond that", "自动增高：2–6 行，超出后内部滚动")).AutoGrow(2, 6)
		pasteInfo := demoText("Paste images or files; plain text still inserts into the notes.", "可粘贴图片或文件；普通文本仍插入备注。")
		note.PasteReader(readComponentClipboard).OnPaste(func(data core.ClipboardData) bool {
			if len(data.Images)+len(data.Files) == 0 {
				return false
			}
			pasteInfo = fmt.Sprintf(demoText("Received %d images and %d file references", "已接收 %d 张图片、%d 个文件引用"), len(data.Images), len(data.Files))
			return true
		}).OnPasteError(func(err error) {
			pasteInfo = demoText("Rich clipboard read failed; falling back to text: ", "富剪贴板读取失败，回退文本：") + err.Error()
		})
		reference := kit.Input(demoText("Reference", "引用")).Clearable()
		tokenRenderer := func(gtx core.C, token kit.InputToken) core.D {
			return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Row().Items(el.Center).Gap(4).Px(4).Rounded(theme.RadiusSm).Bg(theme.Highlight).Child(kit.Icon(kit.IconFile).Size(14).Render(cx), el.Text(token.Display()).TextSize(theme.TextMd).MaxLines(1))
			})).Layout(gtx)
		}
		reference.TokenRenderer(tokenRenderer)
		draft, _ := kit.NewInputContent(demoText("View docs/input.md", "查看 docs/input.md"), kit.InputTokenSpan{Range: kit.InputRange{Start: len(demoText("View ", "查看 ")), End: len(demoText("View docs/input.md", "查看 docs/input.md"))}, Token: kit.InputToken{ID: "input-doc", Text: "docs/input.md", Label: demoText("Input component documentation", "输入组件文档")}})
		_ = reference.SetContent(draft)
		referenceInfo := demoText("Click a reference to view its ID; copying preserves the original path.", "点击引用查看标识；复制时保留原始路径。")
		reference.OnTokenActivate(func(token kit.InputToken) { referenceInfo = token.ID + " · " + token.Text })
		wrapped := kit.TextArea(demoText("References wrap as a unit", "引用整块换行")).AutoGrow(2, 5).TokenRenderer(tokenRenderer)
		prefix := demoText("The document reference following this description wraps as a unit: ", "这段说明后面的文档引用会整体换行： ")
		wrappedDraft, _ := kit.NewInputContent(prefix+"docs/input.md", kit.InputTokenSpan{Range: kit.InputRange{Start: len(prefix), End: len(prefix + "docs/input.md")}, Token: kit.InputToken{ID: "wrapped-doc", Text: "docs/input.md", Label: demoText("Input component reference and usage guide", "输入组件文档与使用说明")}})
		_ = wrapped.SetContent(wrappedDraft)
		off := kit.Input(demoText("Read only", "只读"))
		off.SetValue("SO-1001")
		off.SetReadOnly(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).W(el.Dp(360)).MaxW(el.Full).Child(search.Render(cx), price.Render(cx), phone.Render(cx), el.Text(demoText("Raw phone value: ", "电话原值：")+phone.UnmaskedValue()).TextColor(theme.Muted), pass.Render(cx), bad.Render(cx), note.Render(cx), el.Text(pasteInfo).TextColor(theme.Muted), off.Render(cx), reference.Render(cx), wrapped.Render(cx), el.Text(referenceInfo).TextColor(theme.Muted))
		}))
	})
}

// readComponentClipboard keeps the native-to-UI adaptation in the application.
// mainWindow is the gallery's window; on Wayland the clipboard is read
// through its connection.
var mainWindow *window.Window

func readComponentClipboard(done func(core.ClipboardData, error)) {
	if mainWindow != nil {
		clipboard.UseWaylandDisplay(mainWindow.WaylandDisplay())
	}
	clipboard.Read(func(data clipboard.Data, err error) {
		result := core.ClipboardData{Text: data.Text, Files: data.Files}
		for _, img := range data.Images {
			result.Images = append(result.Images, core.ClipboardImage{MIME: img.MIME, Data: img.Data})
		}
		done(result, err)
	})
}
