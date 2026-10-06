package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"strings"
	"time"
)

func init() {
	registerSection("command_async", "overlays", func() core.Widget {
		message := demoText("Type a query for async search; type error to simulate failure", "输入关键词异步搜索；输入 error 模拟失败")
		palette := kit.Command()
		palette.OnSearch(func(query string, token uint64) {
			go func() {
				delay := 200 * time.Millisecond
				if len(query) == 1 {
					delay = 700 * time.Millisecond
				}
				time.Sleep(delay)
				core.Update(func() {
					if strings.EqualFold(query, "error") {
						palette.SetSearchError(token, demoText("Search failed. Change the query or try again", "搜索失败，请修改关键词或重试"))
						return
					}
					items := make([]kit.CommandItem, 30)
					for i := range items {
						title := fmt.Sprintf(demoText("%s search result %02d", "%s 搜索结果 %02d"), query, i+1)
						items[i] = kit.CommandItem{Title: title, Group: demoText("Remote results", "远程结果"), Action: func() { message = demoText("Selected: ", "选择了：") + title }}
					}
					palette.SetResults(token, items...)
				})
			}()
		})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).Items(el.Start).Child(kit.Button(demoText("Async command search", "异步命令搜索"), palette.Toggle).Render(cx), el.Text(message), palette.Render(cx))
		}))
	})
}
