package deps

import (
	"go/build"
	"path/filepath"
	"slices"
	"testing"
)

// Check that macOS desktop always selects purego with zero cgo/C/ObjC sources,
// while iOS retains its native cgo bindings.
func TestDarwinDefaultBindings(t *testing.T) {
	bindings := []struct {
		dir, pure string
		nometal   bool
	}{
		{"gio/app", "os_darwin_purego.go", false},
		{"gio/app", "os_macos_purego.go", false},
		{"gio/app", "metal_macos_purego.go", false},
		{"gio/app", "gl_macos_purego.go", true},
		{"gio/gpu/headless", "headless_purego_darwin.go", false},
		{"gio/gpu/internal/metal", "metal_purego_darwin.go", false},
		{"gio/internal/cocoainit", "cocoa_purego_darwin.go", false},
		{"gio/internal/gl", "gl_purego_unix.go", false},
		{"ui/theme", "coretext_purego_darwin.go", false},
	}
	for _, arch := range []string{"amd64", "arm64"} {
		for _, tc := range []struct {
			name string
			cgo  bool
		}{
			{"cgo-on", true},
			{"cgo-off", false},
		} {
			t.Run(arch+"/"+tc.name, func(t *testing.T) {
				for _, b := range bindings {
					ctx := build.Default
					ctx.GOOS, ctx.GOARCH, ctx.CgoEnabled = "darwin", arch, tc.cgo
					if b.nometal {
						ctx.BuildTags = []string{"nometal"}
					}
					pkg, err := ctx.ImportDir(filepath.Join("..", "..", b.dir), 0)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Contains(pkg.GoFiles, b.pure) {
						t.Errorf("%s: pure binding %s not selected", b.dir, b.pure)
					}
					if len(pkg.CgoFiles) != 0 || len(pkg.CFiles) != 0 || len(pkg.MFiles) != 0 {
						t.Errorf("%s: unexpected native sources: cgo=%v c=%v m=%v", b.dir, pkg.CgoFiles, pkg.CFiles, pkg.MFiles)
					}
				}
			})
		}
	}

	// Verify iOS retains its cgo bindings and never selects macOS purego files.
	t.Run("ios", func(t *testing.T) {
		ctx := build.Default
		ctx.GOOS, ctx.GOARCH, ctx.CgoEnabled = "ios", "arm64", true
		for _, b := range []struct{ dir, pure, legacy string }{
			{"gio/app", "os_darwin_purego.go", "os_darwin.go"},
			{"gio/gpu/headless", "headless_purego_darwin.go", "headless_darwin.go"},
			{"gio/gpu/internal/metal", "metal_purego_darwin.go", "metal_cgo_darwin.go"},
			{"gio/internal/cocoainit", "cocoa_purego_darwin.go", "cocoa_darwin.go"},
			{"gio/internal/gl", "gl_purego_unix.go", "gl_unix.go"},
		} {
			pkg, err := ctx.ImportDir(filepath.Join("..", "..", b.dir), 0)
			if err != nil {
				t.Fatal(err)
			}
			if slices.Contains(pkg.GoFiles, b.pure) {
				t.Errorf("%s: iOS unexpectedly selected purego file %s", b.dir, b.pure)
			}
			if !slices.Contains(pkg.CgoFiles, b.legacy) {
				t.Errorf("%s: iOS did not select cgo file %s", b.dir, b.legacy)
			}
		}
	})
}
