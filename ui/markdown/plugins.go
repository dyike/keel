package markdown

import (
	"strings"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	gmparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// InlineObject is an indivisible inline widget. Text supplies its selectable
// and copied representation and must be nonempty. The widget receives the
// remaining line width and wraps as a whole when it does not fit. Its bottom
// aligns with the text baseline. Keep application state in the widget itself.
type InlineObject struct {
	Text   string
	Widget core.Widget
}

// Plugin extends one document's Goldmark parser and maps AST nodes to native
// views. Factories run during parsing, under the UI lock; treat node/source as
// read-only and do not mutate Doc. A nil block view or nil inline widget falls
// back to the built-in conversion. Later plugins win for duplicate node kinds.
// HTML conversion and fenced code contents are not passed through these maps.
type Plugin struct {
	Extensions []goldmark.Extender
	Blocks     map[ast.NodeKind]func(ast.Node, []byte) el.View
	Inlines    map[ast.NodeKind]func(ast.Node, []byte) InlineObject
	// Local declares that the plugin's syntax never spans a blank line, as
	// inline objects and per-block renderers do. When every plugin is local
	// the document keeps parsing in blank-line chunks: a streaming append
	// reparses only its last chunk instead of the whole answer.
	Local bool
}

type pluginSet struct {
	parser  gmparser.Parser
	blocks  map[ast.NodeKind]func(ast.Node, []byte) el.View
	inlines map[ast.NodeKind]func(ast.Node, []byte) InlineObject
	local   bool // every plugin is Local
}

// Only blocks that actually embed custom views need to rebuild each frame.
// Registering a heading/quote renderer must not disable caching for every
// ordinary paragraph, list and table in a long document.
func hasCustomContent(b *block) bool {
	if b.custom != nil {
		return true
	}
	objects := func(spans []span) bool {
		for _, s := range spans {
			if s.object != nil {
				return true
			}
		}
		return false
	}
	if objects(b.spans) {
		return true
	}
	if b.tbl != nil {
		for _, cell := range b.tbl.header {
			if objects(cell) {
				return true
			}
		}
		for _, row := range b.tbl.rows {
			for _, cell := range row {
				if objects(cell) {
					return true
				}
			}
		}
	}
	for i := range b.children {
		if hasCustomContent(&b.children[i]) {
			return true
		}
	}
	for i := range b.items {
		for j := range b.items[i].blocks {
			if hasCustomContent(&b.items[i].blocks[j]) {
				return true
			}
		}
	}
	return false
}

// Plugins replaces this document's plugins and reparses its source. Configure
// once, not every Render. Empty input restores the built-in parser. Documents
// with a plugin that is not Local parse as a whole, so custom block syntax can
// cross blank lines.
// Factories may run again on each edit; reuse widgets/views in application code
// when their interactive state must survive reparsing. Render runs every frame.
func (d *Doc) Plugins(plugins ...Plugin) *Doc {
	d.plugins = nil
	if len(plugins) > 0 {
		p := &pluginSet{blocks: make(map[ast.NodeKind]func(ast.Node, []byte) el.View), inlines: make(map[ast.NodeKind]func(ast.Node, []byte) InlineObject), local: true}
		extensions := []goldmark.Extender{extension.GFM, extension.Footnote}
		for _, plugin := range plugins {
			p.local = p.local && plugin.Local
			for _, ext := range plugin.Extensions {
				if ext != nil {
					extensions = append(extensions, ext)
				}
			}
			for k, fn := range plugin.Blocks {
				if fn != nil {
					p.blocks[k] = fn
				}
			}
			for k, fn := range plugin.Inlines {
				if fn != nil {
					p.inlines[k] = fn
				}
			}
		}
		p.parser = goldmark.New(goldmark.WithExtensions(extensions...), goldmark.WithParserOptions(gmparser.WithInlineParsers(util.Prioritized(mathInlineParser{}, 50)), gmparser.WithBlockParsers(util.Prioritized(mathBlockParser{}, 50)))).Parser()
		d.plugins = p
	}
	d.chunks, d.contextual, d.parsedContext = nil, false, ""
	d.extensionRevision++
	d.selection.clear()
	d.ranges.reveal = nil
	d.update()
	return d
}

func (d *Doc) parseSource(src string) []block {
	// Only the document's first chunk, or the whole document when parsed in
	// one context, can start with front matter; split keeps it whole.
	if fm, yaml, rest := splitFrontMatter(src); fm != "" && d.startsWith(src) {
		out := []block{{kind: frontMatter, code: strings.TrimRight(yaml, "\n")}}
		if strings.TrimSpace(rest) != "" {
			out = append(out, d.parseSource(rest)...)
		}
		return out
	}
	if d.plugins == nil {
		return parse(src)
	}
	source := []byte(src)
	node := d.plugins.parser.Parse(text.NewReader(source))
	return blocks(node, source, mathMacros{}, d.plugins)
}

func customBlock(p []*pluginSet, n ast.Node, src []byte) el.View {
	if len(p) > 0 && p[0] != nil {
		if fn := p[0].blocks[n.Kind()]; fn != nil {
			return fn(n, src)
		}
	}
	return nil
}
func customInline(p []*pluginSet, n ast.Node, src []byte) *InlineObject {
	if len(p) > 0 && p[0] != nil {
		if fn := p[0].inlines[n.Kind()]; fn != nil {
			object := fn(n, src)
			if object.Widget != nil && object.Text != "" {
				return &object
			}
		}
	}
	return nil
}

func (r *richBlock) hasInlineObjects() bool {
	for _, rn := range r.runs {
		if rn.object != nil {
			return true
		}
	}
	return false
}

// startsWith reports whether src is the beginning of the document.
func (d *Doc) startsWith(src string) bool {
	return strings.HasPrefix(d.src, src) || strings.HasPrefix(src, d.src)
}
