package kit

import (
	"image/color"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/dyike/keel/ui/theme"
)

// codeStyleName picks a chroma style that reads on the editor background.
func codeStyleName() string {
	if int(theme.CodeBg.R)+int(theme.CodeBg.G)+int(theme.CodeBg.B) < 384 {
		return "github-dark"
	}
	return "github"
}

// highlightCode tokenizes text and returns colored spans per line, with the
// syntax context of each. It reads only its arguments, so it can run off the
// UI goroutine. An unknown language yields nil.
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
		kind := codeKindCode
		switch {
		case tok.Type.InCategory(chroma.Comment):
			kind = codeKindComment
		case tok.Type.InCategory(chroma.LiteralString):
			kind = codeKindString
		}
		start := col
		emit := func() {
			if col > start && (c.A != 0 || kind != codeKindCode) {
				cur = append(cur, codeSpan{start, col, c, kind})
			}
		}
		for _, r := range tok.Value {
			if r == '\n' {
				emit()
				lines = append(lines, cur)
				cur, col, start = nil, 0, 0
				continue
			}
			col++
		}
		emit()
	}
	return append(lines, cur)
}

// codeLocalReach is how far around an edit the editor re-highlights at once,
// before the background pass over the whole file catches up.
const codeLocalReach = 60

// highlightNear re-highlights the lines around [first, last] immediately, so
// what the user just typed is colored without waiting for a full pass. It
// starts at a line that begins at column 0 outside a string or comment, the
// place a lexer's state is most likely back at its root.
func (b *codeBuffer) highlightNear(lang, style string, first, last int) {
	if lexers.Get(lang) == nil {
		return
	}
	start := first
	for start > 0 && first-start < codeLocalReach {
		if l := b.line(start); len(l) > 0 && l[0] != ' ' && l[0] != '\t' {
			if prev := b.spans(start - 1); len(prev) == 0 || prev[len(prev)-1].kind == codeKindCode || prev[len(prev)-1].end < len(b.line(start-1)) {
				break
			}
		}
		start--
	}
	end := min(b.count()-1, last+codeLocalReach)
	var sb strings.Builder
	b.lines.each(start, func(i int, l *codeLine) bool {
		if i > start {
			sb.WriteByte('\n')
		}
		sb.WriteString(string(l.text))
		return i < end
	})
	spans := highlightCode(lang, sb.String(), style)
	// Keep the lines near the end of the window as they were: tokens there
	// may continue past it.
	keep := min(end, last+codeLocalReach/2)
	for i := start; i <= keep && i-start < len(spans); i++ {
		b.lines.at(i).spans = spans[i-start]
	}
}
