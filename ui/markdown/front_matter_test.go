package markdown

import (
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestFrontMatter(t *testing.T) {
	src := "---\ntitle: \"发布说明\"\ntags: [keel, gio]\n\nauthor: yike\n---\n# 正文\n\n第一段。\n\n---\n\nnote: 不是元数据\n"
	d := New(src)
	if got := d.Meta(); got["title"] != "发布说明" || got["tags"] != "[keel, gio]" || got["author"] != "yike" {
		t.Fatalf("meta %v", got)
	}
	if !strings.HasPrefix(d.FrontMatter(), "title:") {
		t.Fatalf("raw front matter %q", d.FrontMatter())
	}
	bs := documentBlocks(d)
	if bs[0].kind != frontMatter || bs[1].kind != heading || plain(bs[1].spans) != "正文" {
		t.Fatalf("blocks: %v %v", bs[0].kind, bs[1].kind)
	}
	// A later --- is a rule, and what follows is text, not metadata.
	var rules, texts int
	for _, b := range bs[1:] {
		if b.kind == rule {
			rules++
		}
		if strings.Contains(plain(b.spans), "note: 不是元数据") {
			texts++
		}
	}
	if rules != 1 || texts != 1 {
		t.Fatalf("later --- misread: rules=%d texts=%d", rules, texts)
	}
	h := uitest.New(el.Root(docView{d}))
	h.Frame()
	if shownText(h, "发布说明") {
		t.Fatal("front matter shown by default")
	}
	d.ShowFrontMatter(true)
	h.Frame()
	if !shownText(h, "发布说明") || !shownText(h, "author") {
		t.Fatal("ShowFrontMatter did not show the table")
	}

	// Unclosed while streaming: shown as written until the closing line.
	s := New("")
	s.SetStreaming(true)
	s.Append("---\ntitle: x\n")
	if s.FrontMatter() != "" || documentBlocks(s)[0].kind == frontMatter {
		t.Fatal("unclosed front matter taken as metadata")
	}
	s.Append("---\n正文\n")
	if s.Meta()["title"] != "x" || documentBlocks(s)[0].kind != frontMatter {
		t.Fatal("closed front matter not recognized")
	}
	if New("no front matter\n---\n").FrontMatter() != "" {
		t.Fatal("front matter found away from the start")
	}
}

func shownText(h *uitest.Harness, text string) bool {
	for _, n := range h.Router.AppendSemantics(nil) {
		if strings.Contains(n.Desc.Label, text) {
			return true
		}
	}
	return false
}
