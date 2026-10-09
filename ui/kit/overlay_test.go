package kit

import (
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// clock drives an el.Root with an injected frame time.
type clock struct{ now time.Time }

func (c *clock) harness(fn viewFunc) *uitest.Harness {
	root := el.Root(fn)
	return uitest.NewFunc(func(gtx core.C) { gtx.Now = c.now; root.Layout(gtx) })
}
func (c *clock) advance(h *uitest.Harness, d time.Duration) { c.now = c.now.Add(d); h.Frame() }

func center(b image.Rectangle) (float32, float32) {
	p := b.Min.Add(b.Size().Div(2))
	return float32(p.X), float32(p.Y)
}

func shown(h *uitest.Harness, name string) bool { _, ok := node(h, name); return ok }

func TestPopoverToggleOutsideAndEscape(t *testing.T) {
	var changes []bool
	p := Popover(viewFunc(func(*el.Context) el.Element { return el.Text("筛选内容") })).OnChange(func(b bool) { changes = append(changes, b) })
	p.Trigger(Button("筛选", p.Toggle))
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(20).Gap(80).Child(p.Render(cx), el.Text("外部"))
	})))
	click(t, h, "筛选")
	h.Frame()
	if !p.Value() || !shown(h, "筛选内容") {
		t.Fatal("trigger did not open")
	}
	click(t, h, "筛选") // the trigger itself toggles; it is not an outside press
	h.Frame()
	if p.Value() || shown(h, "筛选内容") {
		t.Fatal("trigger did not close")
	}
	click(t, h, "筛选")
	click(t, h, "外部")
	h.Frame()
	if p.Value() {
		t.Fatal("outside press did not close")
	}
	click(t, h, "筛选")
	h.Key(key.NameEscape, 0)
	h.Frame()
	if p.Value() {
		t.Fatal("Esc did not close")
	}
	p.SetValue(true)
	h.Frame()
	if len(changes) != 6 || !shown(h, "筛选内容") {
		t.Fatalf("changes %v: SetValue must not notify", changes)
	}
}

func TestTooltipDelayFocusAndEscape(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	tip := WithTooltip(Button("复制", nil), "复制到剪贴板")
	h := c.harness(func(cx *el.Context) el.Element {
		return el.Div().P(60).Gap(40).Child(tip.Render(cx), Button("其他", nil).Render(cx))
	})
	x, y := center(bounds(h, "复制"))
	h.Move(x, y)
	c.advance(h, 300*time.Millisecond)
	if shown(h, "复制到剪贴板") {
		t.Fatal("shown before the delay")
	}
	c.advance(h, 300*time.Millisecond)
	h.Frame()
	if !shown(h, "复制到剪贴板") {
		t.Fatal("not shown after the delay")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if shown(h, "复制到剪贴板") {
		t.Fatal("Esc did not hide")
	}
	ox, oy := center(bounds(h, "其他"))
	h.Move(ox, oy)
	h.Frame()
	click(t, h, "复制") // a press focuses the button: keyboard focus shows at once
	h.Move(ox, oy)
	h.Frame()
	if !shown(h, "复制到剪贴板") {
		t.Fatal("focus inside did not show the tooltip")
	}
	if n, ok := semanticNode(h, "tooltip"); !ok || n.Desc.Label != "复制到剪贴板" {
		t.Fatalf("tooltip semantics: %+v", n.Desc)
	}
}

func TestHoverCardDelaysAndCardHover(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	card := HoverCard(viewFunc(func(*el.Context) el.Element { return el.Text("张三").Name("张三") }),
		viewFunc(func(*el.Context) el.Element { return el.Text("产品经理，北京") }))
	h := c.harness(func(cx *el.Context) el.Element {
		return el.Div().P(20).Gap(200).Child(card.Render(cx), el.Text("远处"))
	})
	x, y := center(bounds(h, "张三"))
	h.Move(x, y)
	c.advance(h, 600*time.Millisecond)
	if shown(h, "产品经理，北京") {
		t.Fatal("opened early")
	}
	c.advance(h, 200*time.Millisecond)
	h.Frame()
	if !shown(h, "产品经理，北京") {
		t.Fatal("did not open")
	}
	cx, cy := center(bounds(h, "产品经理，北京"))
	h.Move(cx, cy) // onto the card: stays open past the close delay
	c.advance(h, time.Second)
	h.Frame()
	if !shown(h, "产品经理，北京") {
		t.Fatal("card hover did not keep it open")
	}
	fx, fy := center(bounds(h, "远处"))
	h.Move(fx, fy)
	c.advance(h, 200*time.Millisecond)
	if !shown(h, "产品经理，北京") {
		t.Fatal("closed before the delay")
	}
	c.advance(h, 200*time.Millisecond)
	h.Frame()
	if shown(h, "产品经理，北京") {
		t.Fatal("did not close")
	}
}

func TestMenuKeyboardSubmenuAndActions(t *testing.T) {
	var ran []string
	do := func(s string) func() { return func() { ran = append(ran, s) } }
	export := Menu().Item("PDF", "", do("pdf")).Item("CSV", "", do("csv"))
	m := Menu().Item("复制", "mod+c", do("copy")).Item("禁用项", "", do("disabled")).Separator().Sub("导出", export).Item("删除", "", do("delete"))
	m.SetItemDisabled("禁用项", true)
	m.Trigger(Button("更多", m.Toggle))
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().P(20).Child(m.Render(cx)) })))
	click(t, h, "更多")
	h.Frame()
	if n, ok := node(h, "菜单"); !ok || n.Desc.Description != "menu" {
		t.Fatal("menu not open")
	}
	// Focus starts on 复制; ↓ skips the disabled item and the separator.
	for _, k := range []key.Name{key.NameDownArrow, key.NameRightArrow} {
		h.Key(k, 0)
		h.Frame()
	}
	if !shown(h, "PDF") {
		t.Fatal("→ did not open the submenu")
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if len(ran) != 1 || ran[0] != "csv" || m.Value() || shown(h, "PDF") {
		t.Fatalf("submenu action: ran=%v open=%v", ran, m.Value())
	}
	click(t, h, "更多")
	h.Frame()
	h.Key(key.NameEnd, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if ran[len(ran)-1] != "delete" {
		t.Fatalf("End/Enter ran %v", ran)
	}
	click(t, h, "更多")
	h.Frame()
	click(t, h, "禁用项")
	h.Frame()
	if len(ran) != 2 || !m.Value() {
		t.Fatal("disabled item ran or closed the menu")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if m.Value() {
		t.Fatal("Esc did not close")
	}
}

func TestMenuSubmenuEscapeClosesOneLevel(t *testing.T) {
	sub := Menu().Item("PDF", "", nil)
	m := Menu().Sub("导出", sub)
	m.Trigger(Button("更多", m.Toggle))
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().P(20).Child(m.Render(cx)) })))
	click(t, h, "更多")
	click(t, h, "导出")
	h.Frame()
	if !shown(h, "PDF") {
		t.Fatal("click did not open the submenu")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if shown(h, "PDF") || !m.Value() {
		t.Fatal("Esc should close only the submenu")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if m.Value() {
		t.Fatal("second Esc should close the menu")
	}
}
