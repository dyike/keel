package el

import (
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

// A themed subtree renders and paints with its palette; the rest of the
// window and the global palette stay as they were.
func TestThemedScopesThePalette(t *testing.T) {
	old := theme.Current()
	dark := theme.Dark()
	var inside, outside, painted [2]bool
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		return Div().Child(
			cx.Themed(dark, ViewFunc(func(*Context) Element {
				inside[0] = theme.Surface == dark.Surface
				return Div().Size(Dp(10)).Decorate(func(_ core.C, draw func()) { painted[0] = theme.Surface == dark.Surface; draw() })
			})),
			ViewFunc(func(*Context) Element {
				outside[0] = theme.Surface == old.Surface
				return Div().Size(Dp(10)).Decorate(func(_ core.C, draw func()) { painted[1] = theme.Surface == old.Surface; draw() })
			}).Render(cx),
		)
	})))
	h.Frame()
	if !inside[0] || !outside[0] || !painted[0] || !painted[1] || theme.Current() != old {
		t.Fatalf("render inside %v outside %v, paint %v, global kept %v", inside[0], outside[0], painted, theme.Current() == old)
	}
}
