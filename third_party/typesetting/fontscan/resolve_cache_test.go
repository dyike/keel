package fontscan

import (
	"testing"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/language"
)

// Keel patch test: the ASCII table, the per-query hash and the early return
// of an unchanged query never change what ResolveFace answers.
func TestResolveFaceCachesMatchFreshMaps(t *testing.T) {
	cacheDir := t.TempDir()
	cached := NewFontMap(nil)
	if err := cached.UseSystemFonts(cacheDir); err != nil {
		t.Skip("no system fonts:", err)
	}
	queries := []Query{
		{Families: []string{"Menlo", "monospace"}},
		{Families: []string{"Helvetica", "sans-serif"}},
		{Families: []string{"Menlo", "monospace"}, Aspect: font.Aspect{Weight: font.WeightBold}},
		{Families: []string{"serif"}},
	}
	scripts := []language.Script{language.Latin, language.Han, language.Arabic}
	runes := []rune("Aa0~ 你好مر")
	for round := 0; round < 2; round++ {
		for _, q := range queries {
			for _, s := range scripts {
				fresh := NewFontMap(nil)
				if err := fresh.UseSystemFonts(cacheDir); err != nil {
					t.Fatal(err)
				}
				fresh.SetQuery(q)
				fresh.SetScript(s)
				cached.SetQuery(q)
				cached.SetQuery(q) // unchanged: must keep working
				cached.SetScript(s)
				for _, r := range runes {
					got, want := cached.ResolveFace(r), fresh.ResolveFace(r)
					if (got == nil) != (want == nil) || got != nil && cached.FontLocation(got.Font) != fresh.FontLocation(want.Font) {
						t.Fatalf("query %v %v script %v rune %q: cached and fresh maps disagree", q.Families, q.Aspect, s, r)
					}
				}
			}
		}
	}
}
