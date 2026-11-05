// Package golang installs Go, Go text template and plain-text highlighting.
// Import ui/highlight instead for the complete Chroma language registry.
package golang

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/codehighlight"
)

func init() {
	registry.Register(chroma.MustNewXMLLexer(embedded, "plaintext.xml"))
	core.SetHighlighter(Chroma{})
}

// Chroma highlights Go, Go text templates and plain text. Its zero value is ready to use.
type Chroma struct{}

var selected = codehighlight.Highlighter{Get: registry.Get, Analyse: registry.Analyse}

func (Chroma) Language(name string) string { return selected.Language(name) }
func (Chroma) Highlight(code string, o core.HighlightOptions) ([]core.CodeToken, bool) {
	return selected.Highlight(code, o)
}
