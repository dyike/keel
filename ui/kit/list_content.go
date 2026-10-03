package kit

import (
	"github.com/dyike/keel/ui/el"
	"strings"
)

type listDisplayRow struct{ index, kind int }

// ListItemContext is a copied source-item snapshot for custom row rendering.
type ListItemContext struct {
	Item               ListItem
	Index              int
	Selected, Disabled bool
}

func (v *ListView) Searchable(on bool) *ListView { v.searchable = on; return v }
func (v *ListView) Query() string                { return v.query }
func (v *ListView) SetQuery(query string) {
	v.query = query
	v.reveal = true
	v.loadRequested = false
	v.rebuildRows()
	if v.onSearch != nil {
		v.onSearch(query)
	}
}
func (v *ListView) OnSearch(fn func(string)) *ListView { v.onSearch = fn; return v }

// RenderItem replaces row contents. Nil results use the default label/icon.
func (v *ListView) RenderItem(fn func(*el.Context, ListItemContext) el.Element) *ListView {
	v.renderItem = fn
	return v
}
func (v *ListView) RenderGroupHeader(fn func(*el.Context, string) el.Element) *ListView {
	v.renderHeader = fn
	return v
}
func (v *ListView) RenderGroupFooter(fn func(*el.Context, string) el.Element) *ListView {
	v.renderFooter = fn
	v.rebuildRows()
	return v
}
func (v *ListView) RowHeight(dp float32) *ListView {
	if dp >= 20 && dp <= 512 {
		v.list.rowH = dp
	}
	return v
}
func (v *ListView) rebuildRows() {
	v.display = nil
	v.positions = make([]int, len(v.items))
	for i := range v.positions {
		v.positions[i] = -1
	}
	query := strings.ToLower(strings.TrimSpace(v.query))
	lastGroup := ""
	last := -1
	footer := func() {
		if last >= 0 && lastGroup != "" && v.renderFooter != nil {
			v.display = append(v.display, listDisplayRow{last, 2})
		}
	}
	for i, item := range v.entries {
		if query != "" && !strings.Contains(strings.ToLower(item.Label+" "+strings.Join(item.Keywords, " ")), query) {
			continue
		}
		if last < 0 || item.Group != lastGroup {
			footer()
			if item.Group != "" {
				v.display = append(v.display, listDisplayRow{i, 1})
			}
		}
		v.positions[i] = len(v.display)
		v.display = append(v.display, listDisplayRow{i, 0})
		lastGroup = item.Group
		last = i
	}
	footer()
	v.list.SetCount(len(v.display))
}
func (v *ListView) OnLoadMore(fn func()) *ListView { v.onLoadMore = fn; return v }
func (v *ListView) SetHasMore(on bool) {
	v.hasMore = on
	if !on {
		v.loadRequested = false
	}
}
func (v *ListView) SetLoading(on bool) {
	v.loading = on
	if on {
		v.loadError = ""
	}
}
func (v *ListView) SetLoadError(message string) {
	v.loadError = message
	v.loading = false
	v.loadRequested = true
}
func (v *ListView) requestMore() {
	if v.disabled || v.loading || v.onLoadMore == nil {
		return
	}
	v.loading = true
	v.loadRequested = true
	v.loadError = ""
	v.onLoadMore()
}

type listLoadKey struct{ view *ListView }

func (v *ListView) loadNearEnd(cx *el.Context) {
	if !v.hasMore || v.loading || v.loadRequested || v.loadError != "" || v.onLoadMore == nil {
		return
	}
	off, view, _ := cx.ScrollState(v.list.ID())
	if view > 0 && off+view+2*v.list.rowH < float32(len(v.display))*v.list.rowH {
		return
	}
	cx.AfterEnabled(autoID("list", v), listLoadKey{v}, 0, func() {
		off, view, _ := cx.ScrollState(v.list.ID())
		if view > 0 && off+view+2*v.list.rowH >= float32(len(v.display))*v.list.rowH {
			v.requestMore()
		}
	})
}
