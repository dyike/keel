package font

type glyphExtents struct {
	valid   bool
	extents GlyphExtents
}

// Cache only pages containing measured glyphs. CJK fonts can have tens of
// thousands of glyphs while a window uses only a few dozen.
const extentsPageSize = 128

type extentsCache struct {
	count int
	pages []*[extentsPageSize]glyphExtents
}

func (ec *extentsCache) get(gid GID) (GlyphExtents, bool) {
	if uint64(gid) >= uint64(ec.count) {
		return GlyphExtents{}, false
	}
	page := int(gid) / extentsPageSize
	if page >= len(ec.pages) || ec.pages[page] == nil {
		return GlyphExtents{}, false
	}
	ge := ec.pages[page][int(gid)%extentsPageSize]
	return ge.extents, ge.valid
}

func (ec *extentsCache) set(gid GID, extents GlyphExtents) {
	if uint64(gid) >= uint64(ec.count) {
		return
	}
	if ec.pages == nil {
		ec.pages = make([]*[extentsPageSize]glyphExtents, (ec.count+extentsPageSize-1)/extentsPageSize)
	}
	page := int(gid) / extentsPageSize
	if ec.pages[page] == nil {
		ec.pages[page] = new([extentsPageSize]glyphExtents)
	}
	ec.pages[page][int(gid)%extentsPageSize] = glyphExtents{valid: true, extents: extents}
}

func (ec *extentsCache) reset() {
	for i := range ec.pages {
		ec.pages[i] = nil
	}
}

func (f *Face) GlyphExtents(glyph GID) (GlyphExtents, bool) {
	if e, ok := f.extentsCache.get(glyph); ok {
		return e, ok
	}
	e, ok := f.glyphExtentsRaw(glyph)
	if ok {
		f.extentsCache.set(glyph, e)
	}
	return e, ok
}
