package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
	"github.com/dyike/keel/ui/window"
)

func gallery() core.Widget {
	volume := widget.Slider("音量", 0, 100).Step(5)
	volume.SetValue(40)
	feedback := widget.Muted("当前音量：40")
	volume.OnChange(func(v float32) { feedback.SetText(fmt.Sprintf("当前音量：%g", v)) })
	temperature := widget.Slider("温度（°C）", -10, 40).Step(.5)
	temperature.SetValue(22.5)
	fixed := widget.Slider("只读预览", 0, 100)
	fixed.SetValue(65)
	fixed.SetDisabled(true)
	name := widget.Input("显示名称").Hint("收起后保留输入内容")
	sections := widget.Accordion().Add("个人资料", name).
		Add("通知设置", widget.Switch("接收通知", true)).
		Add("暂不可用", widget.Text("禁用项目"))
	sections.SetOpen(0, true)
	sections.SetItemDisabled(2, true)
	faq := widget.Accordion().Multiple().
		Add("如何用键盘操作？", widget.Text("Tab 聚焦标题，↑ ↓ 切换标题，Enter 或空格展开，Home / End 跳到首尾。")).
		Add("可以同时展开吗？", widget.Text("这个面板启用了 Multiple，可以同时展开两项。"))
	img := image.NewNRGBA(image.Rect(0, 0, 240, 72))
	colors := []color.NRGBA{{R: 224, G: 96, B: 86, A: 255}, {R: 89, G: 171, B: 137, A: 255}, {R: 87, G: 139, B: 209, A: 255}}
	for y := 0; y < 72; y++ {
		for x := 0; x < 240; x++ {
			img.SetNRGBA(x, y, colors[x/80])
		}
	}
	photo := &widget.ImageView{Asset: widget.ImageData(img), Alt: "红绿蓝色块"}
	status := widget.Muted("点击图片查看反馈")
	photo.OnClick = func() { status.SetText("已点击图片") }
	return layout.Column(
		widget.Heading("组件交互示例"), widget.Muted("拖动滑块、切换折叠面板，也可以用 Tab 和方向键操作。"),
		layout.Card(widget.Heading("滑块"), volume, feedback, temperature, fixed, widget.Checkbox("禁用音量调整", false).OnChange(volume.SetDisabled)),
		layout.Card(widget.Heading("折叠面板"), sections),
		layout.Card(widget.Heading("多项展开"), faq),
		layout.Card(widget.Heading("图片"), photo, status),
	)
}
func main() {
	screenshot := flag.String("screenshot", "", "render a PNG and exit")
	flag.Parse()
	content := gallery()
	if *screenshot != "" {
		if err := window.Screenshot(content, 680, 1040, *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(window.Options{Title: "Keel · 组件", Width: 680, Height: 860, Content: content})
	window.Main()
}
