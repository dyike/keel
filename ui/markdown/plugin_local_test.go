package markdown

import (
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/yuin/goldmark/ast"
)

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
