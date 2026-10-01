package markdown

import (
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

type paletteKey struct {
	revision uint64
	colors   [5]color.NRGBA
	style    string
}

func followColor(c, fallback color.NRGBA) color.NRGBA {
	if c == (color.NRGBA{}) {
		return fallback
	}
	return c
}
func codeStyle() string {
	if CodeStyle != "" {
		return CodeStyle
	}
	if int(theme.CodeBg.R)+int(theme.CodeBg.G)+int(theme.CodeBg.B) < 384 {
		return "github-dark"
	}
	return "github"
}
func currentPaletteKey() paletteKey {
	return paletteKey{theme.Revision(), [5]color.NRGBA{CodeBg, CodeBorder, CodeHover, InlineCode, InlineCodeBg}, CodeStyle}
}
func (d *Doc) refreshPalette() {
	k := currentPaletteKey()
	if d.palette == k {
		return
	}
	d.palette = k
	var visit func(*block)
	visit = func(b *block) {
		if cv, ok := b.view.(*codeView); ok {
			cv.rich = highlight(b.lang, b.code)
		} else {
			b.view = nil
		}
		for i := range b.children {
			visit(&b.children[i])
		}
		for i := range b.items {
			for j := range b.items[i].blocks {
				visit(&b.items[i].blocks[j])
			}
		}
	}
	for i := range d.chunks {
		for j := range d.chunks[i].blocks {
			visit(&d.chunks[i].blocks[j])
		}
	}
}
