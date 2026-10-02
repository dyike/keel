package kit

import (
	"slices"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/locale"
)

// Items that do not fit move into the "…" menu, lowest priority first, and
// come back when the bar widens; Left/Right views always stay.
func TestStatusBarOverflowMenu(t *testing.T) {
	var ran []string
	act := func(s string) func() { return func() { ran = append(ran, s) } }
	bar := StatusBar().Left(text("就绪")).
		Add(StatusItem{Label: "main 分支", Priority: 3}, StatusItem{Label: "UTF-8 编码", Action: act("enc"), Priority: 1}).
		AddRight(StatusItem{Label: "行 12，列 4 位置", Priority: 2}, StatusItem{Label: "Go 语言模式", Action: act("lang")})
	width := float32(640)
	h := uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().Items(el.Start).Child(el.Div().W(el.Dp(width)).Child(bar.Render(cx)))
	})))
	for range 3 {
		h.Frame()
	}
	if got := bar.Hidden(); len(got) != 0 {
		t.Fatalf("wide bar hid %v", got)
	}
	width = 260
	for range 4 {
		h.Frame()
	}
	got := bar.Hidden()
	if len(got) == 0 || !slices.Contains(got, "Go 语言模式") || slices.Contains(got, "main 分支") && !slices.Contains(got, "UTF-8 编码") {
		t.Fatalf("narrow bar hid %v", got)
	}
	if !shown(h, "就绪") {
		t.Fatal("a Left view was hidden")
	}
	click(t, h, locale.Current().More)
	h.Frame()
	click(t, h, "Go 语言模式")
	h.Frame()
	if !slices.Equal(ran, []string{"lang"}) {
		t.Fatalf("overflow action ran %v", ran)
	}
	width = 900
	for range 4 {
		h.Frame()
	}
	if got := bar.Hidden(); len(got) != 0 {
		t.Fatalf("widened bar still hides %v", got)
	}
}
