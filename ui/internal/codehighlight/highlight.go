// Package codehighlight adapts a selected Chroma registry to Keel tokens.
package codehighlight

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/dyike/keel/ui/core"
	"image/color"
)

// Highlighter resolves languages through the selected registry.
type Highlighter struct {
	Get     func(string) chroma.Lexer
	Analyse func(string) chroma.Lexer
}

// Language is chroma's name for a language, alias or extension.
func (h Highlighter) Language(name string) string {
	if l := h.Get(name); l != nil {
		return l.Config().Name
	}
	return ""
}

// Highlight tokenizes code; Style names a chroma style, defaulting to
// github or github-dark.
func (h Highlighter) Highlight(code string, o core.HighlightOptions) ([]core.CodeToken, bool) {
	lexer := h.Get(o.Language)
	if lexer == nil && o.Guess {
		lexer = h.Analyse(code)
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
