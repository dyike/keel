package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const keelModulePath = "github.com/dyike/keel"

// A full semantic version is immutable; queries such as latest, branches and
// shortened versions are not. Numeric prerelease identifiers cannot have zeros.
var releaseSemver = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-([0-9A-Za-z-]+)(\.[0-9A-Za-z-]+)*)?(\+([0-9A-Za-z-]+)(\.[0-9A-Za-z-]+)*)?$`)

func immutableVersion(version string) bool {
	if !releaseSemver.MatchString(version) {
		return false
	}
	base := strings.SplitN(version, "+", 2)[0]
	if _, pre, ok := strings.Cut(base, "-"); ok {
		for _, id := range strings.Split(pre, ".") {
			if len(id) > 1 && id[0] == '0' && strings.Trim(id, "0123456789") == "" {
				return false
			}
		}
	}
	return true
}

// goModuleOutput isolates module resolution from workspace and build flags.
// Offline resolution permits only cached metadata and never invokes a proxy,
// checksum server, direct VCS fetch, or automatic toolchain download.
func goModuleOutput(dir string, offline bool, args ...string) ([]byte, error) {
	overrides := map[string]string{"GOWORK": "off", "GOFLAGS": "", "GO111MODULE": "on"}
	if offline {
		overrides["GOPROXY"] = "off"
		overrides["GOSUMDB"] = "off"
		overrides["GOTOOLCHAIN"] = "local"
		overrides["GONOPROXY"] = "none"
	}
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := overrides[key]; !replaced {
			env = append(env, entry)
		}
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	cmd := exec.Command("go", args...)
	cmd.Dir, cmd.Env = dir, env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("go %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func downloadKeelVersion(version string, offline bool) (exact, dir string, err error) {
	if version == "" {
		version = "latest"
	}
	neutral, err := os.MkdirTemp("", "keel-resolve-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(neutral)
	out, err := goModuleOutput(neutral, offline, "mod", "download", "-json", keelModulePath+"@"+version)
	if err != nil {
		mode := ""
		if offline {
			mode = " (offline: requested release must already be cached)"
		}
		return "", "", fmt.Errorf("resolve Keel %q%s: %w", version, mode, err)
	}
	var result struct {
		Path, Version, Dir, Error string
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", "", fmt.Errorf("decode Keel download: %w", err)
	}
	if result.Error != "" || result.Path != keelModulePath || !immutableVersion(result.Version) || result.Dir == "" {
		return "", "", fmt.Errorf("invalid Keel download metadata: %s", strings.TrimSpace(string(out)))
	}
	return result.Version, result.Dir, nil
}

// resolveKeelVersion turns -keel into the exact release new projects require.
// Gio and go-text ship inside that release, so nothing else is pinned.
func (c *cli) resolveKeelVersion(version string, offline bool) (exact string, err error) {
	if offline && immutableVersion(version) {
		return version, nil
	}
	exact, _, err = downloadKeelVersion(version, offline)
	return exact, err
}

func validateLocalModule(dir, path string) error {
	manifest := filepath.Join(dir, "go.mod")
	info, err := os.Stat(manifest)
	if err != nil {
		return fmt.Errorf("local module %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("local module %s: %s is not a regular file", path, manifest)
	}
	b, err := goModuleOutput(dir, true, "mod", "edit", "-json")
	if err != nil {
		return fmt.Errorf("local module %s: %w", path, err)
	}
	var mod struct{ Module struct{ Path string } }
	if err := json.Unmarshal(b, &mod); err != nil {
		return err
	}
	if mod.Module.Path != path {
		return fmt.Errorf("local directory %s declares %s, expected %s", dir, mod.Module.Path, path)
	}
	return nil
}
