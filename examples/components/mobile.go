package main

import (
	"net/url"
	"sort"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

func (g *gallery) open(name string) {
	if _, ok := g.sections[name]; !ok {
		return
	}
	if section := g.sections[name]; section.build == nil {
		switchGalleryBundle(name, section.category)
		return
	}
	g.nav.SetValue(name)
	g.mobileDetail = true
	if g.onNavigate != nil {
		g.onNavigate(name)
	}
}
func (g *gallery) back() {
	if g.onBack != nil {
		g.onBack()
		return
	}
	g.mobileDetail = false
}
func (g *gallery) language(cx *el.Context) el.Element {
	label := "中文"
	if strings.HasPrefix(locale.Current().Lang, "zh") {
		label = "English"
	}
	return kit.Button(label, func() {
		if strings.HasPrefix(locale.Current().Lang, "zh") {
			locale.Apply(locale.English())
		} else {
			locale.Apply(locale.Chinese())
		}
	}).Variant(kit.ButtonGhost).Size(44).Render(cx)
}
func (g *gallery) mobileIndex(cx *el.Context) el.Element {
	g.search.Size(kit.InputSizeLarge)
	g.themes.Size(44)
	query := strings.ToLower(strings.TrimSpace(g.search.Value()))
	list := el.Div().ID("gallery/mobile-index").Grow().MinH(el.Dp(0)).ScrollY().P(16).Gap(20)
	count := 0
	for _, category := range categoryTitles() {
		var names []string
		for name, section := range g.sections {
			if section.category == category.id && (query == "" || strings.Contains(strings.ToLower(displayName(name)+" "+name+" "+category.title), query)) {
				names = append(names, name)
			}
		}
		if len(names) == 0 {
			continue
		}
		sort.Slice(names, func(i, j int) bool { return displayName(names[i]) < displayName(names[j]) })
		group := el.Div().Gap(8).Child(el.Text(category.title).TextSize(18).Bold().TextColor(theme.Text))
		rows := el.Div().Gap(4)
		for _, name := range names {
			rows.Child(el.Div().ID("gallery/open/"+name).Role("button").Name(displayName(name)).Focusable(true).
				WFull().MinH(el.Dp(52)).Px(12).Py(12).Rounded(10).Bg(theme.Surface).Row().Gap(12).
				Hover(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
				FocusStyle(func(s *el.Style) { s.Bg(theme.SubtleHover) }).
				OnClick(func() { g.open(name) }).Child(
				el.Div().Grow().W(el.Dp(0)).Child(el.Text(displayName(name)).TextSize(16)),
				kit.Icon(kit.IconChevronRight).Size(18).Color(theme.Muted).Render(cx),
			))
			count++
		}
		list.Child(group.Child(rows))
	}
	if count == 0 {
		list.Child(el.Text(demoText("No components found", "没有找到组件")).TextColor(theme.Muted))
	}
	if g.categoryBundle {
		list.Child(el.Div().Gap(8).Child(
			el.Text(demoText("Keep demo state when switching categories with the full gallery.", "使用完整组件库，可保留跨分类操作状态。")).TextSize(12).TextColor(theme.Muted),
			kit.Button(demoText("Load full gallery", "加载完整组件库"), func() { switchGalleryBundle(g.nav.Value(), "all") }).Variant(kit.ButtonSecondary).Size(44).Render(cx),
		))
	}
	return el.Div().Items(el.Stretch).Child(
		el.Div().NoShrink().P(16).Gap(12).Child(
			el.Text(demoText("Keel components", "Keel 组件")).TextSize(24).Bold(),
			el.Div().Row().Gap(8).Child(g.language(cx), el.Div().Grow().W(el.Dp(0)).Child(g.themes.Render(cx))),
			g.search.Render(cx),
		), list,
	)
}
func (g *gallery) mobilePage(cx *el.Context, name, category string, body el.Element) el.Element {
	return el.Div().Items(el.Stretch).Child(
		el.Div().NoShrink().P(12).Gap(8).Child(
			el.Div().Row().Gap(8).Child(
				kit.Button("", g.back).Icon(kit.IconChevronLeft).Name(demoText("Back to components", "返回组件列表")).Variant(kit.ButtonGhost).Size(44).Render(cx),
				el.Div().Grow().W(el.Dp(0)).Gap(2).Child(el.Text(displayName(name)).TextSize(20).Bold(), el.Text(category).TextSize(12).TextColor(theme.Muted)),
			),
		),
		el.Div().H(el.Dp(1)).NoShrink().Bg(theme.Border), body,
	)
}

// Preserve user preferences and unrelated host parameters across runtime loads.
func galleryBundleURL(path, search, name, category, language, palette string) string {
	q, _ := url.ParseQuery(strings.TrimPrefix(search, "?"))
	q.Set("bundle", category)
	q.Set("lang", language)
	q.Set("theme", palette)
	q.Del("section")
	return path + "?" + q.Encode() + "#component=" + url.QueryEscape(name)
}
