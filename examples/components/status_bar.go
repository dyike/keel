package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("status_bar", "controls", func() core.Widget { return el.Embed(newStatusBarGallery()) })
}

// The bars live in the view: an overflowing bar remembers its item widths.
type statusBarGallery struct {
	wide, narrow, editor *kit.StatusBarView
	width                *kit.SliderView
	last                 string
}

func newStatusBarGallery() *statusBarGallery {
	g := &statusBarGallery{}
	text := func(s string) el.View { return el.ViewFunc(func(cx *el.Context) el.Element { return el.Text(s) }) }
	g.wide = kit.StatusBar().Left(text(demoText("Connected", "连接正常 Connected"))).Right(text(demoText("123 records", "共 123 条记录")))
	g.narrow = kit.StatusBar().Left(text(demoText("Offline; waiting to reconnect", "离线，等待重新连接"))).Right(text(demoText("3 tasks awaiting delivery", "3 个任务待发送")))
	act := func(s string) func() { return func() { g.last = s } }
	g.editor = kit.StatusBar().Left(text(demoText("Ready", "就绪"))).
		Add(kit.StatusItem{Label: demoText("main branch", "main 分支"), Action: act(demoText("Switch branch", "切换分支")), Priority: 3},
			kit.StatusItem{Label: demoText("0 errors, 2 warnings", "0 个错误 2 个警告"), Action: act(demoText("Open problems panel", "打开问题面板")), Priority: 2}).
		AddRight(kit.StatusItem{Label: demoText("Line 12, column 4", "行 12，列 4"), Action: act(demoText("Go to line", "跳到行")), Priority: 2},
			kit.StatusItem{Label: demoText("Spaces: 4", "空格: 4"), Action: act(demoText("Indentation settings", "缩进设置")), Priority: 1},
			kit.StatusItem{Label: "UTF-8", Action: act(demoText("Choose encoding", "选择编码"))},
			kit.StatusItem{Label: "Go", Action: act(demoText("Choose language", "选择语言")), Priority: 1})
	g.width = kit.Slider(demoText("Status bar width", "状态栏宽度"), 240, 760).Step(20)
	g.width.SetValue(640)
	return g
}

func (g *statusBarGallery) Render(cx *el.Context) el.Element {
	note := demoText("Drag the slider to narrow the status bar. Items that do not fit move into the … menu by priority.", "拖动滑块缩窄状态栏，放不下的项按优先级收进 … 菜单")
	if g.last != "" {
		note = fmt.Sprintf(demoText("Clicked: %s", "点了：%s"), g.last)
	}
	return el.Div().W(el.Full).Gap(16).Child(
		g.wide.Render(cx),
		el.Div().W(el.Dp(180)).Child(g.narrow.Render(cx)),
		el.Div().W(el.Dp(320)).Child(g.width.Render(cx)),
		el.Div().W(el.Dp(float32(g.width.Value()))).Child(g.editor.Render(cx)),
		el.Text(note).TextSize(theme.TextSm).TextColor(theme.Muted),
	)
}
