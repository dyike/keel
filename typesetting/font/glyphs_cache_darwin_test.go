package font

import (
	ot "github.com/dyike/keel/typesetting/font/opentype"
	"github.com/dyike/keel/typesetting/font/opentype/tables"
	"os"
	"reflect"
	"testing"
)

func TestLazyVariableFontContoursMatchEager(t *testing.T) {
	file, err := os.Open("/System/Library/PrivateFrameworks/FontServices.framework/Resources/Reserved/PingFangUI.ttc")
	if err != nil {
		t.Skip("PingFang unavailable")
	}
	defer file.Close()
	loaders, err := ot.NewLoaders(file)
	if err != nil {
		t.Fatal(err)
	}
	ft, err := NewFont(loaders[0])
	if err != nil {
		t.Fatal(err)
	}
	if ft.glyfOutlines == nil {
		t.Skip("installed font has no glyf outlines")
	}
	raw, _ := loaders[0].RawTable(ot.MustNewTag("glyf"))
	eager, err := tables.ParseGlyf(raw, ft.glyfOutlines.offsets)
	if err != nil {
		t.Fatal(err)
	}
	old := *ft
	old.glyf, old.glyfOutlines = eager, nil
	for _, weight := range []float32{300, 400, 600, 900} {
		a, b := NewFace(ft), NewFace(&old)
		vars := []Variation{{Tag: ot.MustNewTag("wght"), Value: weight}}
		a.SetVariations(vars)
		b.SetVariations(vars)
		for _, r := range "中文字体粗细测试Aa09" {
			gid, ok := ft.Cmap.Lookup(r)
			if !ok {
				continue
			}
			ga, ea := a.glyphDataFromGlyf(gID(gid))
			gb, eb := b.glyphDataFromGlyf(gID(gid))
			if !reflect.DeepEqual(ga, gb) || !reflect.DeepEqual(ea, eb) {
				t.Fatalf("glyph %c weight %v changed", r, weight)
			}
		}
	}
}
