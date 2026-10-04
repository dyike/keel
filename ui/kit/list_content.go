package kit

import (
	"strings"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
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

type listSlots struct {
	empty, noMatches, initial, loading el.View
	err                                func(message string, retry func()) el.View
}

// EmptyContent shows view when the list has no items and no search is
// active. Nil restores the default message.
func (v *ListView) EmptyContent(view el.View) *ListView { v.slots.empty = view; return v }

// NoMatchesContent shows view when a search finds nothing. Nil restores
// the default message.
func (v *ListView) NoMatchesContent(view el.View) *ListView { v.slots.noMatches = view; return v }

// InitialContent shows view in place of the rows while a searchable list
// has an empty query, as for a search that has not started. Nil turns it
// off.
func (v *ListView) InitialContent(view el.View) *ListView { v.slots.initial = view; return v }

// LoadingContent replaces the spinner shown while more rows load.
func (v *ListView) LoadingContent(view el.View) *ListView { v.slots.loading = view; return v }

// ErrorContent replaces the load error message and its Retry button; retry
// requests the rows again.
func (v *ListView) ErrorContent(fn func(message string, retry func()) el.View) *ListView {
	v.slots.err = fn
	return v
}

// stateContent is what shows below the rows: loading, a load error, or an
// empty or unmatched list.
func (v *ListView) stateContent(cx *el.Context) el.Element {
	text := locale.Current()
	slot := func(view el.View, name string) el.Element {
		return el.Div().ID(autoID("list", v) + "/" + name).Items(el.Stretch).Child(view.Render(cx))
	}
	switch {
	case v.loading && v.slots.loading != nil:
		return slot(v.slots.loading, "loading")
	case v.loading:
		return Spinner().Render(cx)
	case v.loadError != "" && v.slots.err != nil:
		if view := v.slots.err(v.loadError, v.requestMore); view != nil {
			return slot(view, "error")
		}
		fallthrough
	case v.loadError != "":
		return el.Div().Items(el.Stretch).Child(el.Text(v.loadError).TextColor(theme.Danger), Button(text.Retry, v.requestMore).Render(cx))
	case len(v.display) > 0:
		return nil
	case v.query == "" && v.slots.empty != nil:
		return slot(v.slots.empty, "empty")
	case v.query != "" && v.slots.noMatches != nil:
		return slot(v.slots.noMatches, "no-matches")
	}
	return el.Text(text.NoMatches).TextColor(theme.Muted)
}
