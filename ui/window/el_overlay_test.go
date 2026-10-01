package window

import (
	"gioui.org/f32"
	"github.com/dyike/keel/ui/el"
	"testing"
)

type overlayTestView struct {
	open, removeTrigger  bool
	first, second, under int
	text                 string
}

func (v *overlayTestView) Render(cx *el.Context) el.Element {
	if v.open {
		cx.Overlay("dialog", el.Modal(el.Div().Role("dialog").Name("dialog").W(el.Dp(200)).P(12).Child(
			el.Div().Name("first").ID("first").OnClick(func() { v.first++ }).H(el.Dp(30)),
			el.Input().Name("editor").ID("editor").Bind(&v.text),
			el.Div().Name("last").ID("last").OnClick(func() { v.second++ }).H(el.Dp(30)),
		)).OnDismiss(func() { v.open = false }))
	}
	return el.Div().Child(
		el.Div().ID("trigger").Name("trigger").Hidden(v.removeTrigger).OnClick(func() { v.open = true }).W(el.Dp(80)).H(el.Dp(40)),
		el.Div().Name("under").OnClick(func() { v.under++ }).W(el.Dp(80)).H(el.Dp(40)),
	)
}
func TestElementOverlayModalSnapshotAndKeyboard(t *testing.T) {
	v := &overlayTestView{}
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(v)})
	w.click(element(t, w, "trigger").center())
	for _, e := range w.snapshot() {
		if e.Name == "trigger" || e.Name == "under" {
			t.Fatalf("modal exposed background: %+v", e)
		}
	}
	element(t, w, "dialog")
	for _, chord := range []string{"enter", "tab"} {
		if err := w.press(chord); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.typeText("中文123"); err != nil {
		t.Fatal(err)
	}
	for _, chord := range []string{"tab", "enter", "tab", "enter", "shift+tab", "enter"} {
		if err := w.press(chord); err != nil {
			t.Fatal(err)
		}
	}
	if v.first != 2 || v.second != 2 || v.text != "中文123" {
		t.Fatalf("trap/keyboard: %+v", v)
	}
	if err := w.press("escape"); err != nil {
		t.Fatal(err)
	}
	element(t, w, "trigger")
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if !v.open {
		t.Fatal("focus not restored to trigger")
	}
	w.click(f32.Pt(10, 60))
	if v.open || v.under != 0 {
		t.Fatalf("scrim click leaked: %+v", v)
	}
}

func TestElementOverlayNonmodalSnapshot(t *testing.T) {
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.Overlay("popover", el.Anchored("anchor", el.Div().Role("dialog").Name("popover").W(el.Dp(100)).H(el.Dp(50))))
		return el.Div().Child(el.Div().ID("anchor").Name("anchor").W(el.Dp(80)).H(el.Dp(30)), el.Text("background"))
	}))})
	elements := w.snapshot()
	index := map[string]int{}
	for i, e := range elements {
		index[e.Name] = i
	}
	for _, name := range []string{"anchor", "background", "popover"} {
		element(t, w, name)
	}
	if index["popover"] < index["background"] {
		t.Fatal("overlay precedes main content in snapshot")
	}
}

func TestElementMissingModalAnchorDoesNotHideMain(t *testing.T) {
	calls := 0
	w := openTest(t, Options{Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		cx.Overlay("missing", el.Anchored("absent", el.Div().Name("hidden layer").Size(el.Dp(50))).Modal().OnDismiss(func() { calls++ }))
		return el.Div().Child(el.Text("main"))
	}))})
	element(t, w, "main")
	w.snapshot()
	if calls != 1 {
		t.Fatalf("missing anchor callbacks %d", calls)
	}
	for _, e := range w.snapshot() {
		if e.Name == "hidden layer" {
			t.Fatal("missing anchor drawn")
		}
	}
}
