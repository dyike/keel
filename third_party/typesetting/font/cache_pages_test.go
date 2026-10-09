package font

import "testing"

func TestExtentsCachePages(t *testing.T) {
	c := extentsCache{count: 65535}
	if c.pages != nil {
		t.Fatal("allocated unused glyph cache")
	}
	e := GlyphExtents{XBearing: 3, YBearing: 4, Width: 5, Height: 6}
	for _, g := range []GID{0, 127, 128, 60000, 65534} {
		c.set(g, e)
		if got, ok := c.get(g); !ok || got != e {
			t.Fatal("cache miss", g)
		}
	}
	for _, g := range []GID{1, 129, 65535, 0xffffffff} {
		c.set(65535, e)
		if _, ok := c.get(g); ok {
			t.Fatal("unexpected cache hit", g)
		}
	}
	pages := 0
	for _, p := range c.pages {
		if p != nil {
			pages++
		}
	}
	if pages != 4 {
		t.Fatal("allocated pages for unused glyphs", pages)
	}
	c.reset()
	for _, g := range []GID{0, 127, 128, 60000, 65534} {
		if _, ok := c.get(g); ok {
			t.Fatal("stale extents after reset", g)
		}
	}
	c.set(128, e)
	if got, ok := c.get(128); !ok || got != e {
		t.Fatal("cannot reuse reset cache")
	}
}
