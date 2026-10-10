package fontscan

import (
	"path/filepath"
	"testing"

	"github.com/go-text/typesetting/font"
)

// macOS keeps PingFang in versioned asset directories; without them CJK fell
// back to Hiragino Sans GB W3. Keel patch.
func TestDarwinAssetFontsResolvePingFang(t *testing.T) {
	if m, _ := filepath.Glob("/System/Library/AssetsV2/com_apple_MobileAsset_Font*/*/AssetData/PingFang.ttc"); len(m) == 0 {
		t.Skip("PingFang is not installed as a font asset")
	}
	fm := NewFontMap(nil)
	if err := fm.UseSystemFonts(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	for _, w := range []font.Weight{font.WeightNormal, font.WeightMedium, font.WeightSemibold} {
		fm.SetQuery(Query{Families: []string{"PingFang SC"}, Aspect: font.Aspect{Weight: w}})
		face := fm.ResolveFace('模')
		if face == nil {
			t.Fatalf("weight %v: no face", w)
		}
		fam, asp := fm.FontMetadata(face.Font)
		if fam != "pingfangsc" || asp.Weight != w {
			t.Fatalf("weight %v: resolved %q %+v", w, fam, asp)
		}
	}
}
