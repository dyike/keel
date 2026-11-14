package fontscan

import (
	"testing"

	"github.com/dyike/keel/typesetting/font"
	"github.com/dyike/keel/typesetting/language"
)

// Keel patch test: the ASCII table, the per-query hash and the early return
// of an unchanged query never change what ResolveFace answers.
func TestResolveFaceCachesMatchUncachedMap(t *testing.T) {
	cacheDir := t.TempDir()
	cached := NewFontMap(nil)
	if err := cached.UseSystemFonts(cacheDir); err != nil {
		t.Skip("no system fonts:", err)
	}
	// Keep the same loading history: last-resort fallback may use the first
	// loaded face, which is not necessarily the same in a newly
	// constructed map. Clear only the query caches in the reference map.
	uncached := NewFontMap(nil)
	if err := uncached.UseSystemFonts(cacheDir); err != nil {
		t.Fatal(err)
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
				uncached.SetQuery(q)
				uncached.SetScript(s)
				cached.SetQuery(q)
				cached.SetQuery(q) // unchanged: must keep working
				cached.SetScript(s)
				for _, r := range runes {
					uncached.lru.Clear()
					uncached.resetRuneCaches()
					got, want := cached.ResolveFace(r), uncached.ResolveFace(r)
					if (got == nil) != (want == nil) || got != nil && cached.FontLocation(got.Font) != uncached.FontLocation(want.Font) {
						t.Fatalf("query %v %v script %v rune %q: cached and uncached maps disagree", q.Families, q.Aspect, s, r)
					}
				}
			}
		}
	}
}

// Keel patch test: SetQuery keeps its own copy of the families. Gio's text
// shaper parses typefaces into a buffer it reuses for the next one.
func TestSetQueryCopiesFamilies(t *testing.T) {
	fm := NewFontMap(nil)
	if err := fm.UseSystemFonts(t.TempDir()); err != nil {
		t.Skip("no system fonts:", err)
	}
	families := []string{"Menlo"}
	fm.SetQuery(Query{Families: families})
	menlo := fm.ResolveFace('a')
	families[0] = "Times New Roman"
	fm.SetQuery(Query{Families: families})
	times := fm.ResolveFace('a')
	if menlo == nil || times == nil {
		t.Skip("Menlo or Times New Roman missing")
	}
	if fm.FontLocation(menlo.Font) == fm.FontLocation(times.Font) {
		t.Fatal("a query changed in place kept the previous query's face")
	}
}
