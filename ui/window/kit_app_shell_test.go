package window

import (
	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestAppShellNavigationAndRetainedInput(t *testing.T) {
	input := kit.Input("Project name")

	calls := 0
	shell := kit.AppShell("Workspace").Transition(0).
		Page(kit.AppPage{ID: "projects", Title: "Projects", Create: func() el.View { return input }}).
		Page(kit.AppPage{ID: "settings", Title: "Settings", Create: func() el.View { return kit.Button("Apply", func() { calls++ }) }})
	w := openTest(t, Options{Width: 800, Height: 500, Content: el.Root(shell)})
	clickNav := func(name string) {
		for _, e := range w.snapshot() {
			if e.Name == name && e.Role == "link" {
				w.click(e.center())
				return
			}
		}
		t.Fatal("missing navigation link", name)
	}
	w.click(element(t, w, "Project name").center())
	if err := w.typeText("Kept draft"); err != nil {
		t.Fatal(err)
	}
	clickNav("Settings")
	if shell.Value() != "settings" {
		t.Fatal("navigation did not select")
	}
	for _, e := range w.snapshot() {
		if e.Name == "Project name" {
			t.Fatal("inactive page exposed to agent")
		}
	}
	w.click(element(t, w, "Apply").center())
	if calls != 1 {
		t.Fatal("new page not interactive")
	}
	clickNav("Projects")
	if input.Value() != "Kept draft" {
		t.Fatal("draft lost")
	}
	element(t, w, "Project name")
}

func TestAppShellCustomSlotsAndDragRegion(t *testing.T) {
	calls := 0
	button := func(name string) el.View { return kit.Button(name, func() { calls++ }) }
	composer := kit.Input("Message")
	shell := kit.AppShell("Workspace").SidebarHeader(button("New chat")).
		Sidebar(button("Custom sidebar")).Header(button("Model")).Actions(button("Share")).
		Content(el.ViewFunc(func(*el.Context) el.Element { return el.Div().HFull().Child(el.Text("Custom content")) })).
		Footer(button("Profile")).ContentFooter(composer)
	w := openTest(t, Options{Width: 1000, Height: 600, Frameless: true, Content: el.Root(shell)})
	w.render()
	area := w.titleArea
	if area[2] <= 0 || area[3] != kit.AppShellTitleBarHeight {
		t.Fatal("missing title drag region", area)
	}
	for _, name := range []string{"New chat", "Custom sidebar", "Model", "Share", "Profile", "Message", "关闭", "最小化", "最大化"} {
		p := element(t, w, name).center()
		if p.X >= area[0] && p.X < area[0]+area[2] && p.Y >= area[1] && p.Y < area[1]+area[3] {
			t.Fatal("interactive slot in drag region", name)
		}
	}
	for _, name := range []string{"New chat", "Custom sidebar", "Model", "Share", "Profile"} {
		w.click(element(t, w, name).center())
	}
	if calls != 5 {
		t.Fatal("custom controls not interactive", calls)
	}
	w.click(element(t, w, "Message").center())
	if err := w.typeText("draft"); err != nil {
		t.Fatal(err)
	}
	if composer.Value() != "draft" {
		t.Fatal("composer did not receive input")
	}
	if element(t, w, "Message").center().Y < 500 {
		t.Fatal("composer not pinned to bottom")
	}
	element(t, w, "Custom content")
}

func TestAppShellDoubleClickOnlyOnBlankTitlebar(t *testing.T) {
	model := kit.Select("", "Assistant", "Writer").Hint("Model")
	model.SetValue("Assistant")
	mode := kit.Select("", "Rail", "Hidden").Hint("Collapse mode")
	mode.SetValue("Hidden")
	shell := kit.AppShell("Workspace").SidebarHeader(kit.Button("New", nil)).
		Header(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(180)).Child(model.Render(cx)) })).
		Actions(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(140)).Child(mode.Render(cx)) })).Content(kit.Button("Body", nil))
	w := openTest(t, Options{Width: 1100, Height: 600, Frameless: true, Content: el.Root(shell)})
	for _, name := range []string{"New", "Model", "Collapse mode", "Body"} {
		point := element(t, w, name).center()
		w.click(point)
		w.click(point)
		if w.Maximized() {
			t.Fatalf("double-click on %s maximized window", name)
		}
		if err := w.press("Escape"); err != nil {
			t.Fatal(err)
		}
	}
	a := w.titleArea
	if a[2] <= 0 {
		t.Fatal("missing blank titlebar region")
	}
	p := f32.Pt(a[0]+a[2]/2, a[1]+a[3]/2)
	w.click(p)
	w.click(p)
	if !w.Maximized() {
		t.Fatal("blank titlebar double-click did not maximize")
	}
	// A fresh input router starts a separate double-click sequence.
	restored := openTest(t, Options{Width: 1100, Height: 600, Frameless: true, Content: el.Root(shell)})
	restored.maximized = true
	restored.render()
	a = restored.titleArea
	p = f32.Pt(a[0]+a[2]/2, a[1]+a[3]/2)
	restored.click(p)
	restored.click(p)
	if restored.Maximized() {
		t.Fatal("blank titlebar double-click did not restore")
	}
}
