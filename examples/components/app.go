package main

import (
	"sort"
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// foundations are the sections that demonstrate el itself, not a component.
var foundations = map[string]bool{"theme": true, "layout": true, "scrollable": true, "focus": true, "time": true, "overlay": true}

var categoryTitles = []struct{ id, title string }{
	{"foundations", "基础能力"},
	{"controls", "基础组件"},
	{"inputs", "输入"},
	{"overlays", "浮层"},
	{"data", "数据"},
	{"shell", "应用外壳"},
}

// gallery is the components app: a searchable sidebar of every section and
// the selected one beside it. Each section is built the first time it is
// shown and kept, so switching away and back preserves its state.
type gallery struct {
	nav      *kit.SidebarView
	search   *kit.InputView
	built    map[string]core.Widget
	sections map[string]demoSection
	dark     bool
}

func newGallery() *gallery {
	g := &gallery{built: map[string]core.Widget{}, sections: map[string]demoSection{}, dark: theme.Current().Bg == theme.Dark().Bg}
	byCategory := map[string][]kit.SidebarItem{}
	for _, s := range demoSections {
		g.sections[s.name] = s
		category := s.category
		if foundations[s.name] {
			category = "foundations"
		}
		byCategory[category] = append(byCategory[category], kit.SidebarItem{ID: s.name, Label: displayName(s.name)})
	}
	g.search = kit.Input("").Placeholder("搜索组件").OnChange(func(q string) { g.nav.Filter(q) }).
		Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
			return kit.Icon(kit.IconSearch).Size(16).Color(theme.Muted).Render(cx)
		}))
	g.nav = kit.Sidebar().Width(232).Header(el.ViewFunc(g.header))
	for _, c := range categoryTitles {
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
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func (g *gallery) header(cx *el.Context) el.Element {
	return el.Div().Gap(12).Items(el.Stretch).Child(
		el.Div().Row().Gap(8).Items(el.Center).Px(4).Child(
			el.Div().Size(el.Dp(24)).Rounded(6).Bg(theme.Primary).Center().TextColor(theme.OnColor).TextSize(13).Bold().Child(el.Text("K")),
			el.Text("Keel 组件").Bold(),
		),
		g.search.Render(cx),
	)
}

func (g *gallery) Render(cx *el.Context) el.Element {
	name := g.nav.Value()
	s, ok := g.sections[name]
	if !ok {
		return el.Div().Row().Grow().Child(g.nav.Render(cx))
	}
	w := g.built[name]
	if w == nil {
		w = s.build()
		g.built[name] = w
	}
	category := s.category
	if foundations[name] {
		category = "foundations"
	}
	for _, c := range categoryTitles {
		if c.id == category {
			category = c.title
		}
	}
	doc := "docs/kit/" + name + ".md"
	if foundations[name] {
		doc = "docs/el.md"
	}
	themeLabel, langLabel := "深色", "English"
	if g.dark {
		themeLabel = "浅色"
	}
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
		kit.Button(themeLabel, func() {
			g.dark = !g.dark
			if g.dark {
				theme.Apply(theme.Dark())
			} else {
				theme.Apply(theme.Light())
			}
		}).Variant(kit.ButtonSecondary).Size(28).Render(cx),
	)
	// A section that fills its window (an el.Root) gets the whole pane and
	// scrolls itself; one sized to its content scrolls inside the pane.
	var body el.Element
	if f, ok := w.(interface{ FillsWindow() bool }); ok && f.FillsWindow() {
		body = el.Div().ID("page/" + name).Grow().MinH(el.Dp(0)).Items(el.Stretch).Child(el.Widget(w).Grow())
	} else {
		body = el.Div().ID("page/" + name).Grow().MinH(el.Dp(0)).ScrollY().P(8).Items(el.Start).Child(el.Widget(w))
	}
	return el.Div().Row().Grow().Items(el.Stretch).Bg(theme.Bg).Child(
		g.nav.Render(cx),
		el.Div().W(el.Dp(1)).NoShrink().Bg(theme.Border),
		el.Div().Grow().W(el.Dp(0)).Items(el.Stretch).Child(
			head,
			el.Div().H(el.Dp(1)).NoShrink().Bg(theme.Border),
			body,
		),
	)
}
