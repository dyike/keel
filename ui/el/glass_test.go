package el

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

func TestGlassSurfaceStylesPaintInLightAndDarkThemes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pal   theme.Palette
		style GlassStyle
		tint  color.NRGBA
		spec  bool
	}{
		{"light-regular", theme.Light(), GlassRegular, color.NRGBA{}, false},
		{"light-clear", theme.Light(), GlassClear, theme.Primary, true},
		{"light-prominent", theme.Light(), GlassProminent, theme.Primary, true},
		{"dark-regular", theme.Dark(), GlassRegular, color.NRGBA{}, false},
		{"dark-clear", theme.Dark(), GlassClear, color.NRGBA{}, true},
		{"dark-prominent", theme.Dark(), GlassProminent, theme.Success, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			theme.Apply(tc.pal)
			t.Cleanup(func() { theme.Apply(theme.Light()) })
			box := Div().W(Dp(120)).H(Dp(36)).Rounded(18).GlassStyle(tc.style).GlassTint(tc.tint).
				Hover(func(s *Style) { s.GlassBoost(0.45) }).
				Active(func(s *Style) { s.GlassBoost(0.85) }).
				Child(Text("Liquid Glass"))
			if tc.spec {
				box.GlassSpecular(0.3, 0.2)
			}
			root := Root(ViewFunc(func(*Context) Element { return box }))
			var ops op.Ops
			gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 2, PxPerSp: 2}}
			gtx.Constraints.Max = image.Pt(300, 120)
			d := root.Layout(gtx)
			if d.Size != image.Pt(300, 120) {
				t.Fatalf("size = %v", d.Size)
			}
			if box.n.style.glass.Style != tc.style {
				t.Fatalf("glass style = %v, want %v", box.n.style.glass.Style, tc.style)
			}
		})
	}
}
