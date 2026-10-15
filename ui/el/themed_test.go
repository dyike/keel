package el

import (
	"image/color"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func TestThemedDefaultTextColorResolvesAtPaint(t *testing.T) {
	parent := textStyle{size: theme.BodySize}
	plain := Text("plain").Bold().node()
	plain.textStyle = plain.style.text.inherit(parent)
	explicit := Text("transparent").TextColor(color.NRGBA{}).node()
	explicit.textStyle = explicit.style.text.inherit(parent)
	var e engine
	for _, p := range []theme.Palette{theme.Light(), theme.Dark()} {
		func() {
			defer theme.Scope(p)()
			if got := e.label(plain, plain.text).Color; got != p.Text {
				t.Fatalf("default text ignored scoped palette: got %v want %v", got, p.Text)
			}
			if got := e.label(explicit, explicit.text).Color; got != (color.NRGBA{}) {
				t.Fatal("scoped palette replaced explicit transparent text")
			}
		}()
	}
	// Exercise Root's inherited default too: capturing the global color there
	// would pass the isolated label checks above but break a local dark theme.
	p := theme.Current()
	p.Text = color.NRGBA{R: 211, G: 42, B: 97, A: 255}
	var root *RootWidget
	painted := false
	root = Root(ViewFunc(func(cx *Context) Element {
		return cx.Themed(p, ViewFunc(func(*Context) Element {
			label := Text("scoped default").Bold()
			return Div().Child(label).Decorate(func(_ core.C, draw func()) {
				painted = true
				if got := root.e.label(label.node(), "scoped default").Color; got != p.Text {
					t.Fatalf("Root captured a global default: got %v want %v", got, p.Text)
				}
				draw()
			})
		}))
	}))
	uitest.New(root).Frame()
	if !painted {
		t.Fatal("scoped text was not painted")
	}
}

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
