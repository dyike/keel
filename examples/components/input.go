package main

import (
	"fmt"
	"github.com/dyike/keel/native/clipboard"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("input", "inputs", func() core.Widget {
		search := kit.Input("搜索").Placeholder("客户名称或单号 123").Clearable().Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Icon(kit.IconSearch).Size(16).Color(theme.Muted).Render(cx)
		}))
		search.ContextMenu(kit.Menu().Item("示例：搜索客户", "", func() { search.SetValue("客户") }).Separator().Item("清空搜索", "", func() { search.SetValue("") }))
		price := kit.Input("单价 · 数字分组").Placeholder("0.00").NumberMask(',', 2).Suffix(el.ViewFunc(func(*el.Context) el.Element { return el.Text("元") }))
		phone := kit.Input("电话 · 模板掩码").Mask("(###)-###-####").Placeholder("(123)-456-7890").Clearable()
		pass := kit.Input("密码").Password()
		bad := kit.Input("邮箱 Email")
		bad.SetValue("not-an-email")
		bad.SetError("邮箱格式不正确")
		note := kit.TextArea("备注").Placeholder("自动增高：2–6 行，超出后内部滚动").AutoGrow(2, 6)
		pasteInfo := "可粘贴图片或文件；普通文本仍插入备注。"
		note.PasteReader(readComponentClipboard).OnPaste(func(data core.ClipboardData) bool {
			if len(data.Images)+len(data.Files) == 0 {
				return false
			}
			pasteInfo = fmt.Sprintf("已接收 %d 张图片、%d 个文件引用", len(data.Images), len(data.Files))
			return true
		}).OnPasteError(func(err error) { pasteInfo = "富剪贴板读取失败，回退文本：" + err.Error() })
		off := kit.Input("只读")
		off.SetValue("SO-1001")
		off.SetReadOnly(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(14).W(el.Dp(360)).MaxW(el.Full).Child(search.Render(cx), price.Render(cx), phone.Render(cx), el.Text("电话原值："+phone.UnmaskedValue()).TextColor(theme.Muted), pass.Render(cx), bad.Render(cx), note.Render(cx), el.Text(pasteInfo).TextColor(theme.Muted), off.Render(cx))
		}))
	})
}

// readComponentClipboard keeps the native-to-UI adaptation in the application.
func readComponentClipboard(done func(core.ClipboardData, error)) {
	clipboard.Read(func(data clipboard.Data, err error) {
		result := core.ClipboardData{Text: data.Text, Files: data.Files}
		for _, img := range data.Images {
			result.Images = append(result.Images, core.ClipboardImage{MIME: img.MIME, Data: img.Data})
		}
		done(result, err)
	})
}
