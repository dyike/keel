package markdown

import (
	"strings"

	"github.com/dyike/keel/ui/el"
)

// CodeBlockContext is a snapshot of a fenced or indented code block. Text is
// the original code, without the Markdown fence. ID remains stable while the
// parser preserves the block's presentation state.
type CodeBlockContext struct {
	ID, Language, Text string
	Wrapped            bool
}

// CodeBlockActions appends controls to the default code header. It is called
// each render, including for code in lists/quotes. Nil removes the extension.
// Handlers may change application state; rendering must not mutate the document.
func (d *Doc) CodeBlockActions(fn func(*el.Context, CodeBlockContext) el.Element) *Doc {
	d.codeActions = fn
	d.extensionRevision++
	return d
}

// CodeBlockRenderer replaces code cards for one language (case insensitive).
// Return nil to retain highlighting and standard actions. An empty language
// targets unlabelled code. Passing nil unregisters the renderer. Custom views
// own their controls and text selection; the source is never executed by Doc.
func (d *Doc) CodeBlockRenderer(language string, fn func(*el.Context, CodeBlockContext) el.Element) *Doc {
	if d.codeRenderers == nil {
		d.codeRenderers = map[string]func(*el.Context, CodeBlockContext) el.Element{}
	}
	language = strings.ToLower(strings.TrimSpace(language))
	if fn == nil {
		delete(d.codeRenderers, language)
	} else {
		d.codeRenderers[language] = fn
	}
	d.extensionRevision++
	return d
}

func (d *Doc) dynamicCode(b *block) bool {
	if b.kind == codeBlock && (d.codeActions != nil || d.codeRenderers[strings.ToLower(b.lang)] != nil) {
		return true
	}
	for i := range b.children {
		if d.dynamicCode(&b.children[i]) {
			return true
		}
	}
	for i := range b.items {
		for j := range b.items[i].blocks {
			if d.dynamicCode(&b.items[i].blocks[j]) {
				return true
			}
		}
	}
	return false
}
