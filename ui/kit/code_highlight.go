package kit

import (
	"image/color"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/dyike/keel/ui/theme"
)

// codeSpan colors runes [start, end) of a line. A zero color means plain text.
type codeSpan struct {
	start, end int
	color      color.NRGBA
}

// codeStyleName picks a chroma style that reads on the editor background.
func codeStyleName() string {
	if int(theme.CodeBg.R)+int(theme.CodeBg.G)+int(theme.CodeBg.B) < 384 {
		return "github-dark"
	}
	return "github"
}

// highlightCode tokenizes the whole text and returns colored spans per line.
// It is slow for large files and runs off the UI goroutine: it reads only its
// arguments. An unknown language yields no spans.
func highlightCode(lang, text, style string) [][]codeSpan {
	lexer := lexers.Get(lang)
	if lexer == nil {
		return nil
	}
	s := styles.Get(style)
	it, err := chroma.Coalesce(lexer).Tokenise(nil, text)
	if err != nil {
		return nil
	}
	var lines [][]codeSpan
	var cur []codeSpan
	col := 0
	for _, tok := range it.Tokens() {
		e := s.Get(tok.Type)
		var c color.NRGBA
		if e.Colour.IsSet() {
			c = color.NRGBA{R: e.Colour.Red(), G: e.Colour.Green(), B: e.Colour.Blue(), A: 0xff}
		}
		start := col
		for _, r := range tok.Value {
			if r == '\n' {
				if col > start && c.A != 0 {
					cur = append(cur, codeSpan{start, col, c})
				}
				lines = append(lines, cur)
				cur, col, start = nil, 0, 0
				continue
			}
			col++
		}
		if col > start && c.A != 0 {
			cur = append(cur, codeSpan{start, col, c})
		}
	}
	return append(lines, cur)
}
