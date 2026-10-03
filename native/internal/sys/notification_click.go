package sys

import "sync"

type notificationClickEntry struct{ fn func() }
type notificationClicks struct {
	sync.Mutex
	entries map[string]*notificationClickEntry
}

var systemNotificationClicks notificationClicks

// Registration precedes Post so an immediate OS response has a handler. A failed
// replacement restores the previous handler, without overwriting a newer post.
func (r *notificationClicks) register(id string, fn func()) func(error) {
	r.Lock()
	if r.entries == nil {
		r.entries = make(map[string]*notificationClickEntry)
	}
	previous := r.entries[id]
	entry := &notificationClickEntry{fn: fn}
	r.entries[id] = entry
	r.Unlock()
	return func(err error) {
		r.Lock()
		defer r.Unlock()
		if r.entries[id] != entry {
			return
		}
		if err != nil {
			if previous == nil {
				delete(r.entries, id)
			} else {
				r.entries[id] = previous
			}
		} else if fn == nil {
			delete(r.entries, id)
		}
	}
}
func (r *notificationClicks) take(id string) func() {
	r.Lock()
	defer r.Unlock()
	entry := r.entries[id]
	delete(r.entries, id)
	if entry != nil {
		return entry.fn
	}
	return nil
}
