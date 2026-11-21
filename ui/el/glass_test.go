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

type fakeNativeGlassWindow struct {
	enabled bool
	regions []core.GlassRegion
}

func (f *fakeNativeGlassWindow) Frameless() bool                 { return false }
func (f *fakeNativeGlassWindow) Focused() bool                   { return true }
func (f *fakeNativeGlassWindow) TitleBarArea(_, _, _, _ float32) {}
func (f *fakeNativeGlassWindow) Minimize()                       {}
func (f *fakeNativeGlassWindow) ToggleMaximize()                 {}
func (f *fakeNativeGlassWindow) Maximized() bool                 { return false }
func (f *fakeNativeGlassWindow) Close()                          {}
func (f *fakeNativeGlassWindow) NativeGlassEnabled() bool        { return f.enabled }
func (f *fakeNativeGlassWindow) AddGlassRegion(r core.GlassRegion) {
	f.regions = append(f.regions, r)
}

func TestGlassRegistersNativeRegionOnTransparentBackdrop(t *testing.T) {
	win := &fakeNativeGlassWindow{enabled: true}
	restore := core.SetCurrentWindow(win)
	defer restore()

	root := Root(ViewFunc(func(*Context) Element {
		return Div().Bg(color.NRGBA{}).Pl(10).Pt(20).Child(
			Div().W(Dp(120)).H(Dp(36)).Rounded(18).
				GlassStyle(GlassProminent).GlassTint(theme.Primary).GlassMerge(12).
				OnClick(func() {}).
				Child(Text("Pill")),
		)
	}))
	var ops op.Ops
	gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 2, PxPerSp: 2}}
	gtx.Constraints.Max = image.Pt(400, 200)
	root.Layout(gtx)

	if len(win.regions) != 1 {
		t.Fatalf("regions = %d, want 1", len(win.regions))
	}
	got := win.regions[0]
	if got.X != 10 || got.Y != 20 || got.Width != 120 || got.Height != 36 {
		t.Fatalf("rect = (%v,%v %vx%v), want (10,20 120x36)", got.X, got.Y, got.Width, got.Height)
	}
	if got.CornerRadius != 18 || got.Tint != theme.Primary || !got.Interactive || got.Spacing != 12 {
		t.Fatalf("region = %+v, unexpected fields", got)
	}
}

func TestGlassFallsBackInsideOpaqueAncestor(t *testing.T) {
	win := &fakeNativeGlassWindow{enabled: true}
	restore := core.SetCurrentWindow(win)
	defer restore()

	root := Root(ViewFunc(func(*Context) Element {
		return Div().Bg(theme.Surface).P(16).Child(
			Div().W(Dp(120)).H(Dp(36)).Rounded(18).Glass().Child(Text("Fallback")),
		)
	}))
	var ops op.Ops
	gtx := core.C{Ops: &ops, Metric: unit.Metric{PxPerDp: 2, PxPerSp: 2}}
	gtx.Constraints.Max = image.Pt(400, 200)
	root.Layout(gtx)

	if len(win.regions) != 0 {
		t.Fatalf("regions = %d, want 0 inside opaque ancestor", len(win.regions))
	}
}
