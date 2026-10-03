package window

import (
	"bytes"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"image/color"
	"image/png"
	"testing"
)

func TestDialogOverlayPixelsAndCloseFocus(t *testing.T) {
	d := kit.Dialog("Options").Width(180).Keyboard(false).OverlayClosable(false).CloseButton(true)
	opens := 0
	b := kit.Button("Open", func() { opens++; d.SetValue(true) }).ID("opener")
	var cx *el.Context
	w := openTest(t, Options{Width: 420, Height: 300, Content: el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(b.Render(ctx), d.Render(ctx))
	}))})
	sample := func() color.NRGBA {
		t.Helper()
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(5, 5)).(color.NRGBA)
	}
	base := sample()
	w.click(element(t, w, "Open").center())
	w.render()
	dim := sample()
	if dim == base {
		t.Fatal("default scrim absent")
	}
	d.Overlay(false)
	if sample() != base {
		t.Fatal("scrim not removed")
	}
	if err := w.press("esc"); err != nil {
		t.Fatal(err)
	}
	if !d.Value() {
		t.Fatal("Esc disabled")
	}
	w.click(element(t, w, locale.Current().Close).center())
	w.render()
	w.render()
	if d.Value() || !cx.Focused("opener") {
		t.Fatal("close did not restore trigger focus")
	}
	if err := w.press("space"); err != nil {
		t.Fatal(err)
	}
	if opens != 2 || !d.Value() {
		t.Fatal("restored trigger not operable")
	}
	d.Overlay(true)
	if sample() != dim {
		t.Fatal("scrim not restored")
	}
}
