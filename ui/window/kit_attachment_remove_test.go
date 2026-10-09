package window

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
)

func TestAttachmentRemoveHoverAndKeyboard(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		t.Run(map[bool]string{false: "horizontal", true: "vertical"}[vertical], func(t *testing.T) {
			removed, opened := 0, 0
			a := kit.Attachment("photo", 123).Vertical(vertical).OnOpen(func() { opened++ }).OnRemove(func() { removed++ })
			w := openTest(t, Options{Width: 340, Height: 400, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Items(el.Start).Child(a.Render(cx)) }))})
			name := locale.Current().Name(locale.Current().Remove, "photo")
			e := element(t, w, name)
			sample := func() []byte {
				t.Helper()
				w.snapshot()
				data, err := w.screenshot()
				if err != nil {
					t.Fatal(err)
				}
				im, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				var pixels []byte
				for y := e.Y; y < e.Y+e.Height; y++ {
					for x := e.X; x < e.X+e.Width; x++ {
						r, g, b, a := im.At(x, y).RGBA()
						pixels = append(pixels, byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8))
					}
				}
				return pixels
			}
			move := func(p f32.Point) {
				w.virt.router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p})
				w.snapshot()
				w.snapshot()
			}
			move(f32.Pt(330, 390))
			hidden := sample()
			a.RemoveOnHover(false)
			shown := sample()
			if bytes.Equal(hidden, shown) {
				t.Fatal("remove did not become visible")
			}
			a.RemoveOnHover(true)
			if !bytes.Equal(hidden, sample()) {
				t.Fatal("hover setting not restored")
			}
			move(f32.Pt(20, 20))
			if !bytes.Equal(shown, sample()) {
				t.Fatal("card hover did not reveal remove")
			}
			move(f32.Pt(330, 390))
			if !bytes.Equal(hidden, sample()) {
				t.Fatal("pointer leave did not hide remove")
			}
			if err := w.press("Tab"); err != nil {
				t.Fatal(err)
			}
			focused := sample()
			a.RemoveOnHover(false)
			expected := sample()
			a.RemoveOnHover(true)
			if !bytes.Equal(focused, expected) || bytes.Equal(hidden, focused) {
				t.Fatal("keyboard focus in card did not reveal remove")
			}
			if err := w.press("Tab"); err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(hidden, sample()) {
				t.Fatal("remove focus invisible")
			}
			if err := w.press("Space"); err != nil {
				t.Fatal(err)
			}
			if removed != 1 || opened != 0 {
				t.Fatal("keyboard activation", removed, opened)
			}
			w.click(element(t, w, name).center())
			if removed != 2 || opened != 0 {
				t.Fatal("remove click reached open", removed, opened)
			}
			a.SetDisabled(true)
			w.click(element(t, w, name).center())
			if removed != 2 {
				t.Fatal("disabled remove activated")
			}
			a.SetDisabled(false)
			a.ShowActions(false)
			for _, item := range w.snapshot() {
				if item.Name == name {
					t.Fatal("hidden actions retained remove")
				}
			}
			a.ShowActions(true)
			restored := element(t, w, name)
			if restored.X != e.X || restored.Y != e.Y || restored.Width != e.Width || restored.Height != e.Height {
				t.Fatal("visibility changed layout", e, restored)
			}
		})
	}
}
