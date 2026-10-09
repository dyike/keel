package kit

import (
	"fmt"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestSettingsGroupNavigationScrollAndFiltering(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, width := range []int{400, 800} {
			t.Run(fmt.Sprintf("%d/%d", scale, width), func(t *testing.T) {
				groups := []SettingGroup{}
				for i := 0; i < 8; i++ {
					groups = append(groups, SettingGroup{Title: fmt.Sprintf("group%d", i), Items: []SettingItem{{Label: fmt.Sprintf("row%d", i), Description: "description"}}})
				}
				v := Settings().GroupNavigation(true).Page(SettingPage{Title: "page", Groups: groups})
				var cx *el.Context
				root := el.Root(viewFunc(func(c *el.Context) el.Element { cx = c; return v.Render(c) }))
				h := uitest.NewFunc(func(gtx core.C) {
					gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					gtx.Constraints = layout.Exact(image.Pt(width*scale, 400*scale))
					root.Layout(gtx)
				})
				if !v.ShowGroup("page", "group7") {
					t.Fatal("target missing")
				}
				for range 3 {
					h.Frame()
				}
				offset, viewport, content := cx.ScrollState(autoID("settings", v) + "/content")
				if offset <= 0 || viewport >= content || bounds(h, "row7").Min.Y >= 400*scale {
					t.Fatal("offscreen target not revealed", offset, viewport, content, bounds(h, "row7"))
				}
				if v.Value() != "page" {
					t.Fatal("group leaked into Value")
				}
				if !v.ShowGroup("page", "group0") {
					t.Fatal("first target")
				}
				for range 3 {
					h.Frame()
				}
				first, _, _ := cx.ScrollState(autoID("settings", v) + "/content")
				if first >= offset {
					t.Fatal("reverse reveal", first, offset)
				}
				click(t, h, "group1")
				for range 3 {
					h.Frame()
				}
				if v.Value() != "page" {
					t.Fatal("navigation exposed internal group key", v.Value())
				}
				v.SetQuery("row0")
				h.Frame()
				if v.ShowGroup("page", "group7") || v.ShowGroup("missing", "group0") {
					t.Fatal("invalid navigation accepted")
				}
				v.GroupNavigation(false)
				h.Frame()
				if v.Value() != "page" || v.pendingGroup != nil {
					t.Fatal("disable navigation changed page")
				}
			})
		}
	}
	v := Settings().Page(SettingPage{Title: "page", Groups: []SettingGroup{{Title: "same", Items: []SettingItem{{Label: "a"}}}, {Title: "same", Items: []SettingItem{{Label: "b"}}}}})
	if v.ShowGroup("page", "same") {
		t.Fatal("ambiguous group accepted")
	}
}

func TestSettingsSizePreservesControl(t *testing.T) {
	field := Input("value")
	field.SetValue("saved")
	v := Settings().Section("page", IconNone, SettingItem{Label: "value", Control: field})
	h := settingsHarness(v)
	heights := []int{}
	for _, size := range []SettingsSize{SettingsSizeXSmall, SettingsSizeSmall, SettingsSizeMedium, SettingsSizeLarge} {
		v.Size(size)
		h.Frame()
		heights = append(heights, bounds(h, "value").Dy())
		if field.Value() != "saved" {
			t.Fatal("size reset control")
		}
	}
	for i := 1; i < len(heights); i++ {
		if heights[i] <= heights[i-1] {
			t.Fatal("row density", heights)
		}
	}
}
