package kit

import (
	"slices"
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// SettingItem is one row of a settings page: a label, an optional
// description, and the control that changes it.
type SettingItem struct {
	Label, Description string
	Control            el.View
}

type settingsSection struct {
	title string
	items []SettingItem
}

// SettingsView is a preferences page: a list of sections on the left, the
// chosen section's rows on the right, and a search box that finds rows in
// every section by label or description.
type SettingsView struct {
	sections []settingsSection
	icons    []IconName
	nav      *SidebarView
	query    string
}

func Settings() *SettingsView {
	v := &SettingsView{nav: Sidebar().Width(180)}
	return v
}

// Section adds a section of rows; its title also names it in the side list.
func (v *SettingsView) Section(title string, icon IconName, items ...SettingItem) *SettingsView {
	v.sections = append(v.sections, settingsSection{title, slices.Clone(items)})
	v.icons = append(v.icons, icon)
	if v.nav.Value() == "" {
		v.nav.SetValue(title)
	}
	return v
}

// Value is the shown section's title; SetValue shows another.
func (v *SettingsView) Value() string         { return v.nav.Value() }
func (v *SettingsView) SetValue(title string) { v.nav.SetValue(title) }

func (v *SettingsView) row(cx *el.Context, it SettingItem, narrow bool) el.Element {
	text := el.Div().Gap(2).Child(el.Text(it.Label))
	if it.Description != "" {
		text.Child(el.Text(it.Description).TextSize(13).TextColor(theme.Muted))
	}
	row := el.Div().Role("group").Name(it.Label).Row().Items(el.Center).Gap(16).Py(12).Child(text)
	if narrow {
		text.WFull()
		row.Col().Items(el.Stretch).Gap(8)
	} else {
		text.Grow().W(el.Dp(0))
	}
	if it.Control != nil {
		if n, ok := it.Control.(interface{ setName(string) }); ok {
			n.setName(it.Label) // the row label names an unlabelled control
		}
		// A fixed column: fields fill it, switches and checkboxes sit at its end.
		control := el.Div().NoShrink().W(el.Dp(240)).MaxW(el.Full).Items(el.End).Child(it.Control.Render(cx))
		if narrow {
			control.Items(el.Start)
		}
		row.Child(control)
	}
	return row
}

func (v *SettingsView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	width, _ := cx.ViewportSize()
	narrow := width < 600
	nav := sidebarSection{}
	for i, s := range v.sections {
		nav.items = append(nav.items, SidebarItem{ID: s.title, Label: s.title, Icon: v.icons[i]})
	}
	v.nav.sections = []sidebarSection{nav}
	q := strings.ToLower(strings.TrimSpace(v.query))
	content := el.Div().Grow().W(el.Dp(0)).ScrollY().Px(24).Py(16).Items(el.Stretch)
	search := searchField(cx, autoID("settings", v)+"/searchbox", autoID("settings", v)+"/search", el.Input().ID(autoID("settings", v)+"/search").Name(text.SearchSettings).Placeholder(text.SearchSettings).Bind(&v.query)).Mb(8)
	content.Child(search)
	found := false
	for _, s := range v.sections {
		if q == "" && s.title != v.nav.Value() {
			continue
		}
		var rows []el.Element
		for i, it := range s.items {
			if q != "" && !strings.Contains(strings.ToLower(it.Label+" "+it.Description), q) {
				continue
			}
			if len(rows) > 0 || i > 0 && q == "" {
				rows = append(rows, el.Div().H(el.Dp(1)).Bg(theme.Border))
			}
			rows = append(rows, v.row(cx, it, narrow))
		}
		if len(rows) == 0 {
			continue
		}
		found = true
		content.Child(el.Div().Pt(8).Pb(4).Child(el.Text(s.title).TextSize(17).Bold()))
		content.Child(el.Div().Role("group").Name(s.title).Items(el.Stretch).Children(rows))
	}
	if !found {
		content.Child(el.Div().Py(24).Child(el.Text(text.NoMatches).TextColor(theme.Muted)))
	}
	if narrow {
		links := el.Div().Role("navigation").Wrap().Gap(4).P(12)
		for i, s := range v.sections {
			title := s.title
			variant := ButtonGhost
			if title == v.Value() {
				variant = ButtonSecondary
			}
			links.Child(Button(title, func() { v.SetValue(title); v.query = "" }).Icon(v.icons[i]).Variant(variant).Render(cx))
		}
		return el.Div().Grow().Items(el.Stretch).Child(links, content.WFull().H(el.Dp(0)))
	}
	return el.Div().Row().Items(el.Stretch).Grow().Child(v.nav.Render(cx), content)
}
