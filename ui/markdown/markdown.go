// Package markdown renders Markdown as ui/el elements, built for AI chat:
// text that streams in a few characters at a time.
//
//	doc := markdown.New("")
//	doc.SetStreaming(true)
//	go func() {
//		for tok := range tokens {
//			core.Update(func() { doc.Append(tok) })
//		}
//		core.Update(func() { doc.SetStreaming(false) })
//	}()
//	... el.Div().Child(doc.Render(cx)) ...
//
// Streaming stays cheap and steady:
//   - The source is split into top-level chunks at blank lines outside code
//     and display-math fences. Each chunk is parsed once and kept; an append
//     only reparses the chunks it changed, normally just the last one.
//   - While streaming, unfinished inline syntax at the end (**bold, `code,
//     ~~strike, [link](url) is closed provisionally, so text does not flash
//     between raw markers and formatting as tokens arrive.
//   - A caret marks where text is arriving.
//
// Supported: paragraphs, headings, emphasis, strikethrough, inline code,
// links, autolinks, fenced code with syntax highlighting and a copy button,
// block quotes, ordered, unordered and task lists, GFM tables, rules, and a
// native subset of TeX mathematics. Images show as their alt text.
package markdown

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/imageload"
	"strings"
)

// Doc is a Markdown document, rendered as an el element. Change it only under
// the UI lock: from callbacks, or from other goroutines through core.Update.
type Doc struct {
	plugins           *pluginSet
	fade              streamFadeState
	preview           documentPreview
	ranges            documentRanges
	codeActions       func(*el.Context, CodeBlockContext) el.Element
	codeRenderers     map[string]func(*el.Context, CodeBlockContext) el.Element
	extensionRevision uint64
	palette           paletteKey
	src               string
	streaming         bool
	chunks            []chunk
	onLink            func(url string)
	selection         documentSelection
	contextual        bool
	parsedContext     string
	pendingAnchor     string
	images            map[string]*imageload.Asset
	imageLoader       imageload.Loader
	showFrontMatter   bool

	parses int // chunks parsed so far, for tests
}

type chunk struct {
	src    string // as parsed: healed while it was the streaming tail
	raw    string // as written
	blocks []block
}

// New creates a document from Markdown source.
func New(src string) *Doc {
	d := &Doc{}
	d.SetSource(src)
	return d
}

// OnLink sets what happens when a link is clicked; by default nothing.
func (d *Doc) OnLink(fn func(url string)) *Doc { d.onLink = fn; return d }

// Source returns the Markdown as written.
func (d *Doc) Source() string { return d.src }

// Streaming reports whether text is still arriving.
func (d *Doc) Streaming() bool { return d.streaming }

// SetSource replaces the document. Unchanged leading chunks keep their parse.
// If the source changes, the selection is cleared; Append preserves it.
func (d *Doc) SetSource(src string) {
	if d.src != src {
		d.selection.clear()
		d.ranges.reveal = nil
	}
	d.src = src
	d.update()
}

// Append adds text at the end, e.g. the next tokens of a streaming answer.
func (d *Doc) Append(s string) {
	if s != "" {
		d.ranges.reveal = nil
	}
	d.src += s
	d.update()
}

// SetStreaming marks whether text is still arriving: while it is, unfinished
// syntax at the end is closed provisionally and a caret shows. Set it to
// false when the answer is complete.
func (d *Doc) SetStreaming(on bool) {
	if d.streaming != on {
		d.streaming = on
		d.update()
	}
}

func (d *Doc) update() {
	if d.plugins != nil && !d.plugins.local || strings.Contains(d.src, "]:") || strings.Contains(d.src, "[^") || hasMathMacros(d.src) {
		d.updateContextual()
		return
	}
	if d.contextual {
		d.chunks = nil
		d.contextual = false
		d.parsedContext = ""
	}
	raws := split(d.src)
	next := make([]chunk, len(raws))
	for i, raw := range raws {
		src := raw
		if d.streaming && i == len(raws)-1 {
			src = heal(raw)
		}
		if i < len(d.chunks) && d.chunks[i].src == src {
			next[i] = d.chunks[i] // unchanged: keep blocks and their drawing state
			continue
		}
		next[i] = chunk{src: src, raw: raw, blocks: d.parseSource(src)}
		if i < len(d.chunks) {
			preserveCodeViews(d.chunks[i].blocks, next[i].blocks)
		}
		d.parses++
	}
	d.chunks = next
}

// split cuts Markdown into top-level chunks at blank lines outside code
// and display-math fences. A blank line followed by an indented line does not split: that line
// continues a list item or an indented code block.
func split(src string) []string {
	// Front matter may hold blank lines: it is one chunk of its own.
	if block, _, rest := splitFrontMatter(src); block != "" {
		return append([]string{block}, split(rest)...)
	}
	var chunks []string
	var cur strings.Builder
	fence := "" // the open fence marker, e.g. "```"
	blank := false
	for _, line := range strings.SplitAfter(src, "\n") {
		if line == "" {
			continue
		}
		trim := strings.TrimSpace(line)
		if fence == "" {
			if blank && trim != "" && !startsIndented(line) && cur.Len() > 0 {
				chunks = append(chunks, cur.String())
				cur.Reset()
			}
			if trim == "$$" {
				fence = "$$"
			} else if m := fenceMarker(trim); m != "" {
				fence = m
			}
		} else if strings.HasPrefix(trim, fence) && strings.TrimLeft(trim, fence[:1]) == "" {
			fence = ""
		}
		blank = fence == "" && trim == ""
		cur.WriteString(line)
	}
	if cur.Len() > 0 {
		chunks = append(chunks, cur.String())
	}
	return chunks
}

func startsIndented(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// fenceMarker returns the run of ``` or ~~~ that opens a fence, or "".
func fenceMarker(trim string) string {
	for _, c := range []string{"`", "~"} {
		n := len(trim) - len(strings.TrimLeft(trim, c))
		if n >= 3 {
			return strings.Repeat(c, n)
		}
	}
	return ""
}

// heal closes inline syntax left open at the end of a streaming chunk, so the
// text renders as it will once the closing marker arrives. Open code fences
// need nothing: the parser runs them to the end.
func heal(src string) string {
	if fenceOpen(src) {
		return src
	}
	// Work on the last paragraph only; earlier ones in the chunk are complete.
	start := strings.LastIndex(strings.TrimRight(src, "\n"), "\n\n") + 1
	tail := src[start:]
	var closers string
	if strings.Count(tail, "`")%2 == 1 {
		closers += "`"
		tail += "`"
	}
	outside := stripMath(stripCode(tail))
	if i := strings.LastIndex(outside, "]("); i >= 0 && !strings.Contains(outside[i:], ")") {
		closers += ")"
	}
	if strings.Count(outside, "~~")%2 == 1 {
		closers += "~~"
	}
	bold := strings.Count(outside, "**")
	singles := strings.Count(strings.ReplaceAll(outside, "**", ""), "*") - listBullets(outside)
	if singles%2 == 1 {
		closers += "*"
	}
	if bold%2 == 1 {
		closers += "**"
	}
	if closers == "" {
		return src
	}
	return strings.TrimRight(src, "\n") + closers
}

func fenceOpen(src string) bool {
	open := ""
	for _, line := range strings.Split(src, "\n") {
		trim := strings.TrimSpace(line)
		if open == "" {
			open = fenceMarker(trim)
			if trim == "$$" {
				open = "$$"
			}
		} else if strings.HasPrefix(trim, open) && strings.TrimLeft(trim, open[:1]) == "" {
			open = ""
		}
	}
	return open != ""
}

// stripCode removes inline code spans, whose markers do not count.
func stripCode(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		if r == '`' {
			in = !in
			continue
		}
		if !in {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// listBullets counts "* " list markers at line starts, which are not emphasis.
func listBullets(s string) int {
	n := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " "), "* ") {
			n++
		}
	}
	return n
}
