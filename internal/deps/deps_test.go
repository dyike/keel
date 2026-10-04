package deps

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const mod = "github.com/dyike/keel"

// allowed lists, for every module, the other Keel packages it may import,
// including indirect imports. Changing this table changes the architecture:
// update the module READMEs and docs/architecture.md too.
var allowed = map[string][]string{
	"ui/core":                  {"ui/internal/loop"},
	"ui/theme":                 {"ui/internal/loop"},
	"ui/locale":                {"ui/internal/loop"},
	"ui/base":                  {},
	"ui/plot":                  {"ui/core", "ui/theme", "ui/internal/loop"},
	"ui/kit":                   {"ui/base", "ui/core", "ui/theme", "ui/locale", "ui/el", "ui/internal/loop", "ui/internal/editorstyle", "ui/internal/inputcontent"},
	"ui/window":                {"ui/core", "ui/theme", "ui/internal/loop"},
	"ui/el":                    {"ui/core", "ui/theme", "ui/locale", "ui/internal/loop", "ui/internal/editorstyle", "ui/internal/inputcontent"},
	"ui/markdown":              {"ui/el", "ui/core", "ui/theme", "ui/locale", "ui/internal/imageload", "ui/internal/loop", "ui/internal/editorstyle", "ui/internal/inputcontent"},
	"ui/internal/imageload":    {"ui/core", "ui/theme", "ui/locale", "ui/internal/loop"},
	"ui/internal/editorstyle":  {},
	"ui/internal/inputcontent": {},
	"native":                   {},
	"native/internal/sys":      {"native"},
	"native/internal/wlclip":   {"native"},
	"native/permission":        {"native", "native/internal/sys"},
	"native/screen":            {"native", "native/internal/sys"},
	"native/input":             {"native", "native/internal/sys"},
	"native/hotkey":            {"native", "native/internal/sys"},
	"native/notification":      {"native", "native/internal/sys"},
	"native/clipboard":         {"native", "native/internal/sys", "native/internal/wlclip"},
	// Talks to apps only through the automation protocol, never Keel's code.
	"cmd/keel-mcp": {},
	"cmd/keel":     {"internal/svgicon"},
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
			case strings.HasPrefix(dep, mod+"/"):
				if rel := strings.TrimPrefix(dep, mod+"/"); !slices.Contains(ok, rel) {
					t.Errorf("%s imports %s, which it is not allowed to depend on", pkg, rel)
				}
			case (strings.HasPrefix(pkg, "native") || strings.HasPrefix(pkg, "cmd/")) && strings.HasPrefix(dep, "gioui.org"):
				t.Errorf("%s imports %s: it must work without the GUI", pkg, dep)
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
		if strings.HasPrefix(rel, "examples/") || strings.HasPrefix(rel, "internal/") || strings.HasPrefix(rel, "ui/internal/") || strings.Contains(rel, "/testdata/") {
			continue
		}
		if _, ok := allowed[rel]; !ok {
			t.Errorf("module %s is missing from the allowed table", rel)
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
		if strings.HasPrefix(dep, mod+"/") && !slices.Contains([]string{mod + "/ui/base", mod + "/ui/core", mod + "/ui/theme", mod + "/ui/locale", mod + "/ui/el"}, dep) {
			t.Errorf("kit directly imports %s", dep)
		}
	}
}
