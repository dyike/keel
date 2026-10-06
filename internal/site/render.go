package main

import (
	"bytes"
	"html/template"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Footnote),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	// Docs may hold HTML (a <br>, a <details>); they are our own files.
	goldmark.WithRendererOptions(gmhtml.WithUnsafe(), renderer.WithNodeRenderers(util.Prioritized(codeRenderer{}, 100))),
)

// render converts one Markdown file and records its title, outline and text.
func (s *site) render(p *page) error {
	src, err := os.ReadFile(filepath.Join(s.root, filepath.FromSlash(p.Src)))
	if err != nil {
		return err
	}
	// Repository language links are replaced by the header switch on the site.
	src = []byte(languageLineRE.ReplaceAllString(string(src), ""))
	ctx := parser.NewContext(parser.WithIDs(&githubIDs{seen: map[string]int{}}))
	doc := md.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	var plain strings.Builder
	var title ast.Node     // the page title, printed by the layout instead
	var dropped []ast.Node // rows and items that point at unpublished pages
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.FencedCodeBlock:
			n.SetAttributeString("copy-label", translateUI(p, "复制", "Copy"))
		case *ast.Heading:
			t := nodeText(n, src)
			id, _ := n.AttributeString("id")
			ids, _ := id.([]byte)
			if n.Level == 1 && p.Title == "" {
				p.Title = t
				title = n
			} else if n.Level <= 3 {
				p.TOC = append(p.TOC, heading{n.Level, string(ids), t})
			}
		case *ast.Link:
			if s.unpublishedLink(p, string(n.Destination)) {
				dropped = append(dropped, container(n))
				return ast.WalkSkipChildren, nil
			}
			n.Destination = []byte(s.link(p, string(n.Destination)))
		case *ast.Image:
			n.Destination = []byte(s.link(p, string(n.Destination)))
		}
		return ast.WalkContinue, nil
	})
	if p.Title == "" {
		p.Title = strings.TrimSuffix(path.Base(p.Src), ".md")
		if path.Base(p.Src) == "README.md" {
			p.Title = path.Dir(p.Src)
		}
	}
	if title != nil {
		title.Parent().RemoveChild(title.Parent(), title)
	}
	for _, n := range dropped {
		if n.Parent() != nil {
			n.Parent().RemoveChild(n.Parent(), n)
		}
	}
	// Search text is what the page shows, after the removals.
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if t, ok := n.(*ast.Text); ok && entering {
			plain.Write(t.Segment.Value(src))
			plain.WriteByte(' ')
		}
		return ast.WalkContinue, nil
	})
	var b bytes.Buffer
	if err := md.Renderer().Render(&b, src, doc); err != nil {
		return err
	}
	p.HTML = template.HTML(s.legacyAnchors(p, b.String()))
	p.Text = strings.Join(strings.Fields(plain.String()), " ")
	return nil
}

// unpublishedLink reports whether dest leads to a page kept off the site.
func (s *site) unpublishedLink(p *page, dest string) bool {
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" {
		return false
	}
	return isUnpublished(path.Clean(path.Join(path.Dir(p.Src), u.Path)))
}

// container is what to drop with a link to an unpublished page: its table
// row or list item, or else its paragraph.
func container(n ast.Node) ast.Node {
	for c := n.Parent(); c != nil; c = c.Parent() {
		switch c.(type) {
		case *east.TableRow, *ast.ListItem:
			return c
		}
	}
	for c := n.Parent(); c != nil; c = c.Parent() {
		if _, ok := c.(*ast.Paragraph); ok {
			return c
		}
	}
	return n
}

func nodeText(n ast.Node, src []byte) string {
	var b strings.Builder
	ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if t, ok := c.(*ast.Text); ok && entering {
			b.Write(t.Segment.Value(src))
		}
		if t, ok := c.(*ast.String); ok && entering {
			b.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// link rewrites a link in page p: Markdown files become their pages, other
// repository files their GitHub view, local images are copied along.
func (s *site) link(p *page, dest string) string {
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || strings.HasPrefix(dest, "#") || dest == "" {
		return dest
	}
	rawTarget := path.Clean(path.Join(path.Dir(p.Src), u.Path))
	target := canonicalSource(rawTarget)
	frag := ""
	if u.Fragment != "" {
		frag = "#" + s.translatedFragment(target, u.Fragment)
	}
	if t, ok := s.pages[target]; ok {
		return relURL(p.Out, t.Out) + frag
	}
	// A directory with a README is that README's page.
	if t, ok := s.pages[path.Join(target, "README.md")]; ok {
		return relURL(p.Out, t.Out) + frag
	}
	// Unpublished Markdown (for example licensing) still links to its
	// actual repository file, including the selected language suffix.
	target = rawTarget
	if strings.HasSuffix(target, ".md") {
		localized := localizedSource(target, s.lang)
		if _, err := os.Stat(filepath.Join(s.root, filepath.FromSlash(localized))); err == nil {
			target = localized
		}
	}
	info, err := os.Stat(filepath.Join(s.root, filepath.FromSlash(target)))
	if err != nil || strings.HasPrefix(target, "../") {
		s.warnings = append(s.warnings, p.Src+": broken link "+dest)
		return dest
	}
	switch strings.ToLower(path.Ext(target)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
		s.copies[target] = true
		return relURL(p.Out, target)
	}
	kind := "blob"
	if info.IsDir() {
		kind = "tree"
	}
	return s.repo + "/" + kind + "/" + s.branch + "/" + target + frag
}

// githubIDs makes heading anchors the way GitHub does, so links written
// for GitHub (often to Chinese headings) work: lower case, letters and
// digits of any script kept, spaces to hyphens, other punctuation dropped.
type githubIDs struct{ seen map[string]int }

func (g *githubIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	var b strings.Builder
	for _, r := range strings.ToLower(string(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	id := b.String()
	if id == "" {
		id = "section"
	}
	if n := g.seen[id]; n > 0 {
		g.seen[id] = n + 1
		id += "-" + strconv.Itoa(n)
	} else {
		g.seen[id] = 1
	}
	return []byte(id)
}

func (g *githubIDs) Put(value []byte) { g.seen[string(value)]++ }

// codeRenderer highlights fenced code with chroma, as CSS classes so light
// and dark styles both apply.
type codeRenderer struct{}

func (codeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, func(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		block := n.(*ast.FencedCodeBlock)
		var code strings.Builder
		for i := 0; i < block.Lines().Len(); i++ {
			seg := block.Lines().At(i)
			code.Write(seg.Value(src))
		}
		lang := string(block.Language(src))
		label, _ := n.AttributeString("copy-label")
		copyLabel, _ := label.(string)
		if err := renderCode(w, lang, code.String(), copyLabel); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkSkipChildren, nil
	})
}

// Markdown snippets and gallery source share highlighting and copy controls.
func renderCode(w io.Writer, lang, code string, labels ...string) error {
	label := "Copy"
	if len(labels) > 0 && labels[0] != "" {
		label = labels[0]
	}
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	it, err := chroma.Coalesce(lexer).Tokenise(nil, code)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, `<div class="code" data-lang="`+template.HTMLEscapeString(lang)+`"><button class="copy" type="button" aria-label="`+template.HTMLEscapeString(label)+`">`+template.HTMLEscapeString(label)+`</button>`); err != nil {
		return err
	}
	if err := formatter.Format(w, styles.Get("github"), it); err != nil {
		return err
	}
	_, err = io.WriteString(w, "</div>\n")
	return err
}

var formatter = chromahtml.New(chromahtml.WithClasses(true), chromahtml.ClassPrefix("c-"))

// codeCSS is the highlighting for both color schemes.
func codeCSS() []byte {
	var b bytes.Buffer
	formatter.WriteCSS(&b, styles.Get("github"))
	b.WriteString("@media (prefers-color-scheme: dark) {\n")
	var dark bytes.Buffer
	formatter.WriteCSS(&dark, styles.Get("github-dark"))
	b.Write(dark.Bytes())
	b.WriteString("}\n")
	return b.Bytes()
}
