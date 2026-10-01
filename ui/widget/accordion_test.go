package widget

import (
	"image"
	"reflect"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

func namedBounds(h *uitest.Harness, name string) image.Rectangle {
	var found image.Rectangle
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Label == name {
			found = n.Desc.Bounds
			return
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	nodes := h.Router.AppendSemantics(nil)
	if len(nodes) > 0 {
		walk(nodes[0])
	}
	return found
}
func clickNamed(t *testing.T, h *uitest.Harness, name string) {
	t.Helper()
	b := namedBounds(h, name)
	if b.Empty() {
		t.Fatalf("missing %s", name)
	}
	h.Click(float32(b.Min.X+b.Dx()/2), float32(b.Min.Y+b.Dy()/2))
}
func TestAccordionSingleMultipleAndPreservedContent(t *testing.T) {
	input := Input("name")
	var changes [][]int
	a := Accordion().Add("first", input).Add("second", Text("details")).OnChange(func(v []int) { changes = append(changes, v) })
	h := uitest.New(a)
	if !namedBounds(h, "name").Empty() {
		t.Fatal("collapsed input exposed")
	}
	clickNamed(t, h, "first")
	clickNamed(t, h, "name")
	h.Type("saved")
	if input.Value() != "saved" {
		t.Fatal("input did not receive typing")
	}
	clickNamed(t, h, "second")
	if a.IsOpen(0) || !a.IsOpen(1) || !namedBounds(h, "name").Empty() {
		t.Fatal("single mode left previous panel open")
	}
	clickNamed(t, h, "first")
	if input.Value() != "saved" {
		t.Fatal("closing discarded input state")
	}
	a.Multiple()
	clickNamed(t, h, "second")
	if !reflect.DeepEqual(a.OpenIndices(), []int{0, 1}) {
		t.Fatalf("multiple mode %v", a.OpenIndices())
	}
	before := len(changes)
	a.SetOpen(0, false)
	h.Frame()
	if len(changes) != before {
		t.Fatal("setter triggered callback")
	}
	changes[0][0] = 99
	if a.OpenIndices()[0] != 1 {
		t.Fatal("callback leaked internal state")
	}
}
func TestAccordionKeyboardSkipsDisabledAndHidesFocus(t *testing.T) {
	a := Accordion().Add("first", Text("one")).Add("disabled", Text("two")).Add("third", Text("three"))
	a.SetItemDisabled(1, true)
	h := uitest.New(a)
	clickNamed(t, h, "first")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if !a.IsOpen(2) || a.IsOpen(0) || a.IsOpen(1) {
		t.Fatalf("keyboard skipped wrong panel: %v", a.OpenIndices())
	}
	h.Key(key.NameHome, 0)
	h.Key(key.NameSpace, 0)
	if !a.IsOpen(0) || a.IsOpen(2) {
		t.Fatalf("home + space failed: %v", a.OpenIndices())
	}
	clickNamed(t, h, "disabled")
	if a.IsOpen(1) {
		t.Fatal("disabled header activated")
	}
	a.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "third")
	h.Key(key.NameEnd, 0)
	h.Key(key.NameReturn, 0)
	if !a.IsOpen(0) || a.IsOpen(2) {
		t.Fatal("disabled accordion changed")
	}
}

func TestAccordionValueAndDisabledContent(t *testing.T) {
	f := Input("").Hint("内容")
	changes := 0
	a := Accordion().Multiple().Add("第一", f).Add("第二", Text("详情")).OnChange(func([]int) { changes++ })
	a.SetValue([]int{1, 0, 99})
	if !reflect.DeepEqual(a.Value(), []int{0, 1}) {
		t.Fatal(a.Value())
	}
	value := a.Value()
	value[0] = 99
	if a.Value()[0] != 0 {
		t.Fatal("value leaked internal storage")
	}
	a.SetValue([]int{0})
	h := uitest.New(a)
	if changes != 0 {
		t.Fatal("setter notified")
	}
	clickNamed(t, h, "内容")
	h.Type("原值")
	a.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "第二")
	h.Key(key.NameSpace, 0)
	h.Key(key.NameEnd, 0)
	clickNamed(t, h, "内容")
	h.Type("x")
	if !reflect.DeepEqual(a.Value(), []int{0}) || f.Value() != "原值" || changes != 0 {
		t.Fatal("disabled accordion changed")
	}
	if h.Router.Source().Focused(&f.editor) {
		t.Fatal("disabled content retained focus")
	}
	a.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "第二")
	if !reflect.DeepEqual(a.Value(), []int{0, 1}) || changes != 1 {
		t.Fatal("accordion did not recover")
	}
	single := Accordion().Add("一", nil).Add("二", nil)
	single.SetValue([]int{99, 1, 0})
	if !reflect.DeepEqual(single.Value(), []int{1}) {
		t.Fatal("single mode did not take first valid index")
	}
}
