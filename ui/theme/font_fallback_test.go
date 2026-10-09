package theme

import (
	"strings"
	"testing"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/fontscan"
	"github.com/go-text/typesetting/language"
)

func TestMonoFaceSharesCJKFallback(t *testing.T) {
	fm := fontscan.NewFontMap(nil)
	if err := fm.UseSystemFonts(""); err != nil {
		t.Skipf("system fonts unavailable: %v", err)
	}
	aspect := font.Aspect{Style: font.StyleNormal, Weight: font.WeightNormal, Stretch: font.StretchNormal}
	for _, r := range "中文终端" {
		fm.SetScript(language.Han)
		fm.SetQuery(fontscan.Query{Families: strings.Split(string(Face), ", "), Aspect: aspect})
		body := fm.ResolveFace(r)
		if body == nil {
			t.Skip("no system CJK face")
		}
		if _, ok := body.NominalGlyph(r); !ok {
			t.Skip("system fonts do not cover CJK")
		}
		fm.SetQuery(fontscan.Query{Families: strings.Split(string(MonoFace), ", "), Aspect: aspect})
		mono := fm.ResolveFace(r)
		if mono == nil || mono.Font != body.Font {
			bodyName, _ := fm.FontMetadata(body.Font)
			var monoName string
			if mono != nil {
				monoName, _ = fm.FontMetadata(mono.Font)
			}
			t.Fatalf("%c body=%q mono=%q: CJK fallback loaded a different font", r, bodyName, monoName)
		}
	}
}
