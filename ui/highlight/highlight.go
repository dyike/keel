// Package highlight colors source code in CodeEditor, TextView and
// Markdown code blocks. Import it for its side effect where the app shows
// code:
//
//	import _ "github.com/dyike/keel/ui/highlight"
//
// It is separate because its lexers (chroma, hundreds of languages) add
// about 4 MB to a binary; without it code shows as plain text.
package highlight

import (
	"image/color"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/dyike/keel/ui/core"
)

func init() { core.SetHighlighter(Chroma{}) }

// Chroma is the core.Highlighter this package installs.
type Chroma struct{}

// Language is chroma's name for a language, alias or extension.
func (Chroma) Language(name string) string {
	if l := lexers.Get(name); l != nil {
		return l.Config().Name
	}
	return ""
}

// Highlight tokenizes code; Style names a chroma style, defaulting to
// github or github-dark.
func (Chroma) Highlight(code string, o core.HighlightOptions) ([]core.CodeToken, bool) {
	lexer := lexers.Get(o.Language)
	if lexer == nil && o.Guess {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		return nil, false
	}
	name := o.Style
	if name == "" {
		name = "github"
		if o.Dark {
			name = "github-dark"
		}
	}
	style := styles.Get(name)
	it, err := chroma.Coalesce(lexer).Tokenise(nil, code)
	if err != nil {
		return nil, false
	}
	var out []core.CodeToken
	for _, tok := range it.Tokens() {
		e := style.Get(tok.Type)
		t := core.CodeToken{Text: tok.Value, Bold: e.Bold == chroma.Yes, Italic: e.Italic == chroma.Yes}
		if e.Colour.IsSet() {
			t.Color = color.NRGBA{R: e.Colour.Red(), G: e.Colour.Green(), B: e.Colour.Blue(), A: 0xff}
		}
		switch {
		case tok.Type.InCategory(chroma.Comment):
			t.Kind = core.CodeTokenComment
		case tok.Type.InCategory(chroma.LiteralString):
			t.Kind = core.CodeTokenString
		}
		out = append(out, t)
	}
	return out, true
}
