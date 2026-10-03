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

func TestBubbleVariantsAndGhost(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	body := el.ViewFunc(func(*el.Context) el.Element { return el.Div().Name("Body").Size(el.Dp(40)) })
	v := kit.Bubble(body).PartStyle(kit.BubblePartContent, func(e *el.DivEl) { e.Role("group").Name("Surface") })
	w := openTest(t, Options{Width: 300, Height: 160, Content: el.Root(v)})
	sample := func() color.NRGBA {
		t.Helper()
		e := element(t, w, "Surface")
		data, err := w.screenshot()
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		return color.NRGBAModel.Convert(im.At(e.X+e.Width/2, e.Y+e.Height/2)).(color.NRGBA)
	}
	for _, p := range []theme.Palette{theme.Light(), theme.Dark()} {
		core.Update(func() { theme.Apply(p) })
		for _, tc := range []struct {
			variant kit.BubbleVariant
			want    color.NRGBA
		}{{kit.BubbleFilled, p.Primary}, {kit.BubbleSecondary, p.Subtle}, {kit.BubbleMuted, p.Subtle}, {kit.BubbleTinted, p.Highlight}, {kit.BubbleOutline, p.Bg}, {kit.BubbleGhost, p.Bg}} {
			v.Variant(tc.variant)
			if got := sample(); got != tc.want {
				t.Fatal("variant paint", tc.variant, got, tc.want)
			}
		}
		v.Variant(kit.BubbleDestructive)
		if got := sample(); got == p.Bg || got == p.Primary {
			t.Fatal("destructive surface absent", got)
		}
		v.Variant(kit.BubbleGhost)
		surface := element(t, w, "Surface")
		content := element(t, w, "Body")
		if surface.Width != 300 || content.X != surface.X || content.Y != surface.Y {
			t.Fatal("ghost padding or width", surface, content)
		}
	}
}
