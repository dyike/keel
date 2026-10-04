package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
)

// configFile describes a project; keel new writes it and keel build reads it.
const configFile = "keel.json"

// Config is keel.json.
type Config struct {
	// Name is shown to people: the window title, the .app and menu entry.
	Name string `json:"name"`
	// AppID is the reverse-DNS identifier: the macOS bundle ID, the Linux
	// app_id and .desktop name. Permissions and notifications are keyed by
	// it, so keep it once released.
	AppID string `json:"appid"`
	// Version is major.minor.patch; Build counts builds of one version.
	Version string `json:"version"`
	Build   int    `json:"build,omitempty"`
	// Binary is the executable's file name.
	Binary string `json:"binary"`
	// Icon is a square PNG, 1024×1024 ideally.
	Icon string `json:"icon"`
	// Main is the main package, relative to keel.json.
	Main string `json:"main"`
}

var (
	appIDPattern   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-]*(\.[A-Za-z][A-Za-z0-9-]*)+$`)
	versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

func (c *Config) validate() error {
	var errs []error
	if c.Name == "" {
		errs = append(errs, errors.New("name is empty"))
	}
	if !appIDPattern.MatchString(c.AppID) {
		errs = append(errs, fmt.Errorf("appid %q is not reverse DNS, like com.example.app", c.AppID))
	}
	if !versionPattern.MatchString(c.Version) {
		errs = append(errs, fmt.Errorf("version %q is not major.minor.patch", c.Version))
	}
	if c.Binary == "" || strings.ContainsAny(c.Binary, `/\ `) {
		errs = append(errs, fmt.Errorf("binary %q is not a plain file name", c.Binary))
	}
	if c.Build < 0 {
		errs = append(errs, errors.New("build is negative"))
	}
	return errors.Join(errs...)
}

// fourPart is the version in gogio's major.minor.patch.build form.
func (c *Config) fourPart() string { return fmt.Sprintf("%s.%d", c.Version, max(c.Build, 1)) }

func loadConfig(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, configFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no %s here; create a project with keel new", configFile)
	}
	if err != nil {
		return nil, err
	}
	c := &Config{Icon: "appicon.png", Main: "."}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("%s: %w", configFile, err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", configFile, err)
	}
	return c, nil
}

func (c *Config) save(dir string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, configFile), append(data, '\n'), 0o644)
}

// keelVersion is this tool's release, which new projects require; a
// development build asks for the latest.
func keelVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Path == "github.com/dyike/keel" {
		if v := bi.Main.Version; v != "" && v != "(devel)" && !strings.Contains(v, "-0.") {
			return v
		}
	}
	return "latest"
}
