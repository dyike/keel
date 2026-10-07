package window

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func useTestMenu(t *testing.T, items ...MenuItem) *MenuBar {
	t.Helper()
	bar, err := NewMenuBar(items...)
	if err != nil {
		t.Fatal(err)
	}
	applicationMenu.Lock()
	old, generation := applicationMenu.current, applicationMenu.generation
	applicationMenu.current = bar
	applicationMenu.generation++
	applicationMenu.Unlock()
	t.Cleanup(func() {
		applicationMenu.Lock()
		applicationMenu.current = old
		applicationMenu.generation = generation
		applicationMenu.Unlock()
	})
	return bar
}
func TestWindowMenuMouseKeyboardAndOutsideClick(t *testing.T) {
	calls := 0
	useTestMenu(t, MenuItem{ID: "file", Title: "File", Children: []MenuItem{
		{Title: "Disabled", Disabled: true, OnSelect: func() { t.Fatal("disabled action") }},
		{ID: "action", Title: "Action", Checked: true, OnSelect: func() { calls++ }},
		{ID: "more", Title: "More", Children: []MenuItem{{ID: "nested", Title: "Nested", OnSelect: func() { calls += 10 }}}},
	}})
	w := newWindow(Options{MenuDisplay: MenuDisplayWindow, Content: views(text("content"))})
	h := uitest.NewFunc(w.layout)
	h.Click(12, 12)
	if len(w.menuView.path) != 1 {
		t.Fatal("menu did not open")
	}
	h.Click(30, 76)
	if calls != 1 || len(w.menuView.path) != 0 {
		t.Fatalf("menu click: calls=%d path=%v", calls, w.menuView.path)
	}
	h.Key("F10", 0)
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameReturn, 0)
	if calls != 11 {
		t.Fatalf("nested keyboard menu: %d", calls)
	}
	h.Key("F10", 0)
	h.Click(390, 280)
	if len(w.menuView.path) != 0 {
		t.Fatal("outside click failed to dismiss")
	}
	h.Key("F10", 0)
	h.Key(key.NameEscape, 0)
	if len(w.menuView.path) != 0 {
		t.Fatal("Escape failed to dismiss")
	}
}
func TestWindowMenuCustomEditorRetainsFocus(t *testing.T) {
	received := core.EditAction("")
	tag := new(struct{})
	bar := useTestMenu(t, MenuItem{ID: "edit", Title: "Edit", Children: []MenuItem{{ID: "copy", Title: "Copy", Action: MenuCopy}}})
	w := newWindow(Options{MenuDisplay: MenuDisplayWindow, Content: core.Func(func(gtx core.C) core.D {
		event.Op(gtx.Ops, tag)
		for {
			_, ok := gtx.Event(key.FocusFilter{Target: tag})
			if !ok {
				break
			}
		}
		gtx.Execute(key.FocusCmd{Tag: tag})
		if action, ok := core.NextEditAction(gtx, tag); ok {
			received = action
		}
		return core.D{Size: gtx.Constraints.Max}
	})})
	h := uitest.NewFunc(w.layout)
	h.Frame()
	h.Click(12, 12)
	h.Click(30, 45)
	if received != core.EditCopy || !h.Router.Source().Focused(tag) {
		t.Fatal("menu lost editor focus or edit action", received)
	}
	// Custom callbacks override standard actions.
	if err := bar.UpdateItem("copy", func(item *MenuItem) { item.OnSelect = func() { received = core.EditPaste } }); err != nil {
		t.Fatal(err)
	}
	h.Frame()
	h.Click(12, 12)
	h.Click(30, 45)
	if received != core.EditPaste {
		t.Fatal("custom edit override failed")
	}
}
func TestWindowMenuSelectAllCutUndoRedo(t *testing.T) {
	value := "original"
	useTestMenu(t, MenuItem{ID: "edit", Title: "Edit", Children: []MenuItem{
		{ID: "all", Title: "Select All", Action: MenuSelectAll}, {ID: "cut", Title: "Cut", Action: MenuCut}, {ID: "undo", Title: "Undo", Action: MenuUndo}, {ID: "redo", Title: "Redo", Action: MenuRedo},
	}})
	w := newWindow(Options{MenuDisplay: MenuDisplayWindow, Content: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element { return el.Input().ID("input").Bind(&value) }))})
	h := uitest.NewFunc(w.layout)
	h.Click(55, 70)
	action := func(row int) { h.Click(12, 12); h.Click(30, float32(45+row*32)); h.Frame() }
	action(0)
	action(1)
	if value != "" {
		t.Fatalf("cut failed: %q", value)
	}
	action(2)
	if value != "original" {
		t.Fatalf("undo failed: %q", value)
	}
	action(3)
	if value != "" {
		t.Fatalf("redo failed: %q", value)
	}
}

func TestWindowMenuRichDocumentUndoNotifies(t *testing.T) {
	doc := new(el.InputDocument)
	if err := doc.SetText("original"); err != nil {
		t.Fatal(err)
	}
	changes := 0
	useTestMenu(t, MenuItem{ID: "edit", Title: "Edit", Children: []MenuItem{
		{ID: "all", Title: "All", Action: MenuSelectAll}, {ID: "cut", Title: "Cut", Action: MenuCut}, {ID: "undo", Title: "Undo", Action: MenuUndo}, {ID: "redo", Title: "Redo", Action: MenuRedo},
	}})
	w := newWindow(Options{MenuDisplay: MenuDisplayWindow, Content: el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Input().ID("input").Document(doc).OnChange(func(string) { changes++ })
	}))})
	h := uitest.NewFunc(w.layout)
	h.Click(55, 70)
	action := func(row int) { h.Click(12, 12); h.Click(30, float32(45+row*32)); h.Frame() }
	action(0)
	action(1)
	action(2)
	action(3)
	if doc.Content().Text() != "" || changes != 3 {
		t.Fatalf("rich menu undo/redo: text=%q notifications=%d", doc.Content().Text(), changes)
	}
}
