package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"strconv"
)

// ComboboxGroup is an ordered section of candidates with a stable nonempty ID.
// Label defaults to ID. Empty groups are not displayed.
type ComboboxGroup struct {
	ID, Label string
	Items     []ComboboxItem
}

// SetGroups replaces candidates with grouped items. Input slices are copied;
// duplicate group IDs and item values keep their first occurrence. Empty group
// IDs and item values are ignored. Selection and disabled overrides survive.
func (v *ComboboxView) SetGroups(groups ...ComboboxGroup) {
	labels, byValue := map[string]string{}, map[string]string{}
	var items []ComboboxItem
	for _, g := range groups {
		if g.ID == "" {
			continue
		}
		if _, seen := labels[g.ID]; seen {
			continue
		}
		label := g.Label
		if label == "" {
			label = g.ID
		}
		labels[g.ID] = label
		for _, item := range g.Items {
			if item.Value == "" {
				continue
			}
			if _, seen := byValue[item.Value]; seen {
				continue
			}
			byValue[item.Value] = g.ID
			items = append(items, item)
		}
	}
	v.SetItems(items...)
	v.groupLabels, v.groupFor = labels, byValue
}

// SetGroupResults delivers grouped asynchronous results on the UI loop.
// Stale, closed and disabled requests are rejected without changing content.
func (v *ComboboxView) SetGroupResults(token uint64, groups ...ComboboxGroup) bool {
	if !v.open || v.disabled || token != v.request || v.onSearch == nil {
		return false
	}
	v.SetGroups(groups...)
	v.loading = false
	v.searchError = ""
	v.active = v.enabledOption(v.matches(), 0, 1)
	return true
}

type comboboxDisplayRow struct {
	option int
	group  string
}

func (v *ComboboxView) buildDisplayRows() {
	v.displayRows = nil
	v.matchKeys = nil
	v.optionRows = make([]int, len(v.filtered))
	counts := make(map[string]int, len(v.filtered))
	previous := ""
	for i, value := range v.filtered {
		group := v.groupFor[value]
		if group != "" && group != previous {
			v.displayRows = append(v.displayRows, comboboxDisplayRow{option: -1, group: group})
			v.matchKeys = append(v.matchKeys, "group:"+group)
		}
		previous = group
		v.optionRows[i] = len(v.displayRows)
		v.displayRows = append(v.displayRows, comboboxDisplayRow{option: i})
		v.matchKeys = append(v.matchKeys, "item:"+strconv.Itoa(len(value))+":"+value+":"+strconv.Itoa(counts[value]))
		counts[value]++
	}
}
func (v *ComboboxView) displayIndex(option int) int {
	v.matches()
	if option < 0 || option >= len(v.optionRows) {
		return -1
	}
	return v.optionRows[option]
}
func (v *ComboboxView) displayRow(cx *el.Context, index int) el.Element {
	v.matches()
	row := v.displayRows[index]
	if row.option >= 0 {
		return v.optionRow(cx, row.option)
	}
	return el.Div().H(el.Dp(v.optionHeight())).Px(theme.SpaceMd * v.sizeRatio()).Justify(el.Center).
		Child(el.Text(v.groupLabels[row.group]).TextSize(float32(theme.TextSm) * v.sizeRatio()).TextColor(theme.Muted).MaxLines(1))
}
