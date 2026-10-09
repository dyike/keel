package theme

import (
	"path/filepath"
	"testing"

	"gioui.org/font"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"
)

// A font set prepares off the main goroutine while the theme changes on it,
// and once used shapes with its faces.
func TestFontSetPreparesElsewhere(t *testing.T) {
	original := Current()
	defer Apply(original)
	cjk, _ := filepath.Glob("/System/Library/AssetsV2/com_apple_MobileAsset_Font*/*.asset/AssetData/PingFang.ttc")
	if len(cjk) == 0 {
		t.Skip("no PingFang")
	}
	keep := func(f font.Font) bool { return f.Typeface == "PingFang SC" && f.Weight == font.Normal }
	done := make(chan *FontSet)
	go func() {
		set, err := PrepareFontFiles(keep, cjk[0])
		if err != nil {
			t.Error(err)
		}
		done <- set
	}()
	for range 20 {
		Apply(Light())
		Apply(Dark())
	}
	set := <-done
	if set == nil || len(set.faces) != 1 {
		t.Fatalf("set %+v", set)
	}
	old := loaded
	defer func() { loaded = old; Material.Shaper = lazyShaper(fallbackFaces()) }()
	loaded = nil
	set.Use()
	if Material.Shaper != set.shaper {
		t.Fatal("the prepared shaper is not in use")
	}
	sh := Material.Shaper
	sh.LayoutString(text.Parameters{Font: font.Font{Typeface: "PingFang SC"}, PxPerEm: fixed.I(16), MaxWidth: 1000}, "中文")
	n := 0
	for g, ok := sh.NextGlyph(); ok; g, ok = sh.NextGlyph() {
		if g.Advance > 0 {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("%d glyphs shaped", n)
	}
}
