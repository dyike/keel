package window

import (
	"fmt"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
)

func TestNavStackAgentNavigationAndRetainedDraft(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	for _, width := range []int{320, 600} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			nav := kit.NavStack()
			draft := kit.Input("Draft")
			nav.Page(kit.NavPage{ID: "home", Title: "Home", Create: func() el.View {
				return el.ViewFunc(func(cx *el.Context) el.Element {
					return el.Div().Gap(8).Child(draft.Render(cx), kit.Button("Open details", func() { nav.Push("detail") }).Render(cx))
				})
			}})
			nav.Page(kit.NavPage{ID: "detail", Title: "Details", Create: func() el.View { return kit.Input("Private notes") }})
			w := openTest(t, Options{Width: width, Height: 600, Content: el.Root(nav)})
			w.click(element(t, w, "Draft").center())
			if err := w.typeText("Preserved"); err != nil {
				t.Fatal(err)
			}
			w.click(element(t, w, "Open details").center())
			back := element(t, w, "Previous page")
			if back.Role != "button" || back.Width < 44 || back.Height < 44 || back.X < 0 || back.X+back.Width > width {
				t.Fatalf("invalid back target: %+v", back)
			}
			for _, e := range w.snapshot() {
				if e.Name == "Draft" {
					t.Fatal("inactive page leaked into snapshot")
				}
			}
			nav.SetDisabled(true)
			if !element(t, w, "Previous page").Disabled {
				t.Fatal("disabled state not exposed")
			}
			w.click(element(t, w, "Previous page").center())
			if len(nav.Value()) != 2 {
				t.Fatal("disabled back activated")
			}
			nav.SetDisabled(false)
			w.click(element(t, w, "Previous page").center())
			if draft.Value() != "Preserved" {
				t.Fatal("draft lost", draft.Value())
			}
			element(t, w, "Draft")
		})
	}
}

func TestNavStackAgentKeyboardBack(t *testing.T) {
	nav := kit.NavStack().Page(kit.NavPage{ID: "home", Title: "Home"}).Page(kit.NavPage{ID: "detail", Title: "Detail"})
	nav.Push("detail")
	w := openTest(t, Options{Width: 320, Height: 480, Content: el.Root(nav)})
	w.snapshot()
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if len(nav.Value()) != 1 {
		t.Fatal("Tab/Enter did not activate back", nav.Value())
	}
}
