package fontscan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dyike/keel/typesetting/font"
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

// Current macOS can place PingFang in FontServices instead of MobileAsset.
func TestDarwinReservedUIFonts(t *testing.T) {
	const path = "/System/Library/PrivateFrameworks/FontServices.framework/Resources/Reserved/PingFangUI.ttc"
	if _, err := os.Stat(path); err != nil {
		t.Skip("reserved PingFang is not installed")
	}
	fm := NewFontMap(nil)
	if err := fm.UseSystemFonts(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	fm.SetQuery(Query{Families: []string{"Helvetica Neue", "PingFang SC"}, Aspect: font.Aspect{Weight: font.WeightNormal}})
	for _, tc := range []struct {
		r      rune
		family string
	}{{'A', "helveticaneue"}, {'中', "pingfangsc"}} {
		face := fm.ResolveFace(tc.r)
		if face == nil {
			t.Fatalf("missing %c", tc.r)
		}
		family, _ := fm.FontMetadata(face.Font)
		if family != tc.family {
			t.Fatalf("%c resolved to %q, want %q", tc.r, family, tc.family)
		}
	}
}

func TestDarwinPingFangNamedWeights(t *testing.T) {
	fm := pingFangUIFixture(t)
	for _, w := range []font.Weight{font.WeightNormal, font.WeightMedium, font.WeightSemibold} {
		fm.SetQuery(Query{Families: []string{"PingFang SC"}, Aspect: font.Aspect{Weight: w}})
		face := fm.ResolveFace('中')
		if face == nil {
			t.Fatal("no Chinese face")
		}
		family, aspect := fm.FontMetadata(face.Font)
		if family != "pingfangsc" || aspect.Weight != w {
			t.Fatalf("want %v: %s %+v", w, family, aspect)
		}
		loc := fm.FontLocation(face.Font)
		if w != font.WeightMedium && (loc.Instance == 0 || len(face.Coords()) == 0) {
			t.Fatalf("weight %v kept base instance: %+v", w, loc)
		}
	}
}
