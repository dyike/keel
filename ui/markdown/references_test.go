package markdown

import (
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func documentBlocks(d *Doc) []block {
	var out []block
	for _, c := range d.chunks {
		out = append(out, c.blocks...)
	}
	return out
}
func TestDocumentWideReferenceLinks(t *testing.T) {
	d := New("[first][shared]\n\nintervening paragraph\n\n[collapsed][] and [shortcut]\n\n[shared]: https://example.com/first\n[collapsed]: https://example.com/second\n[shortcut]: https://example.com/third")
	bs := documentBlocks(d)
	if len(bs) != 3 {
		t.Fatalf("definitions rendered as content: %+v", bs)
	}
	if bs[0].spans[0].link != "https://example.com/first" {
		t.Fatal("forward reference not resolved")
	}
	if bs[2].spans[0].link != "https://example.com/second" || bs[2].spans[2].link != "https://example.com/third" {
		t.Fatalf("collapsed/shortcut references: %+v", bs[2].spans)
	}
	d.SetSource("[mixed CASE]\n\n[mixed case]: https://example.com")
	if documentBlocks(d)[0].spans[0].link != "https://example.com" {
		t.Fatal("reference labels should be case insensitive")
	}
}

func TestFootnotesNumberingAndRichContent(t *testing.T) {
	d := New("First[^b], second[^a], repeat[^b].\n\n[^a]: alpha\n\n[^b]: **bold** note.\n\n    Another paragraph with `code`.\n\n[^unused]: hidden")
	bs := documentBlocks(d)
	if len(bs) != 2 || bs[1].kind != footnoteList {
		t.Fatalf("missing footer: %+v", bs)
	}
	if plain(bs[0].spans) != "First[1], second[2], repeat[1]." {
		t.Fatal("footnotes not numbered by first use")
	}
	notes := bs[1].children
	if len(notes) != 2 || notes[0].level != 1 || notes[1].level != 2 {
		t.Fatalf("wrong note ordering: %+v", notes)
	}
	if len(notes[0].children) != 2 || !notes[0].children[0].spans[0].bold {
		t.Fatal("lost multi-paragraph formatted footnote")
	}
	last := notes[0].children[1].spans
	if last[len(last)-2].link != footnoteRefID(1, 0) || last[len(last)-1].link != footnoteRefID(1, 1) {
		t.Fatalf("missing backlinks to repeated references: %+v", last)
	}
}

func clickReference(t *testing.T, h *uitest.Harness, label string) {
	t.Helper()
	var rect image.Rectangle
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == label && n.Desc.Class.String() == "Button" {
			rect = n.Desc.Bounds
		}
	})
	if rect.Empty() {
		t.Fatalf("link %q not visible", label)
	}
	h.Click(float32(rect.Min.X+rect.Max.X)/2, float32(rect.Min.Y+rect.Max.Y)/2)
	h.Frame()
	h.Frame()
}
func TestFootnoteJumpAndRepeatedReturn(t *testing.T) {
	external := 0
	d := New("第一处[^note]。\n\n" + strings.Repeat("中间段落\n\n", 6) + "第二处[^note]。\n\n" + strings.Repeat("后面的段落\n\n", 6) + "[^note]: 这是脚注正文。")
	d.OnLink(func(string) { external++ })
	h := uitest.New(el.Root(scrollingDocView{d}))
	clickReference(t, h, "[1]")
	clickReference(t, h, "↩2")
	textPoint(t, h, d, "第二处[1]。", 0)
	if external != 0 {
		t.Fatal("internal navigation opened an external URL")
	}
}

func TestReferencesStreamingResolvesAndKeepsViews(t *testing.T) {
	d := New("unchanged\n\n```text\nlong code\n```\n\n[docs][id]")
	d.SetStreaming(true)
	h := uitest.New(el.Root(docView{d}))
	cv := d.chunks[1].blocks[0].view.(*codeView)
	clickCodeButton(t, h, "自动换行")
	d.Append("\n\n[id]: https://example.com")
	h.Frame()
	if d.chunks[1].blocks[0].view != cv || !cv.wrap {
		t.Fatal("forward definition reset code state")
	}
	head := &d.chunks[0].blocks[0]
	if documentBlocks(d)[2].spans[0].link != "https://example.com" {
		t.Fatal("streamed forward reference unresolved")
	}
	d.Append("\n\nnew tail")
	h.Frame()
	if &d.chunks[0].blocks[0] != head {
		t.Fatal("unchanged block lost its cache identity")
	}
	d.SetStreaming(false)
	h.Frame()
	d.SetSource("ordinary\n\nparagraphs")
	if d.contextual || len(d.chunks) != 2 {
		t.Fatal("ordinary document did not restore incremental parsing")
	}
}
