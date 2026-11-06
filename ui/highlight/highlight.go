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
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/codehighlight"
)

func init() { core.SetHighlighter(Chroma{}) }

// Chroma installs the full language registry. Its zero value is ready to use.
type Chroma struct{}

func (Chroma) Language(name string) string { return full.Language(name) }
func (Chroma) Highlight(code string, o core.HighlightOptions) ([]core.CodeToken, bool) {
	return full.Highlight(code, o)
}

var full = codehighlight.Highlighter{Get: lexers.Get, Analyse: lexers.Analyse}
