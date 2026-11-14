package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestPruneKeepsMetadataAndSelectedBuilders(t *testing.T) {
	input := []byte(`package main
import "fmt"
func init(){registerSection("button","general",button);registerSection("editor","inputs",func() Widget {fmt.Println("editor");return editor()})}
func unrelated(){println("untouched")}
`)
	src, err := scan("fixture.go", input)
	if err != nil {
		t.Fatal(err)
	}
	out := prune(src, "general")
	if _, err = parser.ParseFile(token.NewFileSet(), "overlay.go", out, 0); err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{`registerSection("button","general",button)`, `if false {registerSection("editor"`, `registerSection("editor","inputs",nil)`, `func unrelated(){println("untouched")}`} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %s in %s", want, s)
		}
	}
	if string(src.data) != string(input) {
		t.Fatal("mutated input")
	}
}
func TestScanRejectsDynamicOrUnsafeNames(t *testing.T) {
	for _, args := range []string{`name,"general",builder`, `"button","../escape",builder`} {
		if _, err := scan("bad.go", []byte("package main;func init(){registerSection("+args+")}")); err == nil {
			t.Fatal(args)
		}
	}
}
