package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestBubbleReactionActionCorners(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	button := kit.Button("Action", func() {}).Size(32)
	v := kit.Bubble(el.ViewFunc(func(*el.Context) el.Element { return el.Text("A sufficiently wide message surface") })).ReactionActions(button).
		PartStyle(kit.BubblePartReactions, func(e *el.DivEl) { e.Name("Reactions").Role("group") })
	w := openTest(t, Options{Width: 320, Height: 220, Content: el.Root(v)})
	for _, p := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(p) })
		b := element(t, w, "Action")
		r := element(t, w, "Reactions")
		if r.Width != b.Width+2 || b.Y != r.Y+1 || r.Height != b.Height+2 {
			t.Fatal("typed actions retained decorative padding", b, r)
		}
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		corner := color.NRGBAModel.Convert(im.At(b.X+3, b.Y+3)).(color.NRGBA)
		center := color.NRGBAModel.Convert(im.At(b.X+b.Width/2, b.Y+5)).(color.NRGBA)
		if corner == p.Primary || center != p.Primary {
			t.Fatal("pill shape or button variant lost", corner, center)
		}
	}
}
