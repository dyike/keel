package kit

import (
	"github.com/dyike/keel/third_party/gio/layout"
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
