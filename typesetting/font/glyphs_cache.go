// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package font

import (
	"sync"

	"github.com/dyike/keel/typesetting/font/opentype/tables"
)

// Keep immutable headers for metrics, and decode contour points only for
// outlines actually used. Native rasterizers do not need Go contour arrays.
// Font instances share this pointer; Faces may query it concurrently.
type glyphOutlineCache struct {
	raw     []byte
	offsets []uint32
	mu      sync.Mutex
	decoded map[tables.GlyphID]tables.Glyph
}

func (c *glyphOutlineCache) glyph(id tables.GlyphID) tables.Glyph {
	c.mu.Lock()
	defer c.mu.Unlock()
	if g, ok := c.decoded[id]; ok {
		return g
	}
	if int(id)+1 >= len(c.offsets) {
		return tables.Glyph{}
	}
	start, end := c.offsets[id], c.offsets[int(id)+1]
	if start >= end || uint64(end) > uint64(len(c.raw)) {
		return tables.Glyph{}
	}
	g, _, err := tables.ParseGlyph(c.raw[start:end])
	if err != nil {
		return tables.Glyph{}
	}
	if c.decoded == nil {
		c.decoded = make(map[tables.GlyphID]tables.Glyph)
	}
	c.decoded[id] = g
	return g
}
