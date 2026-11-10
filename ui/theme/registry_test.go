package theme

import (
	"image/color"
	"slices"
	"testing"
)

func TestParseThemeOverridesBase(t *testing.T) {
	name, p, err := ParseTheme([]byte(`{"name": "Nord", "base": "dark",
		"colors": {"bg": "#2e3440", "PRIMARY": "#88c0d0", "scrim": "#0008", "chart": ["#fff"]}}`))
	if err != nil || name != "Nord" {
		t.Fatal(name, err)
	}
	dark := Dark()
	if p.Bg != RGB(0x2e3440) || p.Primary != RGB(0x88c0d0) || p.Scrim != (color.NRGBA{A: 0x88}) {
		t.Fatalf("overrides: %+v", p)
	}
	if p.Surface != dark.Surface || p.Chart[0] != RGB(0xffffff) || p.Chart[1] != dark.Chart[1] {
		t.Fatal("unlisted colors come from the base")
	}
	for _, bad := range []string{`{"name": "x", "colors": {"nope": "#fff"}}`, `{"name": "x", "colors": {"bg": "#12"}}`,
		`{"name": "x", "base": "sepia"}`, `{"colors": {}}`, `{"name": "x", "colors": {"chart": ["#fff","#fff","#fff","#fff","#fff","#fff","#fff","#fff","#fff"]}}`} {
		if _, _, err := ParseTheme([]byte(bad)); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
}

func TestRegisterAndNamed(t *testing.T) {
	_, p, _ := ParseTheme([]byte(`{"name": "Paper", "colors": {"bg": "#fbf7ef"}}`))
	Register("test-paper", p)
	Register("test-paper", p) // replacing keeps one entry
	if got, ok := Named("test-paper"); !ok || got.Bg != RGB(0xfbf7ef) {
		t.Fatal("named")
	}
	if names := Names(); !slices.Equal(names[:6], []string{"light", "dark", "nord", "paper", "solarized-dark", "high-contrast"}) || slices.Index(names, "test-paper") != len(names)-1 || len(slices.Compact(slices.Sorted(slices.Values(names)))) != len(names) {
		t.Fatalf("names %v", names)
	}
	if _, ok := Named("missing"); ok {
		t.Fatal("missing theme found")
	}
}

func TestDefaultOutlinesAndThemeFileOptIn(t *testing.T) {
	for _, name := range []string{"light", "dark", "nord", "paper", "solarized-dark", "high-contrast", "aurora"} {
		p, ok := Named(name)
		if !ok || p.Outline.A != 0 || p.Border.A == 0 {
			t.Fatalf("%s must have no default outline and retain visible separators", name)
		}
	}
	_, p, err := ParseTheme([]byte(`{"name":"outlined","colors":{"outline":"#123456"}}`))
	if err != nil || p.Outline != RGB(0x123456) {
		t.Fatal("outline opt-in did not load", p.Outline, err)
	}
	old := Current()
	restore := Scope(p)
	if Current().Outline != p.Outline {
		t.Fatal("outline not scoped")
	}
	restore()
	if Current() != old {
		t.Fatal("outline scope leaked")
	}
}
