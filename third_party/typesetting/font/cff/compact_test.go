package cff

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/go-text/typesetting/font/opentype"
	"github.com/go-text/typesetting/font/opentype/tables"
)

func TestCompactCFF(t *testing.T) {
	// Header, Name INDEX, Top DICT (CharStrings at byte 21), empty
	// String/GlobalSubr indexes, then three valid endchar programs.
	src := []byte{1, 0, 4, 4, 0, 1, 1, 1, 2, 'T', 0, 1, 1, 1, 3, 160, 17, 0, 0, 0, 0, 0, 3, 1, 1, 2, 3, 4, 14, 14, 14}
	compareCFF(t, src)
	if path := os.Getenv("KEEL_CFF_TEST_FILE"); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		compareCFF(t, b)
	}
}

func compareCFF(t *testing.T, src []byte) {
	t.Helper()
	a, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ParseCompact(src)
	if err != nil {
		t.Fatal(err)
	}
	if a.GlyphCount() != b.GlyphCount() || b.Charstrings != nil {
		t.Fatal("incorrect compact representation")
	}
	for i, s := range a.Charstrings {
		if !bytes.Equal(s, b.charstring(i)) {
			t.Fatal("charstring differs", i)
		}
		ag, ab, ae := a.LoadGlyph(tables.GlyphID(i))
		bg, bb, be := b.LoadGlyph(tables.GlyphID(i))
		if (ae == nil) != (be == nil) || !reflect.DeepEqual(ag, bg) || ab != bb {
			t.Fatal("outline differs", i)
		}
		if a.GlyphName(opentype.GID(i)) != b.GlyphName(opentype.GID(i)) {
			t.Fatal("glyph name differs", i)
		}
	}
}

func TestCompactIndexValidation(t *testing.T) {
	for _, tc := range []struct {
		data   []byte
		header indexStart
	}{
		{[]byte{1, 2, 3, 14, 14}, indexStart{2, 1}},
		{[]byte{1, 0}, indexStart{1, 1}},
		{[]byte{1, 8, 14}, indexStart{1, 1}},
		{[]byte{1, 3, 2, 14, 14}, indexStart{2, 1}},
		{nil, indexStart{0xffffffff, 4}},
		{nil, indexStart{100, 0}},
	} {
		a, ar, ae := parseIndexContent(tc.data, tc.header)
		b, br, be := parseIndexView(tc.data, tc.header)
		if (ae == nil) != (be == nil) || ar != br {
			t.Fatal("validation differs")
		}
		if ae == nil {
			for i := range a {
				if !bytes.Equal(a[i], b.at(i)) {
					t.Fatal("index bytes differ")
				}
			}
		}
	}
}
