package deps

import (
	"go/build"
	"path/filepath"
	"slices"
	"testing"
)

// Check that Linux desktop always selects purego with zero cgo/C sources,
// while Android and FreeBSD retain their native bindings.
func TestLinuxDefaultBindings(t *testing.T) {
	bindings := []struct{ dir, pure string }{
		{"app", "os_wayland_purego.go"},
		{"app", "os_x11_purego.go"},
		{"app", "egl_wayland_purego.go"},
		{"app/internal/xkb", "xkb_purego_linux.go"},
		{"internal/egl", "egl_purego_linux.go"},
		{"internal/gl", "gl_purego_unix.go"},
		{"internal/vk", "vulkan_purego.go"},
		{"internal/vk", "vulkan_wayland_purego.go"},
		{"internal/vk", "vulkan_x11_purego.go"},
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
				ctx := build.Default
				ctx.GOOS, ctx.GOARCH, ctx.CgoEnabled = "linux", arch, tc.cgo
				for _, b := range bindings {
					pkg, err := ctx.ImportDir(filepath.Join("..", "..", "gio", b.dir), 0)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Contains(pkg.GoFiles, b.pure) {
						t.Errorf("%s: pure binding %s not selected", b.dir, b.pure)
					}
					if len(pkg.CgoFiles) != 0 || len(pkg.CFiles) != 0 {
						t.Errorf("%s: unexpected native sources: %v %v", b.dir, pkg.CgoFiles, pkg.CFiles)
					}
				}
			})
		}
	}
}
