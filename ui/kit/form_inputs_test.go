package kit

import (
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
)

func TestNumberInputButtonsTypingAndClamp(t *testing.T) {
	var got []float64
	n := NumberInput("数量").Range(0, 10).Step(2).OnChange(func(x float64) { got = append(got, x) })
	h := page(n)
	click(t, h, "增加 数量")
	click(t, h, "增加 数量")
	if n.Value() != 4 {
		t.Fatalf("buttons: %v", n.Value())
	}
	clickClass(t, h, "Editor", "数量")
	h.Key(key.NameEnd, 0)
	h.Type("5") // "45" is out of range while typing
	h.Key(key.NameReturn, 0)
	h.Frame()
	if n.Value() != 10 || len(got) != 3 {
		t.Fatalf("Enter should clamp 45 to 10: %v %v", n.Value(), got)
	}
	click(t, h, "增加 数量") // disabled at the maximum
	if n.Value() != 10 {
		t.Fatal("stepped past the maximum")
	}
}

func TestOtpInputFillsAndCompletes(t *testing.T) {
	done := ""
	o := OtpInput("验证码", 4).OnComplete(func(s string) { done = s })
	h := page(o)
	clickClass(t, h, "Editor", "验证码")
	h.Type("12a34")
	h.Frame()
	if o.Value() != "1234" || done != "1234" {
		t.Fatalf("value %q done %q", o.Value(), done)
	}
	for _, d := range "1234" {
		if !shown(h, string(d)) {
			t.Fatalf("digit %c not shown in a box", d)
		}
	}
}

func TestTimeFieldParsesAndReverts(t *testing.T) {
	tf := TimeField("开始")
	h := page(tf)
	clickClass(t, h, "Editor", "开始")
	h.Key(key.NameEnd, 0)
	for range 5 {
		h.Key(key.NameDeleteBackward, 0)
	}
	h.Type("930")
	h.Key(key.NameReturn, 0)
	h.Frame()
	if tf.Value() != 9*time.Hour+30*time.Minute || tf.text != "09:30" {
		t.Fatalf("930 → %v %q", tf.Value(), tf.text)
	}
	tf.text = "25:00"
	tf.commit()
	if tf.text != "09:30" {
		t.Fatalf("invalid text should revert, got %q", tf.text)
	}
	for in, want := range map[string]time.Duration{"0:05": 5 * time.Minute, "2359": 23*time.Hour + 59*time.Minute} {
		if d, ok := parseClock(in); !ok || d != want {
			t.Errorf("%s → %v %v", in, d, ok)
		}
	}
	for _, bad := range []string{"", "9", "24:00", "12:3", "ab:cd"} {
		if _, ok := parseClock(bad); ok {
			t.Errorf("%q parsed", bad)
		}
	}
}

func TestComboboxFilterEnterAndRevert(t *testing.T) {
	c := Combobox("城市", "北京", "上海", "深圳")
	h := page(c)
	clickClass(t, h, "Editor", "城市")
	h.Type("深")
	h.Frame()
	if shown(h, "北京") || !shown(h, "深圳") {
		t.Fatal("list not filtered")
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if c.Value() != "深圳" {
		t.Fatalf("Enter took %q", c.Value())
	}
	c.text = "火星"
	c.settle()
	if c.Value() != "深圳" || c.text != "深圳" {
		t.Fatal("unknown text should revert without AllowCustom")
	}
	free := Combobox("标签").AllowCustom()
	free.text = "新标签"
	free.settle()
	if free.Value() != "新标签" {
		t.Fatal("AllowCustom should keep typed text")
	}
}

func TestCalendarKeyboardRangeAndBounds(t *testing.T) {
	c := &clock{now: time.Date(2026, 10, 15, 9, 0, 0, 0, time.Local)}
	var start, end time.Time
	cal := Calendar().Range().Bounds(time.Time{}, time.Date(2026, 10, 25, 0, 0, 0, 0, time.Local)).
		OnChange(func(a, b time.Time) { start, end = a, b })
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().P(10).Child(cal.Render(cx)) })
	click(t, h, "2026-10-20")
	click(t, h, "2026-10-12") // the earlier end: the range is ordered
	if start.Day() != 12 || end.Day() != 20 {
		t.Fatalf("range %v – %v", start, end)
	}
	if n, _ := node(h, "2026-10-16"); !n.Desc.Selected {
		t.Fatal("days between are not marked")
	}
	h.Key(key.NameRightArrow, 0) // focus is on the 12th
	h.Key(key.NamePageDown, 0)
	h.Frame()
	if !shown(h, "2026-10-25") || cal.focus.Day() != 25 || cal.month.Month() != time.October {
		t.Fatalf("keyboard focus %v month %v", cal.focus, cal.month)
	}
	click(t, h, "2026-10-26") // after max: ignored
	if s, _ := cal.Value(); s.Day() != 12 {
		t.Fatal("out-of-bounds day was picked")
	}
}

func TestDatePickerOpensPicksCloses(t *testing.T) {
	c := &clock{now: time.Date(2026, 10, 15, 9, 0, 0, 0, time.Local)}
	d := DatePicker("交货日期").Placeholder("选择日期")
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().P(10).Child(d.Render(cx)) })
	clickClass(t, h, "Button", "交货日期")
	h.Frame()
	h.Key(key.NameRightArrow, 0) // focus starts on today
	h.Key(key.NameReturn, 0)
	h.Frame()
	if s, _ := d.Value(); s.Day() != 16 || d.open {
		t.Fatalf("picked %v open=%v", s, d.open)
	}
	if !shown(h, "2026-10-16") {
		t.Fatal("field does not show the date")
	}
}

func TestRatingClickKeysReadOnly(t *testing.T) {
	r := Rating("评分", 5)
	ro := Rating("平均", 5).ReadOnly()
	ro.SetValue(3)
	h := page(r, ro)
	n, _ := semanticNode(h, "slider:0/5")
	b := n.Desc.Bounds
	h.Click(float32(b.Min.X+24*3+12), float32(b.Min.Y+11)) // the 4th star
	h.Key(key.NameLeftArrow, 0)
	if r.Value() != 3 {
		t.Fatalf("value %d", r.Value())
	}
	if _, ok := semanticNode(h, "slider:3/5"); !ok {
		t.Fatal("semantics")
	}
	ro.set(1)
	if ro.Value() != 3 {
		t.Fatal("read-only rating changed")
	}
}

func TestStepperStatesAndNavigation(t *testing.T) {
	s := Stepper("填写", "确认", "完成").Navigable()
	s.SetValue(2)
	h := page(s)
	if desc(h, "填写") != "step:done" || desc(h, "完成") != "step:current" {
		t.Fatalf("states %q %q", desc(h, "填写"), desc(h, "完成"))
	}
	click(t, h, "填写")
	if s.Value() != 0 {
		t.Fatal("navigable step did not go back")
	}
}

func TestFormValidatesFocusesAndNamesControls(t *testing.T) {
	customer, amount := Input(""), NumberInput("").Range(0, 1e6)
	f := Form().
		Field("客户", customer, func() string { return Required(customer.Value(), "请填写客户") }).
		Field("金额", amount, func() string {
			if amount.Value() <= 0 {
				return "金额必须大于 0"
			}
			return ""
		})
	var ok bool
	submit := Button("创建", nil)
	h := uitest_page(func(cx *el.Context) el.Element {
		submit.onClick = func() { ok = f.Validate(cx) }
		return el.Div().P(10).Gap(10).Child(f.Render(cx), submit.Render(cx))
	})
	if _, named := semanticNode(h, "form"); !named || bounds(h, "客户").Empty() {
		t.Fatal("form row label should name the control")
	}
	click(t, h, "创建")
	h.Frame()
	if ok || !shown(h, "请填写客户") || !shown(h, "金额必须大于 0") {
		t.Fatal("validation messages missing")
	}
	h.Type("华东") // focus moved to the first invalid field
	h.Frame()
	if customer.Value() != "华东" || shown(h, "请填写客户") {
		t.Fatalf("focus/clear: %q", customer.Value())
	}
}
