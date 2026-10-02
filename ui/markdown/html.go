package markdown

import (
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/dyike/keel/ui/locale"
	"github.com/yuin/goldmark/ast"
)

// HTML in Markdown is drawn, not shown as source: inline tags style the text
// around them, and HTML blocks become the same blocks Markdown makes. Only
// presentation is kept; scripts, styles, forms and embeds are dropped, and
// nothing is ever run or fetched except images, as for Markdown images.

// dropped elements lose their content too.
var dropped = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Head: true, atom.Title: true, atom.Template: true,
	atom.Iframe: true, atom.Object: true, atom.Embed: true, atom.Noscript: true,
	atom.Form: true, atom.Input: true, atom.Button: true, atom.Select: true, atom.Textarea: true,
	atom.Svg: true, atom.Math: true, atom.Canvas: true, atom.Video: true, atom.Audio: true,
}

// htmlStyle applies an inline element's look to st; ok is false for
// elements that do not change the text's style.
func htmlStyle(a atom.Atom, attr func(string) string, st span) (span, bool) {
	switch a {
	case atom.B, atom.Strong:
		st.bold = true
	case atom.I, atom.Em, atom.Cite, atom.Var, atom.Dfn:
		st.italic = true
	case atom.S, atom.Del, atom.Strike:
		st.strike = true
	case atom.U, atom.Ins:
		st.underline = true
	case atom.Mark:
		st.mark = true
	case atom.Code, atom.Kbd, atom.Samp, atom.Tt:
		st.code = true
	case atom.Sup:
		st.superscript = true
	case atom.Sub:
		st.subscript = true
	case atom.A:
		if href := attr("href"); href != "" {
			st.link = href
		}
	default:
		return st, false
	}
	return st, true
}

// inlineTag is one raw HTML tag inside Markdown text, such as "<b>" or "</b>".
type inlineTag struct {
	atom    atom.Atom
	closing bool
	void    bool // <br>, <img>, or written as <x/>
	attrs   []html.Attribute
}

// parseInlineTag reads one tag. ok is false for comments, unknown element
// names and anything else that is not a tag of a known HTML element, which
// Markdown then shows as written: "Vec<String>" stays readable.
func parseInlineTag(raw string) (inlineTag, bool) {
	z := html.NewTokenizer(strings.NewReader(raw))
	switch z.Next() {
	case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
	default:
		return inlineTag{}, false
	}
	tok := z.Token()
	if tok.DataAtom == 0 { // not an HTML element name
		return inlineTag{}, false
	}
	return inlineTag{atom: tok.DataAtom, closing: tok.Type == html.EndTagToken,
		void:  tok.Type == html.SelfClosingTagToken || tok.DataAtom == atom.Br || tok.DataAtom == atom.Img || tok.DataAtom == atom.Hr || tok.DataAtom == atom.Wbr,
		attrs: tok.Attr}, true
}

func attrOf(attrs []html.Attribute) func(string) string {
	return func(name string) string {
		for _, a := range attrs {
			if a.Key == name {
				return a.Val
			}
		}
		return ""
	}
}

// htmlInline tracks the open inline tags while a paragraph's nodes are
// walked in order, so "<b>bold</b>" styles the text between the two tags.
type htmlInline struct {
	open []openTag
}

type openTag struct {
	atom  atom.Atom
	saved span
}

// tag applies one raw tag: it returns the style for the text that follows,
// a span to insert (a line break or an image), and whether raw was a tag.
func (h *htmlInline) tag(raw string, st span) (span, span, bool) {
	if strings.HasPrefix(raw, "<!") || strings.HasPrefix(raw, "<?") { // comments, declarations
		return st, span{}, true
	}
	t, ok := parseInlineTag(raw)
	if !ok {
		return st, span{}, false
	}
	attr := attrOf(t.attrs)
	switch {
	case t.atom == atom.Br:
		return st, span{text: "\n", bold: st.bold, italic: st.italic, link: st.link}, true
	case t.atom == atom.Img:
		return st, imageSpan(st, attr("src"), attr("alt")), true
	case t.closing:
		for i := len(h.open) - 1; i >= 0; i-- {
			if h.open[i].atom == t.atom {
				st = h.open[i].saved
				h.open = h.open[:i]
				break
			}
		}
		return st, span{}, true
	case t.void:
		return st, span{}, true
	}
	if next, styled := htmlStyle(t.atom, attr, st); styled {
		h.open = append(h.open, openTag{t.atom, st})
		return next, span{}, true
	}
	return st, span{}, true // a known element without a look, such as <span>
}

func imageSpan(st span, src, alt string) span {
	if src == "" {
		return span{}
	}
	st.imageURL, st.imageAlt = src, alt
	st.text = "[" + locale.Current().Name(locale.Current().Image, alt) + "]"
	return st
}

// htmlBlocks turns an HTML block into Markdown blocks.
func htmlBlocks(src string) []block {
	nodes, err := html.ParseFragment(strings.NewReader(src), &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body})
	if err != nil {
		return []block{{kind: codeBlock, lang: "html", code: src}}
	}
	var c htmlConv
	for _, n := range nodes {
		c.node(n)
	}
	c.flush()
	return c.out
}

// htmlConv collects blocks; loose inline content gathers into a paragraph.
type htmlConv struct {
	out    []block
	inline []span
}

func (c *htmlConv) flush() {
	if strings.TrimSpace(plain(c.inline)) != "" {
		c.out = append(c.out, block{kind: paragraph, spans: trimSpans(c.inline)})
	}
	c.inline = nil
}

func (c *htmlConv) node(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		c.inline = append(c.inline, htmlSpans(n, span{})...)
		return
	case html.ElementNode:
	default:
		return
	}
	if dropped[n.DataAtom] {
		return
	}
	switch n.DataAtom {
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		c.flush()
		level, _ := strconv.Atoi(n.Data[1:])
		c.out = append(c.out, block{kind: heading, level: level, spans: trimSpans(childSpans(n, span{}))})
	case atom.P, atom.Div, atom.Section, atom.Article, atom.Header, atom.Footer, atom.Main, atom.Aside, atom.Nav,
		atom.Figure, atom.Figcaption, atom.Center, atom.Address, atom.Dl, atom.Dt, atom.Dd, atom.Fieldset, atom.Body:
		c.flush()
		c.children(n)
		c.flush()
	case atom.Details:
		c.flush()
		c.children(n)
		c.flush()
	case atom.Summary:
		c.flush()
		c.out = append(c.out, block{kind: paragraph, spans: trimSpans(childSpans(n, span{bold: true}))})
	case atom.Blockquote:
		c.flush()
		c.out = append(c.out, block{kind: quote, children: subBlocks(n)})
	case atom.Ul, atom.Ol:
		c.flush()
		b := block{kind: list, ordered: n.DataAtom == atom.Ol, start: 1}
		if s, err := strconv.Atoi(attrOf(n.Attr)("start")); err == nil {
			b.start = s
		}
		for li := n.FirstChild; li != nil; li = li.NextSibling {
			if li.Type == html.ElementNode && li.DataAtom == atom.Li {
				b.items = append(b.items, listItem{blocks: subBlocks(li)})
			}
		}
		c.out = append(c.out, b)
	case atom.Pre:
		c.flush()
		lang := ""
		if code := n.FirstChild; code != nil && code.DataAtom == atom.Code {
			for _, cls := range strings.Fields(attrOf(code.Attr)("class")) {
				if l, ok := strings.CutPrefix(cls, "language-"); ok {
					lang = l
				}
			}
		}
		c.out = append(c.out, block{kind: codeBlock, lang: lang, code: strings.TrimSuffix(textOf(n), "\n")})
	case atom.Hr:
		c.flush()
		c.out = append(c.out, block{kind: rule})
	case atom.Table:
		c.flush()
		if t := htmlTable(n); t != nil {
			c.out = append(c.out, block{kind: table, tbl: t})
		}
	default:
		c.inline = append(c.inline, htmlSpans(n, span{})...)
	}
}

func (c *htmlConv) children(n *html.Node) {
	for k := n.FirstChild; k != nil; k = k.NextSibling {
		c.node(k)
	}
}

func subBlocks(n *html.Node) []block {
	var c htmlConv
	c.children(n)
	c.flush()
	return c.out
}

// htmlSpans is the inline content of n in style st.
func htmlSpans(n *html.Node, st span) []span {
	switch n.Type {
	case html.TextNode:
		text := n.Data
		if !st.code {
			text = collapseSpace(text)
		}
		if text == "" {
			return nil
		}
		st.text = text
		return []span{st}
	case html.ElementNode:
	default:
		return nil
	}
	if dropped[n.DataAtom] {
		return nil
	}
	attr := attrOf(n.Attr)
	switch n.DataAtom {
	case atom.Br:
		return []span{{text: "\n", bold: st.bold, italic: st.italic, link: st.link}}
	case atom.Img:
		if s := imageSpan(st, attr("src"), attr("alt")); s.text != "" {
			return []span{s}
		}
		return nil
	}
	st, _ = htmlStyle(n.DataAtom, attr, st)
	return childSpans(n, st)
}

func childSpans(n *html.Node, st span) []span {
	var out []span
	for k := n.FirstChild; k != nil; k = k.NextSibling {
		out = append(out, htmlSpans(k, st)...)
	}
	return out
}

// collapseSpace folds runs of HTML white space into one space, as browsers
// do. A no-break space is not white space here.
func collapseSpace(s string) string {
	isSpace := func(r rune) bool { return r == ' ' || r == '\n' || r == '\t' || r == '\r' || r == '\f' }
	words := strings.FieldsFunc(s, isSpace)
	if len(words) == 0 {
		if s != "" {
			return " "
		}
		return ""
	}
	out := strings.Join(words, " ")
	if isSpace(rune(s[0])) {
		out = " " + out
	}
	if isSpace(rune(s[len(s)-1])) {
		out += " "
	}
	return out
}

// trimSpans drops the white space at the start and end of a block's text.
func trimSpans(spans []span) []span {
	for len(spans) > 0 && strings.TrimSpace(spans[0].text) == "" && spans[0].imageURL == "" {
		spans = spans[1:]
	}
	for len(spans) > 0 && strings.TrimSpace(spans[len(spans)-1].text) == "" && spans[len(spans)-1].imageURL == "" {
		spans = spans[:len(spans)-1]
	}
	if len(spans) > 0 {
		spans[0].text = strings.TrimLeft(spans[0].text, " ")
		spans[len(spans)-1].text = strings.TrimRight(spans[len(spans)-1].text, " ")
	}
	return spans
}

func textOf(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for k := n.FirstChild; k != nil; k = k.NextSibling {
		b.WriteString(textOf(k))
	}
	return b.String()
}

// htmlTable reads a table's rows; the first row is the header when it is in
// <thead> or made of <th> cells.
func htmlTable(n *html.Node) *tableData {
	var rows [][]*html.Node
	var headed bool
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, head bool) {
		for k := n.FirstChild; k != nil; k = k.NextSibling {
			switch k.DataAtom {
			case atom.Thead:
				walk(k, true)
			case atom.Tbody, atom.Tfoot:
				walk(k, false)
			case atom.Tr:
				var cells []*html.Node
				allTh := true
				for cell := k.FirstChild; cell != nil; cell = cell.NextSibling {
					if cell.DataAtom == atom.Td || cell.DataAtom == atom.Th {
						cells = append(cells, cell)
						allTh = allTh && cell.DataAtom == atom.Th
					}
				}
				if len(rows) == 0 && (head || allTh) {
					headed = true
				}
				rows = append(rows, cells)
			}
		}
	}
	walk(n, false)
	if len(rows) == 0 {
		return nil
	}
	cols := 0
	for _, r := range rows {
		cols = max(cols, len(r))
	}
	d := &tableData{align: make([]cellAlign, cols)}
	for i, cell := range rows[0] {
		switch strings.ToLower(attrOf(cell.Attr)("align")) {
		case "center":
			d.align[i] = alignCenter
		case "right":
			d.align[i] = alignRight
		}
	}
	cellSpans := func(r []*html.Node) [][]span {
		out := make([][]span, cols)
		for i, cell := range r {
			out[i] = trimSpans(childSpans(cell, span{}))
		}
		return out
	}
	if !headed {
		d.header = make([][]span, cols)
	} else {
		d.header, rows = cellSpans(rows[0]), rows[1:]
	}
	for _, r := range rows {
		d.rows = append(d.rows, cellSpans(r))
	}
	return d
}

// closure is an HTML block's closing line, such as "</pre>", which goldmark
// keeps apart from its lines.
func closure(n *ast.HTMLBlock, src []byte) string {
	if n.HasClosure() {
		return "\n" + string(n.ClosureLine.Value(src))
	}
	return ""
}
