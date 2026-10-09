package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"gioui.org/io/key"

	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("headless", "controls", func() core.Widget {
		return el.Embed(&headlessGallery{
			colors: []string{"Red", "Orange", "Olive", "Green", "Teal", "Blue", "Violet"},
			files:  []string{"main.go", "app.go", "theme.go", "README.md", "go.mod", "go.sum"},
		})
	})
}

// headlessGallery draws its own look with el and takes the behavior from
// base: the same keys, typeahead and selection rules as kit's lists.
type headlessGallery struct {
	colors []string
	nav    base.List
	find   base.Typeahead
	at     int

	files []string
	sel   base.Selection[string]
}

func (v *headlessGallery) Render(cx *el.Context) el.Element {
	return el.Div().Gap(20).W(el.Dp(560)).MaxW(el.Full).Child(
		el.Text(demoText("Drawn with el; keyboard navigation, type-ahead, and multiselect rules come from ui/base, matching kit lists.", "外观用 el 自己画，键盘导航、首字母跳转、多选规则来自 ui/base，和 kit 的列表一致。")).TextColor(theme.Muted),
		v.swatches(),
		v.fileList(cx),
	)
}

// swatches is a horizontal single-choice list: ← → move, letters jump.
func (v *headlessGallery) swatches() el.Element {
	v.nav.Count = len(v.colors)
	box := el.Div().ID("swatches").Role("listbox").Name(demoText("Color", "颜色")).Focusable(true).Wrap().Gap(8).P(8).Rounded(theme.RadiusLg).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).Border(1, theme.Border).
		OnKey(func(e el.KeyEvent) bool {
			if t, ok := base.Text(e.Name, e.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt) != 0); ok {
				if e.State == el.KeyPress {
					v.at, _ = v.find.Find(time.Now(), t, v.at, v.nav, func(i int) string { return v.colors[i] })
				}
				return true
			}
			name := e.Name
			switch key.Name(name) {
			case key.NameLeftArrow:
				name = string(key.NameUpArrow)
			case key.NameRightArrow:
				name = string(key.NameDownArrow)
			}
			i, ok := v.nav.Key(name, v.at)
			if ok && e.State == el.KeyPress {
				v.at = i
			}
			return ok
		})
	for i, c := range v.colors {
		hue := swatchColor(c)
		box.Child(el.Div().Role("option").Name(c).Selected(i == v.at).Size(el.Dp(32)).Rounded(theme.RadiusFull).
			Bg(hue).When(i == v.at, func(d *el.DivEl) { d.Border(3, theme.Text) }).
			OnClick(func() { v.at = i }))
	}
	return el.Div().Gap(8).Child(
		el.Text(demoText("Single-select swatches: click, then use ← → or type an initial (for example, o cycles between Orange and Olive).", "单选色板：点一下后用 ← → 或输入首字母（如 o 在 Orange 和 Olive 间循环）")).TextSize(theme.TextSm),
		box,
		el.Text(demoText("Selected: ", "选中：")+v.colors[v.at]).TextSize(theme.TextSm).TextColor(theme.Muted),
	)
}

// fileList is a multi-select list with click, Cmd/Ctrl and Shift rules.
func (v *headlessGallery) fileList(cx *el.Context) el.Element {
	list := el.Div().ID("files").Role("listbox").Name(demoText("Files", "文件")).Gap(2).P(4).Rounded(theme.RadiusLg).Border(1, theme.Border)
	for i, f := range v.files {
		on := v.sel.Has(f)
		row := el.Div().Role("option").Name(f).Selected(on).Row().Px(10).Py(6).Rounded(theme.RadiusMd).
			Hover(func(s *el.Style) { s.Bg(theme.Subtle) }).
			When(on, func(d *el.DivEl) { d.Bg(theme.Highlight).TextColor(theme.PrimaryText) }).
			OnClick(func() {
				mods := cx.ClickModifiers()
				v.sel.Click(v.files, i, mods.Contain(key.ModShift), mods.Contain(key.ModShortcut), nil)
			}).
			Child(el.Text(f).Mono())
		list.Child(row)
	}
	return el.Div().Gap(8).Child(
		el.Text(demoText("Multiselect: click to select, Cmd/Ctrl-click to toggle, Shift-click to select a range", "多选：单击、Cmd/Ctrl 单击切换、Shift 单击选范围")).TextSize(theme.TextSm),
		list,
		el.Text(fmt.Sprintf(demoText("Selected %d: %s", "已选 %d 个：%s"), v.sel.Len(), strings.Join(v.sel.In(v.files), ", "))).TextSize(theme.TextSm).TextColor(theme.Muted),
	)
}

func swatchColor(name string) color.NRGBA {
	return map[string]color.NRGBA{
		"Red": theme.RGB(0xdc2626), "Orange": theme.RGB(0xea580c), "Olive": theme.RGB(0x65a30d), "Green": theme.RGB(0x16a34a),
		"Teal": theme.RGB(0x0d9488), "Blue": theme.RGB(0x2563eb), "Violet": theme.RGB(0x7c3aed),
	}[name]
}
