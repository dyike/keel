package markdown

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestSplitKeepsFencesAndContinuations(t *testing.T) {
	src := "# 标题\n\n段落一\n\n```go\nfunc a() {\n\n\treturn\n}\n```\n\n- 一\n\n  续行\n- 二\n\n段落二\n"
	got := split(src)
	if len(got) != 5 {
		t.Fatalf("%d chunks: %q", len(got), got)
	}
	if !strings.Contains(got[2], "\treturn") || !strings.Contains(got[3], "续行") {
		t.Fatalf("a fence or a list continuation was split: %q", got)
	}
}

func TestHeal(t *testing.T) {
	for in, want := range map[string]string{
		"这是**加粗":                "这是**加粗**",
		"用 `fmt.Print":          "用 `fmt.Print`",
		"~~删除":                  "~~删除~~",
		"见 [文档](https://go.dev": "见 [文档](https://go.dev)",
		"*斜体":                   "*斜体*",
		"* 列表项 **粗":             "* 列表项 **粗**",
		"已完成的**粗体**":            "已完成的**粗体**",
		"```go\nfmt.Println(":   "```go\nfmt.Println(", // an open fence is fine as is
	} {
		if got := heal(in); got != want {
			t.Errorf("heal(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParse(t *testing.T) {
	d := New("# 标题\n\n有 **粗** 和 `代码` 和 [链接](https://go.dev)。\n\n" +
		"```go\nx := 1\n```\n\n> 引用\n\n1. 一\n2. 二\n\n- [x] 完成\n- [ ] 待办\n\n| 名 | 数 |\n|:--|--:|\n| a | 1 |\n\n---\n")
	var kinds []blockKind
	for _, c := range d.chunks {
		for _, b := range c.blocks {
			kinds = append(kinds, b.kind)
		}
	}
	want := []blockKind{heading, paragraph, codeBlock, quote, list, list, table, rule}
	if len(kinds) != len(want) {
		t.Fatalf("kinds %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("kinds %v, want %v", kinds, want)
		}
	}
	p := d.chunks[1].blocks[0]
	if len(p.spans) < 5 || !p.spans[1].bold || !p.spans[3].code || p.spans[5].link != "https://go.dev" {
		t.Fatalf("spans %+v", p.spans)
	}
	tasks := d.chunks[5].blocks[0].items
	if tasks[0].task == nil || !*tasks[0].task || tasks[1].task == nil || *tasks[1].task {
		t.Fatal("task states lost")
	}
	if tb := d.chunks[6].blocks[0].tbl; tb.align[1] != alignRight || plain(tb.rows[0][0]) != "a" {
		t.Fatalf("table %+v", tb)
	}
}

// Streaming a long answer token by token must only reparse the chunk being
// written, not the whole document.
func TestStreamingReparsesOnlyTheTail(t *testing.T) {
	answer := strings.Repeat(sample, 8) // ~2 KB, 9 chunks
	d := New("")
	d.SetStreaming(true)
	appends := 0
	for i := 0; i < len(answer); {
		n := min(4, len(answer)-i)
		for !utf8Boundary(answer, i+n) {
			n++
		}
		d.Append(answer[i : i+n])
		i += n
		appends++
	}
	d.SetStreaming(false)
	chunks := len(d.chunks)
	// Each append reparses the tail; a new chunk can reparse the previous one
	// once more (it lost its healing). Far below appends × chunks.
	if d.parses > appends+2*chunks {
		t.Fatalf("%d parses for %d appends over %d chunks", d.parses, appends, chunks)
	}
	if d.Source() != answer {
		t.Fatal("source differs from what was appended")
	}
}

func utf8Boundary(s string, i int) bool { return i >= len(s) || (s[i]&0xc0) != 0x80 }

type docView struct{ d *Doc }

func (v docView) Render(cx *el.Context) el.Element { return el.Div().P(10).Child(v.d.Render(cx)) }

func TestRenderForAgents(t *testing.T) {
	d := New("说明 [链接](https://go.dev)\n\n```go\nfmt.Println(1)\n```\n\n| 名 | 数 |\n|--|--|\n| a | 1 |\n")
	d.SetStreaming(true)
	var clicked string
	d.OnLink(func(u string) { clicked = u })
	h := uitest.New(el.Root(docView{d}))
	labels := map[string]string{} // label → role descriptions seen, joined
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label != "" {
			labels[n.Desc.Label] += n.Desc.Description + ";"
		}
	})
	for _, want := range []string{"说明 链接", "fmt.Println(1)", "go", "复制", "a | 1"} {
		if _, ok := labels[want]; !ok {
			t.Errorf("no element %q in %v", want, keys(labels))
		}
	}
	if !strings.Contains(labels["go"], "code;") || !strings.Contains(labels["a | 1"], "row;") {
		t.Errorf("roles: go=%q row=%q", labels["go"], labels["a | 1"])
	}
	_ = clicked
}

func TestCopyButton(t *testing.T) {
	d := New("```\ncopy me\n```\n")
	h := uitest.New(el.Root(docView{d}))
	var b = struct{ x, y float32 }{}
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == "复制" {
			b.x, b.y = float32(n.Desc.Bounds.Min.X+n.Desc.Bounds.Max.X)/2, float32(n.Desc.Bounds.Min.Y+n.Desc.Bounds.Max.Y)/2
		}
	})
	h.Click(b.x, b.y)
	cv := d.chunks[0].blocks[0].view.(*codeView)
	if time.Since(cv.copied) > time.Second {
		t.Fatal("copy button did not run")
	}
}

func walk(n input.SemanticNode, fn func(input.SemanticNode)) {
	fn(n)
	for _, c := range n.Children {
		walk(c, fn)
	}
}

func keys(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

// BenchmarkStreamingFrame measures one append plus one frame (render, layout,
// paint) while a long answer streams in. The frame budget at 60 Hz is 16 ms.
func BenchmarkStreamingFrame(b *testing.B) {
	for _, repeats := range []int{8, 32} {
		src := strings.Repeat(sample, repeats)
		b.Run(fmt.Sprintf("%dKB", len(src)/1024), func(b *testing.B) {
			d := New(src)
			d.SetStreaming(true)
			h := uitest.New(el.Root(docView{d}))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				d.Append("字")
				h.Frame()
			}
		})
	}
}

const sample = "## 流式渲染示例\n\n" +
	"这段回答包含 **加粗**、*斜体*、`行内代码` 和 [链接](https://go.dev)，用于测量渲染开销。\n\n" +
	"```go\nfunc main() {\n\tfor i := range 10 {\n\t\tfmt.Println(\"hello\", i)\n\t}\n}\n```\n\n" +
	"- 第一项\n- 第二项，带 `code`\n- 第三项\n\n" +
	"| 指标 | 数值 |\n|:--|--:|\n| 延迟 | 12ms |\n| 吞吐 | 3.4k |\n\n" +
	"> 引用一段话，说明注意事项。\n\n"

// Lines of highlighted code: one line break must make one line, not two.
func TestRichTextLineBreaks(t *testing.T) {
	r := highlight("go", "func main() { // 注释\n\tfmt.Println(\"hi\")\n\n}")
	uitest.New(r)
	var ys []int
	for _, p := range r.rt.pieces {
		if len(ys) == 0 || ys[len(ys)-1] != p.rect.Min.Y {
			ys = append(ys, p.rect.Min.Y)
		}
	}
	// Three lines of code and one empty line: tops at 0, h, 3h.
	if len(ys) != 3 || ys[1] <= 0 || ys[2] != 3*ys[1] {
		t.Fatalf("line tops %v, want [0 h 3h]", ys)
	}
}

// Links are their own elements for agents, and clicking one reports its URL.
func TestLinksForAgents(t *testing.T) {
	var got string
	d := New("先看 [官方文档](https://go.dev/doc) 再说。").OnLink(func(u string) { got = u })
	h := uitest.New(el.Root(docView{d}))
	var link input.SemanticNode
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == "官方文档" {
			link = n
		}
	})
	if link.Desc.Description != "link:https://go.dev/doc" || link.Desc.Class.String() != "Button" {
		t.Fatalf("link node %+v", link.Desc)
	}
	b := link.Desc.Bounds
	h.Click(float32(b.Min.X+b.Max.X)/2, float32(b.Min.Y+b.Max.Y)/2)
	if got != "https://go.dev/doc" {
		t.Fatalf("OnLink got %q", got)
	}
}

func TestSelectAndCopy(t *testing.T) {
	d := New("第一句话。第二句话。")
	h := uitest.New(el.Root(docView{d}))
	r := d.chunks[0].blocks[0].view.(*richBlock)
	// The doc is inset 10dp; drag across the first five characters.
	p := r.rt.pieces[0]
	x0 := float32(10 + p.rect.Min.X + 1)
	x1 := float32(10 + p.rect.Min.X + p.xAt(5))
	y := float32(10 + (p.rect.Min.Y+p.rect.Max.Y)/2)
	h.Drag(x0, y, x1, y)
	if got := r.rt.selectedText(); got != "第一句话。" {
		t.Fatalf("selected %q", got)
	}
	h.Key("A", key.ModShortcut)
	if got := r.rt.selectedText(); got != "第一句话。第二句话。" {
		t.Fatalf("select all gave %q", got)
	}
	h.Click(300, 280) // elsewhere: clears
	if got := r.rt.selectedText(); got != "" {
		t.Fatalf("selection kept after clicking elsewhere: %q", got)
	}
}
