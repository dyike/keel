package deps

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const mod = "github.com/dyike/keel"

// Keel's copies of Gio and go-text are packages of this module, but they are
// foundations, not Keel modules: ui may build on them, native and the CLI
// must not, and they never import Keel's own packages.
const (
	gio         = mod + "/gio"
	typesetting = mod + "/typesetting"
)

func within(pkg, root string) bool { return pkg == root || strings.HasPrefix(pkg, root+"/") }

// allowed lists, for every module, the other Keel packages it may import,
// including indirect imports. Changing this table changes the architecture:
// update the module READMEs and docs/architecture.md too.
var allowed = map[string][]string{
	"ui/core":                  {"ui/internal/loop"},
	"ui/theme":                 {"ui/internal/loop"},
	"ui/locale":                {"ui/internal/loop"},
	"ui/base":                  {},
	"ui/plot":                  {"ui/core", "ui/theme", "ui/internal/loop"},
	"ui/motion":                {"ui/el", "ui/core", "ui/theme", "ui/locale", "ui/internal/loop", "ui/internal/editorstyle", "ui/internal/inputcontent"},
	"ui/highlight":             {"ui/core", "ui/internal/loop"},
	"ui/netimage":              {"ui/core", "ui/internal/loop"},
	"ui/kit":                   {"ui/base", "ui/core", "ui/theme", "ui/locale", "ui/el", "ui/internal/loop", "ui/internal/editorstyle", "ui/internal/inputcontent"},
	"ui/window":                {"ui/core", "ui/theme", "ui/internal/loop", "ui/internal/appkit", "ui/internal/wayland", "internal/appicon"},
	"ui/el":                    {"ui/core", "ui/theme", "ui/locale", "ui/internal/loop", "ui/internal/editorstyle", "ui/internal/inputcontent"},
	"ui/markdown":              {"ui/el", "ui/core", "ui/theme", "ui/locale", "ui/internal/imageload", "ui/internal/loop", "ui/internal/editorstyle", "ui/internal/inputcontent"},
	"ui/internal/imageload":    {"ui/core", "ui/theme", "ui/locale", "ui/internal/loop"},
	"ui/internal/editorstyle":  {},
	"ui/internal/inputcontent": {},
	"native":                   {},
	"native/internal/sys":      {"native"},
	"native/internal/wlclip":   {"native"},
	"native/process":           {"native", "native/internal/sys"},
	"native/permission":        {"native", "native/internal/sys"},
	"native/screen":            {"native", "native/internal/sys"},
	"native/input":             {"native", "native/internal/sys"},
	"native/hotkey":            {"native", "native/internal/sys"},
	"native/notification":      {"native", "native/internal/sys"},
	"native/tray":              {"native", "native/internal/sys"},
	"native/clipboard":         {"native", "native/internal/sys", "native/internal/wlclip"},
	// Talks to apps only through the automation protocol, never Keel's code.
	"cmd/keel-mcp": {},
	"cmd/keel":     {"internal/svgicon", "internal/appicon", "internal/gogio"},
	"cmd/gogio":    {"internal/gogio"},
	// The packager is shared by both commands (internal/gogio/README.md).
	"internal/gogio": {},
}

func TestModuleBoundaries(t *testing.T) {
	for pkg, ok := range allowed {
		out, err := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", mod+"/"+pkg).Output()
		if err != nil {
			t.Fatalf("go list %s: %v", pkg, err)
		}
		for _, dep := range strings.Fields(string(out)) {
			switch {
			case dep == mod+"/"+pkg:
			case within(dep, gio) || within(dep, typesetting):
				if strings.HasPrefix(pkg, "native") || strings.HasPrefix(pkg, "cmd/") || pkg == "internal/gogio" {
					t.Errorf("%s imports %s: it must work without the GUI", pkg, dep)
				}
			case strings.HasPrefix(dep, mod+"/"):
				if rel := strings.TrimPrefix(dep, mod+"/"); !slices.Contains(ok, rel) {
					t.Errorf("%s imports %s, which it is not allowed to depend on", pkg, rel)
				}
			case strings.HasPrefix(dep, "gioui.org/") && !strings.HasPrefix(dep, "gioui.org/shader"),
				dep == "github.com/go-text/typesetting" || strings.HasPrefix(dep, "github.com/go-text/typesetting/"):
				t.Errorf("%s imports upstream %s instead of Keel's root-level copy", pkg, dep)
			}
		}
	}
}

// Every module directory must be listed above, so a new module can't skip the check.
func TestEveryModuleIsListed(t *testing.T) {
	out, err := exec.Command("go", "list", mod+"/...").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		rel := strings.TrimPrefix(p, mod+"/")
		if within(p, gio) || within(p, typesetting) || strings.HasPrefix(rel, "examples/") || strings.HasPrefix(rel, "internal/") || strings.HasPrefix(rel, "ui/internal/") || strings.Contains(rel, "/testdata/") {
			continue
		}
		if _, ok := allowed[rel]; !ok {
			t.Errorf("module %s is missing from the allowed table", rel)
		}
	}
}

// Gio may build on go-text and its own packages, go-text only on itself;
// neither reaches into Keel, though they now share its module path.
func TestFoundationBoundaries(t *testing.T) {
	for _, root := range []string{gio, typesetting} {
		out, err := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", root+"/...").Output()
		if err != nil {
			t.Fatalf("go list %s: %v", root, err)
		}
		for _, dep := range strings.Fields(string(out)) {
			switch {
			case within(dep, typesetting), root == gio && within(dep, gio):
			case strings.HasPrefix(dep, mod+"/"):
				t.Errorf("%s depends on %s", root, dep)
			case strings.HasPrefix(dep, "gioui.org/") && !strings.HasPrefix(dep, "gioui.org/shader"),
				strings.HasPrefix(dep, "github.com/go-text/typesetting"):
				t.Errorf("%s depends on upstream %s", root, dep)
			}
		}
	}
}

// kit may reach implementation helpers transitively through el/core/theme,
// but its own components must stay on their public APIs.
func TestKitDirectDependencies(t *testing.T) {
	out, err := exec.Command("go", "list", "-f", `{{range .Imports}}{{println .}}{{end}}`, mod+"/ui/kit").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, dep := range strings.Fields(string(out)) {
		if strings.HasPrefix(dep, mod+"/") && !within(dep, gio) && !within(dep, typesetting) && !slices.Contains([]string{mod + "/ui/base", mod + "/ui/core", mod + "/ui/theme", mod + "/ui/locale", mod + "/ui/el"}, dep) {
			t.Errorf("kit directly imports %s", dep)
		}
	}
}

// The UI modules stay slim: syntax highlighting (chroma) and network
// images (net/http, TLS) cost about 4 MB each and are opt-in packages, so a
// package below must never import them, directly or not.
func TestHeavyDependenciesAreOptIn(t *testing.T) {
	heavy := []string{"net/http", "crypto/tls", "github.com/alecthomas/chroma/v2"}
	for _, pkg := range []string{"ui/core", "ui/el", "ui/kit", "ui/markdown", "ui/motion", "ui/window"} {
		out, err := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", mod+"/"+pkg).Output()
		if err != nil {
			t.Fatalf("go list %s: %v", pkg, err)
		}
		for _, dep := range strings.Fields(string(out)) {
			if slices.Contains(heavy, dep) {
				t.Errorf("%s pulls in %s; keep it behind ui/highlight or ui/netimage", pkg, dep)
			}
		}
	}
}
