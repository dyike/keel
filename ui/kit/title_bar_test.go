package kit

import (
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

type fakeWindow struct {
	frameless, maximized bool
	calls                []string
}

func (f *fakeWindow) Frameless() bool { return f.frameless }
func (f *fakeWindow) Minimize()       { f.calls = append(f.calls, "min") }
func (f *fakeWindow) ToggleMaximize() { f.maximized = !f.maximized; f.calls = append(f.calls, "max") }
func (f *fakeWindow) Maximized() bool { return f.maximized }
func (f *fakeWindow) Close()          { f.calls = append(f.calls, "close") }

func TestTitleBarButtonsPerPlatform(t *testing.T) {
	for _, goos := range []string{"darwin", "windows"} {
		w := &fakeWindow{frameless: true}
		bar := TitleBar("标题")
		bar.goos = goos
		root := el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Stretch).Child(bar.Render(cx)) }))
		h := uitest.NewFunc(func(gtx core.C) { defer core.SetCurrentWindow(w)(); root.Layout(gtx) })
		closeX, mid := bounds(h, "关闭").Min.X, bounds(h, "标题").Dx()/2 // the bar is named after its title
		if goos == "darwin" && closeX > mid || goos == "windows" && closeX < mid {
			t.Errorf("%s: close at %d, bar middle %d", goos, closeX, mid)
		}
		click(t, h, "最小化")
		click(t, h, "最大化")
		h.Frame()
		click(t, h, "还原")
		click(t, h, "关闭")
		if got := len(w.calls); got != 4 || w.calls[3] != "close" || w.maximized {
			t.Errorf("%s: calls %v maximized %v", goos, w.calls, w.maximized)
		}
	}
}
