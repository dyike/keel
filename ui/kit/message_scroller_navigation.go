package kit

import "github.com/dyike/keel/ui/el"

// ScrollToMessage minimally reveals a stable message ID, including before the
// first render. It takes precedence over tail following and pending end jumps.
// Unknown IDs return false without changing a pending request. The application
// owns unread/bookmark meaning and can call this after installing message keys.
func (v *MessageScrollerView) ScrollToMessage(id string) bool {
	if _, ok := v.list.indices[id]; !ok {
		return false
	}
	v.list.reveal = id
	v.list.endApplied = v.list.end
	return true
}

// IsScrolledUp reports whether the last painted viewport has content below it.
// Before the first paint (or with no scrollable content) it returns false.
func (v *MessageScrollerView) IsScrolledUp(cx *el.Context) bool {
	offset, view, content := cx.ScrollState(v.list.ID())
	return view > 0 && offset+view < content-4
}

// IsFollowingTail reports effective automatic tail following. A pending message
// jump suspends it; reaching the bottom normally resumes it when SetFollow is on.
func (v *MessageScrollerView) IsFollowingTail(cx *el.Context) bool {
	return v.list.followEnd && v.list.reveal == "" && !v.IsScrolledUp(cx)
}
