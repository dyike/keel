package core

import (
	"image/color"
	"sync/atomic"
)

// Syntax highlighting is opt-in. Keel's highlighter (chroma, with rules for
// hundreds of languages) adds about 4 MB to a binary, so code shows as plain
// text until an app links one in:
//
//	import _ "github.com/dyike/keel/ui/highlight"
//
// CodeEditor, TextView and Markdown code blocks all read it from here.

// CodeTokenKind is the syntax context of a token, for editing rules such as
// not pairing brackets inside strings.
type CodeTokenKind uint8

const (
	CodeTokenCode CodeTokenKind = iota
	CodeTokenString
	CodeTokenComment
)

// CodeToken is a run of source text with its style. A zero Color means the
// default text color.
type CodeToken struct {
	Text         string
	Color        color.NRGBA
	Bold, Italic bool
	Kind         CodeTokenKind
}

// HighlightOptions say how to highlight.
type HighlightOptions struct {
	// Language is a name, alias or file extension ("go", "Go", "golang").
	Language string
	// Guess detects the language from the code when Language is empty or
	// unknown.
	Guess bool
	// Style names a color scheme; empty picks one for Dark.
	Style string
	// Dark is true on a dark code background.
	Dark bool
}

// Highlighter tokenizes source code. It is called off the UI goroutine.
type Highlighter interface {
	// Highlight splits code into tokens whose texts concatenate to code.
	// ok is false when the language is unknown (and not guessed): the code
	// then stays plain.
	Highlight(code string, opts HighlightOptions) (tokens []CodeToken, ok bool)
	// Language is the canonical name of a language, "" when unknown.
	Language(name string) string
}

var highlighter atomic.Pointer[Highlighter]

// SetHighlighter installs the syntax highlighter; ui/highlight calls it
// when imported. Nil removes it.
func SetHighlighter(h Highlighter) {
	if h == nil {
		highlighter.Store(nil)
		return
	}
	highlighter.Store(&h)
}

// CurrentHighlighter is the installed highlighter, or nil.
func CurrentHighlighter() Highlighter {
	if p := highlighter.Load(); p != nil {
		return *p
	}
	return nil
}
