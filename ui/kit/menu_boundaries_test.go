package kit

import (
	"fmt"
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestMenuLongListKeyboardRevealsAndRunsLastItem(t *testing.T) {
	ran := -1
	m := Menu()
	for i := range 80 {
		i := i
		m.Item(fmt.Sprintf("Command %d", i), "", func() { ran = i })
		if i%10 == 0 {
			m.Separator()
		}
	}
	m.Trigger(Button("Open", m.Toggle))
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(5).Items(el.Start).Child(m.Render(cx)) }))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(200, 160); root.Layout(gtx) })
	click(t, h, "Open")
	h.Frame()
	h.Key(key.NameEnd, 0)
	h.Frame()
	last := bounds(h, "Command 79")
	if last.Empty() || last.Min.Y < 0 || last.Max.Y > 160 || last.Max.X > 200 {
		t.Fatalf("last item not revealed: %v", last)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if ran != 79 || m.Value() {
		t.Fatalf("action %d open %v", ran, m.Value())
	}
}
func TestMenuFirstEnabledItemStartsVisible(t *testing.T) {
	m := Menu()
	for i := range 50 {
		label := fmt.Sprintf("Row %d", i)
		m.Item(label, "", nil)
		m.SetItemDisabled(label, i < 45)
	}
	m.Trigger(Button("Open", m.Toggle))
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return m.Render(cx) })))
	click(t, h, "Open")
	for range 4 {
		h.Frame()
	}
	r := bounds(h, "Row 45")
	if r.Empty() || r.Min.Y < 0 || r.Max.Y > 300 {
		t.Fatalf("initial enabled row invisible %v", r)
	}
}
func TestMenuNestedCloseDisableAndInvalidSubtrees(t *testing.T) {
	deep := Menu().Item("Deep item", "", nil)
	sub := Menu().Sub("Deeper", deep)
	m := Menu().Sub("Sub", sub).Item("Other", "", nil)
	m.Sub("nil", nil).Sub("self", m)
	deep.Sub("cycle", m)
	m.Sub("repeat", sub)
	if len(m.items) != 2 || len(deep.items) != 1 {
		t.Fatal("invalid subtree accepted")
	}
	m.Trigger(Button("Open", m.Toggle))
	disabled := false
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(m.Render(cx)) })))
	click(t, h, "Open")
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	if !shown(h, "Deep item") {
		t.Fatal("nested open")
	}
	m.SetValue(false)
	h.Frame()
	m.SetValue(true)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	h.Frame()
	if shown(h, "Deep item") || sub.openSub != -1 {
		t.Fatal("stale deep state")
	}
	m.SetItemDisabled("Sub", true)
	h.Frame()
	if shown(h, "Deeper") || m.openSub != -1 {
		t.Fatal("disabled open subtree retained")
	}
	m.SetItemDisabled("Sub", false)
	h.Frame()
	if shown(h, "Deeper") {
		t.Fatal("re-enabled subtree reopened")
	}
	disabled = true
	h.Frame()
	h.Frame()
	if m.Value() {
		t.Fatal("disabled ancestor kept menu")
	}
	disabled = false
	m.SetValue(true)
	h.Frame()
	m.SetDisabled(true)
	h.Frame()
	m.Toggle()
	h.Frame()
	if m.Value() || shown(h, "Other") {
		t.Fatal("disabled menu opened")
	}
}
func TestMenuContextTriggerAndOutsideDismiss(t *testing.T) {
	m := Menu().Item("Action", "", nil)
	m.Trigger(el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().W(el.Dp(100)).H(el.Dp(40)).Role("button").Name("Target").Focusable(true).OnContextMenu(m.Toggle).Child(el.Text("Target"))
	}))
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(10).Items(el.Start).Child(m.Render(cx)) })))
	x, y := center(bounds(h, "Target"))
	p := f32.Pt(x, y)
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonSecondary, Position: p}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p})
	h.Frame()
	h.Frame()
	if !m.Value() || !shown(h, "Action") {
		t.Fatal("context did not open")
	}
	h.Click(390, 290)
	h.Frame()
	if m.Value() {
		t.Fatal("outside did not close")
	}
}
