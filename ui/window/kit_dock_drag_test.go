package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestKitDockDragPreviewAndDrop(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	for _, palette := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(palette) })
		txt := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
		d := kit.Dock(txt("Center")).Panel(kit.DockPanel{ID: "a", Title: "A", View: txt("Body A")}, kit.DockLeft).
			Panel(kit.DockPanel{ID: "b", Title: "B", View: txt("Body B")}, kit.DockRight)
		calls := 0
		d.OnLayoutChange(func(kit.DockLayout) { calls++ })
		w := openTest(t, Options{Width: 800, Height: 500, Content: el.Root(d)})
		var source, target Element
		for _, e := range w.snapshot() {
			if e.Role == "tab" && e.Name == "A" {
				source = e
			}
			if e.Role == "region" && e.Name == "B" {
				target = e
			}
		}
		if source.Width == 0 || target.Height == 0 {
			t.Fatal("missing agent geometry")
		}
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
			return color.NRGBAModel.Convert(im.At(target.X+target.Width/2, target.Y+target.Height-30)).(color.NRGBA)
		}
		before := pixel()
		a := source.center()
		b := f32.Pt(float32(target.X+target.Width/2), float32(target.Y+target.Height-10))
		w.virt.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: a}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: a})
		w.render()
		w.virt.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: b})
		w.render()
		if pixel() == before || len(d.Layout().Left) != 1 || calls != 0 {
			t.Fatal("preview missing or changed layout before release")
		}
		w.virt.router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: b})
		w.render()
		if d.Layout().RightTree == nil || d.Layout().RightTree.Second == nil || d.Layout().RightTree.Second.Active != "a" || calls != 1 {
			t.Fatal("agent split drop")
		}
	}
}
