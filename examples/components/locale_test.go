package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/dyike/keel/ui/locale"
)

func TestDemoTextLanguages(t *testing.T) {
	previous := locale.Current()
	defer locale.Apply(previous)
	locale.Apply(locale.English())
	if got := demoText("Save", "保存"); got != "Save" {
		t.Fatal(got)
	}
	locale.Apply(locale.Chinese())
	if got := demoText("Save", "保存"); got != "保存" {
		t.Fatal(got)
	}
}

// Check actual gallery sources: every Chinese example string must have an
// English counterpart, and formatting arguments must remain compatible.
func TestGalleryTranslationCoverage(t *testing.T) {
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	verbs := regexp.MustCompile(`%[-+# 0]*[0-9]*(?:\.[0-9]+)?[a-zA-Z%]`)
	fs := token.NewFileSet()
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fs, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		translated := map[*ast.BasicLit]bool{}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if !ok || id.Name != "demoText" || len(call.Args) != 2 {
				return true
			}
			en, ok1 := call.Args[0].(*ast.BasicLit)
			zh, ok2 := call.Args[1].(*ast.BasicLit)
			if !ok1 || !ok2 {
				t.Errorf("%s: translation needs two literal strings", fs.Position(call.Pos()))
				return true
			}
			a, _ := strconv.Unquote(en.Value)
			b, _ := strconv.Unquote(zh.Value)
			translated[zh] = true
			if strings.Join(verbs.FindAllString(a, -1), ",") != strings.Join(verbs.FindAllString(b, -1), ",") {
				t.Errorf("%s: format arguments differ: %q / %q", fs.Position(call.Pos()), a, b)
			}
			return true
		})
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || translated[lit] {
				return true
			}
			value, _ := strconv.Unquote(lit.Value)
			if value == "中文" {
				return true
			} // Language selector's native name.
			for _, r := range value {
				if unicode.Is(unicode.Han, r) {
					t.Errorf("%s: untranslated example %q", fs.Position(lit.Pos()), value)
					break
				}
			}
			return true
		})
	}
}
