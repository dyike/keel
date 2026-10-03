package kit

import (
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestLabelMatchesAndMask(t *testing.T) {
	v := Label("中a中a").Highlights("中").Secondary("必填")
	text, r := v.content()
	if text != "中a中a 必填" || len(r) != 3 || r[0].Start != 0 || r[0].End != 1 || r[1].Start != 2 || r[2].Start != 5 || r[2].Color != theme.Muted {
		t.Fatalf("%q %+v", text, r)
	}
	v.HighlightPrefix("a")
	_, r = v.content()
	if len(r) != 1 {
		t.Fatal(r)
	}
	v.HighlightPrefix("中")
	_, r = v.content()
	if len(r) != 2 {
		t.Fatal(r)
	}
	v.Masked(true)
	text, r = v.content()
	if text != "•••• 必填" || len(r) != 1 {
		t.Fatalf("%q %+v", text, r)
	}
	v.SetText("🔒秘密")
	text, _ = v.content()
	if text != "••• 必填" {
		t.Fatal(text)
	}
	v.Masked(false).Secondary("").Highlights("")
	text, r = v.content()
	if text != "🔒秘密" || len(r) != 0 {
		t.Fatalf("%q %+v", text, r)
	}
}
