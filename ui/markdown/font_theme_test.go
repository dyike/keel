package markdown

import (
	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestRichTextFollowsThemeFace(t *testing.T) {
	old := theme.Material.Face
	defer func() { theme.Material.Face = old }()
	for _, face := range []font.Typeface{"Go", "Go Mono"} {
		theme.Material.Face = face
		r := newRich([]span{{text: "中文 English"}, {text: "bold", bold: true}, {text: "code", code: true}}, theme.BodySize, false, theme.Text, nil)
		for _, i := range []int{0, 1} {
			if r.runs[i].font.Typeface != face {
				t.Fatalf("run %d uses %q, want %q", i, r.runs[i].font.Typeface, face)
			}
		}
		if r.runs[1].font.Weight != font.Bold || r.runs[2].font.Typeface != MonoFace {
			t.Fatal("lost emphasis or code font")
		}
	}
}

func TestFontChangeInvalidatesCachedDocument(t *testing.T) {
	old := theme.Material.Face
	defer func() { theme.Material.Face = old }()
	theme.Material.Face = "Go"
	d := New("中文 English")
	d.refreshPalette()
	b := &d.chunks[0].blocks[0]
	b.view = newRich(b.spans, theme.BodySize, false, theme.Text, nil)
	before := keyForBlock(b)
	theme.Material.Face = "Go Mono"
	d.refreshPalette()
	if b.view != nil || before == keyForBlock(b) {
		t.Fatal("font change retained stale text or layout cache")
	}
	if d.Source() != "中文 English" {
		t.Fatal("font change rewrote content")
	}
}
