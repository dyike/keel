package kit

// NotifyKey creates or replaces a notice with a business key local to this
// Notifier. Replacement preserves its ID and queue position and restarts its
// timeout, like Update. An empty key creates a new anonymous notice.
func (v *NotifierView) NotifyKey(key string, n Notice) int {
	if key != "" {
		for _, current := range v.items {
			if current.key == key {
				v.Update(current.id, n)
				return current.id
			}
		}
	}
	id := v.Notify(n)
	v.items[len(v.items)-1].key = key
	return id
}

// DismissKey removes a shown or waiting keyed notice. Empty and unknown keys
// return false. A later NotifyKey for this key creates a new notice ID.
func (v *NotifierView) DismissKey(key string) bool {
	if key == "" {
		return false
	}
	for _, current := range v.items {
		if current.key == key {
			v.Dismiss(current.id)
			return true
		}
	}
	return false
}

// Clear removes all notices present at entry, then calls their OnClose callbacks
// in queue order. Notices added by callbacks survive unless explicitly removed
// by another callback. The return value counts the removed notices.
func (v *NotifierView) Clear() int {
	items := v.items
	v.items = nil
	for _, current := range items {
		if current.OnClose != nil {
			current.OnClose()
		}
	}
	return len(items)
}
