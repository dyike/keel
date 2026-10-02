package theme

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/dyike/keel/ui/internal/loop"
)

func TestWatchThemesReloadsTheThemeInUse(t *testing.T) {
	saved := Current()
	t.Cleanup(func() { Apply(saved); current = "light" })
	dir := t.TempDir()
	file := filepath.Join(dir, "ocean.json")
	write := func(bg string, mod time.Time) {
		if err := os.WriteFile(file, []byte(`{"name": "test-ocean", "base": "dark", "colors": {"bg": "`+bg+`",
			"bgGradient": {"from": "#001020", "to": "#003050", "angle": 90}}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		os.Chtimes(file, mod, mod)
	}
	write("#001020", time.Unix(1000, 0))
	var got [][]string
	w, err := WatchThemes(dir, time.Hour, func(names []string, err error) { got = append(got, names) })
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()
	if err := Use("test-ocean"); err != nil || Bg != RGB(0x001020) || BgGradient.To != RGB(0x003050) {
		t.Fatalf("use: %v bg %v gradient %+v", err, Bg, BgGradient)
	}

	write("#002030", time.Unix(2000, 0))
	changes, err := w.scan()
	if err != nil || len(changes) != 1 {
		t.Fatalf("scan: %v %v", changes, err)
	}
	loop.Post(func() { w.apply(changes, err, true) })
	loop.Drain()
	if Bg != RGB(0x002030) {
		t.Fatalf("the theme in use was not reapplied: %v", Bg)
	}
	if !slices.Equal(got[len(got)-1], []string{"test-ocean"}) {
		t.Fatalf("onChange got %v", got)
	}
	if changes, _ := w.scan(); len(changes) != 0 {
		t.Fatalf("an unchanged file was reloaded: %v", changes)
	}

	os.WriteFile(filepath.Join(dir, "broken.json"), []byte(`{"name": "x", "colors": {"bg": "nope"}}`), 0o644)
	if _, err := w.scan(); err == nil {
		t.Fatal("a broken theme file reported no error")
	}
	if Use("no-such-theme") == nil {
		t.Fatal("Use accepted an unknown theme")
	}
}

func TestParseThemeGradient(t *testing.T) {
	_, p, err := ParseTheme([]byte(`{"name": "g", "colors": {"primaryGradient": {"from": "#f00", "to": "#00f", "angle": 45}}}`))
	if err != nil || p.PrimaryGradient != (Gradient{From: RGB(0xff0000), To: RGB(0x0000ff), Angle: 45}) || !p.BgGradient.IsZero() {
		t.Fatalf("gradient %+v %v", p.PrimaryGradient, err)
	}
	if _, _, err := ParseTheme([]byte(`{"name": "g", "colors": {"bgGradient": "#fff"}}`)); err == nil {
		t.Fatal("a color string was accepted as a gradient")
	}
}
