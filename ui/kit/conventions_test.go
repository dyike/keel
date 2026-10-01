package kit

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

// These tests enforce docs/kit.md so reviews do not have to.

func kitPackage(t *testing.T, dir string) (*ast.Package, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pkgs {
		return p, fset
	}
	t.Fatal("no package in " + dir)
	return nil, nil
}

func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Every component Xxx(...) *XxxView has docs, an example section and an Agent test.
func TestComponentsAreDocumentedAndExercised(t *testing.T) {
	pkg, _ := kitPackage(t, ".")
	windowTests := ""
	files, _ := filepath.Glob("../window/*_test.go")
	for _, f := range files {
		b, _ := os.ReadFile(f)
		windowTests += string(b)
	}
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
				continue
			}
			star, ok := fn.Type.Results.List[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			if id, ok := star.X.(*ast.Ident); !ok || id.Name != fn.Name.Name+"View" {
				continue // a convenience constructor, e.g. VectorIcon
			}
			name := snake(fn.Name.Name)
			for _, p := range []string{"../../docs/kit/" + name + ".md", "../../examples/components/" + name + ".go"} {
				if _, err := os.Stat(p); err != nil {
					t.Errorf("%s: missing %s", fn.Name.Name, p)
				}
			}
			if b, err := os.ReadFile("../../examples/components/" + name + ".go"); err == nil && !strings.Contains(string(b), `registerSection("`+name+`"`) {
				t.Errorf("%s: example section must be named %q", fn.Name.Name, name)
			}
			if !strings.Contains(windowTests, "kit."+fn.Name.Name+"(") {
				t.Errorf("%s: no Agent snapshot test in ui/window", fn.Name.Name)
			}
		}
	}
}

// Enum constants carry their type's name: ToneDanger, AvatarOnline, IconCheck.
func TestEnumConstantsArePrefixed(t *testing.T) {
	pkg, fset := kitPackage(t, ".")
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.CONST {
				continue
			}
			typ := ""
			for _, s := range g.Specs {
				vs := s.(*ast.ValueSpec)
				if id, ok := vs.Type.(*ast.Ident); ok {
					typ = id.Name
				} else if vs.Type != nil || len(vs.Values) > 0 {
					typ = ""
				}
				if typ == "" {
					continue
				}
				// The prefix is the type name, or the name without its last word
				// (IconName → Icon, AvatarStatus → Avatar).
				short := regexp.MustCompile(`[A-Z][a-z0-9]*$`).ReplaceAllString(typ, "")
				for _, n := range vs.Names {
					if n.IsExported() && !strings.HasPrefix(n.Name, typ) && (short == "" || !strings.HasPrefix(n.Name, short)) {
						t.Errorf("%s: %s %s must start with %s", fset.Position(n.Pos()), typ, n.Name, typ)
					}
				}
			}
		}
	}
}

// One way to do each thing: no compatibility shims or aliases in public APIs.
func TestNoCompatibilityAliases(t *testing.T) {
	for _, dir := range []string{".", "../el"} {
		pkg, _ := kitPackage(t, dir)
		d := doc.New(pkg, "", 0)
		var texts []string
		for _, ty := range d.Types {
			texts = append(texts, ty.Doc)
			for _, fn := range append(ty.Funcs, ty.Methods...) {
				texts = append(texts, fn.Name+": "+fn.Doc)
			}
		}
		for _, fn := range d.Funcs {
			texts = append(texts, fn.Name+": "+fn.Doc)
		}
		for _, s := range texts {
			if l := strings.ToLower(s); strings.Contains(l, "compatib") || strings.Contains(l, "alias") || strings.Contains(l, "deprecated") {
				t.Errorf("%s: %q", dir, s)
			}
		}
	}
}
