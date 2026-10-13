package main

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// moduleProxy serves fake Keel releases from a file proxy, with a fresh
// module cache, so resolution runs real Go downloads without the network.
func moduleProxy(t *testing.T) string {
	t.Helper()
	proxy := t.TempDir()
	t.Setenv("GOPROXY", "file://"+filepath.ToSlash(proxy))
	t.Setenv("GOSUMDB", "off")
	t.Setenv("GONOPROXY", "none")
	t.Setenv("GOPRIVATE", "")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	t.Setenv("GOTOOLCHAIN", "local")
	cache := t.TempDir()
	t.Setenv("GOMODCACHE", cache)
	// Go makes extracted module directories read-only.
	t.Cleanup(func() {
		filepath.WalkDir(cache, func(path string, e os.DirEntry, err error) error {
			if err == nil && e.IsDir() {
				return os.Chmod(path, 0o755)
			}
			return err
		})
	})
	return proxy
}

func writeProxyRelease(t *testing.T, proxy, version string) {
	t.Helper()
	dir := filepath.Join(proxy, filepath.FromSlash(keelModulePath), "@v")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	mod := []byte("module " + keelModulePath + "\n\ngo 1.20\n")
	info, _ := json.Marshal(map[string]string{"Version": version, "Time": "2026-10-12T00:27:05Z"})
	for ext, data := range map[string][]byte{".mod": mod, ".info": info} {
		if err := os.WriteFile(filepath.Join(dir, version+ext), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.Create(filepath.Join(dir, version+".zip"))
	if err != nil {
		t.Fatal(err)
	}
	z := zip.NewWriter(f)
	for name, data := range map[string]string{"go.mod": string(mod), "probe.go": "package keel\n"} {
		w, err := z.Create(keelModulePath + "@" + version + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(data))
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()
	list := filepath.Join(dir, "list")
	old, _ := os.ReadFile(list)
	os.WriteFile(list, append(old, version+"\n"...), 0o644)
}

func TestResolveKeelVersion(t *testing.T) {
	proxy := moduleProxy(t)
	writeProxyRelease(t, proxy, "v1.0.0")
	writeProxyRelease(t, proxy, "v1.1.0")
	// Hostile ambient settings must not influence neutral module resolution.
	t.Setenv("GOWORK", filepath.Join(t.TempDir(), "missing.work"))
	t.Setenv("GOFLAGS", "-invalid-fixture-flag")
	c := &cli{dir: t.TempDir()}
	for query, want := range map[string]string{"v1.0.0": "v1.0.0", "latest": "v1.1.0"} {
		if exact, err := c.resolveKeelVersion(query, false); err != nil || exact != want {
			t.Fatalf("resolve %s: %s, %v", query, exact, err)
		}
	}
	if err := os.RemoveAll(proxy); err != nil {
		t.Fatal(err)
	}
	// Even a private-module setting must not enable a direct offline fetch.
	t.Setenv("GONOPROXY", "*")
	if exact, err := c.resolveKeelVersion("v1.0.0", true); err != nil || exact != "v1.0.0" {
		t.Fatalf("cached offline release: %s, %v", exact, err)
	}
	if exact, err := c.resolveKeelVersion("v1.2.0", true); err != nil || exact != "v1.2.0" {
		t.Fatalf("explicit offline release: %s, %v", exact, err)
	}
	if _, err := c.resolveKeelVersion("main", true); err == nil || !strings.Contains(err.Error(), "offline") {
		t.Fatalf("uncached offline query: %v", err)
	}
	for _, v := range []string{"v1.0.0", "v1.0.0-rc.1", "v0.0.0-20261012002705-b2696a6089c4"} {
		if !immutableVersion(v) {
			t.Fatal("rejected", v)
		}
	}
	for _, v := range []string{"latest", "v1", "v1.0", "v1.0.0-01", "main"} {
		if immutableVersion(v) {
			t.Fatal("accepted", v)
		}
	}
}

// New projects require Keel alone: Gio and go-text come inside its module.
func TestNewRequiresOnlyKeel(t *testing.T) {
	moduleProxy(t)
	c, out := newCLI(t.TempDir())
	if c.main([]string{"new", "pinned", "-keel", "v1.0.0", "-offline"}) != 0 {
		t.Fatal(out.String())
	}
	mod, err := os.ReadFile(filepath.Join(c.dir, "pinned", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mod), "require github.com/dyike/keel v1.0.0") || strings.Contains(string(mod), "replace") ||
		strings.Contains(string(mod), "gioui.org") || strings.Contains(string(mod), "go-text") {
		t.Fatalf("unexpected go.mod:\n%s", mod)
	}
	if c.main([]string{"new", "missing", "-offline", "-replace", t.TempDir()}) == 0 {
		t.Fatal("checkout without go.mod accepted")
	}
	if _, err := os.Stat(filepath.Join(c.dir, "missing")); !os.IsNotExist(err) {
		t.Fatal("failed metadata created project")
	}
	checkout := filepath.Join(t.TempDir(), "checkout with spaces")
	os.MkdirAll(checkout, 0o755)
	os.WriteFile(filepath.Join(checkout, "go.mod"), []byte("module "+keelModulePath+"\n"), 0o644)
	if c.main([]string{"new", "local app", "-offline", "-replace", checkout}) != 0 {
		t.Fatal(out.String())
	}
	b, err := goModuleOutput(filepath.Join(c.dir, "local app"), false, "mod", "edit", "-json")
	if err != nil {
		t.Fatalf("quoted local paths invalid: %v", err)
	}
	var edits struct {
		Replace []struct{ Old, New struct{ Path string } }
	}
	json.Unmarshal(b, &edits)
	if len(edits.Replace) != 1 || edits.Replace[0].Old.Path != keelModulePath || edits.Replace[0].New.Path != filepath.ToSlash(checkout) {
		t.Fatalf("replace: %s", b)
	}
	subdirectory := filepath.Join(checkout, "not the module root")
	if err := os.Mkdir(subdirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if c.main([]string{"new", "nested", "-offline", "-replace", subdirectory}) == 0 {
		t.Fatal("checkout subdirectory inherited its parent's module")
	}
	if _, err := os.Stat(filepath.Join(c.dir, "nested")); !os.IsNotExist(err) {
		t.Fatalf("invalid local checkout created destination: %v", err)
	}
}
