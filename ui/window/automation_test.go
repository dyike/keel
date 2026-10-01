package window

import (
	"fmt"
	"image"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func TestAutomationImageRoleAndClick(t *testing.T) {
	clicked := false
	v := &widget.ImageView{Asset: widget.ImageData(image.NewNRGBA(image.Rect(0, 0, 120, 40))), Alt: "产品图", OnClick: func() { clicked = true }}
	w := openTest(t, Options{Content: v})
	e := element(t, w, "产品图")
	if e.Role != "image" || e.Value != "loaded" {
		t.Fatalf("image lost accessible role/state: %+v", e)
	}
	w.click(e.center())
	if !clicked {
		t.Fatal("image click not delivered")
	}
}

// openTest opens a virtual window as automation mode would, without a socket.
func openTest(t *testing.T, o Options) *Window {
	t.Helper()
	w := newWindow(o)
	openVirtual(w, true)
	t.Cleanup(func() { closeVirtual(w) })
	return w
}

func element(t *testing.T, w *Window, name string) Element {
	t.Helper()
	e, err := w.find("", name)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestAutomationScroll(t *testing.T) {
	var rows []core.Widget
	for i := 1; i <= 40; i++ {
		rows = append(rows, widget.Text(fmt.Sprintf("row %d", i)))
	}
	w := openTest(t, Options{Width: 300, Height: 600, Content: layout.Column(rows...)})
	before := element(t, w, "row 12").Y
	w.scroll(element(t, w, "row 1").center(), 300)
	w.snapshot()
	if after := element(t, w, "row 12").Y; after > before-250 {
		t.Fatalf("row 12 moved from y=%d to y=%d; expected about 300dp up", before, after)
	}
	// Scrolled out of view, so no longer listed: only visible elements are.
	for _, e := range w.snapshot() {
		if e.Name == "row 1" {
			t.Fatalf("row 1 is scrolled out of view but still listed at %+v", e)
		}
	}
}

func TestAutomationTabMovesFocus(t *testing.T) {
	a, b := widget.Input("a"), widget.Input("b")
	w := openTest(t, Options{Content: layout.Column(a, b)})
	w.click(element(t, w, "a").center())
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.typeText("x"); err != nil {
		t.Fatal(err)
	}
	if a.Value() != "" || b.Value() != "x" {
		t.Fatalf("a=%q b=%q; Tab should have moved focus to b", a.Value(), b.Value())
	}
}

func TestAutomationCloseFromCallback(t *testing.T) {
	var w *Window
	closed := false
	w = openTest(t, Options{
		Content: widget.Button("close", func() { w.Close() }),
		OnClose: func() { closed = true },
	})
	other := openTest(t, Options{Content: widget.Text("stay")}) // keeps the process alive
	w.click(element(t, w, "close").center())
	if !closed || !w.Closed() || other.Closed() {
		t.Fatalf("closed=%t w.Closed=%t other.Closed=%t", closed, w.Closed(), other.Closed())
	}
}

func TestAutomationDisabledAndChecked(t *testing.T) {
	b := widget.Button("save", nil)
	b.SetDisabled(true)
	w := openTest(t, Options{Content: layout.Column(b, widget.Checkbox("agree", true))})
	if e := element(t, w, "save"); !e.Disabled {
		t.Errorf("disabled button reported enabled: %+v", e)
	}
	if e := element(t, w, "agree"); e.Checked == nil || !*e.Checked {
		t.Errorf("checked box reported %+v", e)
	}
}

// Agent actions must redraw the real windows even when no component callback
// runs: a checkbox without OnChange changes state silently.
func TestAutomationRedrawsRealWindows(t *testing.T) {
	redraws := 0
	key := new(int)
	loop.Register(key, func() { redraws++ })
	defer loop.Unregister(key)
	w := openTest(t, Options{Content: widget.Checkbox("silent", false)})
	redraws = 0
	w.click(element(t, w, "silent").center())
	if redraws == 0 {
		t.Fatal("clicking a checkbox without OnChange did not invalidate real windows")
	}
}

// The first request after launch may be a key press: the shadow has not
// rendered yet, so no handler would receive it.
func TestAutomationShortcutAsFirstRequest(t *testing.T) {
	n := 0
	w := openTest(t, Options{Content: widget.Text("x"), Shortcuts: map[string]func(){"mod+n": func() { n++ }}})
	if err := w.press("mod+n"); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("shortcut fired %d times", n)
	}
}

func TestAutomationSliderAndAccordion(t *testing.T) {
	s := widget.Slider("volume", 0, 100).Step(5)
	a := widget.Accordion().Add("details", widget.Text("inside"))
	w := openTest(t, Options{Content: layout.Column(s, a)})
	e := element(t, w, "volume")
	if e.Role != "slider" || e.Value != "0" {
		t.Fatalf("slider semantics: %+v", e)
	}
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("right"); err != nil {
		t.Fatal(err)
	}
	if s.Value() != 5 || element(t, w, "volume").Value != "5" {
		t.Fatal("Tab did not focus slider or value is stale")
	}
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	e = element(t, w, "details")
	if e.Role != "disclosure" || e.Value != "expanded" || !a.IsOpen(0) {
		t.Fatalf("accordion semantics/keyboard: %+v", e)
	}
	element(t, w, "inside")
	a.SetDisabled(true)
	if !element(t, w, "details").Disabled {
		t.Fatal("disabled header not exposed")
	}
}

func TestAutomationToggleState(t *testing.T) {
	toggle := widget.Toggle("固定工具栏", false)
	w := openTest(t, Options{Content: toggle})
	e := element(t, w, "固定工具栏")
	if e.Role != "toggle" || e.Selected == nil || *e.Selected {
		t.Fatalf("wrong toggle semantics: %+v", e)
	}
	w.click(e.center())
	e = element(t, w, "固定工具栏")
	if e.Selected == nil || !*e.Selected {
		t.Fatal("selected state missing")
	}
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("space"); err != nil {
		t.Fatal(err)
	}
	e = element(t, w, "固定工具栏")
	if e.Selected == nil || *e.Selected {
		t.Fatal("tab/space failed")
	}
	toggle.SetDisabled(true)
	e = element(t, w, "固定工具栏")
	if !e.Disabled {
		t.Fatal("disabled state missing")
	}
}

func TestAutomationExtendedControlStates(t *testing.T) {
	check := widget.Checkbox("全部项目", false)
	check.SetIndeterminate(true)
	check.SetDisabled(true)
	button := widget.Button("上传", nil)
	button.SetLoading(true)
	progress := widget.Progress("处理中")
	progress.SetIndeterminate(true)
	group := widget.ToggleGroup("单选一", "单选二")
	group.SetDisabled(true)
	w := openTest(t, Options{Height: 600, Content: layout.Column(check, button, progress, group)})
	e := element(t, w, "全部项目")
	if e.Role != "checkbox" || e.Value != "mixed" || !e.Disabled {
		t.Fatalf("mixed checkbox missing state: %+v", e)
	}
	if e = element(t, w, "上传"); !e.Disabled {
		t.Fatal("loading button not disabled")
	}
	if e = element(t, w, "处理中"); e.Value != "indeterminate" {
		t.Fatal("progress has fabricated percentage")
	}
	if e = element(t, w, "单选一"); !e.Disabled {
		t.Fatal("group disabled state not propagated")
	}
}

func TestAutomationBadgeValues(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		b           *widget.BadgeView
	}{
		{"150", "99+", widget.Badge(150)}, {"3", "3", widget.Badge(3)}, {"1", "dot", widget.Badge(1).Dot()}, {"2", "icon", widget.Badge(2).Icon(widget.Icon(widget.IconCheck))},
	} {
		t.Run(tc.value, func(t *testing.T) {
			w := openTest(t, Options{Content: tc.b})
			e := element(t, w, tc.name)
			if e.Role != "badge" || e.Value != tc.value {
				t.Fatalf("unexpected badge: %+v", e)
			}
		})
	}
}

func TestAutomationDisabledRadio(t *testing.T) {
	r := widget.RadioGroup("", "立即", "每天")
	w := openTest(t, Options{Content: r})
	r.SetDisabled(true)
	e := element(t, w, "立即")
	if e.Role != "radio" || !e.Disabled {
		t.Fatalf("missing disabled radio: %+v", e)
	}
	w.click(e.center())
	w.press("space")
	if r.Value() != "" {
		t.Fatal("disabled radio selected")
	}
	r.SetDisabled(false)
	e = element(t, w, "立即")
	if e.Disabled {
		t.Fatal("radio remained disabled")
	}
	w.click(e.center())
	if r.Value() != "立即" {
		t.Fatal("radio did not recover")
	}
}

func TestAutomationDisabledSelect(t *testing.T) {
	s := widget.Select("状态", "甲", "乙")
	w := openTest(t, Options{Content: s})
	s.SetDisabled(true)
	e := element(t, w, "状态")
	if e.Role != "select" || !e.Disabled {
		t.Fatalf("missing disabled select: %+v", e)
	}
	w.click(e.center())
	w.press("space")
	w.press("enter")
	for _, e := range w.snapshot() {
		if e.Role == "option" {
			t.Fatal("disabled select opened")
		}
	}
	s.SetDisabled(false)
	e = element(t, w, "状态")
	if e.Disabled {
		t.Fatal("select remained disabled")
	}
	w.click(e.center())
	w.click(element(t, w, "乙").center())
	if s.Value() != "乙" {
		t.Fatal("select did not recover")
	}
}
