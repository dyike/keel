package deps

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Verify target selection, not just host imports. Initializers in these
// packages survive dead-code elimination even when no native API uses them.
func TestNativePlatformDependencies(t *testing.T) {
	for _, target := range []string{"darwin", "windows", "linux"} {
		t.Run(target, func(t *testing.T) {
			cmd := exec.Command("go", "list", "-tags=", "-deps", "-f", "{{.ImportPath}}", mod+"/native/...")
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+target, "GOARCH=amd64", "GOWORK=off")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%v: %s", err, out)
			}
			for _, root := range []string{"github.com/jezek/xgb", "github.com/godbus/dbus/v5"} {
				found := false
				for _, dep := range strings.Fields(string(out)) {
					if within(dep, root) {
						found = true
					}
				}
				if found != (target == "linux") {
					t.Errorf("%s dependency present=%v on %s", root, found, target)
				}
			}
		})
	}
}

func TestDefaultSVGDependencies(t *testing.T) {
	cmd := exec.Command("go", "list", "-tags=", "-deps", "-f", "{{.ImportPath}}", mod+"/ui/kit")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	for _, dep := range strings.Fields(string(out)) {
		if within(dep, "golang.org/x/text/encoding") || dep == "golang.org/x/net/html/charset" || dep == "github.com/srwiley/oksvg" {
			t.Errorf("default kit imports legacy SVG dependency %s", dep)
		}
	}
}

func TestNativeAppDoesNotImportFallbackWidgets(t *testing.T) {
	for _, target := range []string{"darwin", "windows"} {
		cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", gio+"/app")
		cmd.Env = append(os.Environ(), "GOOS="+target, "GOARCH=amd64", "CGO_ENABLED=0", "GOWORK=off")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v: %s", target, err, out)
		}
		for _, dep := range strings.Fields(string(out)) {
			if within(dep, gio+"/widget") || within(dep, gio+"/text") || within(dep, typesetting) {
				t.Errorf("%s app retains client-side title bar dependency %s", target, dep)
			}
		}
	}
}

func TestNoOpenGLDependencies(t *testing.T) {
	for _, target := range []string{"darwin", "linux", "windows"} {
		cmd := exec.Command("go", "list", "-tags=noopengl", "-deps", "-f", "{{.ImportPath}}", gio+"/gpu/headless")
		cmd.Env = append(os.Environ(), "GOOS="+target, "GOARCH=amd64", "CGO_ENABLED=0", "GOWORK=off")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v: %s", target, err, out)
		}
		for _, dep := range strings.Fields(string(out)) {
			if dep == gio+"/gpu/internal/opengl" || dep == gio+"/internal/egl" {
				t.Errorf("%s noopengl retains %s", target, dep)
			}
		}
	}
}
