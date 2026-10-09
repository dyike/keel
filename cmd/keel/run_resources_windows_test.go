//go:build windows

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tc-hib/winres"
)

func TestWindowsCompileRunResourcesAndCleanup(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{Name: "Run DPI", AppID: "com.example.run", Binary: "run-dpi", Version: "1.0.0", Main: ".", Icon: "appicon.png"}
	if err := cfg.save(dir); err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string]string{
		"go.mod":  "module rundpitest\n\ngo 1.26.1\n",
		"main.go": "package main\nfunc main() {}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GOWORK", "off")
	c, _ := newCLI(dir)
	tmp := t.TempDir()
	build := c.compileRun(context.Background(), dir, tmp, 1)
	defer build.cleanup()
	if build.err != nil {
		t.Fatal(build.err)
	}
	file, err := os.Open(build.binary)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rs, err := winres.LoadFromEXE(file)
	if err != nil {
		t.Fatal(err)
	}
	if manifest := string(rs.Get(winres.RT_MANIFEST, winres.ID(1), winres.LCIDDefault)); !strings.Contains(manifest, "permonitorv2") {
		t.Fatal("compileRun did not embed its DPI manifest")
	}
	assertClean := func() {
		t.Helper()
		if paths, _ := filepath.Glob(filepath.Join(dir, "*.syso")); len(paths) != 0 {
			t.Fatal("compileRun leaked resources", paths)
		}
	}
	assertClean()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("invalid Go source"), 0600); err != nil {
		t.Fatal(err)
	}
	failed := c.compileRun(context.Background(), dir, tmp, 2)
	defer failed.cleanup()
	if failed.err == nil {
		t.Fatal("invalid source compiled")
	}
	assertClean()
}
