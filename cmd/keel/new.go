package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/dyike/keel/internal/svgicon"
)

//go:embed template
var templates embed.FS

func (c *cli) newProject(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(c.errw)
	module := fs.String("module", "", "Go module path (default: the directory name)")
	name := fs.String("name", "", "app name shown to people (default: from the directory)")
	appID := fs.String("appid", "", "reverse-DNS app ID (default: com.example.<binary>)")
	keel := fs.String("keel", keelVersion(), "Keel version to require")
	replace := fs.String("replace", "", "use a local Keel checkout at this path (for developing Keel)")
	offline := fs.Bool("offline", false, "write the files only; do not fetch dependencies")
	fs.Usage = func() {
		fmt.Fprintln(c.errw, "Usage: keel new <dir> [flags]")
		fs.PrintDefaults()
	}
	dirArg, rest := splitPositional(args)
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if dirArg == "" && fs.NArg() > 0 {
		dirArg = fs.Arg(0)
	}
	if dirArg == "" {
		fs.Usage()
		return errors.New("missing project directory")
	}
	dir := dirArg
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(c.wd(), dir)
	}
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return fmt.Errorf("%s is not empty", dirArg)
	}
	base := filepath.Base(dir)
	binary := binaryName(base)
	cfg := &Config{
		Name: *name, AppID: *appID, Version: "0.1.0", Build: 1,
		Binary: binary, Icon: "appicon.png", Main: ".",
	}
	if cfg.Name == "" {
		cfg.Name = displayName(base)
	}
	if cfg.AppID == "" {
		cfg.AppID = "com.example." + strings.ReplaceAll(binary, "_", "-")
	}
	if *module == "" {
		*module = binary
	}
	if err := cfg.validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data := map[string]string{"Name": cfg.Name, "Module": *module}
	for out, tmpl := range map[string]string{
		"main.go": "main.go.tmpl", "app.go": "app.go.tmpl",
		".gitignore": "gitignore.tmpl", "README.md": "README.md.tmpl",
	} {
		if err := writeTemplate(filepath.Join(dir, out), tmpl, data); err != nil {
			return err
		}
	}
	icon, err := placeholderIcon()
	if err != nil {
		return fmt.Errorf("placeholder icon: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, cfg.Icon), icon, 0o644); err != nil {
		return err
	}
	if err := cfg.save(dir); err != nil {
		return err
	}
	gomod := "module " + *module + "\n\ngo 1.26\n"
	if *replace != "" {
		abs, err := filepath.Abs(*replace)
		if err != nil {
			return err
		}
		gomod += "\nrequire github.com/dyike/keel v0.0.0\n\nreplace github.com/dyike/keel => " + abs + "\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		return err
	}
	if !*offline {
		if *replace == "" {
			if err := c.command(dir, nil, "go", "get", "github.com/dyike/keel@"+*keel); err != nil {
				return err
			}
		}
		if err := c.command(dir, nil, "go", "mod", "tidy"); err != nil {
			return err
		}
	}
	fmt.Fprintf(c.out, "Created %s (%s, %s).\n\n  cd %s\n  keel run\n", cfg.Name, cfg.AppID, *module, dirArg)
	return nil
}

func writeTemplate(path, name string, data any) error {
	src, err := templates.ReadFile("template/" + name)
	if err != nil {
		return err
	}
	t, err := template.New(name).Parse(string(src))
	if err != nil {
		return err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// splitPositional takes a leading non-flag argument, so both
// "keel new dir -flag" and "keel new -flag dir" work.
func splitPositional(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	return "", args
}

// binaryName turns a directory name into a lowercase executable name.
func binaryName(base string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(base) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case unicode.IsSpace(r) || r == '.':
			b.WriteRune('-')
		}
	}
	s := strings.Trim(b.String(), "-_")
	if s == "" || s[0] >= '0' && s[0] <= '9' {
		s = "app" + s
	}
	return s
}

// displayName turns "my-app" into "My App".
func displayName(base string) string {
	words := strings.FieldsFunc(base, func(r rune) bool { return r == '-' || r == '_' || unicode.IsSpace(r) })
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	if len(words) == 0 {
		return "App"
	}
	return strings.Join(words, " ")
}

// placeholderIcon is the template icon on Apple's grid: an 824px body
// centered in a 1024px canvas, so it sits level with other Dock icons.
func placeholderIcon() ([]byte, error) {
	svg, _ := templates.ReadFile("template/appicon.svg")
	body, err := svgicon.Render(svg, 824)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	canvas := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	draw.Draw(canvas, image.Rect(100, 100, 924, 924), img, image.Point{}, draw.Over)
	var b bytes.Buffer
	if err := png.Encode(&b, canvas); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
