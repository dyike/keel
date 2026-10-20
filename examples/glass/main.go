// Glass demonstrates a native macOS backdrop beneath a transparent Gio UI.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"

	"github.com/dyike/keel/gio/io/system"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type demo struct {
	selected string
	style    window.GlassStyle
	glass    bool
}

func (d *demo) Render(cx *el.Context) el.Element {
	dark := theme.Current().Bg == theme.Dark().Bg
	background := color.NRGBA{}
	if !d.glass {
		background = theme.Bg
	}
	soft := color.NRGBA{R: 255, G: 255, B: 255, A: 30}
	selected := color.NRGBA{R: 255, G: 255, B: 255, A: 105}
	if dark {
		soft = color.NRGBA{A: 24}
		selected = color.NRGBA{R: 255, G: 255, B: 255, A: 28}
	}
	root := el.Div().Bg(background).Child(
		el.Div().H(el.Dp(64)).Row().Items(el.Center).Pl(112).Pr(24).Gap(16).Child(
			el.Div().Grow().H(el.Full).Row().Items(el.Center).Child(
				el.Text(text("Keel / Glass Studio", "Keel / 玻璃工作室")).TextSize(14).Bold(),
			).Decorate(func(gtx core.C, draw func()) {
				if w := core.CurrentWindow(); w != nil {
					origin, visible := cx.PaintGeometry()
					area := visible.Intersect(image.Rectangle{Min: origin, Max: origin.Add(gtx.Constraints.Max)})
					scale := gtx.Metric.PxPerDp
					if scale <= 0 {
						scale = 1
					}
					w.TitleBarArea(float32(area.Min.X)/scale, float32(area.Min.Y)/scale, float32(area.Dx())/scale, float32(area.Dy())/scale)
				}
				stack := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
				system.ActionInputOp(system.ActionMove).Add(gtx.Ops)
				stack.Pop()
				draw()
			}),
			kit.Button(text("Toggle appearance", "切换明暗"), func() {
				if !dark {
					theme.Apply(theme.Dark())
					_ = window.SetNativeAppearance(window.AppearanceDark)
				} else {
					theme.Apply(theme.Light())
					_ = window.SetNativeAppearance(window.AppearanceLight)
				}
			}).Variant(kit.ButtonGhost).Render(cx),
		),
		el.Div().Row().Grow().H(el.Dp(0)).Items(el.Stretch).Child(
			el.Div().W(el.Dp(236)).P(20).Gap(8).Child(
				el.Text(text("WORKSPACE", "工作区")).TextSize(11).Bold().TextColor(theme.Muted).Mb(12),
				d.nav(text("Overview", "概览"), "01", selected, soft),
				d.nav(text("Projects", "项目"), "02", selected, soft),
				d.nav(text("Favorites", "收藏"), "03", selected, soft),
				el.Div().Grow(),
				el.Div().P(16).Gap(8).Rounded(16).Bg(soft).Child(
					el.Text(text("Let the background through", "让背景透进来")).TextSize(13).Bold(),
					el.Text(text("Move the window to watch the sidebar and header react to what is behind them.", "拖动窗口，观察侧栏与顶部随背后内容变化。")).TextSize(12).TextColor(theme.Muted),
				),
				el.Text(text("KEEL · NATIVE macOS", "KEEL · 原生 macOS")).TextSize(10).TextColor(theme.Muted).Mt(12),
			),
			el.Div().Grow().W(el.Dp(0)).Bg(theme.Bg).Rounded(20).M(12).Ml(0).P(32).Gap(24).ScrollY().Child(
				el.Div().Gap(8).Child(
					el.Text(text("Materials by macOS.", "材质，交给 macOS。")).TextSize(29).Bold(),
					el.Text(text("Layout and interaction in Go.", "布局与交互，继续用 Go。")).TextSize(17).TextColor(theme.Muted),
				),
				el.Div().P(24).Gap(14).Rounded(16).Bg(theme.Surface).Border(1, theme.Border).Child(
					el.Text(d.selected+" / "+d.material()).TextSize(16).Bold(),
					el.Text(text("The sidebar and header use native glass. This card stays opaque to keep its content readable.", "左侧与顶部是原生玻璃背景。这张卡片使用实色，保持内容清晰可读。")).TextSize(14).TextColor(theme.Muted),
					el.Div().Row().Gap(8).Child(
						kit.Button(text("Open Clear", "打开 Clear"), func() { openDemo(window.GlassClear, true) }).Render(cx),
						kit.Button(text("Compare Frosted", "对比毛玻璃"), func() { openDemo(window.GlassFrosted, true) }).Variant(kit.ButtonSecondary).Render(cx),
					),
				),
				el.Div().Row().Gap(16).Child(
					d.tile("01", text("Native materials", "原生材质"), "Liquid Glass / Vibrancy"),
					d.tile("02", text("Go interface", "Go 界面"), text("Text, buttons, and navigation are still drawn by Gio", "文字、按钮、导航仍由 Gio 绘制")),
				),
				el.Div().Gap(10).Child(
					el.Text(text("Try it", "试试看")).TextSize(14).Bold(),
					el.Text(text("Use the sidebar, toggle the appearance, or resize the window.\nPlace it over colorful content to make the glass response more obvious.", "点击左侧导航，切换明暗外观，或调整窗口大小。\n把窗口放在有颜色的内容上，玻璃变化更明显。")).TextSize(13).TextColor(theme.Muted),
				),
			),
		),
	)
	return root
}

func (d *demo) nav(name, number string, selected, hover color.NRGBA) el.Element {
	row := el.Div().Row().Items(el.Center).Gap(14).P(12).Rounded(10).Role("button").Name(name).OnClick(func() { d.selected = name }).Hover(func(s *el.Style) { s.Bg(hover) }).Child(
		el.Text(number).TextSize(11).TextColor(theme.Muted),
		el.Text(name).TextSize(14),
	)
	if d.selected == name {
		row.Bg(selected)
	}
	return row
}

func (d *demo) tile(number, title, detail string) el.Element {
	return el.Div().Grow().W(el.Dp(0)).P(20).Gap(10).Rounded(16).Bg(theme.Surface).Border(1, theme.Border).Child(
		el.Text(number).TextSize(23).TextColor(theme.Primary),
		el.Text(title).TextSize(14).Bold(),
		el.Text(detail).TextSize(12).TextColor(theme.Muted),
	)
}

func (d *demo) material() string {
	if !d.glass || !window.GlassSupported() {
		return text("Opaque background", "实色背景")
	}
	if d.style == window.GlassFrosted || !window.LiquidGlassSupported() {
		return text("Frosted / Vibrancy", "毛玻璃 / Vibrancy")
	}
	if d.style == window.GlassClear {
		return text("Liquid Glass · Clear", "Liquid Glass · 透明")
	}
	return text("Liquid Glass · Regular", "Liquid Glass · 常规")
}

func openDemo(style window.GlassStyle, glass bool) {
	d := &demo{selected: text("Overview", "概览"), style: style, glass: glass}
	o := window.Options{
		Title: text("Keel Glass Studio", "Keel 玻璃工作室"), Width: 980, Height: 680, MinWidth: 760, MinHeight: 580,
		Frameless: true, NativeTrafficLights: true,
		TrafficLightLayout: &window.TrafficLightLayout{Height: 64, Left: 20, Spacing: 22},
		Content:            el.Root(d),
	}
	if glass {
		o.Glass = &window.GlassOptions{Style: style, CornerRadius: 20}
	}
	window.Open(o)
}

func main() {
	language := flag.String("lang", "en", "example language: en or zh-CN")
	styleName := flag.String("style", "regular", "regular, clear, or frosted")
	opaque := flag.Bool("opaque", false, "use an opaque window for comparison")
	dark := flag.Bool("dark", false, "start with a dark appearance")
	backdrop := flag.Bool("backdrop", false, "open a colored reference window behind the glass")
	flag.Parse()
	applyLanguage(*language)
	styles := map[string]window.GlassStyle{"regular": window.GlassRegular, "clear": window.GlassClear, "frosted": window.GlassFrosted}
	style, ok := styles[*styleName]
	if !ok {
		log.Fatal(text("-style must be regular, clear, or frosted", "-style 必须是 regular、clear 或 frosted"))
	}
	if *dark {
		theme.Apply(theme.Dark())
		_ = window.SetNativeAppearance(window.AppearanceDark)
	} else {
		_ = window.SetNativeAppearance(window.AppearanceLight)
	}
	if *backdrop {
		window.Open(window.Options{Title: text("Glass backdrop reference", "玻璃背景参考"), Width: 1100, Height: 800, Content: el.Root(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().Row().Items(el.Stretch).BgGradient(theme.Gradient{From: theme.RGB(0x47bbd9), To: theme.RGB(0xc977b4), Angle: 35}).Child(
				el.Div().Grow().BgGradient(theme.Gradient{From: theme.RGB(0x12989b), To: theme.RGB(0x3869db), Angle: 100}),
				el.Div().Grow().BgGradient(theme.Gradient{From: theme.RGB(0xb989c4), To: theme.RGB(0xf4aa86), Angle: 100}),
			)
		}))})
	}
	openDemo(style, !*opaque)
	window.Main()
}
