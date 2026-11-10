// AppShell demonstrates customizable desktop application chrome and content.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type conversation struct {
	messages []string
	draft    *kit.InputView
}

func (c *conversation) Render(cx *el.Context) el.Element {
	body := el.Div().Gap(24).Items(el.Stretch)
	body.Child(el.Text("今天想从哪里开始？").TextSize(22).Bold(), el.Text("把想法写下来，整理下一步要做的事。").TextColor(theme.Muted))
	for _, message := range c.messages {
		body.Child(el.Div().Row().Justify(el.End).Child(el.Div().MaxW(el.Dp(520)).P(16).Rounded(16).Bg(theme.Subtle).Child(el.Text(message))))
	}
	return body
}

func newApp() *kit.AppShellView {
	shell := kit.AppShell("Workspace").SidebarWidth(248).HideSidebarOnCollapse(true)
	chats := map[string]*conversation{}
	for _, p := range []struct{ id, title string }{{"plan", "本周工作计划"}, {"notes", "整理项目笔记"}, {"new", "新会话"}} {
		c := &conversation{draft: kit.TextArea("").Placeholder("在这里输入消息").AutoGrow(2, 5)}
		if p.id == "plan" {
			c.messages = []string{"帮我整理一下这周的工作计划。"}
		}
		chats[p.id] = c
		shell.Page(kit.AppPage{ID: p.id, Title: p.title, Icon: kit.IconFile, Create: func() el.View { return c }})
	}
	nav := kit.Sidebar().Width(248).BorderWidth(0).Collapsible(false).
		Section("本月", kit.SidebarItem{ID: "plan", Label: "本周工作计划", Icon: kit.IconFile}).
		Section("会话", kit.SidebarItem{ID: "notes", Label: "整理项目笔记", Icon: kit.IconFile}, kit.SidebarItem{ID: "new", Label: "新会话", Icon: kit.IconPlus})
	nav.SetValue(shell.Value())
	nav.OnChange(shell.SetValue)
	search := kit.Input("").Placeholder("搜索会话").OnChange(nav.Filter)
	searching := false
	shell.Sidebar(el.ViewFunc(func(cx *el.Context) el.Element {
		nav.SetCollapsed(shell.Collapsed())
		nav.Width(shell.SidebarSize())
		body := el.Div().HFull().Items(el.Stretch)
		if searching && !shell.Collapsed() {
			body.Child(el.Div().P(12).Child(search.Render(cx)))
		}
		return body.Child(el.Div().Grow().MinH(el.Dp(0)).Items(el.Stretch).Child(nav.Render(cx)))
	}))
	shell.SidebarHeader(el.ViewFunc(func(cx *el.Context) el.Element {
		row := el.Div().Row().Gap(4)
		row.Child(kit.Button("", func() { shell.SetValue("new"); nav.SetValue("new") }).Name("新建会话").Icon(kit.IconPlus).Variant(kit.ButtonGhost).Render(cx))
		if !shell.Collapsed() {
			row.Child(kit.Button("", func() { searching = !searching }).Name("搜索会话").Icon(kit.IconSearch).Variant(kit.ButtonGhost).Render(cx))
		}
		return row
	}))
	model := kit.Select("", "工作助手", "写作助手")
	model.SetValue("工作助手")
	shell.Header(el.ViewFunc(func(cx *el.Context) el.Element {
		width, _ := cx.ViewportSize()
		modelWidth := float32(180)
		if width < 760 {
			modelWidth = 120
		}
		return el.Div().W(el.Dp(modelWidth * max(1, cx.TextScale()))).Child(model.Render(cx))
	}))
	collapseMode := kit.Select("", "保留图标栏", "完全隐藏").Hint("折叠方式").OnChange(func(mode string) {
		shell.HideSidebarOnCollapse(mode == "完全隐藏")
	})
	collapseMode.SetValue("完全隐藏")
	shell.Actions(el.ViewFunc(func(cx *el.Context) el.Element {
		row := el.Div().Row().Items(el.Center).Gap(8)
		width, _ := cx.ViewportSize()
		if width >= 900 {
			row.Child(el.Text("折叠方式").TextSize(12).TextColor(theme.Muted))
		}
		return row.Child(el.Div().W(el.Dp(140 * max(1, cx.TextScale()))).Child(collapseMode.Render(cx)))
	}))

	shell.Footer(el.ViewFunc(func(cx *el.Context) el.Element {
		row := el.Div().Row().Items(el.Center).Gap(8).P(16).Child(kit.Icon(kit.IconUser).Render(cx))
		if !shell.Collapsed() {
			row.Child(el.Text("我的工作区"))
		}
		return row
	}))
	shell.ContentFooter(el.ViewFunc(func(cx *el.Context) el.Element {
		c := chats[shell.Value()]
		send := func() {
			if value := strings.TrimSpace(c.draft.Value()); value != "" {
				c.messages = append(c.messages, value)
				c.draft.SetValue("")
			}
		}
		return el.Div().P(20).Items(el.Stretch).Child(el.Div().P(12).Gap(8).Rounded(16).Bg(theme.Surface).Child(
			c.draft.Render(cx), el.Div().Row().Items(el.Center).Justify(el.SpaceBetween).Child(el.Text("消息仅保存在当前会话中").TextSize(12).TextColor(theme.Muted), kit.Button("发送", send).Render(cx))))
	}))
	return shell
}

func main() {
	screenshot := flag.String("screenshot", "", "save a PNG and exit")
	flag.Parse()
	options := window.Options{Title: "Keel · Workspace", Width: 1100, Height: 760, MinWidth: 560, MinHeight: 420, Frameless: true, NativeTrafficLights: true, TrafficLightLayout: &window.TrafficLightLayout{Height: kit.AppShellTitleBarHeight, Left: 16, Spacing: 20}, Content: el.Root(newApp())}
	if *screenshot != "" {
		w := window.OpenHeadless(options)
		defer w.Close()
		data, err := w.Screenshot()
		if err != nil {
			log.Fatal(err)
		}
		if err = os.WriteFile(*screenshot, data, 0644); err != nil {
			log.Fatal(err)
		}
		return
	}
	options.AutoSize = true
	if dir, err := os.UserCacheDir(); err == nil {
		state, err := window.LoadSizeState(filepath.Join(dir, "keel-examples", "workspace-size.json"))
		if err != nil {
			log.Print(err)
		}
		options.SizeState = state
	}
	window.Open(options)
	window.Main()
}
