package kit

import "github.com/dyike/keel/ui/el"

// SetFilter filters source rows before sorting. The predicate receives a copy;
// nil clears filtering. Hidden selections are retained but not copied/exported.
// Update the predicate via SetFilter when external filter criteria change.
func (v *TableView) SetFilter(match func([]string) bool) {
	v.filter = match
	v.resort()
}

// VisibleLen reports the row count after filtering; Len is the source count.
func (v *TableView) VisibleLen() int { return len(v.order) }

// OnLoadMore requests another page near the bottom. The callback runs once
// per delivered dataset and sets Loading before calling fn. Deliver rows and
// clear Loading through core.Update; SetHasMore(false) marks the final page.
func (v *TableView) OnLoadMore(fn func()) *TableView { v.onLoadMore = fn; return v }
func (v *TableView) SetHasMore(more bool) {
	v.hasMore = more
	if !more {
		v.loadRequested = false
	}
}

// SetLoadError stops loading and exposes a retry action. Errors suppress
// automatic retries; a user retry invokes OnLoadMore again.
func (v *TableView) SetLoadError(message string) {
	v.loadError = message
	v.loading = false
	v.loadRequested = true
}
func (v *TableView) requestMore() {
	if v.loading || v.onLoadMore == nil || v.disabled {
		return
	}
	v.loading, v.loadRequested = true, true
	v.loadError = ""
	v.onLoadMore()
}

type tableLoadKey struct{ table *TableView }

func (v *TableView) loadNearEnd(cx *el.Context) {
	if !v.hasMore || v.loading || v.loadError != "" || v.loadRequested || v.onLoadMore == nil {
		return
	}
	off, view, _ := cx.ScrollState(v.list.ID())
	if view == 0 || off+view+2*v.list.rowH >= float32(len(v.order))*v.list.rowH {
		cx.AfterEnabled(autoID("table", v), tableLoadKey{v}, 0, func() {
			off, view, _ := cx.ScrollState(v.list.ID())
			if view > 0 && off+view+2*v.list.rowH >= float32(len(v.order))*v.list.rowH {
				v.requestMore()
			}
		})
	}
}
