package kit

import (
	"github.com/dyike/keel/ui/base"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"slices"
)

func cloneCommandItems(items []CommandItem) []CommandItem {
	out := slices.Clone(items)
	for i := range out {
		out[i].Keywords = slices.Clone(out[i].Keywords)
	}
	return out
}

// OnSelect reports the original SetItems index of a newly highlighted item,
// or -1 when none is available. Group headings do not contribute indices.
// Initial/model-driven selection is delivered after render, never from Render.
func (v *CommandView) OnSelect(fn func(int)) *CommandView { v.onSelect = fn; return v }

// OnConfirm reports the original index after executing the item's Action.
// The callback and index are captured before Action; callback-driven model
// changes do not change which command was confirmed. Items without actions
// also confirm. Programmatic close does not confirm or cancel.
func (v *CommandView) OnConfirm(fn func(int)) *CommandView { v.onConfirm = fn; return v }

// OnCancel reports a user dismissal, after internal close. Escape first clears
// a nonempty searchable query; only the next Escape cancels. Outside clicks
// cancel immediately. Programmatic close/disable does not cancel.
func (v *CommandView) OnCancel(fn func()) *CommandView { v.onCancel = fn; return v }

// OnQuery observes user query edits (including Escape clearing). It does not
// replace local filtering. Unlike OnSearch it does not run on opening/retry.
// Refiltering updates OnSelect before OnQuery.
func (v *CommandView) OnQuery(fn func(string)) *CommandView { v.onQuery = fn; return v }

func (v *CommandView) selectActive(row int) {
	v.active = row
	index := -1
	if row >= 0 && row < len(v.rows) && !v.rowDisabled(row) && !v.loading && v.searchError == "" {
		index = v.rows[row].index
	}
	if index != v.selected {
		v.selected = index
		v.selectionPending = true
		v.selectionVersion++
	}
}
func (v *CommandView) notifySelection() {
	if !v.selectionPending {
		return
	}
	v.selectionPending = false
	index := v.selected
	if v.onSelect != nil {
		v.onSelect(index)
	}
}

type commandSelectionTimer struct {
	view    *CommandView
	version uint64
}

func (v *CommandView) queueSelection(cx *el.Context) {
	if v.selectionPending {
		cx.AfterEnabled(v.FocusID(), commandSelectionTimer{v, v.selectionVersion}, 0, v.notifySelection)
	}
}
func (v *CommandView) queryChanged() {
	query, callback := v.query, v.onQuery
	request := v.request + 1
	v.searchChanged()
	if !v.open || request != v.request {
		if callback != nil {
			callback(query)
		}
		return
	}
	v.buildRows()
	v.selectActive(base.List{Count: len(v.rows), Disabled: v.rowDisabled}.First())
	v.notifySelection()
	if callback != nil {
		callback(query)
	}
}
func (v *CommandView) clearQuery() bool {
	if v.nonsearchable || v.query == "" {
		return false
	}
	v.query = ""
	v.queryChanged()
	return true
}
func (v *CommandView) escape() {
	if !v.clearQuery() {
		v.cancel()
	}
}
func (v *CommandView) cancel() {
	if !v.open {
		return
	}
	callback := v.onCancel
	v.SetValue(false)
	if callback != nil {
		callback()
	}
}

// Query returns the current query, including a hidden non-searchable query.
func (v *CommandView) Query() string { return v.query }

// SetQuery changes an open searchable palette as if typed, including callbacks.
// Closed/non-searchable palettes only store the text. Opening resets it.
func (v *CommandView) SetQuery(query string) {
	if v.query == query {
		return
	}
	v.query = query
	if v.open && !v.nonsearchable {
		v.queryChanged()
	}
}

// SelectedIndex returns the highlighted original item index, or -1. Model
// changes are reconciled on render; group headings/dividers are never selected.
func (v *CommandView) SelectedIndex() int { return v.selected }

// MatchedCount reports matching items, including disabled items, but excludes
// group headings and dividers. Loading does not erase the installed model.
func (v *CommandView) MatchedCount() int {
	v.buildRows()
	count := 0
	for _, row := range v.rows {
		if !row.header && !row.item.Separator {
			count++
		}
	}
	return count
}

// SetLoading controls the spinner for application-owned synchronous models.
// OnSearch/SetResults manage it automatically for token-based remote results.
func (v *CommandView) SetLoading(on bool) { v.loading = on; v.searchError = ""; v.active = -1 }

// IsLoading reports whether the palette is waiting for results.
func (v *CommandView) IsLoading() bool { return v.loading }

// Focused non-editor elements consume their keys before root shortcuts, so
// resolve the same bindings on the focus chain as well as on the editor path.
func (v *CommandView) boundKey(e el.KeyEvent) bool {
	for _, entry := range v.rows {
		item := entry.item
		if entry.header || item.Separator || item.Disabled || item.ActionName == "" {
			continue
		}
		for _, chord := range core.Bindings(item.ActionName) {
			name, mods, err := core.ParseShortcut(chord)
			if err == nil && string(name) == e.Name && mods == e.Modifiers {
				if e.State == el.KeyPress {
					v.run(entry)
				}
				return true
			}
		}
	}
	return false
}
