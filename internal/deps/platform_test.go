package deps

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The interface builds for the browser. Go's os/user has no js/wasm
// implementation unless osusergo is set, and Gio's font scanning imports it,
// so web builds pass -tags osusergo (docs/web.md).
func TestUIBuildsForWebAssembly(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-compiles the interface")
	}
	cmd := exec.Command("go", "build", "-tags", "osusergo", "-o", os.DevNull, "../../ui/...", "../../examples/hello", "../../examples/components")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("GOOS=js GOARCH=wasm go build: %v\n%s", err, strings.TrimSpace(string(out)))
	}
}
