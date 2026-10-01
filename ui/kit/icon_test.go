package kit

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestIconBounds(t *testing.T) {
	for _, scale := range []int{1, 2} {
		h := renderView(viewFunc(func(cx *el.Context) el.Element {
			return el.Div().Name("icon").Child(Icon(IconInfo).Size(24).Render(cx))
		}), 40, scale)
		if r := bounds(h, "icon"); r.Dx() != 24*scale || r.Dy() != 24*scale {
			t.Fatalf("%v", r)
		}
	}
}
func TestIconTheme(t *testing.T) {
	v := Icon(IconInfo)
	h := render(viewFunc(func(cx *el.Context) el.Element { return v.Render(cx) }))
	old := theme.Current()
	defer theme.Apply(old)
	theme.Apply(theme.Dark())
	h.Frame()
	if v.color != nil {
		t.Fatal("default color captured")
	}
}
