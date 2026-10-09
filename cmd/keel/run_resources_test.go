package main

import (
	"debug/pe"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tc-hib/winres"
)

// Inspect a real linked PE, rather than only checking the resource generator.
// This runs on macOS/Linux too and catches missing or ignored .syso objects.
func TestRunResourcesWindowsExecutable(t *testing.T) {
	for _, artwork := range []bool{false, true} {
		name := "legacy"
		if artwork {
			name = "with-icon"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := &Config{Name: "DPI test", AppID: "com.example.dpi", Binary: "dpi", Version: "1.2.3", Main: "cmd/app", Icon: "appicon.png"}
			mainDir := filepath.Join(dir, "cmd", "app")
			if err := os.MkdirAll(mainDir, 0700); err != nil {
				t.Fatal(err)
			}
			if artwork {
				if err := writePNG(filepath.Join(dir, cfg.Icon), solid(64)); err != nil {
					t.Fatal(err)
				}
			}
			for path, data := range map[string]string{
				"go.mod":          "module dpitest\n\ngo 1.26.1\n",
				"cmd/app/main.go": "package main\nfunc main() {}\n",
			} {
				if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(path)), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cleanup, err := prepareRunResources(dir, cfg, "windows", "amd64")
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()
			binary := filepath.Join(dir, "dpi.exe")
			cmd := exec.Command("go", "build", "-o", binary, "-ldflags", runLinkFlags(cfg, "windows"), "./cmd/app")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0", "GOWORK=off")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("link: %v\n%s", err, output)
			}
			file, err := os.Open(binary)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			rs, err := winres.LoadFromEXE(file)
			if err != nil {
				t.Fatal(err)
			}
			manifest := string(rs.Get(winres.RT_MANIFEST, winres.ID(1), winres.LCIDDefault))
			for _, want := range []string{"permonitorv2", "Microsoft.Windows.Common-Controls", "longPathAware"} {
				if !strings.Contains(manifest, want) {
					t.Fatalf("manifest missing %q: %s", want, manifest)
				}
			}
			foundIcon := false
			rs.WalkType(winres.RT_GROUP_ICON, func(_ winres.Identifier, _ uint16, _ []byte) bool { foundIcon = true; return false })
			if foundIcon != artwork {
				t.Fatalf("embedded icon=%v, artwork=%v", foundIcon, artwork)
			}
			info, err := pe.Open(binary)
			if err != nil {
				t.Fatal(err)
			}
			defer info.Close()
			if info.OptionalHeader.(*pe.OptionalHeader64).Subsystem != pe.IMAGE_SUBSYSTEM_WINDOWS_GUI {
				t.Fatal("development executable opens a console window")
			}
			cleanup()
			if paths, _ := filepath.Glob(filepath.Join(mainDir, "*.syso")); len(paths) != 0 {
				t.Fatal("temporary resources leaked", paths)
			}
		})
	}
}

func TestRunResourcesPreserveExistingObjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom_windows_amd64.syso")
	if err := os.WriteFile(path, []byte("user resources"), 0600); err != nil {
		t.Fatal(err)
	}
	cleanup, err := prepareRunResources(dir, &Config{Main: "."}, "windows", "amd64")
	cleanup()
	if !errors.Is(err, errOtherSyso) {
		t.Fatalf("conflicting resources accepted: %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "user resources" {
		t.Fatal("existing resources were changed")
	}
}
