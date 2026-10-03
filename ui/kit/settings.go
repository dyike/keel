package kit

import (
	"slices"
	"strconv"
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
	// DescriptionContent replaces plain description rendering; Description remains searchable.
	DescriptionContent el.View
	Keywords           []string
	Content            el.View // Replaces the whole row body.
	Vertical, Disabled bool
	// Reset restores application-owned defaults; disabled rows are skipped.
	Reset func()
}

type settingsSection struct {
	title      string
	items      []SettingItem
	groups     []SettingGroup
	suffix     el.View
	resettable bool
}

// SettingGroup collects rows under an optional title and an outside footer.
type SettingGroup struct {
	Title  string
	Items  []SettingItem
	Footer el.View
	// Nil inherits the settings group variant.
	Variant *GroupBoxVariant
}

// SettingPage is one navigation destination. Titles must be unique.
type SettingPage struct {
	Title       string
	Icon        IconName
	Groups      []SettingGroup
	TitleSuffix el.View
	Resettable  bool
}

// SettingsView is a preferences page: a list of sections on the left, the
// chosen section's rows on the right, and a search box that finds rows in
// every section by label or description.
type SettingsView struct {
	sections []settingsSection
	icons    []IconName
	nav      *SidebarView
	query    string
	variant  GroupBoxVariant
	spacing  float32
}

func Settings() *SettingsView {
	v := &SettingsView{nav: Sidebar().Width(180)}
	return v
}

// Section adds a section of rows; its title also names it in the side list.
func (v *SettingsView) Section(title string, icon IconName, items ...SettingItem) *SettingsView {
	v.sections = append(v.sections, settingsSection{title: title, items: cloneSettingItems(items)})
	v.icons = append(v.icons, icon)
	if v.nav.Value() == "" {
		v.nav.SetValue(title)
	}
	return v
}

// Page adds an owned snapshot of a page and its groups. Views and callbacks remain shared.
func (v *SettingsView) Page(page SettingPage) *SettingsView {
	groups := slices.Clone(page.Groups)
	for i := range groups {
		groups[i].Items = cloneSettingItems(groups[i].Items)
		if groups[i].Variant != nil {
			variant := *groups[i].Variant
			groups[i].Variant = &variant
		}
	}
	v.sections = append(v.sections, settingsSection{title: page.Title, groups: groups, suffix: page.TitleSuffix, resettable: page.Resettable})
	v.icons = append(v.icons, page.Icon)
	if v.Value() == "" {
		v.SetValue(page.Title)
	}
	return v
}
func cloneSettingItems(items []SettingItem) []SettingItem {
	items = slices.Clone(items)
	for i := range items {
		items[i].Keywords = slices.Clone(items[i].Keywords)
	}
	return items
}

// GroupVariant sets the default group surface.
func (v *SettingsView) GroupVariant(variant GroupBoxVariant) *SettingsView {
	if variant <= GroupBoxOutline {
		v.variant = variant
	}
	return v
}

// RowSpacing sets row vertical padding in dp; zero restores the theme default.
func (v *SettingsView) RowSpacing(dp float32) *SettingsView {
	if dp >= 0 && dp < 1024 {
		v.spacing = dp
	}
	return v
}

// Query returns the current search text. SetQuery also reconciles page selection.
func (v *SettingsView) Query() string         { return v.query }
func (v *SettingsView) SetQuery(query string) { v.query = query; v.reconcileSearch() }
func settingMatches(it SettingItem, q string) bool {
	return q == "" || strings.Contains(strings.ToLower(it.Label+" "+it.Description+" "+strings.Join(it.Keywords, " ")), q)
}
func (s settingsSection) allGroups() []SettingGroup {
	if s.groups != nil {
		return s.groups
	}
	return []SettingGroup{{Items: s.items}}
}
func (s settingsSection) matches(q string) bool {
	for _, g := range s.allGroups() {
		for _, it := range g.Items {
			if settingMatches(it, q) {
				return true
			}
		}
	}
	return false
}
func (v *SettingsView) reconcileSearch() {
	q := strings.ToLower(strings.TrimSpace(v.query))
	if q == "" {
		return
	}
	first := ""
	for _, s := range v.sections {
		if !s.matches(q) {
			continue
		}
		if s.title == v.Value() {
			return
		}
		if first == "" {
			first = s.title
		}
	}
	if first != "" {
		v.SetValue(first)
	}
}

// ResetPage invokes each enabled row's reset callback, including filtered-out rows.
// Callbacks are captured before invocation so reentrant page additions do not join this reset.
func (v *SettingsView) ResetPage(title string) {
	var callbacks []func()
	for _, s := range v.sections {
		if s.title == title && s.resettable {
			for _, g := range s.allGroups() {
				for _, it := range g.Items {
					if !it.Disabled && it.Reset != nil {
						callbacks = append(callbacks, it.Reset)
					}
				}
			}
			break
		}
	}
	for _, fn := range callbacks {
		fn()
	}
}

// Value is the shown section's title; SetValue shows another.
func (v *SettingsView) Value() string         { return v.nav.Value() }
func (v *SettingsView) SetValue(title string) { v.nav.SetValue(title) }

func (v *SettingsView) row(cx *el.Context, it SettingItem, narrow bool) el.Element {
	if it.Content != nil {
		return el.Div().Role("group").Name(it.Label).Disabled(it.Disabled).Child(it.Content.Render(cx))
	}
	narrow = narrow || it.Vertical
	text := el.Div().Gap(theme.SpaceXxs).Child(el.Text(it.Label))
	if it.DescriptionContent != nil {
		text.Child(it.DescriptionContent.Render(cx))
	} else if it.Description != "" {
		text.Child(el.Text(it.Description).TextSize(theme.TextMd).TextColor(theme.Muted))
	}
	row := el.Div().Role("group").Name(it.Label).Row().Items(el.Center).Gap(theme.SpaceXl).Py(theme.SpaceLg).Child(text)
	row.Disabled(it.Disabled)
	if v.spacing > 0 {
		row.Py(v.spacing)
	}
	if narrow {
		text.WFull()
		row.Col().Items(el.Stretch).Gap(theme.SpaceMd)
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
	v.reconcileSearch()
	q := strings.ToLower(strings.TrimSpace(v.query))
	nav := sidebarSection{}
	for i, s := range v.sections {
		if q != "" && !s.matches(q) {
			continue
		}
		nav.items = append(nav.items, SidebarItem{ID: s.title, Label: s.title, Icon: v.icons[i]})
	}
	v.nav.sections = []sidebarSection{nav}
	id := autoID("settings", v)
	content := el.Div().Grow().W(el.Dp(0)).ScrollY().Px(theme.Space2xl).Py(theme.SpaceXl).Items(el.Stretch)
	search := searchField(cx, id+"/searchbox", id+"/search", el.Input().ID(id+"/search").Name(text.SearchSettings).Placeholder(text.SearchSettings).Bind(&v.query)).Mb(8)
	content.Child(search)
	found := false
	for pi, s := range v.sections {
		if s.title != v.Value() {
			continue
		}
		pageID := id + "/page/" + strconv.Itoa(pi)
		var groups []el.Element
		for gi, g := range s.allGroups() {
			var rows []el.Element
			for ii, it := range g.Items {
				if !settingMatches(it, q) {
					continue
				}
				if len(rows) > 0 {
					rows = append(rows, el.Div().H(el.Dp(1)).Bg(theme.Border))
				}
				rows = append(rows, el.Div().ID(pageID+"/group/"+strconv.Itoa(gi)+"/row/"+strconv.Itoa(ii)).Child(v.row(cx, it, narrow)))
			}
			if len(rows) == 0 {
				continue
			}
			variant := v.variant
			if g.Variant != nil {
				variant = *g.Variant
			}
			group := GroupBox(g.Title).Variant(variant).Footer(g.Footer).Child(el.ViewFunc(func(*el.Context) el.Element { return el.Div().Items(el.Stretch).Children(rows) }))
			groups = append(groups, el.Div().ID(pageID+"/group/"+strconv.Itoa(gi)).Items(el.Stretch).Child(group.Render(cx)))
		}
		if len(groups) == 0 {
			continue
		}
		found = true
		header := el.Div().Row().Items(el.Center).Gap(theme.SpaceMd).Child(el.Text(s.title).TextSize(theme.TextLg).Bold())
		if s.suffix != nil {
			header.Child(s.suffix.Render(cx))
		}
		if s.resettable {
			header.Child(Button(text.ResetSettings, func() { v.ResetPage(s.title) }).Variant(ButtonGhost).Render(cx))
		}
		content.Child(el.Div().ID(pageID).Role("group").Name(s.title).Items(el.Stretch).Gap(theme.SpaceLg).Child(header).Children(groups))
	}
	if !found {
		content.Child(el.Div().Py(theme.Space2xl).Child(el.Text(text.NoMatches).TextColor(theme.Muted)))
	}
	if narrow {
		links := el.Div().Role("navigation").Wrap().Gap(theme.SpaceXs).P(theme.SpaceLg)
		for i, s := range v.sections {
			if q != "" && !s.matches(q) {
				continue
			}
			title := s.title
			variant := ButtonGhost
			if title == v.Value() {
				variant = ButtonSecondary
			}
			links.Child(Button(title, func() { v.SetValue(title) }).Icon(v.icons[i]).Variant(variant).Render(cx))
		}
		return el.Div().Grow().Items(el.Stretch).Child(links, content.WFull().H(el.Dp(0)))
	}
	return el.Div().Row().Items(el.Stretch).Grow().Child(v.nav.Render(cx), content)
}
