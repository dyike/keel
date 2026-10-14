package fontscan

import (
	"os"
	"testing"

	"github.com/dyike/keel/typesetting/font"
)

func TestNamedInstancesShareImmutableTables(t *testing.T) {
	const path = "/System/Library/PrivateFrameworks/FontServices.framework/Resources/Reserved/PingFangUI.ttc"
	if _, err := os.Stat(path); err != nil {
		t.Skip("variable PingFang unavailable")
	}
	fm := NewFontMap(nil)
	if err := fm.UseSystemFonts(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	resolve := func(weight font.Weight) *font.Face {
		fm.SetQuery(Query{Families: []string{"PingFang SC"}, Aspect: font.Aspect{Weight: weight}})
		return fm.ResolveFace('中')
	}
	regular := resolve(font.WeightNormal)
	bold := resolve(font.WeightSemibold)
	if regular == nil || bold == nil {
		t.Fatal("missing face")
	}
	if regular.Font == bold.Font {
		t.Fatal("native font identities must remain independent")
	}
	a, b := fm.FontLocation(regular.Font), fm.FontLocation(bold.Font)
	if a.File != b.File || a.Index != b.Index || a.Instance == b.Instance {
		t.Fatalf("expected two instances of the same font: %+v %+v", a, b)
	}
	if len(regular.GSUB.Lookups) == 0 || len(bold.GSUB.Lookups) == 0 {
		t.Fatal("missing layout tables")
	}
	if &regular.GSUB.Lookups[0] != &bold.GSUB.Lookups[0] {
		t.Fatal("named instances parsed duplicate immutable tables")
	}
	_, wa := fm.FontMetadata(regular.Font)
	_, wb := fm.FontMetadata(bold.Font)
	if wa.Weight != font.WeightNormal || wb.Weight != font.WeightSemibold {
		t.Fatal("shared tables corrupted weight metadata")
	}
}
