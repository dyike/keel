package window

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"
	"time"

	"github.com/dyike/keel/gio/io/input"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func testMenuIcon(t *testing.T) *MenuIcon {
	t.Helper()
	art := image.NewNRGBA(image.Rect(10, 20, 42, 52))
	art.SetNRGBA(10, 20, color.NRGBA{R: 200, A: 255})
	icon, err := NewMenuIcon(art, true)
	if err != nil {
		t.Fatal(err)
	}
	art.SetNRGBA(10, 20, color.NRGBA{B: 255, A: 255})
	decoded, err := png.Decode(bytes.NewReader(icon.png))
	if err != nil || color.NRGBAModel.Convert(decoded.At(0, 0)).(color.NRGBA).R != 200 {
		t.Fatal("icon aliases source or lost its nonzero origin", err)
	}
	return icon
}

func TestMenuIconSnapshotsAndUpdates(t *testing.T) {
	icon := testMenuIcon(t)
	copy := icon.Image().(*image.NRGBA)
	copy.SetNRGBA(0, 0, color.NRGBA{})
	if !icon.Template() || icon.pixels.NRGBAAt(0, 0).R != 200 {
		t.Fatal("mutable image snapshot")
	}
	bar, err := NewMenuBar(MenuItem{ID: "run", Title: "Run", Icon: icon})
	if err != nil {
		t.Fatal(err)
	}
	if err := bar.UpdateItem("run", func(item *MenuItem) { item.Icon = new(MenuIcon) }); err == nil {
		t.Fatal("accepted uninitialized icon")
	}
	if bar.Items()[0].Icon != icon || !bytes.Equal(wireMenuItems(bar.Items())[0].PNG, icon.png) {
		t.Fatal("invalid update modified image")
	}
	if _, err := NewMenuBar(MenuItem{Separator: true, Icon: icon}); err == nil {
		t.Fatal("separator icon accepted")
	}
	for _, art := range []image.Image{nil, image.NewNRGBA(image.Rect(0, 0, 0, 1)), image.NewNRGBA(image.Rect(0, 0, 257, 1))} {
		if _, err := NewMenuIcon(art, false); err == nil {
			t.Fatal("invalid icon accepted")
		}
	}
}

func TestWindowMenuAnimationStopsAndHonorsReducedMotion(t *testing.T) {
	bar := useTestMenu(t, MenuItem{ID: "file", Title: "File", Children: []MenuItem{{ID: "run", Title: "Run", Icon: testMenuIcon(t)}}})
	if err := bar.SetWindowAnimation(MenuAnimation{Duration: time.Second, Offset: 6}); err != nil {
		t.Fatal(err)
	}
	old := theme.ReducedMotion
	theme.SetReducedMotion(false)
	defer theme.SetReducedMotion(old)
	w := newWindow(Options{MenuDisplay: MenuDisplayWindow, Content: views(text("content"))})
	now := time.Unix(1000, 0)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; w.layout(gtx) })
	h.Click(12, 12)
	var ops op.Ops
	// Consume wakeups as a real event loop would. Use a separate router so
	// pointer/focus events in the harness cannot mask animation scheduling.
	var router input.Router
	gtx := core.C{Ops: &ops, Now: now, Source: router.Source()}
	if p := w.menuView.opening(gtx, "file"); p != 0 {
		t.Fatal("opening did not start at zero", p)
	}
	if _, pending := router.WakeupTime(); !pending {
		t.Fatal("opening did not request a frame")
	}
	now = now.Add(time.Second / 2)
	h.Frame()
	gtx.Now = now
	if p := w.menuView.opening(gtx, "file"); p <= 0 || p >= 1 {
		t.Fatal("missing intermediate frame", p)
	}
	router.WakeupTime()
	now = now.Add(time.Second / 2)
	h.Frame()
	gtx.Now = now
	if p := w.menuView.opening(gtx, "file"); p != 1 {
		t.Fatal("animation did not settle", p)
	}
	if _, pending := router.WakeupTime(); pending {
		t.Fatal("settled menu keeps requesting frames")
	}
	theme.SetReducedMotion(true)
	w.menuView.animationPath = ""
	if p := w.menuView.opening(gtx, "file"); p != 1 {
		t.Fatal("reduced motion ignored")
	}
	if !w.MenuFeatures().WindowAnimation || !w.MenuFeatures().TemplateIcons {
		t.Fatal("window capabilities")
	}
	for _, animation := range []MenuAnimation{{Duration: -1}, {Offset: float32(math.NaN())}, {Offset: float32(math.Inf(1))}} {
		if err := bar.SetWindowAnimation(animation); err == nil {
			t.Fatal("invalid animation accepted")
		}
	}
}

func TestMenuIconCacheKeepsEnabledAndDisabledVariants(t *testing.T) {
	icon := testMenuIcon(t)
	var view applicationMenuView
	var ops op.Ops
	gtx := core.C{Ops: &ops}
	view.icon(gtx, icon, false)
	view.icon(gtx, icon, true)
	enabled, disabled := menuViewIconKey{icon, false}, menuViewIconKey{icon, true}
	before := view.icons[enabled].image
	allocs := testing.AllocsPerRun(100, func() {
		ops.Reset()
		view.icon(gtx, icon, false)
		view.icon(gtx, icon, true)
	})
	if allocs != 0 || len(view.icons) != 2 || view.icons[enabled].image != before {
		t.Fatalf("shared template icon recolored every frame: allocs=%v variants=%d", allocs, len(view.icons))
	}
	text := theme.Text
	defer func() { theme.Text = text }()
	theme.Text = color.NRGBA{R: 75, G: 130, B: 240, A: 255}
	view.icon(gtx, icon, false)
	if view.icons[enabled].color != theme.Text || view.icons[enabled].image == before || view.icons[disabled].color != theme.Muted {
		t.Fatal("theme changes did not refresh the appropriate icon variant")
	}
}

func TestMenuIconAutomationAndCancellation(t *testing.T) {
	calls := 0
	icon := testMenuIcon(t)
	bar := useTestMenu(t, MenuItem{ID: "file", Title: "File", Children: []MenuItem{
		{ID: "run", Title: "Icon action", Icon: icon, OnSelect: func() { calls++ }},
		{ID: "disabled", Title: "Disabled icon", Icon: icon, Disabled: true},
		{ID: "checked", Title: "Checked icon", Icon: icon, Checked: true},
	}})
	w := openTest(t, Options{MenuDisplay: MenuDisplayWindow, Content: views(text("content"))})
	registerApplicationWindow(w)
	t.Cleanup(func() { unregisterApplicationWindow(w) })
	w.press("f10")
	if e := element(t, w, "Icon action"); e.Role != "menuitem" {
		t.Fatal("menu item hidden from automation", e)
	}
	if e := element(t, w, "Disabled icon"); !e.Disabled {
		t.Fatal("disabled icon lost its semantics")
	}
	if e := element(t, w, "Checked icon"); e.Checked == nil || !*e.Checked {
		t.Fatal("checkmark lost its semantics")
	}
	w.click(element(t, w, "Icon action").center())
	if calls != 1 || len(w.menuView.path) != 0 {
		t.Fatal("icon action or dismissal failed")
	}
	w.press("f10")
	bar.CancelTracking(false)
	deadline := time.Now().Add(time.Second)
	for len(w.menuView.path) > 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
		w.render()
	}
	if len(w.menuView.path) != 0 {
		t.Fatal("CancelTracking did not reset the window menu")
	}
}
