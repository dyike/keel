package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("emoji_picker", "inputs", func() core.Widget {
		message := kit.Input(demoText("Message", "消息"))
		message.SetValue(demoText("Lunch together? ", "一起吃午饭？"))
		picker := kit.EmojiPicker().Popup(true).Name(demoText("Add emoji", "添加表情"))
		picker.SetRecent([]string{"👍", "😂", "❤️", "🎉", "🙏", "🔥"})
		picker.OnSelect(func(s kit.EmojiSelection) { message.SetValue(message.Value() + s.Emoji) })
		panel := kit.EmojiPicker().Height(264)
		selected := ""
		panel.OnSelect(func(s kit.EmojiSelection) { selected = s.Emoji + " · " + s.Name })
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(16).Gap(16).ScrollY().Items(el.Start).
				Child(message.Render(cx), picker.Render(cx), el.Text(selected), panel.Render(cx))
		}))
	})
}
