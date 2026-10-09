package window

import (
	"bytes"
	"image/color"
	"image/png"
	"runtime"
	"testing"

	"gioui.org/f32"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestTitleBarAreaDoubleClickAndDisabledOwner(t *testing.T) {
	bar := kit.TitleBar("title").Leading(kit.Button("left", nil)).Trailing(kit.Button("right", nil))
	disabled := false
	w := openTest(t, Options{Width: 600, Height: 200, Frameless: true, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(bar.Render(cx)) }))})
	w.render()
	a := w.titleArea
	if a[2] <= 0 || a[3] <= 0 {
		t.Fatal("title region not registered", a)
	}
	for _, name := range []string{"关闭", "最小化", "最大化", "left", "right"} {
		p := element(t, w, name).center()
		if p.X >= a[0] && p.X < a[0]+a[2] && p.Y >= a[1] && p.Y < a[1]+a[3] {
			t.Fatal("control in title drag region", name)
		}
	}
	p := f32.Pt(a[0]+a[2]/2, a[1]+a[3]/2)
	w.click(p)
	w.click(p)
	if !w.Maximized() {
		t.Fatal("headless double click did not maximize")
	}
	disabled = true
	w.render()
	if w.titleArea != [4]float32{} {
		t.Fatal("disabled region remained registered")
	}
}

func TestTitleBarInactivePixels(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS traffic lights")
	}
	w := openTest(t, Options{Width: 400, Height: 160, Frameless: true, Content: el.Root(kit.TitleBar("title"))})
	at := element(t, w, "关闭").center()
	pixel := func() color.NRGBA {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(int(at.X), int(at.Y))).(color.NRGBA)
	}
	active := pixel()
	w.setFocused(false)
	w.render()
	if w.Focused() || pixel() == active || pixel() != theme.Border {
		t.Fatal("inactive traffic light did not dim", pixel(), active)
	}
	w.setFocused(true)
	w.render()
	if !w.Focused() || pixel() != active {
		t.Fatal("active look did not restore")
	}
}
