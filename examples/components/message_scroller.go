package main

import (
	"strconv"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("message_scroller", "data", func() core.Widget {
		var msgs []string
		for i := 20; i < 40; i++ {
			msgs = append(msgs, "消息 #"+strconv.Itoa(i))
		}
		older := 20
		var sc *kit.MessageScrollerView
		sc = kit.MessageScroller(func(cx *el.Context) []el.Element {
			var out []el.Element
			for i, m := range msgs {
				m := m
				msg := kit.Message("AI", el.ViewFunc(func(*el.Context) el.Element { return el.Text(m) }))
				if i%2 == 0 {
					msg.User()
				}
				out = append(out, msg.Render(cx))
			}
			return out
		}).OnReachTop(func() {
			if older == 0 {
				return
			}
			var batch []string
			for i := older - 10; i < older; i++ {
				batch = append(batch, "更早的消息 #"+strconv.Itoa(i))
			}
			older -= 10
			msgs = append(batch, msgs...)
			sc.HistoryPrepended()
		})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Child(
				el.Text("向上滚到顶部加载更早的消息，视图不会跳动").TextColor(theme.Muted).Px(20).Py(10),
				sc.Render(cx),
				el.Div().P(12).Child(kit.Button("发送新消息", func() {
					msgs = append(msgs, "新消息 #"+strconv.Itoa(len(msgs)))
					sc.ScrollToEnd()
				}).Render(cx)))
		}))
	})
}
