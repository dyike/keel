// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"strings"
	"unsafe"
)

// Keel patch: a paragraph count does not bound the storage of long documents.
// Keep recent visible layouts above the ordinary budget to avoid evicting a
// working set just before its next frame. Both bytes and entries have hard caps.
const (
	layoutCacheBytes = 8 << 20
	layoutLiveBytes  = 64 << 20
	layoutLiveFrames = 8
)

type layoutEntry struct {
	prev, next *layoutEntry
	key        layoutKey
	value      document
	bytes      int
	used       uint64
}

type layoutCache struct {
	m              map[layoutKey]*layoutEntry
	oldest, newest *layoutEntry
	bytes          int
	frame          uint64
}

func (c *layoutCache) Get(k layoutKey) (document, bool) {
	if e := c.m[k]; e != nil {
		c.unlink(e)
		c.link(e)
		return e.value, true
	}
	return document{}, false
}

func (c *layoutCache) Put(k layoutKey, v document) {
	if e := c.m[k]; e != nil {
		c.remove(e)
	}
	n := layoutBytes(k, v)
	// A single oversized paragraph must not evict the entire working set.
	limit := layoutCacheBytes
	if c.frame != 0 {
		limit = layoutLiveBytes / 2
	}
	if n > limit {
		return
	}
	if c.m == nil {
		c.m = make(map[layoutKey]*layoutEntry)
	}
	// LayoutString splits a document into substring keys. Own each key so
	// one short cached paragraph cannot retain the whole original document.
	k.str = strings.Clone(k.str)
	k.truncator = strings.Clone(k.truncator)
	e := &layoutEntry{key: k, value: v, bytes: n}
	c.m[k] = e
	c.bytes += n
	c.link(e)
	c.trim()
}

func (c *layoutCache) beginFrame() {
	c.frame++
	c.trim()
}

func (c *layoutCache) trim() {
	for c.oldest != nil && (len(c.m) > maxSize || c.bytes > layoutLiveBytes ||
		c.bytes > layoutCacheBytes && (c.frame == 0 || c.frame-c.oldest.used >= layoutLiveFrames)) {
		c.remove(c.oldest)
	}
}

func (c *layoutCache) unlink(e *layoutEntry) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.oldest = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.newest = e.prev
	}
	e.prev, e.next = nil, nil
}

func (c *layoutCache) link(e *layoutEntry) {
	e.used, e.prev = c.frame, c.newest
	if c.newest != nil {
		c.newest.next = e
	} else {
		c.oldest = e
	}
	c.newest = e
}

func (c *layoutCache) remove(e *layoutEntry) {
	c.unlink(e)
	delete(c.m, e.key)
	c.bytes -= e.bytes
}

func layoutBytes(k layoutKey, d document) int {
	n := int(unsafe.Sizeof(layoutEntry{})) + len(k.str) + len(k.truncator) + len(k.font.Typeface) + len(k.locale.Language)
	n += cap(d.lines) * int(unsafe.Sizeof(line{}))
	for _, l := range d.lines {
		n += cap(l.runs)*int(unsafe.Sizeof(runLayout{})) + cap(l.visualOrder)*int(unsafe.Sizeof(int(0)))
		for _, r := range l.runs {
			n += cap(r.Glyphs) * int(unsafe.Sizeof(glyph{}))
		}
	}
	return n
}
