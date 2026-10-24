package main

import (
	"bytes"
	"go/format"
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

func testCheckout(t *testing.T) string {
	t.Helper()
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return repo
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
	for v, want := range map[string]bool{
		"go1.26.1": true, "go1.26.4": true, "go1.27rc1": true, "go1.27.0": true,
		"go1.26.0": false, "go1.26rc2": false, "go1.25.9": false, "devel": false,
	} {
		if goAtLeast(v, "1.26.1") != want {
			t.Errorf("goAtLeast(%q, 1.26.1) = %v", v, !want)
		}
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
	if code := c.main([]string{"new", "my-notes", "-offline", "-replace", testCheckout(t), "-appid", "dev.keel.notes"}); code != 0 {
		t.Fatal(out.String())
	}
	dir := filepath.Join(root, "my-notes")
	for _, f := range []string{"main.go", "app.go", "go.mod", "keel.json", "appicon.png", ".gitignore", "README.md", "README.zh-CN.md"} {
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
	if p := plan("-target", "linux"); !strings.Contains(p, "-X github.com/dyike/keel/gio/app.ID=dev.keel.notes") || !strings.Contains(p, "dev.keel.notes.desktop") {
		t.Fatal("linux plan:\n" + p)
	}
	if p := plan("-target", "windows"); !strings.Contains(p, "go build -trimpath -ldflags -s -w -H=windowsgui") {
		t.Fatal("release builds strip symbols and paths:\n" + p)
	}
	if p := plan("-target", "windows", "-debug"); strings.Contains(p, "-s -w") || !strings.Contains(p, "-ldflags -H=windowsgui") {
		t.Fatal("-debug keeps symbols:\n" + p)
	}
	if p := plan("-target", "darwin"); !strings.Contains(p, "-ldflags -s -w -o") {
		t.Fatal("darwin release flags:\n" + p)
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
	// The generated example test drives the app in memory.
	test := exec.Command("go", "test", ".")
	test.Dir, test.Env = cmd.Dir, cmd.Env
	if b, err := test.CombinedOutput(); err != nil {
		t.Fatalf("generated app_test.go fails: %v\n%s", err, b)
	}
	// Gio and go-text come from the checkout's Keel module; no upstream
	// core module is required or replaced.
	list := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}} {{with .Module}}{{.Path}}{{end}}", ".")
	list.Dir = cmd.Dir
	b, err := list.Output()
	if err != nil {
		t.Fatal(err)
	}
	foundGio := false
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		pkg, module, _ := strings.Cut(line, " ")
		if strings.HasPrefix(pkg, "github.com/dyike/keel/gio/") {
			foundGio = true
			if module != "github.com/dyike/keel" {
				t.Fatalf("%s is in module %s", pkg, module)
			}
		}
		if module == "gioui.org" || module == "github.com/go-text/typesetting" {
			t.Fatalf("app depends on upstream %s", pkg)
		}
	}
	if !foundGio {
		t.Fatal("app does not use Keel's Gio")
	}
}

// The generated app speaks the chosen language, tells the truth about
// keel run's watching, and is gofmt-clean in both languages.
func TestNewLanguage(t *testing.T) {
	for lang, want := range map[string][]string{
		"":      {"Enter your name", "keel run rebuilds and restarts"},
		"en":    {"Enter your name", "keel run rebuilds and restarts"},
		"zh-CN": {"输入你的名字", "keel run 会自动重新编译并重启"},
	} {
		root := t.TempDir()
		c, out := newCLI(root)
		args := []string{"new", "app", "-offline", "-replace", testCheckout(t)}
		if lang != "" {
			args = append(args, "-lang", lang)
		}
		if code := c.main(args); code != 0 {
			t.Fatal(out.String())
		}
		src, err := os.ReadFile(filepath.Join(root, "app", "app.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range want {
			if !strings.Contains(string(src), w) {
				t.Errorf("-lang %q: app.go lacks %q", lang, w)
			}
		}
		if formatted, err := format.Source(src); err != nil || !bytes.Equal(formatted, src) {
			t.Errorf("-lang %q: app.go is not gofmt-clean: %v", lang, err)
		}
	}
	c, _ := newCLI(t.TempDir())
	if c.main([]string{"new", "app", "-offline", "-lang", "fr", "-replace", testCheckout(t)}) == 0 {
		t.Fatal("an unknown -lang must be refused")
	}
}

// Generated projects and keel doctor require the Go version Keel does.
func TestMinGoVersionMatchesModule(t *testing.T) {
	mod, err := os.ReadFile(filepath.Join(testCheckout(t), "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "\ngo "+minGoVersion+"\n") {
		t.Fatalf("minGoVersion %s is not the go directive of Keel's go.mod", minGoVersion)
	}
	root := t.TempDir()
	c, out := newCLI(root)
	if c.main([]string{"new", "app", "-offline", "-replace", testCheckout(t)}) != 0 {
		t.Fatal(out.String())
	}
	gen, _ := os.ReadFile(filepath.Join(root, "app", "go.mod"))
	if !strings.Contains(string(gen), "\ngo "+minGoVersion+"\n") {
		t.Fatalf("generated go.mod:\n%s", gen)
	}
}

// -h prints a command's flags and succeeds; every spelling of version works.
func TestHelpAndVersion(t *testing.T) {
	for _, cmd := range []string{"new", "run", "build", "icon", "doctor"} {
		c, out := newCLI(t.TempDir())
		if code := c.main([]string{cmd, "-h"}); code != 0 || strings.Contains(out.String(), "help requested") || !strings.Contains(out.String(), "-") {
			t.Errorf("keel %s -h: exit %d\n%s", cmd, code, out.String())
		}
	}
	for _, arg := range []string{"version", "-version", "--version"} {
		c, out := newCLI(t.TempDir())
		if code := c.main([]string{arg}); code != 0 || !strings.HasPrefix(out.String(), "keel ") || strings.Contains(out.String(), "latest") {
			t.Errorf("keel %s: exit %d %q", arg, code, out.String())
		}
	}
}

func TestDropLines(t *testing.T) {
	var b bytes.Buffer
	d := &dropLines{w: &b, suffix: ": replacing existing signature"}
	for _, chunk := range []string{"dist/A.app: replacing exi", "sting signature\nwarning: keep\r\n", "x.app: replacing existing signature\n", "tail"} {
		d.Write([]byte(chunk))
	}
	d.Flush()
	if b.String() != "warning: keep\r\ntail" {
		t.Fatalf("%q", b.String())
	}
}

// keel new writes the example test only when the required Keel has
// windowtest, and the CI workflow only on request.
func TestNewExampleTestAndCI(t *testing.T) {
	gen := func(args ...string) string {
		t.Helper()
		root := t.TempDir()
		c, out := newCLI(root)
		if code := c.main(append([]string{"new", "app", "-offline"}, args...)); code != 0 {
			t.Fatal(out.String())
		}
		return filepath.Join(root, "app")
	}
	exists := func(p string) bool { _, err := os.Stat(p); return err == nil }

	dir := gen("-replace", testCheckout(t))
	mod, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	if !exists(filepath.Join(dir, "app_test.go")) || exists(filepath.Join(dir, ".github")) || strings.Contains(string(mod), "\ntool ") {
		t.Fatal("default project: want app_test.go, no workflow, no tool directive")
	}

	dir = gen("-replace", testCheckout(t), "-ci", "github")
	want, _ := templates.ReadFile("template/github-build.yml")
	got, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "build.yml"))
	if err != nil || !bytes.Equal(got, want) || !bytes.Contains(got, []byte("${{ matrix.os }}")) {
		t.Fatalf("workflow not copied verbatim: %v", err)
	}
	mod, _ = os.ReadFile(filepath.Join(dir, "go.mod"))
	if !strings.Contains(string(mod), "\ntool github.com/dyike/keel/cmd/keel\n") {
		t.Fatalf("go.mod lacks the keel tool:\n%s", mod)
	}

	// An older release has neither windowtest nor KeepState: no example test
	// and no KeepState, either of which would not compile.
	dir = gen("-keel", "v0.3.8")
	if exists(filepath.Join(dir, "app_test.go")) {
		t.Fatal("app_test.go written for a release without windowtest")
	}
	for _, f := range []string{"main.go", "app.go"} {
		src, _ := os.ReadFile(filepath.Join(dir, f))
		if bytes.Contains(src, []byte("KeepState")) || bytes.Contains(src, []byte("a.state")) {
			t.Fatalf("%s uses KeepState with v0.3.8", f)
		}
		if formatted, err := format.Source(src); err != nil || !bytes.Equal(formatted, src) {
			t.Fatalf("%s for v0.3.8 is not gofmt-clean", f)
		}
	}

	c, _ := newCLI(t.TempDir())
	if c.main([]string{"new", "app", "-offline", "-ci", "gitlab", "-replace", testCheckout(t)}) == 0 {
		t.Fatal("an unknown -ci must be refused")
	}
}
