package golang_test

import (
	"reflect"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/highlight"
	"github.com/dyike/keel/ui/highlight/golang"
)

func TestParityWithFullRegistry(t *testing.T) {
	compact, full := golang.Chroma{}, highlight.Chroma{}
	for _, name := range []string{"Go", "go", "golang", "main.go", "text", "plain", "notes.txt", "go-text-template"} {
		if got, want := compact.Language(name), full.Language(name); got != want {
			t.Errorf("%s: %q != %q", name, got, want)
		}
	}
	samples := []string{
		"package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"你好\") }\n",
		"// comment\n/* block */\nvar x = 0xFF + 3.14i\nvar r = '字'\n",
		"var tpl = `Hello {{ .Name }} {{if .OK}}ok{{end}}`\n",
		"func broken( {\n\"unterminated",
	}
	for _, sample := range samples {
		for _, lang := range []string{"go", "text", "go-text-template"} {
			for _, dark := range []bool{false, true} {
				for _, style := range []string{"", "monokai", "unknown-style"} {
					opts := core.HighlightOptions{Language: lang, Dark: dark, Style: style, Guess: lang == ""}
					got, ok := compact.Highlight(sample, opts)
					want, wantOK := full.Highlight(sample, opts)
					if ok != wantOK || !reflect.DeepEqual(got, want) {
						t.Fatalf("lang=%q dark=%v style=%q sample=%q: token mismatch", lang, dark, style, sample)
					}
				}
			}
		}
	}
	opts := core.HighlightOptions{Guess: true}
	got, ok := compact.Highlight(samples[0], opts)
	want, wantOK := full.Highlight(samples[0], opts)
	if ok != wantOK || !reflect.DeepEqual(got, want) {
		t.Fatal("Go detection differs")
	}
	if compact.Language("python") != "" {
		t.Fatal("unexpected full language registry")
	}
	if _, ok := compact.Highlight("print('hi')", core.HighlightOptions{Language: "python"}); ok {
		t.Fatal("unsupported language should fall back to plain text")
	}
}
