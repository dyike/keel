package markdown

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"testing"
)

func TestMarkdownThemeRefresh(t *testing.T) {
	old := theme.Current()
	defer theme.Apply(old)
	d := New("`code`\n\n```go\nvar x = 1\n```\n")
	h := uitest.New(el.Embed(d))
	theme.Apply(theme.Dark())
	h.Frame()
	if c := followColor(CodeBg, theme.CodeBg); int(c.R)+int(c.G)+int(c.B) > 150 {
		t.Fatal("light code background")
	}
	if codeStyle() != "github-dark" {
		t.Fatal("light highlighter")
	}
	var code *codeView
	for i := range d.chunks {
		for j := range d.chunks[i].blocks {
			if c, ok := d.chunks[i].blocks[j].view.(*codeView); ok {
				code = c
			}
		}
	}
	if code == nil {
		t.Fatal("missing code view")
	}
	code.wrap = true
	theme.Apply(theme.Light())
	h.Frame()
	if !code.wrap {
		t.Fatal("theme lost wrapping")
	}
	override := color.NRGBA{R: 9, A: 255}
	if followColor(override, theme.CodeBg) != override {
		t.Fatal("override ignored")
	}
}
