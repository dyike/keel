package kit

import (
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/locale"
	"image"
	"testing"
)

func TestSettingsPageSearchResetAndSnapshots(t *testing.T) {
	resets, disabled := 0, 0
	keywords := []string{"needle"}
	variant := GroupBoxOutline
	groups := []SettingGroup{
		{Title: "first group", Variant: &variant, Footer: el.ViewFunc(func(*el.Context) el.Element { return el.Text("footer marker") }), Items: []SettingItem{{Label: "first row", Keywords: keywords, Reset: func() { resets++ }}}},
		{Title: "second group", Items: []SettingItem{{Label: "second row", Reset: func() { resets++ }}, {Label: "disabled row", Disabled: true, Reset: func() { disabled++ }}}},
	}
	v := Settings().Section("other", IconUser, SettingItem{Label: "other row"}).Page(SettingPage{Title: "page", Groups: groups, Resettable: true})
	keywords[0] = "mutated"
	groups[0].Items[0].Label = "mutated"
	variant = GroupBoxNormal
	v.SetQuery(" NEEDLE ")
	if v.Value() != "page" || v.sections[1].groups[0].Variant == &variant || *v.sections[1].groups[0].Variant != GroupBoxOutline {
		t.Fatal("snapshot/search", v.Value())
	}
	h := settingsHarness(v)
	if !shown(h, "first row") || !shown(h, "footer marker") || shown(h, "second row") {
		t.Fatal("filtered groups/footer", shown(h, "first row"), shown(h, "footer marker"), shown(h, "second row"))
	}
	click(t, h, "重置此页")
	h.Frame()
	if resets != 2 || disabled != 0 {
		t.Fatal("page-wide reset", resets, disabled)
	}
	v.SetQuery("missing")
	h.Frame()
	if v.Value() != "page" || shown(h, "footer marker") || !shown(h, locale.Current().NoMatches) {
		t.Fatal("empty search changed selection or retained content")
	}

	v.SetQuery("")
	h.Frame()
	if v.Value() != "page" || !shown(h, "second row") {
		t.Fatal("clear search")
	}
}

func TestSettingsCustomAndDisabledRows(t *testing.T) {
	calls := 0
	v := Settings().Page(SettingPage{Title: "page", Groups: []SettingGroup{{Items: []SettingItem{
		{Label: "custom", Keywords: []string{"find"}, Disabled: true, Content: Button("blocked action", func() { calls++ })},
		{Label: "rich", Description: "searchable description", DescriptionContent: el.ViewFunc(func(*el.Context) el.Element { return el.Text("rich description") }), Vertical: true, Control: Input("")},
	}}}})
	h := settingsHarness(v)
	click(t, h, "blocked action")
	h.Frame()
	if calls != 0 {
		t.Fatal("disabled custom row invoked callback")
	}
	v.SetQuery("description")
	h.Frame()
	if !shown(h, "rich description") || shown(h, "blocked action") {
		t.Fatal("description search")
	}
	v.SetQuery("find")
	h.Frame()
	if !shown(h, "blocked action") || shown(h, "rich description") {
		t.Fatal("custom keyword search")
	}
}

func settingsHarness(v *SettingsView) *uitest.Harness {
	root := el.Root(v)
	return uitest.NewFunc(func(gtx core.C) { gtx.Constraints = layout.Exact(image.Pt(800, 1000)); root.Layout(gtx) })
}

func TestSettingsSidebarSearchFiltersPages(t *testing.T) {
	v := Settings().SearchInSidebar(true).Section("General", IconSettings, SettingItem{Label: "Theme"}).Section("Connection", IconUser, SettingItem{Label: "Remote address"})
	h := settingsHarness(v)
	if !shown(h, locale.Current().SearchSettings) || !shown(h, "Theme") {
		t.Fatal("missing initial search or page")
	}
	v.SetQuery("Remote")
	h.Frame()
	if v.Value() != "Connection" || !shown(h, "Remote address") || shown(h, "Theme") {
		t.Fatal("sidebar search failed to select matching page")
	}
	v.SetQuery("")
	h.Frame()
	if !shown(h, "General") {
		t.Fatal("clearing search did not restore navigation")
	}
}

func TestSettingsNavigationCategories(t *testing.T) {
	v := Settings().SearchInSidebar(true).Page(SettingPage{Title: "General", Category: "Settings", Groups: []SettingGroup{{Items: []SettingItem{{Label: "Theme"}}}}}).Page(SettingPage{Title: "Phone", Category: "Remote", Groups: []SettingGroup{{Items: []SettingItem{{Label: "Pair device"}}}}})
	h := settingsHarness(v)
	if len(v.nav.sections) != 2 || v.nav.sections[0].title != "Settings" || v.nav.sections[1].title != "Remote" {
		t.Fatal("missing navigation categories")
	}
	v.SetQuery("Pair")
	h.Frame()
	if len(v.nav.sections) != 1 || v.nav.sections[0].title != "Remote" {
		t.Fatal("empty category retained after filtering")
	}
}

func TestSettingsHorizontalNavigationPreservesDraftAndQuery(t *testing.T) {
	value := "draft"
	v := Settings().SearchInSidebar(true).AvailableWidth(700).
		Section("General", IconSettings, SettingItem{Label: "Name", Control: el.ViewFunc(func(*el.Context) el.Element { return el.Input().Bind(&value) })}).
		Section("Connection", IconUser, SettingItem{Label: "Remote address"})
	h := settingsHarness(v)
	v.SetValue("Connection")
	v.SetQuery("Remote")
	h.Frame()
	for _, width := range []float32{1100, 700, 460, 1100} {
		v.AvailableWidth(width)
		h.Frame()
		if v.Value() != "Connection" || v.Query() != "Remote" || value != "draft" || !shown(h, "Remote address") {
			t.Fatalf("layout change reset state at %v", width)
		}
	}
	v.SetQuery("")
	v.AvailableWidth(700)
	h.Frame()
	click(t, h, "General")
	h.Frame()
	if v.Value() != "General" || !shown(h, "Name") {
		t.Fatal("horizontal tab did not select page")
	}
}
