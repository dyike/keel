package kit

import (
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
	"image"
	"testing"
	"time"
)

func TestAppShellLazyPagesAndFade(t *testing.T) {
	creates, changes := 0, 0
	makePage := func() el.View { creates++; return Input("Name") }
	v := AppShell("App").Page(AppPage{ID: "a", Title: "Alpha", Create: makePage}).Page(AppPage{ID: "b", Title: "Beta", Create: makePage}).Page(AppPage{ID: "off", Disabled: true}).OnChange(func(string) { changes++ })
	v.Page(AppPage{ID: "a"}).Page(AppPage{})
	if len(v.pages) != 3 || creates != 0 {
		t.Fatal("registration eagerly built pages or accepted duplicate")
	}
	c := &clock{now: time.Unix(100, 0)}
	h := c.harness(func(cx *el.Context) el.Element { return v.Render(cx) })
	h.Frame()
	first := v.page("a").root
	if creates != 1 || v.fade.value != 1 {
		t.Fatal("first frame should not animate", creates, v.fade)
	}
	v.SetValue("off")
	v.SetValue("missing")
	if v.Value() != "a" {
		t.Fatal("invalid selection")
	}
	v.nav.choose("b")
	h.Frame()
	if creates != 2 || changes != 1 || v.fade.value != 0 {
		t.Fatal("switch", creates, changes, v.fade)
	}
	c.advance(h, 75*time.Millisecond)
	if v.fade.value <= 0 || v.fade.value >= 1 {
		t.Fatal("missing interpolation", v.fade)
	}
	v.SetValue("a")
	h.Frame()
	c.advance(h, 200*time.Millisecond)
	if creates != 2 || v.page("a").root != first || changes != 1 || v.fade.value != 1 {
		t.Fatal("retention or settlement")
	}
}

func TestAppShellReducedMotion(t *testing.T) {
	old := theme.ReducedMotion
	defer theme.SetReducedMotion(old)
	theme.SetReducedMotion(true)
	v := AppShell("App").Page(AppPage{ID: "a"}).Page(AppPage{ID: "b"})
	c := &clock{now: time.Unix(100, 0)}
	h := c.harness(func(cx *el.Context) el.Element { return v.Render(cx) })
	v.SetValue("b")
	h.Frame()
	if v.fade.value != 1 {
		t.Fatal("reduced motion animated", v.fade)
	}
}

func TestAppShellRestoresScroll(t *testing.T) {
	var offset float32
	scroll := true
	content := el.ViewFunc(func(cx *el.Context) el.Element {
		if scroll {
			cx.ScrollTo("app-shell/content", 120)
			scroll = false
		}
		offset, _, _ = cx.ScrollState("app-shell/content")
		return el.Div().H(el.Dp(1000)).NoShrink().Child(el.Text("Long page"))
	})
	v := AppShell("App").Transition(0).Page(AppPage{ID: "a", Create: func() el.View { return content }}).Page(AppPage{ID: "b"})
	c := &clock{now: time.Unix(100, 0)}
	h := c.harness(func(cx *el.Context) el.Element { return v.Render(cx) })
	// The first frame establishes the scroll extent.
	scroll = true
	h.Frame()
	h.Frame()
	if offset != 120 {
		t.Fatal("did not scroll", offset)
	}
	v.SetValue("b")
	h.Frame()
	v.SetValue("a")
	h.Frame()
	h.Frame()
	if offset != 120 {
		t.Fatal("scroll was reset", offset)
	}
}

func TestAppShellResponsivePreference(t *testing.T) {
	v := AppShell("App").Page(AppPage{ID: "a", Title: "Alpha"})
	root := el.Root(v)
	width := 800
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(width, 500); root.Layout(gtx) })
	if v.Collapsed() {
		t.Fatal("wide window collapsed")
	}
	width = 320
	h.Frame()
	if !v.Collapsed() || v.nav.collapsible {
		t.Fatal("narrow window did not force icon rail")
	}
	width = 800
	h.Frame()
	if v.Collapsed() {
		t.Fatal("automatic collapse overwrote user preference")
	}
	v.SetCollapsed(true)
	h.Frame()
	width = 320
	h.Frame()
	width = 800
	h.Frame()
	if !v.Collapsed() {
		t.Fatal("user collapse lost")
	}
	v.SetCollapsed(false)
	v.CollapseBelow(0)
	width = 320
	h.Frame()
	if v.Collapsed() {
		t.Fatal("disabled automatic collapsing")
	}
}

func TestAppShellDesktopWindowControls(t *testing.T) {
	var windows map[string]image.Rectangle
	for _, goos := range []string{"darwin", "windows", "linux"} {
		t.Run(goos, func(t *testing.T) {
			w := &fakeWindow{frameless: true}
			shell := AppShell("Workspace").Page(AppPage{ID: "home", Title: "Home"})
			shell.chrome.goos = goos
			root := el.Root(shell)
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Constraints.Max = image.Pt(900, 500)
				defer core.SetCurrentWindow(w)()
				root.Layout(gtx)
			})
			rects := map[string]image.Rectangle{}
			for _, name := range []string{"关闭", "最小化", "最大化"} {
				rects[name] = bounds(h, name)
				if rects[name].Empty() {
					t.Fatal("missing window control", name)
				}
			}
			if goos == "windows" {
				windows = rects
			}
			if goos == "linux" {
				for name, b := range rects {
					if b != windows[name] {
						t.Fatal("Windows/Linux geometry differs", name, b, windows[name])
					}
				}
			}
			if goos == "darwin" && rects["关闭"].Min.X >= 220 {
				t.Fatal("traffic lights not in sidebar")
			}
			click(t, h, "最小化")
			click(t, h, "最大化")
			h.Frame()
			click(t, h, "还原")
			click(t, h, "关闭")
			if len(w.calls) != 4 || w.maximized {
				t.Fatal("window actions", w.calls)
			}
			if goos == "darwin" {
				shell.SetCollapsed(true)
				h.Frame()
				if bounds(h, "关闭") != rects["关闭"] {
					t.Fatal("sidebar collapse moved traffic lights", bounds(h, "关闭"), rects["关闭"])
				}
			}
		})
	}
}

func TestAppShellActionRegions(t *testing.T) {
	w := &fakeWindow{frameless: true}
	shell := AppShell("App").SidebarHeader(Button("New", nil)).Header(Button("Model", nil)).Actions(Button("Action", nil)).Content(Button("Body", nil))
	root := el.Root(shell)
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints.Max = image.Pt(1000, 600)
		defer core.SetCurrentWindow(w)()
		root.Layout(gtx)
	})
	for _, name := range []string{"New", "Model", "Action", "Body", "收起侧栏", "关闭"} {
		b := bounds(h, name)
		x, y := center(b)
		if act, ok := h.Router.ActionAt(f32.Pt(x, y)); ok {
			t.Errorf("%s has native action %v %v", name, act, b)
		}
	}
	for range 3 {
		for _, name := range []string{"New", "Model", "Action", "Body"} {
			click(t, h, name)
			click(t, h, name)
		}
	}
	if len(w.calls) != 0 {
		t.Fatal("buttons invoked window controls", w.calls)
	}
}

func TestAppShellSidebarResize(t *testing.T) {
	shell := AppShell("App").SidebarWidth(248).SidebarLimits(200, 420)
	root := el.Root(shell)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(1000, 600); root.Layout(gtx) })
	x, y := center(bounds(h, "调整大小"))
	h.Drag(x, y, x+60, y)
	if shell.SidebarSize() != 308 {
		t.Fatalf("drag width = %v", shell.SidebarSize())
	}
	shell.SetCollapsed(true)
	h.Frame()
	shell.SetCollapsed(false)
	h.Frame()
	if shell.SidebarSize() != 308 {
		t.Fatal("collapse lost dragged width")
	}
	x, y = center(bounds(h, "调整大小"))
	h.Drag(x, y, x+500, y)
	if shell.SidebarSize() != 420 {
		t.Fatalf("upper limit = %v", shell.SidebarSize())
	}
	x, y = center(bounds(h, "调整大小"))
	h.Drag(x, y, x-500, y)
	if shell.SidebarSize() != 200 {
		t.Fatalf("lower limit = %v", shell.SidebarSize())
	}
}

type nativeControlsWindow struct{ fakeWindow }

func (*nativeControlsWindow) NativeTitleBarInset() float32 { return 80 }

func TestNativeControlsNotDuplicated(t *testing.T) {
	for _, shellMode := range []bool{false, true} {
		w := &nativeControlsWindow{fakeWindow: fakeWindow{frameless: true}}
		var view el.View
		if shellMode {
			s := AppShell("App").SidebarHeader(Button("New", nil))
			s.chrome.goos = "darwin"
			view = s
		} else {
			b := TitleBar("App").Leading(Button("New", nil))
			b.goos = "darwin"
			view = b
		}
		root := el.Root(view)
		h := uitest.NewFunc(func(gtx core.C) { defer core.SetCurrentWindow(w)(); root.Layout(gtx) })
		for _, name := range []string{"关闭", "最小化", "最大化"} {
			if !bounds(h, name).Empty() {
				t.Fatalf("native controls duplicated: %s", name)
			}
		}
		if bounds(h, "New").Min.X < 80 {
			t.Fatal("toolbar overlaps system controls")
		}
	}
}

func TestAppShellFullyHiddenSidebar(t *testing.T) {
	shell := AppShell("App").SidebarWidth(280).HideSidebarOnCollapse(true).
		Sidebar(Button("Navigation", nil)).Footer(Button("Profile", nil)).Content(Button("Content", nil))
	root := el.Root(shell)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(1000, 600); root.Layout(gtx) })
	expanded := bounds(h, "Content")
	click(t, h, "收起侧栏")
	h.Frame()
	for _, name := range []string{"Navigation", "Profile", "调整大小"} {
		if !bounds(h, name).Empty() {
			t.Fatalf("hidden sidebar exposes %s", name)
		}
	}
	if b := bounds(h, "Content"); b.Min.X != 0 || b.Dx() <= expanded.Dx() {
		t.Fatalf("content did not reclaim sidebar: %v -> %v", expanded, b)
	}
	click(t, h, "展开侧栏")
	h.Frame()
	if bounds(h, "Navigation").Empty() || bounds(h, "Profile").Empty() || shell.SidebarSize() != 280 {
		t.Fatal("sidebar did not restore")
	}
	shell.HideSidebarOnCollapse(false)
	shell.SetCollapsed(true)
	h.Frame()
	if bounds(h, "Navigation").Empty() || bounds(h, "Content").Min.X != 56 {
		t.Fatal("default icon rail not preserved")
	}
}

func TestAppShellCollapsedToolbarKeepsControlWidths(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		shell := AppShell("App").HideSidebarOnCollapse(hidden).
			SidebarHeader(Button("", nil).Name("New conversation").Icon(IconPlus).Variant(ButtonGhost))
		shell.chrome.goos = "linux"
		root := el.Root(shell)
		h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(1000, 600); root.Layout(gtx) })
		expandedToggle, expandedNew := bounds(h, "收起侧栏"), bounds(h, "New conversation")
		shell.SetCollapsed(true)
		h.Frame()
		toggle, newButton := bounds(h, "展开侧栏"), bounds(h, "New conversation")
		if toggle.Dx() != expandedToggle.Dx() || toggle.Dx() != 32 || newButton.Dx() != expandedNew.Dx() || newButton.Empty() {
			t.Fatalf("hidden=%v: controls shrank: %v / %v -> %v / %v", hidden, expandedToggle, expandedNew, toggle, newButton)
		}
		if toggle.Max.X > newButton.Min.X {
			t.Fatalf("overlapping controls: %v / %v", toggle, newButton)
		}
	}
}
