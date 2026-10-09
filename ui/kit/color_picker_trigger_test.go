package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"image/color"
	"testing"
)

func TestColorPickerSizesAndPopup(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, size := range []ColorPickerSize{ColorPickerSizeXSmall, ColorPickerSizeSmall, ColorPickerSizeMedium, ColorPickerSizeLarge} {
			t.Run(fmt.Sprintf("%d/%d", scale, size), func(t *testing.T) {
				red := color.NRGBA{R: 255, A: 255}
				p := ColorPicker().Size(size).Label("Accent").Popup(true).Swatches(red)
				original := p.Value()
				calls := 0
				p.OnChange(func(color.NRGBA) { calls++ })
				h := renderView(p, 500, scale)
				if shown(h, locale.Current().ColorShade) {
					t.Fatal("closed panel visible")
				}
				clickClass(t, h, "Button", "Accent")
				h.Frame()
				if !p.IsOpen() {
					t.Fatal("trigger failed")
				}
				r := bounds(h, locale.Current().ColorShade)
				if r.Dx() != int(p.metrics().width)*scale || r.Dy() != int(p.metrics().square)*scale {
					t.Fatal("panel metrics", r, p.metrics())
				}
				click(t, h, "#FF0000")
				h.Frame()
				if p.Value() != red || calls != 1 {
					t.Fatal("popup color callback", p.Value(), calls)
				}
				h.Key(key.NameEscape, 0)
				h.Frame()
				if p.IsOpen() || shown(h, locale.Current().ColorShade) {
					t.Fatal("escape failed")
				}
				p.SetValue(original)
				p.Icon(IconCheck).SetOpen(true)
				h.Frame()
				p.Size(ColorPickerSizeLarge)
				p.Size(ColorPickerSize(255))
				h.Frame()
				if !p.IsOpen() || p.Value() != original || calls != 1 || p.size != ColorPickerSizeLarge {
					t.Fatal("configuration lost value")
				}
				p.SetDisabled(true)
				h.Frame()
				p.SetOpen(true)
				if p.IsOpen() {
					t.Fatal("disabled panel remained open")
				}
				p.SetDisabled(false)
				p.Popup(false)
				h.Frame()
				if !shown(h, locale.Current().ColorShade) {
					t.Fatal("inline mode unavailable")
				}
			})
		}
	}
}

func TestColorPickerPopupFocusAndDraftCancellation(t *testing.T) {
	p := ColorPicker().Popup(true).Label("Accent")
	var cx *el.Context
	h := renderView(viewFunc(func(c *el.Context) el.Element { cx = c; return p.Render(c) }), 400, 1)
	id := autoID("color", p)
	cx.Focus(id + "/trigger")
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !p.IsOpen() || !cx.Focused(id+"/shade") {
		t.Fatal("keyboard opening did not focus shade")
	}
	before := p.Value()
	calls := 0
	p.OnChange(func(color.NRGBA) { calls++ })
	clickClass(t, h, "Editor", "HEX")
	h.Type("#FF0000")
	h.Key(key.NameEscape, 0)
	h.Frame()
	if p.IsOpen() || p.Value() != before || calls != 0 || p.hex != hexOf(before, false) || !cx.Focused(id+"/trigger") {
		t.Fatal("Escape did not cancel draft and restore focus", p.Value(), calls, p.hex)
	}
	p.SetOpen(true)
	h.Frame()
	click(t, h, locale.Current().ColorShade)
	calls = 0
	h.Key(key.NameRightArrow, 0)
	if calls != 1 {
		t.Fatal("popup keyboard color edit missing", calls)
	}
	h.Click(390, 390)
	h.Frame()
	if p.IsOpen() {
		t.Fatal("outside click did not close")
	}
}
