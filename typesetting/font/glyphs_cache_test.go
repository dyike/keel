package font

import (
	"bytes"
	"reflect"
	"sync"
	"testing"

	ot "github.com/dyike/keel/typesetting/font/opentype"
	"github.com/dyike/keel/typesetting/font/opentype/tables"
	"golang.org/x/image/font/gofont/goregular"
)

func TestLazyContoursMatchEagerOutlines(t *testing.T) {
	ld, err := ot.NewLoader(bytes.NewReader(goregular.TTF))
	if err != nil {
		t.Fatal(err)
	}
	ft, err := NewFont(ld)
	if err != nil {
		t.Fatal(err)
	}
	if ft.glyfOutlines == nil || len(ft.glyfOutlines.decoded) != 0 {
		t.Fatal("font retained eager contour data")
	}
	raw, _ := ld.RawTable(ot.MustNewTag("glyf"))
	eager, err := tables.ParseGlyf(raw, ft.glyfOutlines.offsets)
	if err != nil {
		t.Fatal(err)
	}
	old := *ft
	old.glyf, old.glyfOutlines = eager, nil
	headers, counts, err := tables.ParseGlyfHeaders(raw, ft.glyfOutlines.offsets)
	if err != nil {
		t.Fatal(err)
	}
	a, b := NewFace(ft), NewFace(&old)
	for i := range eager {
		header := eager[i]
		header.Data = nil
		if headers[i] != header || counts[i] != pointNumbersCount(eager[i]) {
			t.Fatalf("header/point count changed for glyph %d", i)
		}
		if ft.glyf[i].Data != nil {
			t.Fatal("header retained contours")
		}
		ga, ea := a.glyphDataFromGlyf(gID(i))
		gb, eb := b.glyphDataFromGlyf(gID(i))
		if !reflect.DeepEqual(ga, gb) || !reflect.DeepEqual(ea, eb) {
			t.Fatalf("outline changed for glyph %d", i)
		}
		if xa, ok := a.getExtentsFromGlyf(gID(i)); ok {
			xb, _ := b.getExtentsFromGlyf(gID(i))
			if xa != xb {
				t.Fatalf("metrics changed for glyph %d", i)
			}
		}
	}
	// Multiple Faces share the Font's cache without sharing Face scratch state.
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range eager {
				ft.glyfOutlines.glyph(tables.GlyphID(i))
			}
		}()
	}
	wg.Wait()
}

func BenchmarkGlyfInitialization(b *testing.B) {
	ld, err := ot.NewLoader(bytes.NewReader(goregular.TTF))
	if err != nil {
		b.Fatal(err)
	}
	ft, err := NewFont(ld)
	if err != nil {
		b.Fatal(err)
	}
	raw, _ := ld.RawTable(ot.MustNewTag("glyf"))
	for _, headers := range []bool{false, true} {
		name := "eager"
		if headers {
			name = "headers"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if headers {
					_, _, err = tables.ParseGlyfHeaders(raw, ft.glyfOutlines.offsets)
				} else {
					_, err = tables.ParseGlyf(raw, ft.glyfOutlines.offsets)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
