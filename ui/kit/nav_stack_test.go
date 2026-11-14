package kit

import (
	"slices"
	"testing"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/locale"
)

func TestNavStackNavigationRetainsParentAndRecreatesPoppedPage(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	parent := Input("Draft")
	var detail *InputView
	creates, changes := 0, 0
	nav := NavStack()
	nav.Page(NavPage{ID: "home", Title: "Home", Create: func() el.View {
		return el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Gap(8).Child(parent.Render(cx), Button("Details", func() { nav.Push("detail") }).Render(cx))
		})
	}}).Page(NavPage{ID: "detail", Title: "Details", Create: func() el.View { creates++; detail = Input("Notes"); return detail }}).OnChange(func([]string) { changes++ })
	h := uitest.New(el.Root(nav))
	clickClass(t, h, "Editor", "Draft")
	h.Frame()
	h.Type("Kept draft")

	click(t, h, "Details")
	h.Frame()
	if creates != 1 || changes != 1 {
		t.Fatal(creates, changes)
	}
	if _, ok := node(h, "Draft"); ok {
		t.Fatal("inactive parent exposed")
	}
	clickClass(t, h, "Editor", "Notes")
	h.Type("Temporary note")
	click(t, h, "Previous page")
	h.Frame()
	if parent.Value() != "Kept draft" || changes != 2 {
		t.Fatal("parent state lost", parent.Value(), changes, nav.Value())
	}
	click(t, h, "Details")
	h.Frame()
	if creates != 2 || detail.Value() != "" {
		t.Fatal("popped page was not recreated")
	}

}

func TestNavStackAssignmentsAndCallbacks(t *testing.T) {
	nav := NavStack().Page(NavPage{ID: "root"}).Page(NavPage{ID: "detail"}).Page(NavPage{ID: "off", Disabled: true})
	nav.Page(NavPage{ID: "root"}).Page(NavPage{})
	if len(nav.pages) != 3 {
		t.Fatal("duplicate/empty registration accepted")
	}
	changes := 0
	nav.OnChange(func(path []string) { changes++; path[0] = "mutated" })
	nav.Push("detail")
	if !slices.Equal(nav.Value(), []string{"root", "detail"}) {
		t.Fatal("callback aliased state")
	}
	path := nav.Value()
	path[0] = "changed"
	for _, bad := range [][]string{{"root", "missing"}, {"off"}} {
		nav.SetValue(bad)
	}
	if !slices.Equal(nav.Value(), []string{"root", "detail"}) {
		t.Fatal("invalid assignment modified path")
	}
	nav.SetDisabled(true)
	if nav.Pop() || nav.Push("root") {
		t.Fatal("disabled navigation changed path")
	}
	nav.SetValue([]string{"root"})
	if changes != 1 || !slices.Equal(nav.Value(), []string{"root"}) {
		t.Fatal("restoration notified or disabled assignment failed")
	}
	nav.SetDisabled(false)
	if nav.Pop() || nav.Push("missing") || nav.Push("off") {
		t.Fatal("invalid navigation accepted")
	}
	nav.SetValue(nil)
	if len(nav.Value()) != 0 || changes != 1 {
		t.Fatal("clear failed or notified")
	}
}

func TestNavStackScrollAndPrefixRestoration(t *testing.T) {
	offset := float32(0)
	scroll := false
	created := 0
	nav := NavStack().Page(NavPage{ID: "root", Create: func() el.View {
		created++
		return el.ViewFunc(func(cx *el.Context) el.Element {
			if scroll {
				cx.ScrollTo("nav-stack/content", 120)
				scroll = false
			}
			offset, _, _ = cx.ScrollState("nav-stack/content")
			return el.Div().H(el.Dp(1200)).NoShrink().Child(el.Text("Long page"))
		})
	}}).Page(NavPage{ID: "detail"})
	h := uitest.New(el.Root(nav))
	scroll = true
	h.Frame()
	h.Frame()
	if offset != 120 {
		t.Fatal("not scrolled", offset)
	}
	nav.SetValue([]string{"root", "detail"})
	h.Frame()
	nav.SetValue([]string{"root"})
	h.Frame()
	h.Frame()
	if offset != 120 || created != 1 {
		t.Fatal("prefix state lost", offset, created)
	}
}

func TestNavStackAncestorDisabled(t *testing.T) {
	disabled := false
	calls := 0
	nav := NavStack().Page(NavPage{ID: "root"}).Page(NavPage{ID: "detail", Create: func() el.View { return Button("Action", func() { calls++ }) }})
	nav.Push("detail")
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(nav.Render(cx)) })))
	disabled = true
	h.Frame()
	click(t, h, locale.Current().PrevPage)
	click(t, h, "Action")
	h.Frame()
	if len(nav.Value()) != 2 || calls != 0 {
		t.Fatal("disabled ancestor did not block input")
	}
	disabled = false
	h.Frame()
	click(t, h, "Action")
	if calls != 1 {
		t.Fatal("reenabling lost page input")
	}
}

func TestNavStackKeyboardBack(t *testing.T) {
	nav := NavStack().Page(NavPage{ID: "root"}).Page(NavPage{ID: "detail"})
	nav.Push("detail")
	h := uitest.New(el.Root(nav))
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if len(nav.Value()) != 1 {
		t.Fatal("keyboard back did not pop", nav.Value())
	}
}
