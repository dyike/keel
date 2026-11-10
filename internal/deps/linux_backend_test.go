package deps

import (
	"go/build"
	"path/filepath"
	"slices"
	"testing"
)

// Check selection without requiring native headers or a cross C compiler.
func TestLinuxDefaultBindings(t *testing.T) {
	bindings := []struct{ dir, pure, legacy string }{
		{"app", "os_wayland_purego.go", "os_wayland_cgo.go"},
		{"app", "os_x11_purego.go", "os_x11_cgo.go"},
		{"app", "egl_wayland_purego.go", "egl_wayland.go"},
		{"app/internal/xkb", "xkb_purego_linux.go", "xkb_cgo.go"},
		{"internal/egl", "egl_purego_linux.go", "egl_unix.go"},
		{"internal/gl", "gl_purego_unix.go", "gl_unix.go"},
		{"internal/vk", "vulkan_purego.go", "vulkan.go"},
		{"internal/vk", "vulkan_wayland_purego.go", "vulkan_wayland.go"},
		{"internal/vk", "vulkan_x11_purego.go", "vulkan_x11.go"},
	}
	for _, arch := range []string{"amd64", "arm64"} {
		for _, tc := range []struct {
			name   string
			cgo    bool
			tags   []string
			legacy bool
		}{
			{"default", true, nil, false},
			{"no-cgo", false, nil, false},
			{"legacy", true, []string{"giocgo"}, true},
			{"purego-override", true, []string{"giocgo", "purego"}, false},
			{"legacy-without-cgo", false, []string{"giocgo"}, false},
		} {
			t.Run(arch+"/"+tc.name, func(t *testing.T) {
				ctx := build.Default
				ctx.GOOS, ctx.GOARCH, ctx.CgoEnabled, ctx.BuildTags = "linux", arch, tc.cgo, tc.tags
				for _, b := range bindings {
					pkg, err := ctx.ImportDir(filepath.Join("..", "..", "gio", b.dir), 0)
					if err != nil {
						t.Fatal(err)
					}
					if got := slices.Contains(pkg.GoFiles, b.pure); got == tc.legacy {
						t.Errorf("%s: pure binding selected = %v", b.dir, got)
					}
					if got := slices.Contains(pkg.CgoFiles, b.legacy); got != tc.legacy {
						t.Errorf("%s: cgo binding selected = %v", b.dir, got)
					}
					if !tc.legacy && (len(pkg.CgoFiles) != 0 || len(pkg.CFiles) != 0) {
						t.Errorf("%s: unexpected native sources: %v %v", b.dir, pkg.CgoFiles, pkg.CFiles)
					}
				}
			})
		}
	}
}
