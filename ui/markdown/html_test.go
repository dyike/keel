package markdown

import (
	"strings"
	"testing"
)

func TestInlineHTMLStylesText(t *testing.T) {
	b := parse("a <b>bold <i>both</i></b> <u>under</u> H<sub>2</sub>O x<sup>2</sup> <mark>hi</mark> <kbd>Ctrl</kbd>" +
		" <a href=\"https://x.dev\">link</a><br>next <img src=\"p.png\" alt=\"cat\"> Vec<String> <!-- note -->")
	if len(b) != 1 {
		t.Fatalf("blocks %d", len(b))
	}
	find := func(text string) span {
		for _, s := range b[0].spans {
			if s.text == text {
				return s
			}
		}
		t.Fatalf("no span %q in %+v", text, b[0].spans)
		return span{}
	}
	if s := find("bold "); !s.bold || s.italic {
		t.Errorf("bold %+v", s)
	}
	if s := find("both"); !s.bold || !s.italic {
		t.Errorf("nested %+v", s)
	}
	if !find("under").underline || !find("2").subscript || !find("hi").mark || !find("Ctrl").code {
		t.Error("u, sub, mark or kbd lost its style")
	}
	if find("link").link != "https://x.dev" {
		t.Error("<a> lost its href")
	}
	all := plain(b[0].spans)
	if !strings.Contains(all, "link\nnext") {
		t.Errorf("<br> is not a line break: %q", all)
	}
	if !strings.Contains(all, "Vec<String>") {
		t.Errorf("an unknown tag disappeared: %q", all)
	}
	if strings.Contains(all, "<b>") || strings.Contains(all, "note") {
		t.Errorf("tags or comments shown as text: %q", all)
	}
	var img bool
	for _, s := range b[0].spans {
		img = img || s.imageURL == "p.png" && s.imageAlt == "cat"
	}
	if !img {
		t.Error("<img> is not an image")
	}
	if s := find(" "); s.bold { // after </b> the style ends
		t.Errorf("a closed tag kept styling: %+v", s)
	}
}

func TestHTMLBlocksBecomeMarkdownBlocks(t *testing.T) {
	src := `<div align="center">
<h2>Title <em>here</em></h2>
<p>First <b>para</b>.</p>
<ul><li>one</li><li>two <code>x</code></li></ul>
<table><tr><th>Name</th><th align="right">N</th></tr><tr><td>a</td><td>1</td></tr></table>
<pre><code class="language-go">fmt.Println(1)
</code></pre>
<script>alert(1)</script>
<blockquote>quoted</blockquote>
<hr>
</div>`
	bs := parse(src)
	if len(bs) != 1 || bs[0].kind != group {
		t.Fatalf("want one group, got %+v", bs)
	}
	var kinds []blockKind
	for _, c := range bs[0].children {
		kinds = append(kinds, c.kind)
	}
	want := []blockKind{heading, paragraph, list, table, codeBlock, quote, rule}
	if len(kinds) != len(want) {
		t.Fatalf("kinds %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("kinds %v, want %v", kinds, want)
		}
	}
	c := bs[0].children
	if c[0].level != 2 || plain(c[0].spans) != "Title here" || !c[0].spans[1].italic {
		t.Errorf("heading %+v", c[0])
	}
	if len(c[2].items) != 2 || plain(c[2].items[1].blocks[0].spans) != "two x" {
		t.Errorf("list %+v", c[2].items)
	}
	if tb := c[3].tbl; plain(tb.header[0]) != "Name" || tb.align[1] != alignRight || plain(tb.rows[0][1]) != "1" {
		t.Errorf("table %+v", tb)
	}
	if c[4].lang != "go" || c[4].code != "fmt.Println(1)" {
		t.Errorf("code %q %q", c[4].lang, c[4].code)
	}
	for _, b := range c {
		if strings.Contains(plain(b.spans), "alert") {
			t.Error("script content shown")
		}
	}
}

func TestDetailsWithMarkdownInside(t *testing.T) {
	bs := parse("<details>\n<summary>More</summary>\n\nHidden **text**\n\n</details>\n")
	if len(bs) < 2 || plain(bs[0].spans) != "More" || !bs[0].spans[0].bold {
		t.Fatalf("summary %+v", bs)
	}
	if plain(bs[1].spans) != "Hidden text" {
		t.Fatalf("markdown inside details %+v", bs[1])
	}
}
