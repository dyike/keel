package kit

import (
	"slices"
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func desc(h *uitest.Harness, name string) string { n, _ := node(h, name); return n.Desc.Description }

func TestCheckboxToggleMixedDisabled(t *testing.T) {
	var changes []bool
	c := Checkbox("全选", false).OnChange(func(b bool) { changes = append(changes, b) })
	h := page(c)
	click(t, h, "全选")
	h.Key(key.NameSpace, 0)
	if c.Value() || len(changes) != 2 {
		t.Fatalf("click+space: %v %v", c.Value(), changes)
	}
	c.SetMixed(true)
	h.Frame()
	if desc(h, "全选") != "checkbox:mixed" {
		t.Fatalf("mixed semantics %q", desc(h, "全选"))
	}
	click(t, h, "全选")
	if !c.Value() || len(changes) != 3 || desc(h, "全选") == "checkbox:mixed" {
		t.Fatal("click on mixed should check")
	}
	c.SetDisabled(true)
	h.Frame()
	click(t, h, "全选")
	if !c.Value() || len(changes) != 3 {
		t.Fatal("disabled checkbox toggled")
	}
}

func TestSwitchToggles(t *testing.T) {
	s := Switch("通知", false)
	h := page(s)
	click(t, h, "通知")
	if n, _ := node(h, "通知"); !s.Value() || !n.Desc.Selected {
		t.Fatal("switch did not turn on")
	}
	s.SetValue(false)
	h.Frame()
	if n, _ := node(h, "通知"); n.Desc.Selected {
		t.Fatal("SetValue not shown")
	}
}

func TestRadioGroupArrowsMoveChoice(t *testing.T) {
	var got []string
	r := RadioGroup("付款", "转账", "支票", "现金").OnChange(func(s string) { got = append(got, s) })
	h := page(r)
	click(t, h, "支票")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	h.Key(key.NameDownArrow, 0) // wraps to the first
	h.Frame()
	if r.Value() != "转账" || !slices.Equal(got, []string{"支票", "现金", "转账"}) {
		t.Fatalf("value %q changes %v", r.Value(), got)
	}
	if n, _ := node(h, "转账"); !n.Desc.Selected {
		t.Fatal("selected state missing")
	}
	r.SetDisabled(true)
	h.Frame()
	click(t, h, "支票")
	if r.Value() != "转账" {
		t.Fatal("disabled group changed")
	}
}

func TestToggleAndGroup(t *testing.T) {
	bold := Toggle("加粗", false)
	g := ToggleGroup("左", "中", "右")
	m := ToggleGroup("A", "B").Multiple()
	h := page(bold, g, m)
	click(t, h, "加粗")
	click(t, h, "中")
	click(t, h, "右")
	click(t, h, "A")
	click(t, h, "B")
	if !bold.Value() || !slices.Equal(g.Value(), []string{"右"}) || !slices.Equal(m.Value(), []string{"A", "B"}) {
		t.Fatalf("toggle %v group %v multiple %v", bold.Value(), g.Value(), m.Value())
	}
	click(t, h, "右") // pressing the active option clears a single group
	if len(g.Value()) != 0 {
		t.Fatal("single group did not clear")
	}
}

func TestInputClearErrorAndDisabled(t *testing.T) {
	var changes []string
	in := Input("客户").Placeholder("客户名称").Clearable().OnChange(func(s string) { changes = append(changes, s) })
	h := page(in)
	h.Click(center(bounds(h, "客户名称")))
	h.Type("华东")
	h.Frame() // the clear button shows once the edit reaches the view
	if in.Value() != "华东" || len(changes) != 1 {
		t.Fatalf("typing: %q %v", in.Value(), changes)
	}
	click(t, h, "清空 客户")
	h.Frame()
	if in.Value() != "" || changes[len(changes)-1] != "" {
		t.Fatal("clear did not empty the field")
	}
	in.SetError("必填")
	h.Frame()
	if !shown(h, "必填") {
		t.Fatal("error message missing")
	}
	in.SetDisabled(true)
	h.Frame()
	h.Click(center(bounds(h, "客户名称")))
	h.Type("x")
	if in.Value() != "" {
		t.Fatal("disabled input accepted text")
	}
}

func TestSelectKeyboardAndSearch(t *testing.T) {
	var got []string
	s := Select("状态", "待付款", "已付款", "已发货").OnChange(func(v string) { got = append(got, v) })
	h := page(s)
	clickRole(t, h, "select", "状态")
	h.Frame()
	if _, ok := semanticNode(h, "listbox"); !ok {
		t.Fatal("list did not open")
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if s.Value() != "已付款" || len(got) != 1 {
		t.Fatalf("keyboard choice %q %v", s.Value(), got)
	}
	if _, ok := semanticNode(h, "select:已付款"); !ok {
		t.Fatal("select semantics missing the value")
	}
	f := Select("城市", "北京", "上海", "深圳").Searchable()
	h = page(f)
	clickRole(t, h, "select", "城市")
	h.Frame()
	h.Type("深")
	h.Key(key.NameReturn, 0)
	h.Frame()
	if f.Value() != "深圳" {
		t.Fatalf("search choice %q", f.Value())
	}
}

func TestSliderDragAndKeys(t *testing.T) {
	var last float64
	s := Slider("音量", 0, 100).Step(10).OnChange(func(v float64) { last = v })
	h := uitest.New(el.Embed(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(216)).Child(s.Render(cx))
	})))
	n, _ := semanticNode(h, "slider:0")
	b := n.Desc.Bounds
	y := float32(b.Min.Y + b.Dy()/2)
	h.Drag(float32(b.Min.X+8), y, float32(b.Min.X+8+100), y) // half of 200dp of travel
	if s.Value() != 50 || last != 50 {
		t.Fatalf("drag value %v", s.Value())
	}
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NamePageUp, 0)
	if s.Value() != 100 {
		t.Fatalf("keys value %v", s.Value())
	}
	h.Key(key.NameHome, 0)
	if s.Value() != 0 {
		t.Fatal("Home")
	}
}
