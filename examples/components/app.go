package main

import (
	"sort"
	"strings"

	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// categoryTitles orders the sidebar: components grouped by purpose, then
// motion, then the el capabilities they are built on.
func categoryTitles() []struct{ id, title string } {
	return []struct{ id, title string }{
		{"general", demoText("General", "通用")},
		{"inputs", demoText("Form & input", "表单输入")},
		{"display", demoText("Data display", "数据展示")},
		{"data", demoText("Lists & tables", "列表与表格")},
		{"charts", demoText("Charts", "图表")},
		{"chat", demoText("Chat & messages", "消息与对话")},
		{"feedback", demoText("Feedback", "反馈")},
		{"navigation", demoText("Navigation", "导航")},
		{"overlays", demoText("Overlay", "浮层")},
		{"shell", demoText("Layout & shell", "布局与外壳")},
		{"motion", demoText("Motion", "动效")},
		{"foundations", demoText("Core capabilities", "基础能力")},
	}
}

// sectionDoc is the document a section demonstrates.
func sectionDoc(s demoSection) string {
	switch {
	case s.name == "attachment_vertical":
		return "docs/kit/attachment.md"
	case s.name == "command_async":
		return "docs/kit/command.md"
	case s.name == "table_data", s.name == "table_static":
		return "docs/kit/table.md"
	case s.name == "text_view":
		return "docs/el.md"
	case s.name == "headless":
		return "docs/base.md"
	case s.category == "foundations", s.name == "motion":
		return "docs/el.md"
	case strings.HasPrefix(s.name, "motion_"):
		return "ui/motion/README.md"
	}
	return "docs/kit/" + s.name + ".md"
}

// gallery is the components app: a searchable sidebar of every section and
// the selected one beside it. Each section is built the first time it is
// shown and kept, so switching away and back preserves its state.
type gallery struct {
	nav              *kit.SidebarView
	search           *kit.InputView
	built            map[string]core.Widget
	sections         map[string]demoSection
	themes           *kit.SelectView
	navWidth         float32 // a width the user dragged the sidebar to; 0 keeps the automatic width
	languageRevision uint64
}

func newGallery() *gallery {
	g := &gallery{languageRevision: locale.Revision(), built: map[string]core.Widget{}, sections: map[string]demoSection{}}
	g.themes = kit.Select("", theme.Names()...).OnChange(func(name string) { theme.Use(name) })
	g.themes.SetValue(theme.CurrentName())
	byCategory := map[string][]kit.SidebarItem{}
	for _, s := range demoSections {
		g.sections[s.name] = s
		byCategory[s.category] = append(byCategory[s.category], kit.SidebarItem{ID: s.name, Label: displayName(s.name)})
	}
	g.search = kit.Input("").Placeholder(demoText("Search components", "搜索组件")).OnChange(func(q string) { g.nav.Filter(q) }).
		Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Icon(kit.IconSearch).Size(16).Color(theme.Muted).Render(cx)
		}))
	g.nav = kit.Sidebar().Width(232).AutoWidth(true).Resizable(180, 420).OnResize(func(dp float32) { g.navWidth = dp }).
		SectionTitleWeight(font.SemiBold).SectionTitleSize(16).Header(el.ViewFunc(g.header))
	for _, c := range categoryTitles() {
		items := byCategory[c.id]
		sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
		g.nav.Section(c.title, items...)
	}
	g.nav.SetValue("button")
	return g
}

// displayName turns a section name into its component name: color_picker
// becomes ColorPicker.
func displayName(name string) string {
	switch name {
	case "motion":
		return demoText("Element transitions", "元素过渡")
	case "motion_values":
		return demoText("Springs & keyframes", "弹簧与关键帧")
	case "motion_gestures":
		return demoText("Hover, press & drag", "悬停、按压与拖拽")
	case "motion_layout":
		return demoText("Presence & layout", "进退场与布局")
	case "motion_scroll":
		return demoText("Scroll-linked motion", "滚动联动")
	case "motion_presets":
		return demoText("Spring presets", "弹簧预设")
	case "motion_pages":
		return demoText("Page transitions", "页面切换")
	case "motion_pose":
		return demoText("Transforms & anchors", "变换与锚定浮层")
	case "motion_timeline":
		return demoText("Timeline", "时间线")
	case "motion_effects":
		return demoText("Cursor glow & confetti", "指针光晕与彩纸")
	}
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func (g *gallery) header(cx *el.Context) el.Element {
	logo := el.Div().NoShrink().Size(el.Dp(24)).Rounded(6).Bg(theme.Primary).Center().TextColor(theme.OnColor).TextSize(13).Bold().Child(el.Text("K"))
	if g.nav.Collapsed() {
		return el.Div().Gap(12).Items(el.Center).Child(
			logo,
			kit.Button("", func() { g.nav.SetCollapsed(false) }).Name(demoText("Search components", "搜索组件")).Icon(kit.IconSearch).Variant(kit.ButtonGhost).Size(32).Render(cx),
		)
	}
	return el.Div().Gap(12).Items(el.Stretch).Child(
		el.Div().Row().Gap(8).Items(el.Center).Px(4).Child(logo, el.Text(demoText("Keel components", "Keel 组件")).Bold()),
		g.search.Render(cx),
	)
}

func (g *gallery) Render(cx *el.Context) el.Element {
	if g.languageRevision != locale.Revision() {
		selected, query := g.nav.Value(), g.search.Value()
		next := newGallery()
		if g.navWidth > 0 {
			next.navWidth = g.navWidth
			next.nav.Width(g.navWidth)
		}
		next.nav.SetValue(selected)
		next.search.SetValue(query)
		next.nav.Filter(query)
		*g = *next
	}
	// Section headings use body text so they stand out from the items;
	// set every frame so they follow theme changes.
	g.nav.SectionTitleColor(theme.Text)
	name := g.nav.Value()
	s, ok := g.sections[name]
	if !ok {
		return el.Div().Row().Items(el.Stretch).Grow().Child(g.nav.Render(cx))
	}
	w := g.built[name]
	if w == nil {
		w = s.build()
		g.built[name] = w
	}
	category := s.category
	for _, c := range categoryTitles() {
		if c.id == category {
			category = c.title
		}
	}
	doc := sectionDoc(s)
	langLabel := "English"
	if !strings.HasPrefix(locale.Current().Lang, "zh") {
		langLabel = "中文"
	}
	head := el.Div().Row().Items(el.Center).Gap(12).Px(24).Py(14).NoShrink().Child(
		el.Div().Grow().W(el.Dp(0)).Gap(2).Child(
			el.Text(displayName(name)).TextSize(20).Bold(),
			el.Text(category+" · "+doc).TextSize(12).TextColor(theme.Muted),
		),
		kit.Button(langLabel, func() {
			if strings.HasPrefix(locale.Current().Lang, "zh") {
				locale.Apply(locale.English())
			} else {
				locale.Apply(locale.Chinese())
			}
		}).Variant(kit.ButtonGhost).Size(28).Render(cx),
		el.Div().W(el.Dp(140)).Child(g.themes.Render(cx)),
	)
	// A section that fills its window (an el.Root) gets the whole pane and
	// scrolls itself; one sized to its content scrolls inside the pane.
	var body el.Element
	if f, ok := w.(interface{ FillsWindow() bool }); ok && f.FillsWindow() {
		body = el.Div().ID("page/" + name).Grow().MinH(el.Dp(0)).Items(el.Stretch).Child(el.Widget(w).Grow())
	} else {
		body = el.Div().ID("page/" + name).Grow().MinH(el.Dp(0)).ScrollY().P(8).Items(el.Start).Child(el.Widget(w))
	}
	return el.Div().Row().Grow().Items(el.Stretch).Child(
		g.nav.Render(cx),
		el.Div().Grow().W(el.Dp(0)).Items(el.Stretch).Child(
			head,
			el.Div().H(el.Dp(1)).NoShrink().Bg(theme.Border),
			body,
		),
	)
}
