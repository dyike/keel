// SPDX-License-Identifier: Unlicense OR MIT
package text

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

func TestLayoutCacheByteBudget(t *testing.T) {
	var c layoutCache
	doc := document{lines: []line{{runs: []runLayout{{Glyphs: make([]glyph, 16000)}}}}}
	// Fewer than 1000 paragraphs can still exceed the byte budget.
	for i := 0; i < 100; i++ {
		c.Put(layoutKey{str: fmt.Sprint(i)}, doc)
	}
	if c.bytes > layoutCacheBytes || len(c.m) >= 100 {
		t.Fatalf("retained %d bytes in %d layouts", c.bytes, len(c.m))
	}
	if _, ok := c.Get(layoutKey{str: "99"}); !ok {
		t.Fatal("lost most recent layout")
	}
	if _, ok := c.Get(layoutKey{str: "0"}); ok {
		t.Fatal("kept oldest layout")
	}
}

func TestLayoutCacheProtectsVisibleWorkingSet(t *testing.T) {
	var c layoutCache
	c.beginFrame()
	doc := document{lines: []line{{runs: []runLayout{{Glyphs: make([]glyph, 16000)}}}}}
	for i := 0; i < 20; i++ {
		c.Put(layoutKey{str: fmt.Sprint(i)}, doc)
	}
	if c.bytes <= layoutCacheBytes {
		t.Fatal("test working set does not exceed ordinary budget")
	}
	for frame := 0; frame < 20; frame++ {
		c.beginFrame()
		for i := 0; i < 20; i++ {
			if _, ok := c.Get(layoutKey{str: fmt.Sprint(i)}); !ok {
				t.Fatal("evicted a visible layout")
			}
		}
	}
	for i := 0; i < layoutLiveFrames; i++ {
		c.beginFrame()
	}
	if c.bytes > layoutCacheBytes {
		t.Fatal("unused working set did not shrink")
	}
}

func TestLayoutCacheHardLimits(t *testing.T) {
	var c layoutCache
	c.beginFrame()
	doc := document{lines: []line{{runs: []runLayout{{Glyphs: make([]glyph, 16000)}}}}}
	for i := 0; i < 200; i++ {
		c.Put(layoutKey{str: fmt.Sprint(i)}, doc)
	}
	if c.bytes > layoutLiveBytes {
		t.Fatalf("live budget: %d", c.bytes)
	}
	for i := 0; i < maxSize+30; i++ {
		c.Put(layoutKey{str: fmt.Sprint(i)}, document{})
	}
	if len(c.m) > maxSize {
		t.Fatal("entry cap exceeded")
	}
	n := c.bytes
	c.Put(layoutKey{str: "1001"}, document{})
	if c.bytes != n {
		t.Fatal("replacement counted twice")
	}
}

func TestLayoutCacheOwnsParagraphKeys(t *testing.T) {
	source := strings.Repeat("document", 1<<16)
	key := layoutKey{str: source[100:110]}
	var c layoutCache
	c.Put(key, document{})
	e := c.m[key]
	if unsafe.StringData(e.key.str) == unsafe.StringData(key.str) {
		t.Fatal("short paragraph retains original document")
	}
	if _, ok := c.Get(key); !ok {
		t.Fatal("copied key not found")
	}
}

func TestLayoutCacheOversizePreservesWorkingSet(t *testing.T) {
	var c layoutCache
	c.Put(layoutKey{str: "keep"}, document{})
	// Only size accounting reads the spare capacity.
	huge := document{lines: make([]line, 0, layoutCacheBytes/int(unsafe.Sizeof(line{}))+1)}
	c.Put(layoutKey{str: "huge"}, huge)
	if _, ok := c.Get(layoutKey{str: "keep"}); !ok {
		t.Fatal("oversized layout evicted ordinary content")
	}
	if _, ok := c.Get(layoutKey{str: "huge"}); ok {
		t.Fatal("retained oversized layout")
	}
}
