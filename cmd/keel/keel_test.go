package main

import (
	"bytes"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func newCLI(dir string) (*cli, *bytes.Buffer) {
	var out bytes.Buffer
	return &cli{out: &out, errw: &out, dir: dir}, &out
}

func TestNames(t *testing.T) {
	for in, want := range map[string][2]string{
		"my-notes":  {"my-notes", "My Notes"},
		"Photo Lab": {"photo-lab", "Photo Lab"},
		"todo_app":  {"todo_app", "Todo App"},
		"2048":      {"app2048", "2048"},
		"记事本":       {"app", "记事本"},
		"Notes.v2":  {"notes-v2", "Notes.v2"},
	} {
		if got := binaryName(in); got != want[0] {
			t.Errorf("binaryName(%q) = %q, want %q", in, got, want[0])
		}
		if got := displayName(in); got != want[1] {
			t.Errorf("displayName(%q) = %q, want %q", in, got, want[1])
		}
	}
	if !goAtLeast("go1.26.4", 1, 26) || !goAtLeast("go1.27rc1", 1, 26) || goAtLeast("go1.25.9", 1, 26) || goAtLeast("devel", 1, 26) {
		t.Fatal("goAtLeast")
	}
}

func TestConfigValidation(t *testing.T) {
	ok := Config{Name: "Notes", AppID: "com.example.notes", Version: "1.2.3", Binary: "notes"}
	if err := ok.validate(); err != nil || ok.fourPart() != "1.2.3.1" {
		t.Fatal(err, ok.fourPart())
	}
	for _, bad := range []Config{
		{Name: "", AppID: "com.example.notes", Version: "1.2.3", Binary: "notes"},
		{Name: "N", AppID: "notes", Version: "1.2.3", Binary: "notes"},
		{Name: "N", AppID: "com.example.notes", Version: "1.2", Binary: "notes"},
		{Name: "N", AppID: "com.example.notes", Version: "1.2.3", Binary: "my notes"},
	} {
		if bad.validate() == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
}

func TestNewAndBuildPlans(t *testing.T) {
	root := t.TempDir()
	c, out := newCLI(root)
	if code := c.main([]string{"new", "my-notes", "-offline", "-appid", "dev.keel.notes"}); code != 0 {
		t.Fatal(out.String())
	}
	dir := filepath.Join(root, "my-notes")
	for _, f := range []string{"main.go", "app.go", "go.mod", "keel.json", "appicon.png", ".gitignore", "README.md"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatal("missing", f)
		}
	}
	cfg, err := loadConfig(dir)
	if err != nil || cfg.Name != "My Notes" || cfg.AppID != "dev.keel.notes" || cfg.Binary != "my-notes" {
		t.Fatalf("config %+v %v", cfg, err)
	}
	f, _ := os.Open(filepath.Join(dir, "appicon.png"))
	img, err := png.Decode(f)
	f.Close()
	if err != nil || img.Bounds().Dx() != 1024 || img.Bounds().Dy() != 1024 {
		t.Fatal("icon size", err)
	}
	if _, _, _, a := img.At(2, 2).RGBA(); a != 0xffff {
		t.Fatal("the artwork is full bleed; keel build shapes it per platform")
	}
	if code := c.main([]string{"new", "my-notes", "-offline"}); code == 0 {
		t.Fatal("a non-empty directory must be refused")
	}

	// Build plans, printed with -n.
	c, out = newCLI(dir)
	plan := func(args ...string) string {
		t.Helper()
		out.Reset()
		if code := c.main(append([]string{"build", "-n"}, args...)); code != 0 {
			t.Fatal(out.String())
		}
		return out.String()
	}
	if p := plan("-target", "darwin", "-arch", "arm64,amd64"); !strings.Contains(p, "-target macos -arch arm64,amd64 -appid dev.keel.notes -version 0.1.0.1") ||
		!strings.Contains(p, "My Notes.app") || !strings.Contains(p, "codesign --force --deep --sign -") {
		t.Fatal("darwin plan:\n" + p)
	}
	if p := plan("-target", "darwin", "-sign", "Developer ID Application: X"); !strings.Contains(p, "--options runtime --timestamp --sign Developer ID Application: X") {
		t.Fatal("signed plan:\n" + p)
	}
	if p := plan("-target", "windows"); !strings.Contains(p, "zz_keel_windows_amd64.syso") || !strings.Contains(p, "-H=windowsgui") || !strings.Contains(p, "my-notes.exe") || !strings.Contains(p, "my-notes.ico") {
		t.Fatal("windows plan:\n" + p)
	}
	if p := plan("-target", "linux"); !strings.Contains(p, "-X gioui.org/app.ID=dev.keel.notes") || !strings.Contains(p, "dev.keel.notes.desktop") {
		t.Fatal("linux plan:\n" + p)
	}
	if p := plan("-target", "js"); !strings.Contains(p, "-target js -tags osusergo") {
		t.Fatal("js plan:\n" + p)
	}
	out.Reset()
	if code := c.main([]string{"build", "-n", "-target", "beos"}); code == 0 {
		t.Fatal("unknown target accepted")
	}
}

// The generated project compiles against this checkout of Keel.
func TestNewProjectCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a whole app")
	}
	repo, _ := filepath.Abs("../..")
	root := t.TempDir()
	c, out := newCLI(root)
	c.out, c.errw = &bytes.Buffer{}, out
	if code := c.main([]string{"new", "hello-keel", "-replace", repo}); code != 0 {
		t.Skip("go mod tidy failed (offline?):", out.String())
	}
	cmd := exec.Command("go", "vet", ".")
	cmd.Dir = filepath.Join(root, "hello-keel")
	if runtime.GOOS == "linux" {
		cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	}
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated project does not build: %v\n%s", err, b)
	}
}
