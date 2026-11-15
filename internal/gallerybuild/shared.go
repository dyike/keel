package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Build away from the publication directory: analysis files and independent
// manifests must never be mistaken for assets of the shared runtime.
func buildShared(out, binaryen string) error {
	if binaryen == "" {
		return fmt.Errorf("shared builds require -binaryen-dir or BINARYEN_DIR (Binaryen version_133); use -mode independent for the Go-only fallback")
	}
	for _, tool := range []string{"wasm-opt", "wasm-split"} {
		version, err := exec.Command(filepath.Join(binaryen, tool), "--version").CombinedOutput()
		fields := strings.Fields(string(version))
		if err != nil || len(fields) < 3 || fields[1] != "version" || fields[2] != "133" {
			return fmt.Errorf("%s requires Binaryen version_133: %s (%v)", tool, version, err)
		}
	}
	// Refuse mixed/partial overwrites; a fresh destination is explicit and safe.
	if entries, err := os.ReadDir(out); err == nil && len(entries) != 0 {
		return fmt.Errorf("shared output %s must be empty; choose a fresh directory", out)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	temp, err := os.MkdirTemp("", "keel-shared-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	analysis, split := filepath.Join(temp, "analysis"), filepath.Join(temp, "split")
	commands := [][]string{
		{"go", "run", "./internal/gallerybuild", "-analysis", "-out", analysis},
		{"python3", "scripts/analyze-gallery-wasm.py", "--input", analysis, "--binaryen-dir", binaryen, "--out", filepath.Join(analysis, "ownership.json")},
		{"python3", "scripts/probe-shared-wasm.py", "--ownership-build", analysis, "--binaryen-dir", binaryen, "--out", split, "--optimize"},
	}
	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", args[0], err)
		}
	}
	data, err := os.ReadFile(filepath.Join(split, "shared-build.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Primary, FullWasm, FullScript, Loader, Startup string
		Categories                                     map[string]string
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	files := []string{manifest.Primary, manifest.FullWasm, manifest.FullScript, manifest.Loader, manifest.Startup, "sizes.json"}
	for _, file := range manifest.Categories {
		files = append(files, file)
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(split, file))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, file), data, 0644); err != nil {
			return err
		}
	}
	// The manifest is the commit marker. Failed copies cannot be published.
	return os.WriteFile(filepath.Join(out, "shared-build.json"), data, 0644)
}
