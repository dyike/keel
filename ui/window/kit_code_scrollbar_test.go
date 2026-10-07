package window

import (
	"bytes"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestCodeEditorScrollbarStaysVisibleAtEnd(t *testing.T) {
	ed := kit.CodeEditor(strings.Repeat("\n", 100)).Height(80).Name("scroll source")
	w := openTest(t, Options{Width: 260, Height: 150, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Items(el.Start).Child(el.Div().W(el.Dp(220)).Child(ed.Render(cx)))
	}))})
	w.scroll(element(t, w, "scroll source").center(), 10000)
	data, err := w.screenshot()
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	// At the end the minimum 24px thumb must remain fully in the viewport.
	// Empty lines keep text and decorations out of this pixel column.
	count := 0
	for y := 0; y < 80; y++ {
		if color.NRGBAModel.Convert(im.At(215, y)).(color.NRGBA) != theme.CodeBg {
			count++
		}
	}
	if count < 22 || count > 24 {
		t.Fatalf("visible thumb height %d, want a full 24px rounded thumb", count)
	}
}
