package window

import (
	"bytes"
	"image/color"
	"image/png"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestKitSlotsFollowRuntimePalette(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	for _, tc := range []struct {
		name string
		host func(el.View) el.View
	}{
		{"empty", func(v el.View) el.View { return kit.Empty("Empty").Action(v) }},
		{"status_left", func(v el.View) el.View { return kit.StatusBar().Left(v) }},
		{"status_right", func(v el.View) el.View { return kit.StatusBar().Right(v) }},
		{"description", func(v el.View) el.View { return kit.DescriptionList().ItemView("Slot", v) }},
		{"group", func(v el.View) el.View { return kit.GroupBox("Group").Child(v) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core.Update(func() { theme.Apply(theme.Light()) })
			// Retain the host and slot across theme changes; only Render reads color.
			v := tc.host(el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Name("theme-slot").Role("image").Size(el.Dp(20)).Bg(theme.Muted)
			}))
			w := openTest(t, Options{Width: 300, Height: 240, Content: el.Embed(v)})
			for _, p := range []theme.Palette{theme.Light(), theme.Dark(), theme.Light()} {
				core.Update(func() { theme.Apply(p) })
				data, err := w.screenshot()
				if err != nil {
					t.Fatal(err)
				}
				im, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				slot := element(t, w, "theme-slot")
				got := color.NRGBAModel.Convert(im.At(slot.X+slot.Width/2, slot.Y+slot.Height/2))
				if slot.Width == 0 || slot.Height == 0 || got != p.Muted {
					t.Fatalf("slot did not follow palette: bounds=%+v color=%v want=%v", slot, got, p.Muted)
				}
			}
		})
	}
}
