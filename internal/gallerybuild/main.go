// gallerybuild publishes a complete gallery and independently linked category
// programs. It edits temporary Go overlays, never the checked-out examples.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type registration struct {
	name, category string
	start, end     int
}
type source struct {
	path          string
	data          []byte
	registrations []registration
}
type group struct {
	Wasm     string   `json:"wasm"`
	Script   string   `json:"script"`
	Sections []string `json:"sections"`
}
type manifest struct {
	Groups map[string]group `json:"groups"`
}

func scan(path string, data []byte) (source, error) {
	src := source{path: path, data: data}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, path, data, 0)
	if err != nil {
		return src, err
	}
	ast.Inspect(f, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}
		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		name, ok := call.Fun.(*ast.Ident)
		if !ok || name.Name != "registerSection" || len(call.Args) != 3 {
			return true
		}
		first, a := call.Args[0].(*ast.BasicLit)
		second, b := call.Args[1].(*ast.BasicLit)
		if !a || !b {
			err = fmt.Errorf("%s: registration names must be literals", path)
			return false
		}
		label, e1 := strconv.Unquote(first.Value)
		category, e2 := strconv.Unquote(second.Value)
		if e1 != nil || e2 != nil || !safeName(label) || !safeName(category) {
			err = fmt.Errorf("%s: invalid registration", path)
			return false
		}
		src.registrations = append(src.registrations, registration{label, category, fs.Position(stmt.Pos()).Offset, fs.Position(stmt.End()).Offset})
		return false
	})
	return src, err
}
func safeName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '_' && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}
func prune(src source, category string) []byte {
	data := string(src.data)
	// Retain metadata for the full index, but don't reference excluded builders.
	// The compiler/linker can then discard their code and resources.
	for i := len(src.registrations) - 1; i >= 0; i-- {
		r := src.registrations[i]
		if r.category == category {
			continue
		}
		replacement := "registerSection(" + strconv.Quote(r.name) + "," + strconv.Quote(r.category) + ",nil)"
		// Keep the old expression type-checked, so its imports remain valid. Its
		// constant-false branch contributes no builder roots to the linker.
		data = data[:r.start] + "if false {" + data[r.start:r.end] + "};" + replacement + data[r.end:]
	}
	return []byte(data)
}
func run() error {
	out := flag.String("out", "demo", "output directory")
	categories := flag.String("categories", "", "comma-separated categories; empty builds all")
	flag.Parse()
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(root, "examples/components/*.go"))
	if err != nil {
		return err
	}
	var sources []source
	groups := map[string][]string{}
	seen := map[string]bool{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src, err := scan(path, data)
		if err != nil {
			return err
		}
		if len(src.registrations) == 0 {
			continue
		}
		sources = append(sources, src)
		for _, r := range src.registrations {
			if seen[r.name] {
				return fmt.Errorf("duplicate section %s", r.name)
			}
			seen[r.name] = true
			groups[r.category] = append(groups[r.category], r.name)
		}
	}
	if len(groups) == 0 {
		return fmt.Errorf("run from the Keel repository root")
	}
	var names []string
	if *categories == "" {
		for name := range groups {
			names = append(names, name)
		}
	} else {
		names = strings.Split(*categories, ",")
	}
	sort.Strings(names)
	for _, name := range names {
		if _, ok := groups[name]; !ok {
			return fmt.Errorf("unknown category %q", name)
		}
	}
	temp, err := os.MkdirTemp("", "keel-gallery-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	build := func(dest, overlay, category string) error {
		tags := "osusergo"
		switch category {
		case "", "inputs", "display", "shell", "feedback", "chat":
		default:
			tags += ",keel_gallery_small"
		}
		cmd := exec.Command("go", "run", "./cmd/gogio", "-target", "js", "-tags", tags, "-ldflags=-s -w", "-o", dest, "./examples/components")
		cmd.Env = os.Environ()
		// GOFLAGS also reaches gogio's subprocess build. The outer tool remains
		// native; setting GOOS here would attempt to execute a wasm gogio binary.
		flags := os.Getenv("GOFLAGS")
		if overlay != "" {
			flags += " " + strconv.Quote("-overlay="+overlay)
		}
		cmd.Env = append(cmd.Env, "GOFLAGS="+flags, "CGO_ENABLED=0")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}
	if err := build(*out, "", ""); err != nil {
		return err
	}
	m := manifest{Groups: map[string]group{}}
	for _, name := range names {
		replace := map[string]string{}
		for i, src := range sources {
			path := filepath.Join(temp, fmt.Sprintf("%s-%d.go", name, i))
			if err := os.WriteFile(path, prune(src, name), 0600); err != nil {
				return err
			}
			replace[src.path] = path
		}
		data, _ := json.Marshal(struct{ Replace map[string]string }{replace})
		overlay := filepath.Join(temp, name+".json")
		if err := os.WriteFile(overlay, data, 0600); err != nil {
			return err
		}
		dest := filepath.Join(*out, "groups", name)
		fmt.Println("Building gallery category:", name)
		if err := build(dest, overlay, name); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		sort.Strings(groups[name])
		m.Groups[name] = group{Wasm: "groups/" + name + "/main.wasm", Script: "groups/" + name + "/wasm.js", Sections: groups[name]}
	}
	data, _ := json.MarshalIndent(m, "", "  ")
	return os.WriteFile(filepath.Join(*out, "groups.json"), data, 0644)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
