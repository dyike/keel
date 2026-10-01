package theme

import (
	"github.com/dyike/keel/ui/internal/loop"
	"testing"
)

func TestApplyPaletteAndInvalidateAllWindows(t *testing.T) {
	loop.Lock()
	defer loop.Unlock()
	original := Current()
	defer Apply(original)
	if original != Light() {
		t.Fatal("default colors differ from light preset")
	}
	material, shaper := Material, Material.Shaper
	counts := [2]int{}
	first, second := new(int), new(int)
	loop.Register(first, func() { counts[0]++ })
	loop.Register(second, func() { counts[1]++ })
	defer loop.Unregister(first)
	defer loop.Unregister(second)
	for i, palette := range []Palette{Dark(), Light()} {
		before := Revision()
		Apply(palette)
		if Current() != palette || Revision() != before+1 {
			t.Fatal("palette or cache revision did not update")
		}
		if Material != material || Material.Shaper != shaper {
			t.Fatal("theme switch replaced font resources")
		}
		if Material.Fg != palette.Text || Material.Bg != palette.Surface || Material.ContrastBg != palette.Primary || Material.ContrastFg != palette.OnColor {
			t.Fatal("Gio palette is stale")
		}
		if counts != [2]int{i + 1, i + 1} {
			t.Fatalf("windows were not both invalidated: %v", counts)
		}
	}
	custom := Dark()
	custom.Success = RGB(0xabcdef)
	Apply(custom)
	custom.Success = RGB(0)
	if Success != RGB(0xabcdef) || Dark().Success == Success {
		t.Fatal("palette does not have value semantics")
	}
}
