package markdown

import (
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/yuin/goldmark/ast"
)

func TestLocalPluginCustomContentRemainsDynamic(t *testing.T) {
	for _, source := range []string{
		"# Heading\n\nplain",
		"before `tag` after",
		"> before `tag` after",
		"- before `tag` after",
		"| name | value |\n|---|---|\n| before | `tag` |",
		"| `tag` | value |\n|---|---|\n| before | after |",
	} {
		t.Run(source, func(t *testing.T) {
			height := float32(20)
			p := Plugin{Local: true,
				Blocks: map[ast.NodeKind]func(ast.Node, []byte) el.View{
					ast.KindHeading: func(ast.Node, []byte) el.View {
						return el.ViewFunc(func(*el.Context) el.Element { return el.Div().H(el.Dp(height)).Child(el.Text("custom heading")) })
					},
				},
				Inlines: map[ast.NodeKind]func(ast.Node, []byte) InlineObject{
					ast.KindCodeSpan: func(ast.Node, []byte) InlineObject {
						return InlineObject{Text: "tag", Widget: core.Func(func(gtx core.C) core.D { return core.D{Size: gtx.Constraints.Constrain(image.Pt(60, int(height)))} })}
					},
				},
			}
			d := New(source).Plugins(p)
			root := el.Embed(d)
			var size image.Point
			h := uitest.NewFunc(func(gtx core.C) { size = root.Layout(gtx).Size })
			before, text, parses := size, d.RenderedText(), d.parses
			height = 80
			h.Frame()
			if size.Y < before.Y+50 {
				t.Fatalf("custom content froze in a cached parent: %v -> %v", before, size)
			}
			if d.RenderedText() != text || d.parses != parses {
				t.Fatal("dynamic layout changed text or reparsed source")
			}
		})
	}
}

// A Local plugin keeps blank-line chunking: streaming with an app's inline or
// block renderers reparses the tail, not the whole answer on every token.
func TestLocalPluginStreamsIncrementally(t *testing.T) {
	headings := 0
	local := Plugin{Local: true, Blocks: map[ast.NodeKind]func(ast.Node, []byte) el.View{
		ast.KindHeading: func(ast.Node, []byte) el.View { headings++; return nil },
	}}
	answer := strings.Repeat("## Title\n\n"+sample, 6)
	for _, p := range []Plugin{local, {Blocks: local.Blocks}} {
		d := New("").Plugins(p)
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
		if p.Local {
			if d.contextual || d.parses > appends+2*len(d.chunks) {
				t.Fatalf("local plugin: %d parses for %d appends over %d chunks", d.parses, appends, len(d.chunks))
			}
			if headings == 0 {
				t.Fatal("chunked parsing skipped the plugin's block factory")
			}
		} else if !d.contextual {
			t.Fatal("a plugin that is not Local must parse the whole document")
		}
		if d.Source() != answer {
			t.Fatal("source differs from what was appended")
		}
	}
}
