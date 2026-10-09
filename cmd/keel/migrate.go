package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Keel carries its own copies of Gio and go-text. Apps that imported the
// upstream packages next to Keel's must use the copies, or their types would
// not match the ones Keel's API takes.
var migratedImports = []struct{ from, to string }{
	{"gioui.org/", "github.com/dyike/keel/third_party/gio/"},
	{"github.com/go-text/typesetting/", "github.com/dyike/keel/third_party/typesetting/"},
}

// migrateImport returns the path Keel's copy has for an upstream import.
// gioui.org/shader stays: it is a separate module Keel still requires.
func migrateImport(path string) (string, bool) {
	if path == "gioui.org/shader" || strings.HasPrefix(path, "gioui.org/shader/") {
		return "", false
	}
	for _, m := range migratedImports {
		if strings.HasPrefix(path, m.from) {
			return m.to + strings.TrimPrefix(path, m.from), true
		}
		if path == strings.TrimSuffix(m.from, "/") {
			return strings.TrimSuffix(m.to, "/"), true
		}
	}
	return "", false
}

// migrateSource rewrites the import paths of one Go file; it returns nil
// when nothing changes. Only import specs change, through their positions in
// the file, so comments and formatting stay as they are.
func migrateSource(name string, src []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	last := 0
	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		to, ok := migrateImport(path)
		if !ok {
			continue
		}
		start, end := fset.Position(spec.Path.Pos()).Offset, fset.Position(spec.Path.End()).Offset
		out.Write(src[last:start])
		out.WriteString(strconv.Quote(to))
		last = end
	}
	if last == 0 {
		return nil, nil
	}
	out.Write(src[last:])
	return out.Bytes(), nil
}

func (c *cli) migrate(args []string) error {
	fset := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fset.SetOutput(c.errw)
	fset.Usage = func() {
		fmt.Fprintln(c.errw, "usage: keel migrate [dir]\n\nRewrites gioui.org and go-text imports to Keel's copies under\ngithub.com/dyike/keel/third_party, then runs go mod tidy.")
	}
	if err := fset.Parse(args); err != nil {
		return err
	}
	dir := c.wd()
	if fset.NArg() > 0 {
		dir = fset.Arg(0)
	}
	changed := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); path != dir && (strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" || name == "dist") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out, err := migrateSource(path, src)
		if err != nil || out == nil {
			return err
		}
		if c.dryRun {
			fmt.Fprintln(c.out, "rewrite", path)
		} else if err := os.WriteFile(path, out, 0o644); err != nil {
			return err
		}
		changed++
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "%d files use Keel's Gio and go-text\n", changed)
	if changed == 0 {
		return nil
	}
	return c.command(dir, nil, "go", "mod", "tidy")
}
