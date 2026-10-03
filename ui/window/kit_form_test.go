package window

import (
	"testing"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func kitPage(views ...el.View) Options {
	return Options{Width: 480, Height: 400, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		box := el.Div().P(16).Gap(12).Items(el.Start)
		for _, v := range views {
			box.Child(v.Render(cx))
		}
		return box
	}))}
}

func TestKitChoiceControlsSnapshot(t *testing.T) {
	c := kit.Checkbox("同意条款", true)
	s := kit.Switch("通知", false)
	r := kit.RadioGroup("付款", "转账", "现金")
	r.SetValue("现金")
	tg := kit.Toggle("加粗", true)
	g := kit.ToggleGroup("左", "右")
	w := openTest(t, kitPage(c, s, r, tg, g))
	for name, want := range map[string]string{"同意条款": "checkbox", "通知": "switch", "现金": "radio", "加粗": "toggle", "左": "toggle"} {
		if e := element(t, w, name); e.Role != want {
			t.Errorf("%s: role %q want %q", name, e.Role, want)
		}
	}
	if e := element(t, w, "现金"); e.Checked == nil || !*e.Checked {
		t.Fatal("radio checked state")
	}
	w.click(element(t, w, "通知").center())
	if !s.Value() {
		t.Fatal("agent click did not toggle the switch")
	}
}

func TestKitTextControlsSnapshot(t *testing.T) {
	in := kit.Input("客户")
	in.SetValue("华东")
	n := kit.NumberInput("数量")
	o := kit.OtpInput("验证码", 4)
	tf := kit.TimeField("开始")
	tf.SetValue(9 * time.Hour)
	cb := kit.Combobox("城市", "北京")
	w := openTest(t, kitPage(in, n, o, tf, cb))
	for _, tc := range []struct{ name, role, value string }{
		{"客户", "textbox", "华东"}, {"数量", "textbox", "0"}, {"验证码", "textbox", ""}, {"开始", "textbox", "09:00"}, {"城市", "combobox", ""},
	} {
		var found bool
		for _, e := range w.snapshot() {
			if e.Name == tc.name && e.Role == tc.role && e.Value == tc.value {
				found = true
			}
		}
		if !found {
			t.Errorf("missing %s %q value %q", tc.role, tc.name, tc.value)
		}
	}
	w.click(element(t, w, "增加 数量").center())
	if n.Value() != 1 {
		t.Fatal("agent click did not step the number")
	}
}

func TestKitPickersSnapshot(t *testing.T) {
	sel := kit.Select("状态", "待付款", "已发货")
	sl := kit.Slider("音量", 0, 10)
	cal := kit.Calendar()
	cal.SetValue(time.Date(2026, 10, 8, 0, 0, 0, 0, time.Local), time.Time{})
	dp := kit.DatePicker("交货日期")
	rt := kit.Rating("评分", 5)
	st := kit.Stepper("填写", "完成")
	w := openTest(t, Options{Width: 640, Height: 600, Content: kitPage(sel, sl, dp, rt, st, cal).Content})
	w.click(element(t, w, "状态").center())
	w.click(element(t, w, "已发货").center())
	if sel.Value() != "已发货" {
		t.Fatalf("select via agent: %q", sel.Value())
	}
	if e := element(t, w, "2026-10-08"); e.Role != "gridcell" || e.Selected == nil || !*e.Selected {
		t.Fatalf("calendar cell: %+v", e)
	}
	for name, want := range map[string]string{"音量": "slider", "评分": "slider", "填写": "step", "交货日期": "button"} {
		found := false
		for _, e := range w.snapshot() {
			if e.Name == name && e.Role == want {
				found = true
			}
		}
		if !found {
			t.Errorf("missing %s %q", want, name)
		}
	}
}

func TestKitFormSnapshot(t *testing.T) {
	name := kit.Input("")
	f := kit.Form().Field("姓名", name, func() string { return kit.Required(name.Value(), "请填写姓名") })
	var cx *el.Context
	w := openTest(t, Options{Width: 480, Height: 300, Content: el.Root(el.ViewFunc(func(c *el.Context) el.Element {
		cx = c
		return el.Div().P(16).Gap(12).Child(f.Render(c), kit.Button("提交", func() { f.Validate(cx) }).Render(c))
	}))})
	if e := element(t, w, "姓名"); e.Role != "text" && e.Role != "textbox" {
		t.Fatalf("form label: %+v", e)
	}
	w.click(element(t, w, "提交").center())
	element(t, w, "请填写姓名")
}

func TestKitVerticalStepperSnapshot(t *testing.T) {
	st := kit.Stepper().Vertical().Size(32).Navigable()
	st.SetEntries(kit.StepperItem{Label: "Locked", Icon: kit.IconLock, Disabled: true}, kit.StepperItem{Label: "Ready", Icon: kit.IconInbox})
	st.SetValue(2)
	w := openTest(t, Options{Width: 320, Height: 300, Content: kitPage(st).Content})
	locked, ready := element(t, w, "Locked"), element(t, w, "Ready")
	if locked.Role != "step" || locked.Value != "done" || !locked.Disabled || ready.Disabled || ready.Y <= locked.Y {
		t.Fatalf("steps: %+v %+v", locked, ready)
	}
	w.click(locked.center())
	if st.Value() != 2 {
		t.Fatal("disabled step changed")
	}
	w.click(ready.center())
	if st.Value() != 1 || element(t, w, "Ready").Value != "current" {
		t.Fatal("step did not become current")
	}
}

func TestKitDescriptionListColumnsSnapshot(t *testing.T) {
	calls := 0
	list := kit.DescriptionList().Columns(2).Vertical().Bordered(true).
		Item("Order", "SO-123").Item("Customer", "Ada").Separator().
		ItemView("Actions", kit.Button("Open order", func() { calls++ })).Span(2)
	w := openTest(t, Options{Width: 420, Height: 300, Content: kitPage(list).Content})
	a, b := element(t, w, "Order：SO-123"), element(t, w, "Customer：Ada")
	if a.Role != "text" || b.X <= a.X || a.Y != b.Y {
		t.Fatalf("columns: %+v %+v", a, b)
	}
	w.click(element(t, w, "Open order").center())
	if calls != 1 {
		t.Fatal("rich item click failed")
	}
	list.Columns(1)
	if element(t, w, "Customer：Ada").Y <= element(t, w, "Order：SO-123").Y {
		t.Fatal("reflow failed")
	}
}
