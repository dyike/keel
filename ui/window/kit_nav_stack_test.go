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

func TestNavStackAgentModalAtMobileSizes(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	for _, width := range []int{320, 390, 600} {
		for _, scale := range []float32{1, 1.6} {
			t.Run(fmt.Sprintf("%d/%.1f", width, scale), func(t *testing.T) {
				nav := kit.NavStack()
				draft := kit.Input("Draft")
				draft.SetValue("Preserved")
				nav.Page(kit.NavPage{ID: "home", Title: "Home", Create: func() el.View {
					return el.ViewFunc(func(cx *el.Context) el.Element {
						return el.Div().Gap(12).Child(draft.Render(cx), kit.Button("Open modal", func() { nav.Push("edit") }).Size(44).Render(cx))
					})
				}})
				nav.Page(kit.NavPage{ID: "edit", Title: "Edit this project and keep the draft", Modal: true, Create: func() el.View { return kit.Input("Modal notes") }})
				w := openTest(t, Options{Width: width, Height: 600, TextScale: scale, Content: el.Root(nav)})
				w.click(element(t, w, "Open modal").center())
				close := element(t, w, "Close")
				if close.Width < 44 || close.Height < 44 || close.X < 0 || close.X+close.Width > width {
					t.Fatal("invalid modal close target", close)
				}
				notes := element(t, w, "Modal notes")
				if notes.X < 0 || notes.X+notes.Width > width {
					t.Fatal("input outside modal", notes)
				}
				w.click(notes.center())
				if err := w.typeText("Private"); err != nil {
					t.Fatal(err)
				}
				if draft.Value() != "Preserved" {
					t.Fatal("modal input reached parent")
				}
				if err := w.press("esc"); err != nil {
					t.Fatal(err)
				}
				if len(nav.Value()) != 1 {
					t.Fatal("Escape did not dismiss", nav.Value())
				}
				element(t, w, "Draft")
				w.click(element(t, w, "Open modal").center())
				w.click(element(t, w, "Close").center())
				if len(nav.Value()) != 1 || draft.Value() != "Preserved" {
					t.Fatal("close lost parent")
				}
			})
		}
	}
}
