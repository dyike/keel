package fontscan

import (
	"os"
	"testing"

	"github.com/dyike/keel/typesetting/font"
	ot "github.com/dyike/keel/typesetting/font/opentype"
)

// Isolate the variable fixture from installed static PingFang asset faces.
// Normal system fallback tests still exercise the full system font index.
func pingFangUIFixture(t *testing.T) *FontMap {
	t.Helper()
	const path = "/System/Library/PrivateFrameworks/FontServices.framework/Resources/Reserved/PingFangUI.ttc"
	f, err := os.Open(path)
	if err != nil {
		t.Skip("variable PingFang unavailable")
	}
	defer f.Close()
	loaders, err := ot.NewLoaders(f)
	if err != nil {
		t.Fatal(err)
	}
	fm := NewFontMap(nil)
	variable := false
	for i, ld := range loaders {
		fp, _, err := newFootprintFromLoader(ld, false, scanBuffer{})
		if err != nil {
			t.Fatal(err)
		}
		if fp.Family != "pingfangsc" {
			continue
		}
		if _, err := ld.RawTable(ot.MustNewTag("fvar")); err != nil {
			continue
		}
		variable = true
		fp.Location = Location{File: path, Index: uint16(i)}
		fm.appendFootprints(fp)
		fm.appendFootprints(namedWeightFootprints(ld, fp)...)
	}
	if !variable {
		t.Skip("reserved PingFang SC fixture is not variable on this OS")
	}
	return fm
}

func TestNamedInstancesShareImmutableTables(t *testing.T) {
	fm := pingFangUIFixture(t)
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
