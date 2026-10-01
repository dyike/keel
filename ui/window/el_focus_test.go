package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"image/color"
	"image/png"
	"testing"
)

type focusTestView struct {
	first, second int
	text          string
}

func (v *focusTestView) Render(cx *el.Context) el.Element {
	return el.Div().Gap(8).Child(
		el.Div().ID("first").Name("first").Focusable(true).P(12).OnClick(func() { v.first++ }).Child(el.Text("first")),
		el.Div().Focusable(true).Hidden(true).Child(el.Text("hidden")),
		el.Div().ID("second").Name("second").Focusable(true).P(12).OnClick(func() { v.second++ }).Child(el.Text("second")),
		el.Input().ID("editor").Name("editor").Bind(&v.text),
	)
}
func TestElementNativeKeyboardFocus(t *testing.T) {
	v := &focusTestView{}
	w := openTest(t, Options{Content: el.Embed(v)})
	w.click(element(t, w, "first").center())
	for _, chord := range []string{"space", "tab", "enter", "tab"} {
		if err := w.press(chord); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.typeText("中文123"); err != nil {
		t.Fatal(err)
	}
	if v.first != 2 || v.second != 1 || v.text != "中文123" {
		t.Fatalf("first=%d second=%d input=%q", v.first, v.second, v.text)
	}
	if err := w.press("shift+tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("space"); err != nil {
		t.Fatal(err)
	}
	if v.second != 2 {
		t.Fatal("reverse Tab did not return to second")
	}
}

type focusColorView struct{}

func (focusColorView) Render(cx *el.Context) el.Element {
	return el.Div().Child(cx.Cache("focus-color", func() el.Element {
		return el.Div().ID("color").Name("color").Focusable(true).P(12).
			TextColor(color.NRGBA{B: 255, A: 255}).FocusStyle(func(s *el.Style) { s.TextColor(color.NRGBA{R: 255, A: 255}) }).Child(el.Text("MMMM"))
	}), el.Div().Name("other").Focusable(true).P(12).Child(el.Text("other")))
}
func TestFocusTextColorRestoresCachedTree(t *testing.T) {
	w := openTest(t, Options{Width: 200, Height: 120, Content: el.Root(focusColorView{})})
	check := func(red bool) {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for y := 12; y < 40; y++ {
			for x := 12; x < 80; x++ {
				c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				if red && c.R > 200 && c.G < 80 && c.B < 80 || !red && c.B > 200 && c.R < 80 && c.G < 80 {
					count++
				}
			}
		}
		if count < 10 {
			t.Fatalf("missing text color red=%v: %d pixels", red, count)
		}
	}
	check(false)
	w.click(element(t, w, "color").center())
	check(true)
	w.click(element(t, w, "other").center())
	check(false)
}

type disabledView struct {
	disabled bool
	calls    int
}

func (v *disabledView) Render(cx *el.Context) el.Element {
	return el.Div().Disabled(v.disabled).Child(el.Div().Name("禁用按钮").OnClick(func() { v.calls++ }).Child(el.Text("禁用按钮")), el.Input().Name("禁用输入"))
}
func TestElementDisabledSnapshot(t *testing.T) {
	v := &disabledView{disabled: true}
	w := openTest(t, Options{Content: el.Embed(v)})
	for _, name := range []string{"禁用按钮", "禁用输入"} {
		if !element(t, w, name).Disabled {
			t.Fatal("missing disabled", name)
		}
	}
	w.click(element(t, w, "禁用按钮").center())
	if v.calls != 0 {
		t.Fatal("disabled click")
	}
	v.disabled = false
	if element(t, w, "禁用按钮").Disabled {
		t.Fatal("stale disabled")
	}
	w.click(element(t, w, "禁用按钮").center())
	if v.calls != 1 {
		t.Fatal("restore failed")
	}
}
