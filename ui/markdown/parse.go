package markdown

import (
	"github.com/dyike/keel/ui/locale"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	gmparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// The intermediate form: what a chunk of Markdown shows, independent of how it
// is drawn. It is built once per chunk and reused across frames.

type blockKind uint8

const (
	paragraph blockKind = iota
	heading
	codeBlock
	quote
	list
	table
	rule
	footnoteList
	footnoteItem
	group // several blocks from one HTML block, drawn one after another
)

type block struct {
	kind     blockKind
	anchor   string
	level    int    // heading
	spans    []span // paragraph, heading
	lang     string // codeBlock
	code     string
	children []block // quote
	ordered  bool    // list
	start    int
	items    []listItem
	tbl      *tableData

	view any // drawing state that must survive frames (rich text, copy feedback)
}

type listItem struct {
	task   *bool // nil: not a task item
	blocks []block
}

type cellAlign uint8

const (
	alignLeft cellAlign = iota
	alignCenter
	alignRight
)

type tableData struct {
	align  []cellAlign
	header [][]span
	rows   [][][]span
}

type span struct {
	text                       string
	bold, italic, code, strike bool
	link                       string
	math                       *mathExpr
	display                    bool
	anchor                     string
	superscript, subscript     bool
	underline, mark            bool // from HTML <u> and <mark>
	imageURL                   string
	imageAlt                   string
}

var parser = goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote), goldmark.WithParserOptions(gmparser.WithInlineParsers(util.Prioritized(mathInlineParser{}, 50)), gmparser.WithBlockParsers(util.Prioritized(mathBlockParser{}, 50)))).Parser()

// parse turns one chunk of Markdown into blocks.
func parse(src string) []block {
	source := []byte(src)
	doc := parser.Parse(text.NewReader(source))
	return blocks(doc, source, mathMacros{})
}

func blocks(parent ast.Node, src []byte, macros mathMacros) []block {
	var out []block
	for n := parent.FirstChild(); n != nil; n = n.NextSibling() {
		if b, ok := convert(n, src, macros); ok {
			out = append(out, b)
		}
	}
	return out
}

func convert(n ast.Node, src []byte, macros mathMacros) (block, bool) {
	switch n := n.(type) {
	case *east.FootnoteList:
		return block{kind: footnoteList, children: blocks(n, src, macros)}, true
	case *east.Footnote:
		children := blocks(n, src, macros)
		anchorFirst(children, footnoteID(n.Index))
		return block{kind: footnoteItem, level: n.Index, children: children}, true
	case *east.FootnoteBacklink:
		return block{kind: paragraph, spans: []span{backlinkSpan(n)}}, true
	case *displayMath:
		source := lines(n, src)
		s := span{text: "$$\n" + source, display: true}
		if n.closed {
			s.text += "\n$$"
			s.math = parseMathIn(source, macros)
		}
		return block{kind: paragraph, spans: []span{s}}, true
	case *ast.Paragraph, *ast.TextBlock:
		return block{kind: paragraph, spans: inlines(n, src, span{}, macros)}, true
	case *ast.Heading:
		return block{kind: heading, level: n.Level, spans: inlines(n, src, span{}, macros)}, true
	case *ast.FencedCodeBlock:
		lang := ""
		if n.Info != nil {
			lang = strings.Fields(string(n.Info.Segment.Value(src)) + " ")[0]
		}
		return block{kind: codeBlock, lang: lang, code: lines(n, src)}, true
	case *ast.CodeBlock:
		return block{kind: codeBlock, code: lines(n, src)}, true
	case *ast.Blockquote:
		return block{kind: quote, children: blocks(n, src, macros)}, true
	case *ast.List:
		b := block{kind: list, ordered: n.IsOrdered(), start: n.Start}
		for it := n.FirstChild(); it != nil; it = it.NextSibling() {
			item := listItem{blocks: blocks(it, src, macros)}
			item.task = taskState(it)
			b.items = append(b.items, item)
		}
		return b, true
	case *ast.ThematicBreak:
		return block{kind: rule}, true
	case *ast.HTMLBlock:
		// An HTML block may hold several blocks; the first goes here and the
		// rest follow it as a quote-less group.
		bs := htmlBlocks(lines(n, src) + closure(n, src))
		switch len(bs) {
		case 0:
			return block{}, false
		case 1:
			return bs[0], true
		}
		return block{kind: group, children: bs}, true
	case *east.Table:
		return block{kind: table, tbl: tableOf(n, src, macros)}, true
	}
	return block{}, false
}

func lines(n ast.Node, src []byte) string {
	var b strings.Builder
	ls := n.Lines()
	for i := 0; i < ls.Len(); i++ {
		seg := ls.At(i)
		b.Write(seg.Value(src))
	}
	return strings.TrimRight(b.String(), "\n")
}

// taskState finds a GFM task checkbox at the start of a list item.
func taskState(item ast.Node) *bool {
	first := item.FirstChild()
	if first == nil {
		return nil
	}
	if cb, ok := first.FirstChild().(*east.TaskCheckBox); ok {
		v := cb.IsChecked
		return &v
	}
	return nil
}

func tableOf(t *east.Table, src []byte, macros mathMacros) *tableData {
	d := &tableData{}
	for _, a := range t.Alignments {
		switch a {
		case east.AlignCenter:
			d.align = append(d.align, alignCenter)
		case east.AlignRight:
			d.align = append(d.align, alignRight)
		default:
			d.align = append(d.align, alignLeft)
		}
	}
	for r := t.FirstChild(); r != nil; r = r.NextSibling() {
		var cells [][]span
		for c := r.FirstChild(); c != nil; c = c.NextSibling() {
			cells = append(cells, inlines(c, src, span{}, macros))
		}
		if _, ok := r.(*east.TableHeader); ok {
			d.header = cells
		} else {
			d.rows = append(d.rows, cells)
		}
	}
	return d
}

// inlines flattens inline content into styled spans, merging neighbours with
// the same style.
func inlines(n ast.Node, src []byte, style span, macros mathMacros) []span {
	var out []span
	add := func(s span) {
		if s.text == "" {
			return
		}
		if k := len(out) - 1; k >= 0 {
			last := &out[k]
			if last.imageURL == "" && s.imageURL == "" && last.anchor == "" && s.anchor == "" && last.superscript == s.superscript && last.subscript == s.subscript && last.underline == s.underline && last.mark == s.mark && last.math == nil && s.math == nil && last.bold == s.bold && last.italic == s.italic && last.code == s.code && last.strike == s.strike && last.link == s.link {
				last.text += s.text
				return
			}
		}
		out = append(out, s)
	}
	var walk func(n ast.Node, st span)
	walk = func(n ast.Node, st span) {
		var tags htmlInline // inline HTML tags style their later siblings
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch c := c.(type) {
			case *east.FootnoteLink:
				add(span{text: "[" + strconv.Itoa(c.Index) + "]", link: footnoteID(c.Index), anchor: footnoteRefID(c.Index, c.RefIndex), superscript: true})
			case *east.FootnoteBacklink:
				add(backlinkSpan(c))
			case *mathInline:
				s := st
				s.text, s.math, s.display = c.raw, parseMathIn(c.source, macros), c.display
				add(s)
			case *ast.Text:
				s := st
				s.text = string(c.Segment.Value(src))
				add(s)
				switch {
				case c.HardLineBreak():
					add(span{text: "\n", bold: st.bold, italic: st.italic, link: st.link})
				case c.SoftLineBreak():
					// A soft break is a space, except between CJK characters.
					if !endsCJK(s.text) {
						b := st
						b.text = " "
						add(b)
					}
				}
			case *ast.String:
				s := st
				s.text = string(c.Value)
				add(s)
			case *ast.CodeSpan:
				s := st
				s.code = true
				var b strings.Builder
				for t := c.FirstChild(); t != nil; t = t.NextSibling() {
					if tt, ok := t.(*ast.Text); ok {
						b.Write(tt.Segment.Value(src))
					}
				}
				s.text = b.String()
				add(s)
			case *ast.Emphasis:
				s := st
				if c.Level >= 2 {
					s.bold = true
				} else {
					s.italic = true
				}
				walk(c, s)
			case *east.Strikethrough:
				s := st
				s.strike = true
				walk(c, s)
			case *ast.Link:
				s := st
				s.link = string(c.Destination)
				walk(c, s)
			case *ast.AutoLink:
				s := st
				s.link = string(c.URL(src))
				s.text = string(c.Label(src))
				add(s)
			case *ast.Image:
				s := st
				s.imageURL = string(c.Destination)
				var alt strings.Builder
				ast.Walk(c, func(t ast.Node, entering bool) (ast.WalkStatus, error) {
					if tt, ok := t.(*ast.Text); entering && ok {
						alt.Write(tt.Segment.Value(src))
					}
					return ast.WalkContinue, nil
				})
				s.imageAlt = alt.String()
				s.text = "[" + locale.Current().Name(locale.Current().Image, s.imageAlt) + "]"
				add(s)
			case *east.TaskCheckBox:
				// drawn as the list marker instead
			case *ast.RawHTML:
				var raw strings.Builder
				for i := 0; i < c.Segments.Len(); i++ {
					seg := c.Segments.At(i)
					raw.Write(seg.Value(src))
				}
				next, insert, ok := tags.tag(raw.String(), st)
				if !ok { // not a known element, e.g. the <T> in List<T>
					s := st
					s.text = raw.String()
					add(s)
					continue
				}
				add(insert)
				st = next
			default:
				walk(c, st)
			}
		}
	}
	walk(n, style)
	return out
}

func endsCJK(s string) bool {
	r, _ := utf8.DecodeLastRuneInString(s)
	return unicode.Is(unicode.Han, r) || unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Hangul) || (r >= 0x3000 && r <= 0x303f) || (r >= 0xff00 && r <= 0xffef)
}

func plain(spans []span) string {
	var b strings.Builder
	for _, s := range spans {
		b.WriteString(s.text)
	}
	return b.String()
}
