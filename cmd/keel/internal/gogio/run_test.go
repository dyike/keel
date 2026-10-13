package gogio

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// Link symbols and resource lookups follow the Gio root, not gioui.org.
func TestGioRoot(t *testing.T) {
	gioRoot = GioRoot
	flags = flag.NewFlagSet("test", flag.ContinueOnError)
	defineFlags(flags)
	flags.Parse([]string{".", "-run", "arg"})
	if got := getLdFlags("dev.keel.notes"); !strings.Contains(got, "-X "+GioRoot+"/app.ID=dev.keel.notes") ||
		!strings.Contains(got, "-X "+GioRoot+"/app/internal/log.appID=dev.keel.notes") ||
		!strings.Contains(got, "-X "+GioRoot+"/app.extraArgs=-run|arg") || strings.Contains(got, "gioui.org") {
		t.Fatal(got)
	}
	// The root is a package of this module, with the Android and iOS
	// sources gogio copies from it.
	dir := filepath.Join("..", "..", "..", "..", "gio", "app")
	for _, f := range []string{"GioActivity.java", "framework_ios.h", "permission/camera/main.go"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunArguments(t *testing.T) {
	var stderr bytes.Buffer
	if err := Run([]string{"-target", "windows", "."}, Options{Stderr: &stderr}); err == nil || !strings.Contains(err.Error(), "invalid -target windows") {
		t.Fatalf("windows resources are Keel's own: %v", err)
	}
	if err := Run([]string{"-target", "js"}, Options{Stderr: &stderr}); err == nil || !strings.Contains(err.Error(), "specify a package") {
		t.Fatal(err)
	}
	if err := Run([]string{"-nosuchflag"}, Options{Stderr: &stderr}); err == nil {
		t.Fatal("unknown flag accepted")
	}
}

func TestWebPackageIdentity(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a WebAssembly package")
	}
	repo, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(t.TempDir(), "app with spaces")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	module := "module example.com/packageprobe\n\ngo 1.26.1\n\nrequire github.com/dyike/keel v0.0.0\n\nreplace github.com/dyike/keel => " + strconv.Quote(filepath.ToSlash(repo)) + "\n"
	source := fmt.Sprintf("package main\nimport (\"fmt\"; \"os\"; %q)\nfunc main() { fmt.Println(app.ID, os.Args) }\n", GioRoot+"/app")
	for name, data := range map[string]string{"go.mod": module, "main.go": source} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var stderr bytes.Buffer
	err = Run([]string{"-target", "js", "-tags", "osusergo", "-name", "Identity Probe", "-appid", "dev.keel.packageprobe", ".", "fixture-launch-argument"}, Options{
		Dir: directory, Env: append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod", "CGO_ENABLED=0"), Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("package: %v\n%s", err, &stderr)
	}
	output := filepath.Join(directory, "Identity Probe")
	for _, name := range []string{"main.wasm", "wasm.js", "index.html"} {
		data, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || len(data) == 0 {
			t.Fatalf("%s: %v", name, err)
		}
		if name == "main.wasm" {
			for _, value := range []string{"dev.keel.packageprobe", "fixture-launch-argument"} {
				if !bytes.Contains(data, []byte(value)) {
					t.Errorf("linked WebAssembly omits %q", value)
				}
			}
		}
		if name == "index.html" && !bytes.Contains(data, []byte("<title>Identity Probe</title>")) {
			t.Errorf("HTML title: %s", data)
		}
	}
}

func TestMobileResourceDiscovery(t *testing.T) {
	repo, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, platform := range []string{"android", "ios"} {
		t.Run(platform, func(t *testing.T) {
			patterns := []string{GioRoot + "/app"}
			if platform == "android" {
				patterns = append(patterns, GioRoot+"/app/permission/camera", GioRoot+"/app/permission/microphone")
			}
			pkgs, err := packages.Load(&packages.Config{
				Mode: packages.NeedName | packages.NeedFiles | packages.NeedModule,
				Dir:  repo, Env: append(os.Environ(), "GOWORK=off", "GOFLAGS=", "GOOS="+platform, "GOARCH=arm64", "CGO_ENABLED=1"),
			}, patterns...)
			if err != nil {
				t.Fatal(err)
			}
			found := make(map[string]bool)
			for _, pkg := range pkgs {
				if len(pkg.Errors) != 0 || len(pkg.GoFiles) == 0 || pkg.Module == nil || pkg.Module.Path != "github.com/dyike/keel" {
					t.Fatalf("%s: files=%v module=%v errors=%v", pkg.PkgPath, pkg.GoFiles, pkg.Module, pkg.Errors)
				}
				found[pkg.PkgPath] = true
				if pkg.PkgPath == GioRoot+"/app" {
					resource := "GioActivity.java"
					if platform == "ios" {
						resource = "framework_ios.h"
					}
					if _, err := os.Stat(filepath.Join(filepath.Dir(pkg.GoFiles[0]), resource)); err != nil {
						t.Fatal(err)
					}
				} else {
					permission := strings.TrimPrefix(pkg.PkgPath, GioRoot+"/app/permission/")
					permissions, _ := getPermissions([]string{permission})
					if !strings.Contains(strings.Join(permissions, ","), AndroidPermissions[permission][0]) {
						t.Fatalf("permission %s was not mapped: %v", permission, permissions)
					}
				}
			}
			for _, pattern := range patterns {
				if !found[pattern] {
					t.Errorf("package discovery missed %s", pattern)
				}
			}
		})
	}
}
