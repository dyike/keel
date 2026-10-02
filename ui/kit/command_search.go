package kit

// OnSearch replaces local fuzzy filtering with application-provided results.
// Each request has a new token. Use core.Update to deliver asynchronous results
// through SetResults/SetSearchError; stale tokens and closed palettes are ignored.
func (v *CommandView) OnSearch(fn func(query string, token uint64)) *CommandView {
	v.onSearch = fn
	v.cached = false
	if v.open {
		v.searchChanged()
	}
	return v
}
func (v *CommandView) searchChanged() {
	v.request++
	v.active = -1
	v.searchError = ""
	v.loading = v.onSearch != nil
	if v.onSearch != nil {
		v.onSearch(v.query, v.request)
	}
}
func (v *CommandView) SetResults(token uint64, items ...CommandItem) bool {
	if !v.open || token != v.request || v.onSearch == nil {
		return false
	}
	v.SetItems(items...)
	v.loading = false
	v.searchError = ""
	return true
}
func (v *CommandView) SetSearchError(token uint64, message string) bool {
	if !v.open || token != v.request || v.onSearch == nil {
		return false
	}
	v.loading = false
	v.searchError = message
	v.active = -1
	return true
}
