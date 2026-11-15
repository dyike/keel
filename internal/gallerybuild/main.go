// gallerybuild publishes a shared-runtime gallery with lazy category modules.
// It edits temporary Go overlays, never the checked-out examples.
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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type registration struct {
	name, category   string
	start, end       int
	argStart, argEnd int
	funcLit          bool
}
type source struct {
	path          string
	data          []byte
	registrations []registration
}
type group struct {
	Wasm     string   `json:"wasm"`
	Script   string   `json:"script"`
	Category string   `json:"category,omitempty"`
	Sections []string `json:"sections"`
}
type manifest struct {
	Groups     map[string]group    `json:"groups"`
	Categories map[string][]string `json:"categories,omitempty"`
}

func shardGroup(section, category string) string {
	switch section {
	case "input", "input_group", "otp_input":
		return "inputs"
	case "code_editor":
		return "code_editor"
	case "emoji_picker":
		return "emoji_picker"
	case "calendar", "color_picker", "date_picker", "time_field":
		return "pickers"
	case "checkbox", "combobox", "form", "number_input", "questionnaire", "radio", "radio_group", "rating", "select", "slider", "switch":
		return "choices"
	case "text_view":
		return "text_view"
	case "avatar", "avatar_group", "carousel", "image":
		return "media"
	case "table", "table_data", "table_static":
		return "table"
	case "diff", "terminal":
		return "terminal"
	case "plot", "plot_primitives":
		return "plot"
	case "attachment", "attachment_group", "attachment_vertical":
		return "attachment"
	case "dock", "gooey_dock":
		return "dock"
	case "nav_stack", "sidebar":
		return "nav_stack"
	case "command", "command_async", "menu":
		return "command"
	case "popover", "sheet":
		return "popover"
	case "resizable", "resizable_group", "settings":
		return "settings"
	case "motion_gestures", "motion_layout", "motion_pages", "motion_pose", "motion_presets", "motion_scroll", "motion_timeline", "motion_values":
		return "motion_lab"
	default:
		return category
	}
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
		_, funcLit := call.Args[2].(*ast.FuncLit)
		src.registrations = append(src.registrations, registration{
			name:     label,
			category: category,
			start:    fs.Position(stmt.Pos()).Offset,
			end:      fs.Position(stmt.End()).Offset,
			argStart: fs.Position(call.Args[2].Pos()).Offset,
			argEnd:   fs.Position(call.Args[2].End()).Offset,
			funcLit:  funcLit,
		})
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
	return rewrite(src, func(r registration) bool { return r.category == category })
}
func pruneShard(src source, shard string) []byte {
	return rewrite(src, func(r registration) bool {
		return shard == "*" || shardGroup(r.name, r.category) == shard
	})
}
func rewrite(src source, keep func(registration) bool) []byte {
	data := string(src.data)
	var lifted strings.Builder
	// Retain metadata for the full index, but don't reference excluded builders.
	// Lift anonymous builder closures to top-level named functions so symbols do
	// not depend on init closure numbering and can be split into child modules.
	for i := len(src.registrations) - 1; i >= 0; i-- {
		r := src.registrations[i]
		call := data[r.start:r.end]
		if r.funcLit {
			fn := "section_" + r.name
			lifted.WriteString("\nfunc ")
			lifted.WriteString(fn)
			lifted.WriteString(string(src.data[r.argStart+len("func") : r.argEnd]))
			lifted.WriteByte('\n')
			call = string(src.data[r.start:r.argStart]) + fn + string(src.data[r.argEnd:r.end])
		}
		if keep(r) {
			data = data[:r.start] + call + data[r.end:]
			continue
		}
		replacement := "registerSection(" + strconv.Quote(r.name) + "," + strconv.Quote(r.category) + ",nil)"
		// Keep the old expression type-checked, so its imports remain valid. Its
		// constant-false branch contributes no builder roots to the linker.
		data = data[:r.start] + "if false {" + call + "};" + replacement + data[r.end:]
	}
	return []byte(data + lifted.String())
}
func run() error {
	out := flag.String("out", "demo", "output directory")
	categories := flag.String("categories", "", "comma-separated categories; empty builds all")
	analysis := flag.Bool("analysis", false, "keep WASM symbols, identical optional imports and a builder-free baseline for ownership analysis")
	mode := flag.String("mode", "shared", "publication mode: shared or independent")
	binaryen := flag.String("binaryen-dir", os.Getenv("BINARYEN_DIR"), "Binaryen version_133 bin directory (or BINARYEN_DIR)")
	flag.Parse()
	if *mode != "shared" && *mode != "independent" {
		return fmt.Errorf("unknown build mode %q", *mode)
	}
	if !*analysis && *categories == "" && *mode == "shared" {
		return buildShared(*out, *binaryen)
	}
	if *analysis && *categories != "" {
		return fmt.Errorf("ownership analysis requires all categories; omit -categories")
	}
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
	groupCategory := map[string]string{}
	categoryGroups := map[string][]string{}
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
			key := r.category
			if *analysis {
				key = shardGroup(r.name, r.category)
			}
			groups[key] = append(groups[key], r.name)
			groupCategory[key] = r.category
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
		cat := groupCategory[name]
		categoryGroups[cat] = append(categoryGroups[cat], name)
	}
	temp, err := os.MkdirTemp("", "keel-gallery-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	gogioBin := filepath.Join(temp, "gogio")
	if runtime.GOOS == "windows" {
		gogioBin += ".exe"
	}
	buildGogio := exec.Command("go", "build", "-o", gogioBin, "./cmd/gogio")
	buildGogio.Env = append(os.Environ(), "CGO_ENABLED=0")
	buildGogio.Stdout, buildGogio.Stderr = os.Stdout, os.Stderr
	if err := buildGogio.Run(); err != nil {
		return fmt.Errorf("build gogio: %w", err)
	}
	writeOverlay := func(name string, transform func(source) []byte) (string, error) {
		replace := map[string]string{}
		for i, src := range sources {
			path := filepath.Join(temp, fmt.Sprintf("%s-%03d-%s", name, i, filepath.Base(src.path)))
			if err := os.WriteFile(path, transform(src), 0600); err != nil {
				return "", err
			}
			replace[src.path] = path
		}
		data, _ := json.Marshal(struct{ Replace map[string]string }{replace})
		overlay := filepath.Join(temp, name+".json")
		if err := os.WriteFile(overlay, data, 0600); err != nil {
			return "", err
		}
		return overlay, nil
	}
	build := func(dest, overlay, category string) error {
		tags := "osusergo,keel_lazy_highlight"
		switch {
		case *analysis:
		case category == "" || category == "inputs" || category == "display" || category == "data":
		default:
			tags += ",keel_gallery_small"
		}
		ldflags := "-s -w"
		if *analysis {
			ldflags = ""
		}
		cmd := exec.Command(gogioBin, "-target", "js", "-tags", tags, "-ldflags="+ldflags, "-o", dest, "./examples/components")
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
	fullOverlay := ""
	if *analysis {
		fullOverlay, err = writeOverlay("full", func(src source) []byte { return pruneShard(src, "*") })
		if err != nil {
			return err
		}
	}
	if err := build(*out, fullOverlay, ""); err != nil {
		return err
	}
	m := manifest{Groups: map[string]group{}}
	if *analysis {
		m.Categories = categoryGroups
	}
	buildNames := append([]string(nil), names...)
	if *analysis {
		buildNames = append(buildNames, "baseline")
	}
	var (
		wg       sync.WaitGroup
		sem      = make(chan struct{}, min(4, max(1, runtime.NumCPU())))
		firstErr error
		errOnce  sync.Once
	)
	for _, name := range buildNames {
		name := name
		overlay, err := writeOverlay(name, func(src source) []byte {
			if *analysis {
				return pruneShard(src, name)
			}
			return prune(src, name)
		})
		if err != nil {
			return err
		}
		dest := filepath.Join(*out, "groups", name)
		if name != "baseline" {
			sort.Strings(groups[name])
			m.Groups[name] = group{Wasm: "groups/" + name + "/main.wasm", Script: "groups/" + name + "/wasm.js", Category: groupCategory[name], Sections: groups[name]}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fmt.Println("Building gallery category:", name)
			if err := build(dest, overlay, name); err != nil {
				errOnce.Do(func() { firstErr = fmt.Errorf("%s: %w", name, err) })
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
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
