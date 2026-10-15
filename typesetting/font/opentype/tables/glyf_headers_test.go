package tables

import (
	"bytes"
	"testing"

	ot "github.com/dyike/keel/typesetting/font/opentype"
	"golang.org/x/image/font/gofont/goregular"
)

func FuzzGlyfHeadersMatchEager(f *testing.F) {
	// Composite glyph with one component and byte-sized arguments.
	f.Add([]byte{255, 255, 0, 0, 0, 0, 0, 10, 0, 10, 0, 0, 0, 1, 0, 0})
	ld, err := ot.NewLoader(bytes.NewReader(goregular.TTF))
	if err != nil {
		f.Fatal(err)
	}
	raw, _ := ld.RawTable(ot.MustNewTag("glyf"))
	// Glyph records include simple, composite and empty outlines.
	locaRaw, _ := ld.RawTable(ot.MustNewTag("loca"))
	headRaw, _ := ld.RawTable(ot.MustNewTag("head"))
	head, _, _ := ParseHead(headRaw)
	maxpRaw, _ := ld.RawTable(ot.MustNewTag("maxp"))
	maxp, _, _ := ParseMaxp(maxpRaw)
	loca, _ := ParseLoca(locaRaw, int(maxp.NumGlyphs), head.IndexToLocFormat == 1)
	for i := 0; i+1 < len(loca); i++ {
		if i < 10 || i%50 == 0 {
			f.Add(raw[loca[i]:loca[i+1]])
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return // Empty loca records are skipped by both parsers.
		}
		g, _, eagerErr := ParseGlyph(data)
		h, counts, err := ParseGlyfHeaders(data, []uint32{0, uint32(len(data))})
		if (err == nil) != (eagerErr == nil) {
			t.Fatalf("validation differs: headers=%v eager=%v", err, eagerErr)
		}
		if err != nil {
			return
		}
		count := 0
		switch d := g.Data.(type) {
		case SimpleGlyph:
			count = len(d.Points)
		case CompositeGlyph:
			count = len(d.Glyphs)
		}
		g.Data = nil
		if h[0] != g || counts[0] != count {
			t.Fatal("glyph header/point count differs")
		}
	})
}

func TestGlyfHeadersRejectInvalidOffsets(t *testing.T) {
	for _, offsets := range [][]uint32{nil, {10, 5}, {0, 1}, {100, 100}} {
		if _, _, err := ParseGlyfHeaders(nil, offsets); err == nil {
			t.Fatalf("accepted offsets %v", offsets)
		}
	}
}
