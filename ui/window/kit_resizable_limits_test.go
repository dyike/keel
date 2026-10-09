package window

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestResizableMaximumAgentAndKeyboard(t *testing.T) {
	calls := 0
	split := kit.Resizable(nil, nil).Min(40, 40).Max(180, 240).OnChange(func(float32) { calls++ })
	w := openTest(t, Options{Width: 400, Height: 200, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().W(el.Dp(400)).H(el.Dp(200)).Items(el.Stretch).Child(split.Render(cx))
	}))})
	w.snapshot()
	w.snapshot()
	w.press("Tab")
	w.press(string(key.NameHome))
	if split.Value() != 154 {
		t.Fatal("second maximum", split.Value())
	}
	e := element(t, w, locale.Current().Resize)
	if e.Value != "154" {
		t.Fatal("semantic size", e)
	}
	w.press(string(key.NameEnd))
	if split.Value() != 180 || calls != 2 {
		t.Fatal("first maximum", split.Value(), calls)
	}
	split.Max(160, 0)
	w.snapshot()
	if split.Value() != 160 || calls != 2 {
		t.Fatal("configuration callback", split.Value(), calls)
	}
	split.SetDisabled(true)
	w.snapshot()
	w.press(string(key.NameHome))
	if split.Value() != 160 || calls != 2 {
		t.Fatal("disabled resized")
	}
}
