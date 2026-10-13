package deps

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSingleModuleSources(t *testing.T) {
	for _, tree := range []string{"ui", "native", "cmd", "gio", "typesetting", "third_party"} {
		root := filepath.Join("..", "..", tree)
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Name() == "go.mod" || entry.Name() == "go.sum" {
				t.Errorf("nested module file: %s", path)
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			source, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(filepath.Join("..", ".."), filepath.Dir(path))
			if err != nil {
				return err
			}
			pkg := mod + "/" + filepath.ToSlash(relative)
			for _, imported := range source.Imports {
				dependency, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					return err
				}
				if within(dependency, "gioui.org") && !within(dependency, "gioui.org/shader") || within(dependency, "github.com/go-text/typesetting") {
					t.Errorf("%s imports upstream %s", path, dependency)
				}
				if within(dependency, mod+"/third_party") {
					t.Errorf("%s imports an obsolete foundation path: %s", path, dependency)
				}
				if within(pkg, gio) || within(pkg, typesetting) {
					if within(dependency, mod) && !within(dependency, typesetting) && !(within(pkg, gio) && within(dependency, gio)) {
						t.Errorf("%s imports non-foundation %s", path, dependency)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
