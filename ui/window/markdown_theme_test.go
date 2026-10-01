package window

import (
	"bytes"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/markdown"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"image/png"
	"testing"
)

func TestMarkdownCodePixelsFollowTheme(t *testing.T) {
	old := theme.Current()
	defer core.Update(func() { theme.Apply(old) })
	w := openTest(t, Options{Width: 400, Height: 240, Content: el.Embed(markdown.New("`inline`\n\n```go\nvar x = 1\n```"))})
	for _, p := range []theme.Palette{theme.Dark(), theme.Light()} {
		core.Update(func() { theme.Apply(p) })
		b, e := w.screenshot()
		if e != nil {
			t.Fatal(e)
		}
		im, e := png.Decode(bytes.NewReader(b))
		if e != nil {
			t.Fatal(e)
		}
		count := 0
		for y := 0; y < im.Bounds().Dy(); y++ {
			for x := 0; x < im.Bounds().Dx(); x++ {
				if color.NRGBAModel.Convert(im.At(x, y)) == p.CodeBg {
					count++
				}
			}
		}
		if count < 1000 {
			t.Fatalf("code background not rendered: %d", count)
		}
	}
}
