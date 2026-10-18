// SPDX-License-Identifier: Unlicense OR MIT

// Package dirty records which areas of a mutable image changed, so each
// renderer holding a texture of it can upload only those areas. Keel patch.
package dirty

import (
	"image"
	"sync"
)

// history is how many changes a Tracker remembers. A renderer that falls
// further behind uploads the whole image instead.
const history = 32

// Tracker is the identity of a mutable image and its change log.
type Tracker struct {
	// Pixels, if set, guards the image's pixels: renderers hold it while
	// they read them, so the owner may write from any goroutine.
	Pixels sync.Locker

	mu      sync.Mutex
	version uint64
	log     [history]image.Rectangle // log[v%history] is the area changed by version v
}

// Invalidate records that pixels in r (image-local coordinates) changed.
func (t *Tracker) Invalidate(r image.Rectangle) {
	if r.Empty() {
		return
	}
	t.mu.Lock()
	t.version++
	t.log[t.version%history] = r
	t.mu.Unlock()
}

// Version reports the current version.
func (t *Tracker) Version() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.version
}

// Since returns the union of areas changed after version from and the
// current version. ok is false when the log no longer reaches back to from;
// the caller must then treat the whole image as changed.
func (t *Tracker) Since(from uint64) (changed image.Rectangle, version uint64, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.version-from > history || from > t.version {
		return image.Rectangle{}, t.version, false
	}
	for v := from + 1; v <= t.version; v++ {
		changed = changed.Union(t.log[v%history])
	}
	return changed, t.version, true
}
