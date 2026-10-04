package editorstyle

// An inline object is represented by one private-use rune with a blank glyph.
// Its advance participates in Gio's own line breaking, caret and selection
// geometry; callers paint the actual object in Editor.Regions afterwards.
// This file builds only the font. It does not replace source text or define the
// application's clipboard/IME representation.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"

	"gioui.org/font"
	fontapi "github.com/go-text/typesetting/font"
)

type InlineGlyph struct {
	Rune    rune
	Advance uint16 // font design units, at most 32767 for go-text's signed metrics
}
type InlineFontMetrics struct {
	UnitsPerEm      uint16 // 16..16384
	Ascent, Descent int16  // positive distances above/below the baseline
}

type inlineFace struct{ font *fontapi.Font }

func (f inlineFace) Face() *fontapi.Face { return fontapi.NewFace(f.font) }

// InlineFont creates an owned, outline-free font containing only the supplied
// private-use runes. Include it ahead of normal text fonts in an editor-local
// shaper. Never register this font globally: rune mappings are document-local.
// Glyphs need not be sorted; duplicate/non-private runes and invalid metrics fail.
func InlineFont(name font.Typeface, metrics InlineFontMetrics, glyphs []InlineGlyph) (font.FontFace, error) {
	if name == "" || metrics.UnitsPerEm < 16 || metrics.UnitsPerEm > 16384 || metrics.Ascent <= 0 || metrics.Descent < 0 || len(glyphs) == 0 || len(glyphs) > 65534 {
		return font.FontFace{}, fmt.Errorf("editorstyle: invalid inline font metrics")
	}
	glyphs = slices.Clone(glyphs)
	slices.SortFunc(glyphs, func(a, b InlineGlyph) int { return int(a.Rune - b.Rune) })
	for i, g := range glyphs {
		private := g.Rune >= 0xe000 && g.Rune <= 0xf8ff || g.Rune >= 0xf0000 && g.Rune <= 0xffffd || g.Rune >= 0x100000 && g.Rune <= 0x10fffd
		if !private || g.Advance == 0 || g.Advance > 32767 || i > 0 && glyphs[i-1].Rune == g.Rune {
			return font.FontFace{}, fmt.Errorf("editorstyle: invalid inline glyph")
		}
	}
	data := inlineFontData(metrics, glyphs)
	face, err := fontapi.ParseTTF(bytes.NewReader(data))
	if err != nil {
		return font.FontFace{}, fmt.Errorf("editorstyle: inline font: %w", err)
	}
	return font.FontFace{Font: font.Font{Typeface: name}, Face: inlineFace{face.Font}}, nil
}

func inlineFontData(m InlineFontMetrics, glyphs []InlineGlyph) []byte {
	u16 := binary.BigEndian.PutUint16
	u32 := binary.BigEndian.PutUint32
	n := len(glyphs) + 1 // glyph zero is the absent .notdef glyph
	head := make([]byte, 54)
	u32(head, 0x00010000)
	u32(head[12:], 0x5f0f3cf5)
	u16(head[18:], m.UnitsPerEm)
	u16(head[50:], 1)
	hhea := make([]byte, 36)
	u32(hhea, 0x00010000)
	u16(hhea[4:], uint16(m.Ascent))
	u16(hhea[6:], uint16(-m.Descent))
	u16(hhea[18:], 1)
	u16(hhea[34:], uint16(n))
	hmtx := make([]byte, n*4)
	for i, g := range glyphs {
		u16(hmtx[(i+1)*4:], g.Advance)
		if g.Advance > binary.BigEndian.Uint16(hhea[10:]) {
			u16(hhea[10:], g.Advance)
		}
	}
	maxp := make([]byte, 32)
	u32(maxp, 0x00010000)
	u16(maxp[4:], uint16(n))
	u16(maxp[14:], 1)
	// Unicode full-repertoire cmap, format 12. Each rune has one blank glyph.
	cmap := make([]byte, 12+16+len(glyphs)*12)
	u16(cmap[2:], 1)
	u16(cmap[4:], 3)
	u16(cmap[6:], 10)
	u32(cmap[8:], 12)
	u16(cmap[12:], 12)
	u32(cmap[16:], uint32(len(cmap)-12))
	u32(cmap[24:], uint32(len(glyphs)))
	for i, g := range glyphs {
		at := 28 + i*12
		u32(cmap[at:], uint32(g.Rune))
		u32(cmap[at+4:], uint32(g.Rune))
		u32(cmap[at+8:], uint32(i+1))
	}
	// Nonzero extents distinguish objects from trailing whitespace in go-text's
	// line wrapper. Zero contours keep the glyph invisible; content is painted separately.
	glyf := make([]byte, n*12)
	loca := make([]byte, (n+1)*4)
	for i := 0; i <= n; i++ {
		u32(loca[i*4:], uint32(i*12))
	}
	for i, g := range glyphs {
		at := (i + 1) * 12
		u16(glyf[at+4:], uint16(-m.Descent))
		u16(glyf[at+6:], g.Advance)
		u16(glyf[at+8:], uint16(m.Ascent))
	}
	tables := map[string][]byte{"head": head, "hhea": hhea, "hmtx": hmtx, "maxp": maxp, "cmap": cmap, "loca": loca, "glyf": glyf}
	tags := make([]string, 0, len(tables))
	for tag := range tables {
		tags = append(tags, tag)
	}
	slices.Sort(tags)
	data := make([]byte, 12+16*len(tags))
	u32(data, 0x00010000)
	u16(data[4:], uint16(len(tags)))
	pow, log := 1, 0
	for pow*2 <= len(tags) {
		pow *= 2
		log++
	}
	u16(data[6:], uint16(pow*16))
	u16(data[8:], uint16(log))
	u16(data[10:], uint16(len(tags)*16-pow*16))
	headOffset := 0
	for i, tag := range tags {
		table := tables[tag]
		at := 12 + i*16
		copy(data[at:], tag)
		u32(data[at+4:], inlineChecksum(table))
		u32(data[at+8:], uint32(len(data)))
		u32(data[at+12:], uint32(len(table)))
		if tag == "head" {
			headOffset = len(data)
		}
		data = append(data, table...)
		for len(data)%4 != 0 {
			data = append(data, 0)
		}
	}
	u32(data[headOffset+8:], 0xb1b0afba-inlineChecksum(data))
	return data
}
func inlineChecksum(data []byte) uint32 {
	var sum uint32
	for len(data) >= 4 {
		sum += binary.BigEndian.Uint32(data)
		data = data[4:]
	}
	var last [4]byte
	copy(last[:], data)
	return sum + binary.BigEndian.Uint32(last[:])
}
