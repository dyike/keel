package kit

import (
	"strconv"
	"strings"

	"github.com/dyike/keel/ui/theme"
)

// SettingsSize controls settings labels, row padding, gaps and control columns.
// Application-provided controls retain their own size configuration.
type SettingsSize uint8

const (
	SettingsSizeMedium SettingsSize = iota
	SettingsSizeXSmall
	SettingsSizeSmall
	SettingsSizeLarge
)

func (v *SettingsView) Size(size SettingsSize) *SettingsView {
	if size <= SettingsSizeLarge {
		v.size = size
	}
	return v
}
func (v *SettingsView) GroupNavigation(on bool) *SettingsView {
	page := v.Value()
	v.groupNavigation = on
	if !on {
		v.SetValue(page)
	}
	return v
}

type settingsGroupTarget struct {
	page  string
	group int
}

// ShowGroup selects a page and requests scrolling to its named group. It
// returns false for a missing, ambiguous, or search-filtered group. It does not
// clear search. GroupNavigation need not be enabled for programmatic navigation.
func (v *SettingsView) ShowGroup(page, title string) bool {
	q := strings.ToLower(strings.TrimSpace(v.query))
	found := -1
	for _, s := range v.sections {
		if s.title != page {
			continue
		}
		for i, g := range s.allGroups() {
			if g.Title != title || !settingGroupMatches(g, q) {
				continue
			}
			if found >= 0 {
				return false
			}
			found = i
		}
	}
	if found < 0 {
		return false
	}
	v.SetValue(page)
	v.pendingGroup = &settingsGroupTarget{page, found}
	return true
}
func settingGroupMatches(g SettingGroup, q string) bool {
	for _, it := range g.Items {
		if settingMatches(it, q) {
			return true
		}
	}
	return false
}
func (v *SettingsView) groupID(page, group int) string {
	id := autoID("settings-group", v) + "/" + strconv.Itoa(page) + "/" + strconv.Itoa(group)
	for {
		collision := false
		for _, s := range v.sections {
			if s.title == id {
				collision = true
				break
			}
		}
		if !collision {
			return id
		}
		id += "/"
	}
}
func (v *SettingsView) settingsMetrics() (label, description, padding, gap, column float32) {
	switch v.size {
	case SettingsSizeXSmall:
		return theme.TextSm, theme.TextXs, theme.SpaceSm, theme.SpaceMd, 180
	case SettingsSizeSmall:
		return theme.TextControl, theme.TextSm, theme.SpaceMd, theme.SpaceLg, 210
	case SettingsSizeLarge:
		return theme.TextLg, theme.TextBody, theme.SpaceXl, theme.Space2xl, 280
	default:
		return theme.TextBody, theme.TextMd, theme.SpaceLg, theme.SpaceXl, 240
	}
}
