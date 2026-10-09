package deps

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// Framework text goes through ui/locale so apps can switch languages; only
// glyph probes used to measure CJK line height may spell Chinese directly.
func TestNoHardcodedFrameworkText(t *testing.T) {
	probes := map[string]bool{"国": true, "国Ag": true}
	for _, dir := range []string{"ui/core", "ui/el", "ui/kit", "ui/window", "ui/markdown", "ui/internal/editorstyle", "ui/internal/imageload"} {
		files, _ := filepath.Glob(filepath.Join("..", "..", dir, "*.go"))
		for _, f := range files {
			// Native QA fixtures (-tags keelnativeqa) type sample IME text.
			if strings.HasSuffix(f, "_test.go") || strings.HasPrefix(filepath.Base(f), "native_qa_") {
				continue
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, f, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				s, _ := strconv.Unquote(lit.Value)
				if probes[s] {
					return true
				}
				for _, r := range s {
					if unicode.Is(unicode.Han, r) {
						t.Errorf("%s: %s: use ui/locale for framework text", fset.Position(lit.Pos()), lit.Value)
						break
					}
				}
				return true
			})
		}
	}
}
